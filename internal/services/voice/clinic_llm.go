package voice

import (
	"context"
	"strings"

	"hello/internal/services/aimodel"

	"github.com/gogf/gf/v2/os/glog"
)

type clinicStreamCallbacks struct {
	OnThinkingDelta func(delta string) error
	OnAnswerDelta   func(delta string) error
}

// streamClinicLLMHeld 在调用方已持 clinic 闸门槽位时发起流式请求。
// 优先调用 Python 微服务的 /v1/clinic/stream 接口，失败时回退到原有 LLM 直连逻辑。
func (s *ClinicService) streamClinicLLMHeld(ctx context.Context, profile aimodel.Profile, baby clinicBabyProfile, summaryJSON, question string, prior []map[string]string, cb clinicStreamCallbacks) (thinking, answer string, err error) {
	// 先尝试调用 Python 微服务进行流式诊疗
	deviceNo := baby.deviceNo
	if deviceNo != "" {
		pythonClient := pythonAIClientFromCfg()
		pythonThinking, pythonAnswer, pythonErr := pythonClient.ClinicStream(ctx, &ClinicStreamRequest{
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
		})
		if pythonErr == nil {
			return pythonThinking, pythonAnswer, nil
		}
		// Python 服务调用失败，回退到原有 LLM 直连逻辑
		glog.Warningf(ctx, "[Python AI] 诊疗流式调用失败，回退到 LLM 直连。deviceNo=%s err=%v", deviceNo, pythonErr)
	}
	// 原有逻辑：直接调用 LLM 流式接口
	messages := buildClinicLLMMessages(s.cfg, baby, summaryJSON, question, prior)
	timeout := s.cfg.LLMTimeoutSeconds
	if timeout <= 0 {
		timeout = 120
	}
	result, err := aimodel.InvokeStreamWithHeldProfile(ctx, profile, aimodel.ChatRequest{
		Messages:        messages,
		ThinkingEnabled: s.cfg.ThinkingEnabled,
		ReasoningEffort: s.cfg.ReasoningEffort,
		TimeoutSec:      timeout,
	}, aimodel.StreamCallbacks{
		OnThinkingDelta: cb.OnThinkingDelta,
		OnAnswerDelta:   cb.OnAnswerDelta,
	})
	if err != nil {
		return "", "", mapVoiceLLMError(err)
	}
	return result.Thinking, result.Answer, nil
}

// streamClinicLLM 经 LaneClinic 调用上游流式接口（内部 Acquire）。
func (s *ClinicService) streamClinicLLM(ctx context.Context, baby clinicBabyProfile, summaryJSON, question string, prior []map[string]string, cb clinicStreamCallbacks) (thinking, answer string, err error) {
	profile, err := aimodel.LoadProfile(ctx, aimodel.LaneClinic)
	if err != nil {
		return "", "", err
	}
	return s.streamClinicLLMHeld(ctx, profile, baby, summaryJSON, question, prior, cb)
}

func buildClinicLLMMessages(cfg AIClinicConfig, baby clinicBabyProfile, summaryJSON, question string, prior []map[string]string) []aimodel.Message {
	system := strings.TrimSpace(cfg.SystemPrompt)
	if system == "" {
		system = "你是「胖宝」AI 育儿助手。"
	}
	system += "\n\n" + baby.promptLine()
	system += "\n\n近7天喂养摘要（JSON，含 by_event 聚合与有备注记录）：\n" + summaryJSON

	messages := make([]aimodel.Message, 0, 2+len(prior)+1)
	messages = append(messages, aimodel.Message{Role: "system", Content: system})
	for _, m := range prior {
		role := strings.TrimSpace(m["role"])
		content := strings.TrimSpace(m["content"])
		if role == "" || content == "" {
			continue
		}
		messages = append(messages, aimodel.Message{Role: role, Content: content})
	}
	messages = append(messages, aimodel.Message{Role: "user", Content: strings.TrimSpace(question)})
	return messages
}
