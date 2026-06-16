package voice

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	gctx "github.com/gogf/gf/v2/os/gctx"
)

// ClinicService 胖宝诊疗业务（WS 鉴权、摘要、LLM、会话 Redis）。
type ClinicService struct {
	cfg        AIClinicConfig
	httpClient *http.Client
}

var (
	clinicOnce sync.Once
	clinicSvc  *ClinicService
)

// Clinic 返回胖宝诊疗单例；endpoint/apiKey 来自 voice-chat.shared.yaml deepseek。
func Clinic() *ClinicService {
	clinicOnce.Do(func() {
		voiceCfg := loadVoiceConfig(gctx.New())
		cfg := loadAIClinicConfig(gctx.New(), voiceCfg)
		timeout := cfg.LLMTimeoutSeconds
		if timeout <= 0 {
			timeout = 120
		}
		clinicSvc = &ClinicService{
			cfg: cfg,
			httpClient: &http.Client{
				Timeout: time.Duration(timeout+10) * time.Second,
			},
		}
	})
	return clinicSvc
}

// BuildSessionSync 构建 auth_ok 后立即下发的 session_sync 帧（见 clinic_session.go 注释）。
func (s *ClinicService) BuildSessionSync(ctx context.Context, wxID int64) (SessionSyncPayload, error) {
	return BuildSessionSync(ctx, wxID, s.cfg.SessionTTLSeconds)
}

// EmitTurnCancelled 下发 turn_cancelled 帧；供 WS handler（supersede/cancel）复用。
func EmitTurnCancelled(writeJSON func(v interface{}) error, turnID, reason string) error {
	return writeJSON(map[string]interface{}{
		"type":   "turn_cancelled",
		"turnId": turnID,
		"reason": reason,
	})
}

// HandleQuestion 在独立 goroutine 中处理单轮 question：限流 check → 额度 check → 摘要 → LLM 流 → 成功时 consume/限流/session。
// turnCtx 被取消（cancel/supersede/disconnect）时 MUST NOT 写 answer_done、不 consume clinic_ai、不 append session。
func (s *ClinicService) HandleQuestion(turnCtx context.Context, wxID int64, deviceNo, question, turnID string, writeJSON func(v interface{}) error) error {
	question = strings.TrimSpace(question)
	if question == "" {
		return writeJSON(map[string]interface{}{"type": "error", "code": 400, "message": "问题不能为空"})
	}
	// 提问前只检查窗口计数；INCR 推迟至 answer_done 成功，cancel/supersede 不占用限流额度。
	if err := checkClinicRateLimit(turnCtx, wxID, s.cfg); err != nil {
		return clinicWriteQuotaErr(writeJSON, err)
	}
	if err := CheckClinicAIQuota(turnCtx, wxID); err != nil {
		return clinicWriteQuotaErr(writeJSON, err)
	}
	summary, err := s.ensureClinicSummary(turnCtx, wxID, deviceNo)
	if err != nil {
		return writeJSON(map[string]interface{}{"type": "error", "code": 500, "message": "喂养摘要加载失败"})
	}
	if err := turnCtx.Err(); err != nil {
		return nil
	}
	sess, _, _ := loadClinicSession(turnCtx, wxID)
	prior := clinicSessionMessages(sess)
	thinking, answer, streamErr := s.streamClinicLLM(turnCtx, summary, question, prior, clinicStreamCallbacks{
		OnThinkingDelta: func(delta string) error {
			if err := turnCtx.Err(); err != nil {
				return err
			}
			return writeJSON(map[string]interface{}{"type": "thinking_delta", "delta": delta, "turnId": turnID})
		},
		OnAnswerDelta: func(delta string) error {
			if err := turnCtx.Err(); err != nil {
				return err
			}
			return writeJSON(map[string]interface{}{"type": "answer_delta", "delta": delta, "turnId": turnID})
		},
	})
	if streamErr != nil {
		// cancel/supersede/disconnect 中断流：静默结束，不下发 error、不写 answer_done、不扣费。
		if errors.Is(streamErr, context.Canceled) || errors.Is(streamErr, context.DeadlineExceeded) {
			return nil
		}
		return writeJSON(map[string]interface{}{"type": "error", "code": 500, "message": streamErr.Error()})
	}
	if err := turnCtx.Err(); err != nil {
		return nil
	}
	// 仅 answer_done 成功路径扣 clinic_ai 额度、递增限流、写入 session。
	if err := ConsumeClinicAIQuota(turnCtx, wxID); err != nil {
		return clinicWriteQuotaErr(writeJSON, err)
	}
	_ = recordClinicRateLimitOnSuccess(turnCtx, wxID, s.cfg)
	if err := appendClinicTurn(turnCtx, wxID, s.cfg, deviceNo, question, answer); err != nil {
		// 扣减已成功；session 写失败仅记录，仍返回 answer_done
	}
	return writeJSON(map[string]interface{}{
		"type":     "answer_done",
		"turnId":   turnID,
		"thinking": thinking,
		"answer":   answer,
	})
}

func clinicWriteQuotaErr(writeJSON func(v interface{}) error, err error) error {
	if qe, ok := err.(*VoiceAIQuotaError); ok {
		return writeJSON(map[string]interface{}{"type": "error", "code": qe.Code, "message": qe.Message})
	}
	return writeJSON(map[string]interface{}{"type": "error", "code": 500, "message": err.Error()})
}
