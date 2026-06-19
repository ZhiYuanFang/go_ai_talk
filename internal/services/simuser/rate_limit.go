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
	outboundLimiterOnce sync.Once
	outboundLimiter     *rate.Limiter
)

func outboundRateLimiter() *rate.Limiter {
	outboundLimiterOnce.Do(func() {
		rps := envFloat("SIM_UCG_RATE_LIMIT_RPS", 2.0)
		if rps <= 0 {
			rps = 2.0
		}
		burst := envInt("SIM_UCG_RATE_LIMIT_BURST", 4)
		if burst <= 0 {
			burst = 4
		}
		outboundLimiter = rate.NewLimiter(rate.Limit(rps), burst)
	})
	return outboundLimiter
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
