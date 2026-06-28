package voice

import (
	"context"
	"strings"

	"hello/internal/services/aimodel"
)

type clinicStreamCallbacks struct {
	OnThinkingDelta func(delta string) error
	OnAnswerDelta   func(delta string) error
}

// streamClinicLLMHeld 在调用方已持 clinic 闸门槽位时发起流式请求。
func (s *ClinicService) streamClinicLLMHeld(ctx context.Context, profile aimodel.Profile, baby clinicBabyProfile, summaryJSON, question string, prior []map[string]string, cb clinicStreamCallbacks) (thinking, answer string, err error) {
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
