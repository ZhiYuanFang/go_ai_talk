package voice

import (
	"context"
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

// HandleQuestion 处理 auth_ok 后的 question 帧：限流 → 额度 check → 摘要 → LLM 流 → consume → 写 session。
func (s *ClinicService) HandleQuestion(ctx context.Context, wxID int64, deviceNo, question string, writeJSON func(v interface{}) error) error {
	question = strings.TrimSpace(question)
	if question == "" {
		return writeJSON(map[string]interface{}{"type": "error", "code": 400, "message": "问题不能为空"})
	}
	if err := checkClinicRateLimit(ctx, wxID, s.cfg); err != nil {
		return clinicWriteQuotaErr(writeJSON, err)
	}
	if err := CheckClinicAIQuota(ctx, wxID); err != nil {
		return clinicWriteQuotaErr(writeJSON, err)
	}
	summary, err := s.ensureClinicSummary(ctx, wxID, deviceNo)
	if err != nil {
		return writeJSON(map[string]interface{}{"type": "error", "code": 500, "message": "喂养摘要加载失败"})
	}
	sess, _, _ := loadClinicSession(ctx, wxID)
	prior := clinicSessionMessages(sess)
	thinking, answer, streamErr := s.streamClinicLLM(ctx, summary, question, prior, clinicStreamCallbacks{
		OnThinkingDelta: func(delta string) error {
			return writeJSON(map[string]interface{}{"type": "thinking_delta", "delta": delta})
		},
		OnAnswerDelta: func(delta string) error {
			return writeJSON(map[string]interface{}{"type": "answer_delta", "delta": delta})
		},
	})
	if streamErr != nil {
		return writeJSON(map[string]interface{}{"type": "error", "code": 500, "message": streamErr.Error()})
	}
	if err := ConsumeClinicAIQuota(ctx, wxID); err != nil {
		return clinicWriteQuotaErr(writeJSON, err)
	}
	if err := appendClinicTurn(ctx, wxID, s.cfg, deviceNo, question, answer); err != nil {
		// 扣减已成功；session 写失败仅记录，仍返回 answer_done
	}
	return writeJSON(map[string]interface{}{
		"type":     "answer_done",
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
