package voice

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"hello/internal/services/aimodel"
	"hello/internal/services/contracts"

	gctx "github.com/gogf/gf/v2/os/gctx"
)

// ClinicService 胖宝诊疗/陪伴业务（WS 鉴权、额度、限流、Python LLM 流转发）。
// 业务说明：不再在 Go 侧缓存对话 turns、喂养摘要或宝宝画像；
// UI 历史由客户端本地负责，多轮上下文由 Python companion_session 负责。
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

// EmitTurnCancelled 下发 turn_cancelled 帧；供 WS handler（supersede/cancel）复用。
func EmitTurnCancelled(writeJSON func(v interface{}) error, turnID, reason string) error {
	return writeJSON(map[string]interface{}{
		"type":   "turn_cancelled",
		"turnId": turnID,
		"reason": reason,
	})
}

// HandleQuestion 在独立 goroutine 中处理单轮 question：限流 check → 额度 check → Python LLM 流 → 成功时 consume/限流。
// turnCtx 被取消（cancel/supersede/disconnect）时 MUST NOT 写 answer_done、不 consume clinic_ai。
// Args: wxID 鉴权绑定主键；deviceNo 透传 Python（不作 Go 侧摘要/画像拉取）；question/turnID 来自客户端帧。
func (s *ClinicService) HandleQuestion(turnCtx context.Context, wxID int64, deviceNo, question, turnID string, writeJSON func(v interface{}) error) error {
	question = strings.TrimSpace(question)
	if question == "" {
		return writeJSON(map[string]interface{}{"type": "error", "code": 400, "message": "问题不能为空"})
	}
	// 提问前只检查窗口计数；INCR 推迟至 answer_done 成功，cancel/supersede 不占用限流额度。
	if err := checkClinicRateLimit(turnCtx, wxID, s.cfg); err != nil {
		return clinicWriteQuotaErr(writeJSON, err)
	}
	quotaSnap, err := CheckClinicAIQuotaSnapshot(turnCtx, wxID)
	if err != nil {
		return clinicWriteQuotaErr(writeJSON, err)
	}
	clinicDegraded := quotaSnap.Degraded && !quotaSnap.Allowed
	var profile aimodel.Profile
	if quotaSnap.Allowed {
		profile, err = aimodel.LoadProfile(turnCtx, aimodel.LaneClinic)
		if err != nil {
			return writeJSON(map[string]interface{}{"type": "error", "code": 500, "message": err.Error()})
		}
	} else if clinicDegraded {
		profile = aimodel.DegradedClinicProfile()
	} else {
		return clinicWriteQuotaErr(writeJSON, &VoiceAIQuotaError{Code: contracts.CodeAIQuotaExhausted, Message: contracts.ErrAIQuotaExhausted.Error()})
	}
	release, err := aimodel.Acquire(turnCtx, profile)
	if err != nil {
		return clinicWriteQuotaErr(writeJSON, mapVoiceLLMError(err))
	}
	defer release()
	if err := turnCtx.Err(); err != nil {
		return nil
	}
	// 直接转发 Python；喂养上下文/画像/多轮由 Python 侧处理，Go 不拼 prompt。
	thinking, answer, answerID, streamErr := s.streamClinicLLMHeld(turnCtx, profile, deviceNo, question, clinicStreamCallbacks{
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
	// 仅正常额度路径扣 clinic_ai；降速 fallback 不 consume。
	if quotaSnap.Allowed {
		if err := ConsumeClinicAIQuota(turnCtx, wxID); err != nil {
			return clinicWriteQuotaErr(writeJSON, err)
		}
	}
	_ = recordClinicRateLimitOnSuccess(turnCtx, wxID, s.cfg)
	return writeJSON(map[string]interface{}{
		"type":     "answer_done",
		"turnId":   turnID,
		"thinking": thinking,
		"answer":   answer,
		"answerId": answerID,
	})
}

func clinicWriteQuotaErr(writeJSON func(v interface{}) error, err error) error {
	if qe, ok := err.(*VoiceAIQuotaError); ok {
		return writeJSON(map[string]interface{}{"type": "error", "code": qe.Code, "message": qe.Message})
	}
	return writeJSON(map[string]interface{}{"type": "error", "code": 500, "message": err.Error()})
}
