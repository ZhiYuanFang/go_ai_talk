package cachekit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type Cache interface {
	// Ping 用于启动期与运行期依赖探测，失败即表示 Redis 不可用。
	Ping(ctx context.Context) error
	Get(ctx context.Context, key string) (string, bool, error)
	MGet(ctx context.Context, keys []string) (map[string]string, error)
	Set(ctx context.Context, key, value string) error
	SetEX(ctx context.Context, key, value string, ttl time.Duration) error
	SetNXEX(ctx context.Context, key, value string, ttl time.Duration) (bool, error)
	Exists(ctx context.Context, key string) (bool, error)
	Del(ctx context.Context, key string) error
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, delta int64) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (int, error)
	Persist(ctx context.Context, key string) error
	HashIncrBy(ctx context.Context, key, field string, delta int64) (int64, error)
	HashSet(ctx context.Context, key, field, value string) error
	HashGet(ctx context.Context, key, field string) (string, bool, error)
	HashGetAll(ctx context.Context, key string) (map[string]string, error)
	ListPush(ctx context.Context, key, value string) error
	ListLen(ctx context.Context, key string) (int64, error)
	ListRange(ctx context.Context, key string, start, end int64) ([]string, error)
	ListIndex(ctx context.Context, key string, index int64) (string, error)
	ListSet(ctx context.Context, key string, index int64, value string) error
	SetAdd(ctx context.Context, key string, members ...string) error
	SetIsMember(ctx context.Context, key, member string) (bool, error)
	SetMembers(ctx context.Context, key string) ([]string, error)
	SetRemove(ctx context.Context, key string, members ...string) error
	SortedSetAdd(ctx context.Context, key string, score float64, member string) error
	GeoAdd(ctx context.Context, key, member string, lng, lat float64) error
	GeoRemove(ctx context.Context, key string, members ...string) error
	GeoSearchByRadiusWithDist(ctx context.Context, key string, lng, lat, radiusKm float64, count int) ([]GeoMemberDist, error)
	GeoPosBatch(ctx context.Context, key string, members []string) (map[string]struct{}, error)
	SortedSetRemove(ctx context.Context, key string, members ...string) error
	SortedSetScores(ctx context.Context, key string, members []string) (map[string]float64, error)
	SortedSetRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ZSetMemberScore, error)
	SortedSetCard(ctx context.Context, key string) (int64, error)
	SetIsMemberBatch(ctx context.Context, key string, members []string) (map[string]bool, error)
	SetAddWithExpire(ctx context.Context, key string, ttl time.Duration, members ...string) error
	// Eval 执行 Lua 脚本，用于跨键原子补丁更新。
	Eval(ctx context.Context, script string, keys []string, args []string) (string, error)
}

type RedisCache struct{}

func NewRedisCache() *RedisCache {
	return &RedisCache{}
}

func (c *RedisCache) Ping(ctx context.Context) error {
	if _, err := g.Redis().Do(ctx, "PING"); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, bool, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", false, ErrInvalidKey
	}
	raw, err := g.Redis().Do(ctx, "GET", key)
	if err != nil {
		return "", false, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	v := strings.TrimSpace(raw.String())
	if v == "" {
		// 空字符串统一视为“不存在”，避免调用方区分 nil/empty。
		return "", false, nil
	}
	return v, true, nil
}

func (c *RedisCache) MGet(ctx context.Context, keys []string) (map[string]string, error) {
	normalized, err := normalizeMGetKeys(keys)
	if err != nil {
		return nil, err
	}
	if len(normalized) == 0 {
		return map[string]string{}, nil
	}
	if !mgetKeysSpanMultipleSlots(normalized) {
		return c.mgetOnce(ctx, normalized)
	}
	// 跨 slot：单机可一次 MGET；Cluster 会 CROSSSLOT，再按 slot 分组。
	out, err := c.mgetOnce(ctx, normalized)
	if err == nil {
		return out, nil
	}
	if !isCrossSlotErr(err) {
		return nil, err
	}
	return c.mgetByClusterSlot(ctx, normalized)
}

func normalizeMGetKeys(keys []string) ([]string, error) {
	normalized := make([]string, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, ErrInvalidKey
		}
		normalized = append(normalized, key)
	}
	return normalized, nil
}

func isCrossSlotErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToUpper(err.Error()), "CROSSSLOT")
}

// mgetKeysSpanMultipleSlots 客户端预判 MGET 是否跨 Cluster slot（含 hash tag 规则）。
func mgetKeysSpanMultipleSlots(keys []string) bool {
	if len(keys) <= 1 {
		return false
	}
	slot := redisClusterSlot(keys[0])
	for _, key := range keys[1:] {
		if redisClusterSlot(key) != slot {
			return true
		}
	}
	return false
}

func (c *RedisCache) mgetOnce(ctx context.Context, keys []string) (map[string]string, error) {
	args := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		args = append(args, key)
	}
	ret, err := g.Redis().Do(ctx, "MGET", args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return parseMGetResult(keys, ret.Array()), nil
}

func (c *RedisCache) mgetByClusterSlot(ctx context.Context, keys []string) (map[string]string, error) {
	groups := make(map[int][]string)
	order := make([]int, 0)
	for _, key := range keys {
		slot := redisClusterSlot(key)
		if _, ok := groups[slot]; !ok {
			order = append(order, slot)
		}
		groups[slot] = append(groups[slot], key)
	}
	out := make(map[string]string, len(keys))
	for _, slot := range order {
		batch, err := c.mgetOnce(ctx, groups[slot])
		if err != nil {
			return nil, err
		}
		for k, v := range batch {
			out[k] = v
		}
	}
	return out, nil
}

func parseMGetResult(keys []string, vals []interface{}) map[string]string {
	out := make(map[string]string, len(keys))
	for idx, raw := range vals {
		if idx >= len(keys) || raw == nil {
			continue
		}
		v := strings.TrimSpace(g.NewVar(raw).String())
		if v == "" {
			continue
		}
		out[keys[idx]] = v
	}
	return out
}

func (c *RedisCache) SetEX(ctx context.Context, key, value string, ttl time.Duration) error {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" {
		return ErrInvalidKey
	}
	if value == "" {
		return ErrEmptyValue
	}
	if ttl <= 0 {
		return ErrInvalidTTL
	}
	// 使用秒级 TTL，统一跨服务缓存过期语义。
	if _, err := g.Redis().Do(ctx, "SET", key, value, "EX", int(ttl.Seconds())); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

func (c *RedisCache) SetNXEX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" {
		return false, ErrInvalidKey
	}
	if value == "" {
		return false, ErrEmptyValue
	}
	if ttl <= 0 {
		return false, ErrInvalidTTL
	}
	ret, err := g.Redis().Do(ctx, "SET", key, value, "NX", "EX", int(ttl.Seconds()))
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return strings.EqualFold(strings.TrimSpace(ret.String()), "OK"), nil
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return false, ErrInvalidKey
	}
	ret, err := g.Redis().Do(ctx, "EXISTS", key)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return ret.Int() > 0, nil
}

func (c *RedisCache) Del(ctx context.Context, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidKey
	}
	if _, err := g.Redis().Do(ctx, "DEL", key); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

func (c *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return 0, ErrInvalidKey
	}
	ret, err := g.Redis().Do(ctx, "INCR", key)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return ret.Int64(), nil
}

func (c *RedisCache) IncrBy(ctx context.Context, key string, delta int64) (int64, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return 0, ErrInvalidKey
	}
	if delta == 0 {
		return 0, ErrInvalidIncrBy
	}
	ret, err := g.Redis().Do(ctx, "INCRBY", key, delta)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return ret.Int64(), nil
}

func (c *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidKey
	}
	if ttl <= 0 {
		return ErrInvalidTTL
	}
	if _, err := g.Redis().Do(ctx, "EXPIRE", key, int(ttl.Seconds())); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

func (c *RedisCache) Eval(ctx context.Context, script string, keys []string, args []string) (string, error) {
	script = strings.TrimSpace(script)
	if script == "" {
		return "", ErrEmptyValue
	}
	cmd := make([]interface{}, 0, 1+len(keys)+len(args))
	cmd = append(cmd, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			return "", ErrInvalidKey
		}
		cmd = append(cmd, key)
	}
	for _, arg := range args {
		cmd = append(cmd, strings.TrimSpace(arg))
	}
	ret, err := g.Redis().Do(ctx, "EVAL", append([]interface{}{script}, cmd...)...)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return strings.TrimSpace(ret.String()), nil
}

