package aimodel

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hello/internal/platform/cachekit"
)

const gateWaitingTTL = 300

var gateCache = cachekit.Default()

// Acquire 在调用上游前获取 model 闸门槽位；队列满返回 ErrQueueFull。
func Acquire(ctx context.Context, profile Profile) (release func(), err error) {
	normalizeProfile(&profile)
	model := NormalizeModel(profile.Model)
	if model == "" {
		return func() {}, fmt.Errorf("aimodel: model 为空")
	}
	waitKey := cachekit.AILLMGateWaitingKey(model)
	inflightKey := cachekit.AILLMGateInflightKey(model)
	maxWait := profile.MaxWaiters
	maxInflight := profile.MaxInFlight
	if maxWait <= 0 {
		maxWait = 0
	}

	joined := false
	defer func() {
		if err != nil && joined {
			_, _ = gateCache.Decr(ctx, waitKey)
		}
	}()

	if maxWait > 0 {
		waiting, rErr := gateInt(ctx, waitKey)
		if rErr != nil {
			return nil, rErr
		}
		if waiting >= maxWait {
			return nil, NewQueueFullError()
		}
		wn, wErr := gateCache.Incr(ctx, waitKey)
		if wErr != nil {
			return nil, wErr
		}
		if wn == 1 {
			_ = gateCache.Expire(ctx, waitKey, gateWaitingTTL*time.Second)
		}
		if int(wn) > maxWait {
			_, _ = gateCache.Decr(ctx, waitKey)
			return nil, NewQueueFullError()
		}
		joined = true
	} else {
		inflight, rErr := gateInt(ctx, inflightKey)
		if rErr != nil {
			return nil, rErr
		}
		if inflight >= maxInflight {
			return nil, NewQueueFullError()
		}
	}

	ttl := profile.TimeoutSec
	if ttl <= 0 {
		ttl = 120
	}
	ttl += 30

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
				_, _ = gateCache.Decr(ctx, waitKey)
				joined = false
			}
			return func() {
				releaseCtx := context.Background()
				n, _ := gateInt(releaseCtx, inflightKey)
				if n > 0 {
					_, _ = gateCache.Decr(releaseCtx, inflightKey)
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

func gateInt(ctx context.Context, key string) (int, error) {
	raw, ok, err := gateCache.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}
	n, _ := strconv.Atoi(strings.TrimSpace(raw))
	return n, nil
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
	ret, err := gateCache.Eval(ctx, script, []string{inflightKey}, []string{strconv.Itoa(maxInflight), strconv.Itoa(ttlSec)})
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(ret) == "1", nil
}

// GateKeyForModel 返回规范化 model 的 inflight 键（可观测用）。
func GateKeyForModel(model string) string {
	return cachekit.AILLMGateInflightKey(NormalizeModel(model))
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
