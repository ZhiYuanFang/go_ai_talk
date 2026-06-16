package voice

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
)

func clinicRateRedisKey(wxID int64) string {
	return fmt.Sprintf("%s%d", clinicRateKeyPrefix, wxID)
}

// checkClinicRateLimit 固定窗口限流检查（只读 GET，不 INCR）；超限返回 CodeClinicRateLimited（42901）。
// 限流计数仅在 answer_done 成功后由 recordClinicRateLimitOnSuccess 递增；cancel/supersede 不计入。
func checkClinicRateLimit(ctx context.Context, wxID int64, cfg AIClinicConfig) error {
	window := cfg.RateLimitWindowSeconds
	maxReq := cfg.RateLimitMaxRequests
	if window <= 0 || maxReq <= 0 {
		return nil
	}
	key := clinicRateRedisKey(wxID)
	n, err := g.Redis().Do(ctx, "GET", key)
	if err != nil {
		return err
	}
	count := n.Int()
	if count >= maxReq {
		return &VoiceAIQuotaError{Code: CodeClinicRateLimited, Message: "请求过于频繁，请稍后再试"}
	}
	return nil
}

// recordClinicRateLimitOnSuccess 在 turn 成功完成（answer_done）后递增限流计数；cancel 路径 MUST NOT 调用。
func recordClinicRateLimitOnSuccess(ctx context.Context, wxID int64, cfg AIClinicConfig) error {
	window := cfg.RateLimitWindowSeconds
	maxReq := cfg.RateLimitMaxRequests
	if window <= 0 || maxReq <= 0 {
		return nil
	}
	key := clinicRateRedisKey(wxID)
	n, err := g.Redis().Do(ctx, "INCR", key)
	if err != nil {
		return err
	}
	count := n.Int()
	if count == 1 {
		_, _ = g.Redis().Do(ctx, "EXPIRE", key, window)
	}
	return nil
}
