package simuser

import "hello/internal/services/aimodel"

// 任务级 LLM 采样温度；调参需 redeploy sim-user-service。
const (
	simTempRegisterNickname = 0.90 // T1 昵称：略高以增加差异
	simTempComment          = 0.85 // T2 评论：口语多样
	simTempPostImageText    = 0.80 // T3 图文配文：稍稳
	simTempPostVideoText    = 0.80 // T4 视频文案：稍稳
	simTempPostDebateText   = 0.80 // T7 辩论话题：稍稳
	simTempDebateComment    = 0.85 // T8 辩论论点：口语多样
	simTempChatReply        = 0.88 // T5 未读回复：自然变化
)

// simChatRequest 构造带任务温度的 chat 请求。
func simChatRequest(temp float64, maxTokens int, messages []aimodel.Message) aimodel.ChatRequest {
	t := temp
	return aimodel.ChatRequest{
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: &t,
	}
}
