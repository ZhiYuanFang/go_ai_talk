package voice

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
)

func clinicRateRedisKey(wxID int64) string {
	return fmt.Sprintf("%s%d", clinicRateKeyPrefix, wxID)
}

// checkClinicRateLimit 固定窗口限流；超限返回 CodeClinicRateLimited（42901），不消耗 clinic_ai 额度。
func checkClinicRateLimit(ctx context.Context, wxID int64, cfg AIClinicConfig) error {
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
	if count > maxReq {
		return &VoiceAIQuotaError{Code: CodeClinicRateLimited, Message: "请求过于频繁，请稍后再试"}
	}
	return nil
}
