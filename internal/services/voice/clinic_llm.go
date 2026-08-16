package voice

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"hello/internal/services/aimodel"
	"hello/internal/services/contracts"

	"github.com/gogf/gf/v2/os/glog"
)

type clinicStreamCallbacks struct {
	OnThinkingDelta func(delta string) error
	OnAnswerDelta   func(delta string) error
	OnDone          func(answerID string) error
}

// streamClinicLLMHeld 在调用方已持 clinic 闸门槽位时经 OpenClaw Clinic agent 流式回复。
func (s *ClinicService) streamClinicLLMHeld(ctx context.Context, modelCfg *PythonModelCfg, deviceNo, question string, cb clinicStreamCallbacks) (thinking, answer, answerID string, err error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return "", "", "", errors.New("诊疗服务设备号缺失")
	}
	sessionKey := "clinic:" + deviceNo
	var onDelta func(string) error
	if cb.OnAnswerDelta != nil {
		onDelta = cb.OnAnswerDelta
	} else if cb.OnThinkingDelta != nil {
		onDelta = cb.OnThinkingDelta
	}
	full, streamErr := OpenClawFromCfg().StreamChat(
		ctx, "openclaw/clinic", openClawBackendModel(modelCfg), sessionKey, question, onDelta,
	)
	if streamErr != nil {
		glog.Warningf(ctx, "[OpenClaw] 诊疗流式失败。deviceNo=%s err=%v", deviceNo, streamErr)
		return "", "", "", streamErr
	}
	answer = strings.TrimSpace(full)
	answerID = newClinicAnswerID()
	if cb.OnDone != nil {
		if doneErr := cb.OnDone(answerID); doneErr != nil {
			return thinking, answer, answerID, doneErr
		}
	}
	return thinking, answer, answerID, nil
}

// streamClinicLLM 经统一权益选模后调用 Gateway 流式接口。
func (s *ClinicService) streamClinicLLM(ctx context.Context, wxID int64, deviceNo, question string, cb clinicStreamCallbacks) (thinking, answer, answerID string, err error) {
	ent, runtime, modelCfg, _ := ResolveLaneModel(ctx, wxID, aimodel.LaneClinic, contracts.AIQuotaClinicAI, PrivilegeAccount)
	if modelCfg != nil {
		release, acqErr := aimodel.Acquire(ctx, runtime)
		if acqErr != nil {
			return "", "", "", acqErr
		}
		defer release()
	}
	thinking, answer, answerID, err = s.streamClinicLLMHeld(ctx, modelCfg, deviceNo, question, cb)
	if err == nil {
		ConsumeVoiceFeatureIfNeeded(ctx, wxID, contracts.AIQuotaClinicAI, ent)
	}
	return thinking, answer, answerID, err
}

func newClinicAnswerID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "clinic-local"
	}
	return "clinic-" + hex.EncodeToString(b[:])
}
