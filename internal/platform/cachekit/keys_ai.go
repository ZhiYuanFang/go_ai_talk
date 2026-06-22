package cachekit

import (
	"fmt"
)

const aiQuotaUsageKeyPrefix = "ai:usage:"

// AIQuotaUsageKey 跨 voice/ucg 共享月度 AI 用量计数；TTL 90 天。
func AIQuotaUsageKey(feature string, wxID int64, monthBucket string) string {
	return fmt.Sprintf("%s%s:%d:%s", aiQuotaUsageKeyPrefix, feature, wxID, monthBucket)
}

// AILLMGateWaitingKey LLM 闸门排队计数；TTL 300s。
func AILLMGateWaitingKey(model string) string {
	return fmt.Sprintf("ai:llm:gate:%s:waiting", model)
}

// AILLMGateInflightKey LLM 闸门在途计数；TTL 由 profile timeout 决定。
func AILLMGateInflightKey(model string) string {
	return fmt.Sprintf("ai:llm:gate:%s:inflight", model)
}
