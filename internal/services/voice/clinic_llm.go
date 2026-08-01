package voice

import (
	"context"
	"errors"
	"strings"

	"hello/internal/services/aimodel"

	"github.com/gogf/gf/v2/os/glog"
)

type clinicStreamCallbacks struct {
	OnThinkingDelta func(delta string) error
	OnAnswerDelta   func(delta string) error
	OnDone          func(answerID string) error
}

// streamClinicLLMHeld 在调用方已持 clinic 闸门槽位时发起流式请求。
// 业务说明：仅向 Python 透传 question、device_no、model；不传 Go 侧摘要/画像/prior。
// Python 不可用时直接返回错误，由上层返回可诊断 error 帧。
// Args: deviceNo 为 auth 锁定设备号；question 为用户提问正文。
func (s *ClinicService) streamClinicLLMHeld(ctx context.Context, profile aimodel.Profile, deviceNo, question string, cb clinicStreamCallbacks) (thinking, answer, answerID string, err error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return "", "", "", errors.New("诊疗服务设备号缺失")
	}
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
	if pythonErr != nil {
		glog.Warningf(ctx, "[Python AI] 诊疗流式调用失败。deviceNo=%s err=%v", deviceNo, pythonErr)
		return "", "", "", pythonErr
	}
	return pythonResp.Thinking, pythonResp.Answer, pythonResp.AnswerID, nil
}

// streamClinicLLM 经 LaneClinic 调用上游流式接口（内部 Acquire）；供非 WS 复用入口。
func (s *ClinicService) streamClinicLLM(ctx context.Context, deviceNo, question string, cb clinicStreamCallbacks) (thinking, answer, answerID string, err error) {
	profile, err := aimodel.LoadProfile(ctx, aimodel.LaneClinic)
	if err != nil {
		return "", "", "", err
	}
	release, err := aimodel.Acquire(ctx, profile)
	if err != nil {
		return "", "", "", err
	}
	defer release()
	return s.streamClinicLLMHeld(ctx, profile, deviceNo, question, cb)
}
