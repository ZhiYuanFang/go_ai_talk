package voice

import (
	"context"
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

// streamClinicLLMHeld 在调用方已持 clinic 闸门槽位时发起流式请求（modelCfg 可为 nil=omit）。
func (s *ClinicService) streamClinicLLMHeld(ctx context.Context, modelCfg *PythonModelCfg, deviceNo, question string, cb clinicStreamCallbacks) (thinking, answer, answerID string, err error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return "", "", "", errors.New("诊疗服务设备号缺失")
	}
	pythonClient := PythonAIClientFromCfg()
	pythonResp, pythonErr := pythonClient.ClinicStream(ctx, &ClinicStreamRequest{
		Question: question,
		DeviceNo: deviceNo,
		Model:    modelCfg,
	}, &ClinicStreamCallback{
		OnThinking: cb.OnThinkingDelta,
		OnAnswer:   cb.OnAnswerDelta,
		OnDone:     cb.OnDone,
	})
	if pythonErr != nil {
		glog.Warningf(ctx, "[Python AI] 诊疗流式调用失败。deviceNo=%s err=%v", deviceNo, pythonErr)
		return "", "", "", pythonErr
	}
	return pythonResp.Thinking, pythonResp.Answer, pythonResp.AnswerID, nil
}

// streamClinicLLM 经统一权益选模后调用上游流式接口。
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
