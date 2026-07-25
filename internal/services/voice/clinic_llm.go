package voice

import (
	"context"
	"errors"

	"hello/internal/services/aimodel"

	"github.com/gogf/gf/v2/os/glog"
)

type clinicStreamCallbacks struct {
	OnThinkingDelta func(delta string) error
	OnAnswerDelta   func(delta string) error
	OnDone          func(answerID string) error
}

// streamClinicLLMHeld 在调用方已持 clinic 闸门槽位时发起流式请求。
// 仅调用 Python 微服务，Python 不可用时直接返回错误，由上层返回降级提示语。
func (s *ClinicService) streamClinicLLMHeld(ctx context.Context, profile aimodel.Profile, baby clinicBabyProfile, summaryJSON, question string, prior []map[string]string, cb clinicStreamCallbacks) (thinking, answer, answerID string, err error) {
	deviceNo := baby.deviceNo
	if deviceNo != "" {
		pythonClient := PythonAIClientFromCfg()
		pythonResp, pythonErr := pythonClient.ClinicStream(ctx, &ClinicStreamRequest{
			Question: question,
			DeviceNo: deviceNo,
			Model: PythonModelCfg{
				Provider:    string(profile.Provider),
				Name:        profile.Model,
				MaxInFlight: 1,
			},
		}, &ClinicStreamCallback{
			OnThinking: cb.OnThinkingDelta,
			OnAnswer:   cb.OnAnswerDelta,
			OnDone:     cb.OnDone,
		})
		if pythonErr == nil {
			return pythonResp.Thinking, pythonResp.Answer, pythonResp.AnswerID, nil
		}
		glog.Warningf(ctx, "[Python AI] 诊疗流式调用失败。deviceNo=%s err=%v", deviceNo, pythonErr)
		return "", "", "", pythonErr
	}
	return "", "", "", errors.New("诊疗服务设备号缺失")
}

// streamClinicLLM 经 LaneClinic 调用上游流式接口（内部 Acquire）。
func (s *ClinicService) streamClinicLLM(ctx context.Context, baby clinicBabyProfile, summaryJSON, question string, prior []map[string]string, cb clinicStreamCallbacks) (thinking, answer, answerID string, err error) {
	profile, err := aimodel.LoadProfile(ctx, aimodel.LaneClinic)
	if err != nil {
		return "", "", "", err
	}
	return s.streamClinicLLMHeld(ctx, profile, baby, summaryJSON, question, prior, cb)
}
