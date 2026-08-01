package voice

import (
	"context"
	"strconv"
	"time"

	"hello/internal/platform/cachekit"
)

// clinicCache Clinic 限流等 KV 访问统一经 cachekit（带 LoggingObserver）。
// 业务说明：原定义在 clinic_session.go；会话/摘要删除后迁至此，供 rate 键使用。
var clinicCache = cachekit.Default()

// checkClinicRateLimit 固定窗口限流检查（只读 GET，不 INCR）；超限返回 CodeClinicRateLimited（42901）。
func checkClinicRateLimit(ctx context.Context, wxID int64, cfg AIClinicConfig) error {
	window := cfg.RateLimitWindowSeconds
	maxReq := cfg.RateLimitMaxRequests
	if window <= 0 || maxReq <= 0 {
		return nil
	}
	key := cachekit.VoiceClinicRateKey(wxID)
	raw, ok, err := clinicCache.Get(ctx, key)
	if err != nil {
		return err
	}
	count := 0
	if ok {
		count, _ = strconv.Atoi(raw)
	}
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
	key := cachekit.VoiceClinicRateKey(wxID)
	count, err := clinicCache.Incr(ctx, key)
	if err != nil {
		return err
	}
	if count == 1 {
		_ = clinicCache.Expire(ctx, key, time.Duration(window)*time.Second)
	}
	return nil
}
