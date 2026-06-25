package simuser

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/time/rate"
)

var (
	outboundLimiterMu sync.RWMutex
	outboundLimiter   *rate.Limiter
	activeRateLimit   *RateLimitSettings
)

// RateLimitSettings 出站 HTTP 限速配置。
type RateLimitSettings struct {
	RPS   float64
	Burst int
}

// setActiveRateLimit Admin/DB 更新后重置进程内 limiter。
func setActiveRateLimit(rl RateLimitSettings) {
	if rl.RPS <= 0 {
		rl.RPS = 2.0
	}
	if rl.Burst <= 0 {
		rl.Burst = 4
	}
	outboundLimiterMu.Lock()
	activeRateLimit = &RateLimitSettings{RPS: rl.RPS, Burst: rl.Burst}
	outboundLimiter = rate.NewLimiter(rate.Limit(rl.RPS), rl.Burst)
	outboundLimiterMu.Unlock()
}

// LoadRateLimitSettings 读取生效限速：内存 active > DB runtime > env。
func LoadRateLimitSettings() RateLimitSettings {
	outboundLimiterMu.RLock()
	if activeRateLimit != nil {
		rl := *activeRateLimit
		outboundLimiterMu.RUnlock()
		return rl
	}
	outboundLimiterMu.RUnlock()

	db, err := LoadRuntimeFromDB(context.Background())
	if err == nil && db.RateLimitRps > 0 {
		burst := db.RateLimitBurst
		if burst <= 0 {
			burst = 4
		}
		return RateLimitSettings{RPS: db.RateLimitRps, Burst: burst}
	}
	rps := envFloat("SIM_UCG_RATE_LIMIT_RPS", 2.0)
	if rps <= 0 {
		rps = 2.0
	}
	burst := envInt("SIM_UCG_RATE_LIMIT_BURST", 4)
	if burst <= 0 {
		burst = 4
	}
	return RateLimitSettings{RPS: rps, Burst: burst}
}

func outboundRateLimiter() *rate.Limiter {
	outboundLimiterMu.RLock()
	lim := outboundLimiter
	outboundLimiterMu.RUnlock()
	if lim != nil {
		return lim
	}
	rl := LoadRateLimitSettings()
	setActiveRateLimit(rl)
	outboundLimiterMu.RLock()
	lim = outboundLimiter
	outboundLimiterMu.RUnlock()
	return lim
}

// waitOutboundRate 阻塞至限速许可可用；ctx 取消时返回。
func waitOutboundRate(ctx context.Context, kind string) error {
	if err := outboundRateLimiter().Wait(ctx); err != nil {
		return err
	}
	g.Log().Debugf(ctx, "[simuser] rate-limit passed kind=%s", kind)
	return nil
}

func envFloat(key string, def float64) float64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil || f <= 0 {
		return def
	}
	return f
}

func envInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
