package aimodel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	gateWaitingKeyFmt  = "ai:llm:gate:%s:waiting"
	gateInflightKeyFmt = "ai:llm:gate:%s:inflight"
	gateWaitingTTL     = 300
)

// Acquire 在调用上游前获取 model 闸门槽位；队列满返回 ErrQueueFull。
// release MUST 在 defer 中调用，覆盖整段上游 HTTP/流式生命周期。
func Acquire(ctx context.Context, profile Profile) (release func(), err error) {
	normalizeProfile(&profile)
	model := NormalizeModel(profile.Model)
	if model == "" {
		return func() {}, fmt.Errorf("aimodel: model 为空")
	}
	waitKey := fmt.Sprintf(gateWaitingKeyFmt, model)
	inflightKey := fmt.Sprintf(gateInflightKeyFmt, model)
	maxWait := profile.MaxWaiters
	maxInflight := profile.MaxInFlight
	if maxWait <= 0 {
		// maxWaiters=0 表示不允许排队，仅尝试立即抢槽。
		maxWait = 0
	}

	joined := false
	defer func() {
		if err != nil && joined {
			_, _ = g.Redis().Do(ctx, "DECR", waitKey)
		}
	}()

	if maxWait > 0 {
		n, rErr := g.Redis().Do(ctx, "GET", waitKey)
		if rErr != nil {
			return nil, rErr
		}
		if n.Int() >= maxWait {
			return nil, NewQueueFullError()
		}
		wn, wErr := g.Redis().Do(ctx, "INCR", waitKey)
		if wErr != nil {
			return nil, wErr
		}
		if wn.Int() == 1 {
			_, _ = g.Redis().Do(ctx, "EXPIRE", waitKey, gateWaitingTTL)
		}
		if wn.Int() > maxWait {
			_, _ = g.Redis().Do(ctx, "DECR", waitKey)
			return nil, NewQueueFullError()
		}
		joined = true
	} else {
		// 无缓冲池：当前 inflight 已满则立即拒绝。
		in, rErr := g.Redis().Do(ctx, "GET", inflightKey)
		if rErr != nil {
			return nil, rErr
		}
		if in.Int() >= maxInflight {
			return nil, NewQueueFullError()
		}
	}

	ttl := profile.TimeoutSec
	if ttl <= 0 {
		ttl = 120
	}
	ttl += 30 // 持槽 TTL 兜底，防止进程崩溃泄漏

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		got, aErr := tryAcquireInflight(ctx, inflightKey, maxInflight, ttl)
		if aErr != nil {
			return nil, aErr
		}
		if got {
			if joined {
				_, _ = g.Redis().Do(ctx, "DECR", waitKey)
				joined = false
			}
			return func() {
				releaseCtx := context.Background()
				v, _ := g.Redis().Do(releaseCtx, "GET", inflightKey)
				if v.Int() > 0 {
					_, _ = g.Redis().Do(releaseCtx, "DECR", inflightKey)
				}
			}, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func tryAcquireInflight(ctx context.Context, inflightKey string, maxInflight, ttlSec int) (bool, error) {
	script := `
local cur = tonumber(redis.call('GET', KEYS[1]) or '0')
if cur >= tonumber(ARGV[1]) then
  return 0
end
redis.call('INCR', KEYS[1])
redis.call('EXPIRE', KEYS[1], tonumber(ARGV[2]))
return 1
`
	res, err := g.Redis().Do(ctx, "EVAL", script, 1, inflightKey, maxInflight, ttlSec)
	if err != nil {
		return false, err
	}
	return res.Int() == 1, nil
}

// GateKeyForModel 返回规范化 model 的 inflight 键（可观测用）。
func GateKeyForModel(model string) string {
	return fmt.Sprintf(gateInflightKeyFmt, NormalizeModel(model))
}

// ProviderKeyEnv 返回 provider 对应的环境变量名（仅用于错误提示）。
func ProviderKeyEnv(provider Provider) string {
	switch provider {
	case ProviderZhipu:
		return "GLM_API_KEY"
	case ProviderDeepSeek:
		return "DEEPSEEK_API_KEY"
	case ProviderDashScope:
		return "UCG_DASHSCOPE_API_KEY"
	default:
		return strings.TrimSpace(string(provider)) + "_API_KEY"
	}
}
