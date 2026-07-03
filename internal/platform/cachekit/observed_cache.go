package cachekit

import (
	"context"
	"time"
)

type observedCache struct {
	base     Cache
	observer CacheObserver
}

func WithObserver(base Cache, observer CacheObserver) Cache {
	if observer == nil {
		observer = NoopObserver{}
	}
	return &observedCache{
		base:     base,
		observer: observer,
	}
}

func (c *observedCache) Ping(ctx context.Context) error {
	begin := time.Now()
	err := c.base.Ping(ctx)
	// 统一把调用耗时与错误上报给观察器，便于接日志/指标。
	c.observer.OnOperation(ctx, "ping", "-", time.Since(begin), err)
	return err
}

func (c *observedCache) Get(ctx context.Context, key string) (string, bool, error) {
	begin := time.Now()
	val, ok, err := c.base.Get(ctx, key)
	c.observer.OnOperation(ctx, "get", key, time.Since(begin), err)
	return val, ok, err
}

func (c *observedCache) MGet(ctx context.Context, keys []string) (map[string]string, error) {
	begin := time.Now()
	// 多 key 观测以首 key 为标识，避免超长日志。
	marker := "-"
	if len(keys) > 0 {
		marker = keys[0]
	}
	vals, err := c.base.MGet(ctx, keys)
	c.observer.OnOperation(ctx, "mget", marker, time.Since(begin), err)
	return vals, err
}

func (c *observedCache) SetEX(ctx context.Context, key, value string, ttl time.Duration) error {
	begin := time.Now()
	err := c.base.SetEX(ctx, key, value, ttl)
	c.observer.OnOperation(ctx, "setex", key, time.Since(begin), err)
	return err
}

func (c *observedCache) Set(ctx context.Context, key, value string) error {
	begin := time.Now()
	err := c.base.Set(ctx, key, value)
	c.observer.OnOperation(ctx, "set", key, time.Since(begin), err)
	return err
}

func (c *observedCache) SetNXEX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	begin := time.Now()
	ok, err := c.base.SetNXEX(ctx, key, value, ttl)
	c.observer.OnOperation(ctx, "setnxex", key, time.Since(begin), err)
	return ok, err
}

func (c *observedCache) Exists(ctx context.Context, key string) (bool, error) {
	begin := time.Now()
	ok, err := c.base.Exists(ctx, key)
	c.observer.OnOperation(ctx, "exists", key, time.Since(begin), err)
	return ok, err
}

func (c *observedCache) Del(ctx context.Context, key string) error {
	begin := time.Now()
	err := c.base.Del(ctx, key)
	c.observer.OnOperation(ctx, "del", key, time.Since(begin), err)
	return err
}

func (c *observedCache) Incr(ctx context.Context, key string) (int64, error) {
	begin := time.Now()
	val, err := c.base.Incr(ctx, key)
	c.observer.OnOperation(ctx, "incr", key, time.Since(begin), err)
	return val, err
}

func (c *observedCache) Decr(ctx context.Context, key string) (int64, error) {
	begin := time.Now()
	val, err := c.base.Decr(ctx, key)
	c.observer.OnOperation(ctx, "decr", key, time.Since(begin), err)
	return val, err
}

func (c *observedCache) IncrBy(ctx context.Context, key string, delta int64) (int64, error) {
	begin := time.Now()
	val, err := c.base.IncrBy(ctx, key, delta)
	c.observer.OnOperation(ctx, "incrby", key, time.Since(begin), err)
	return val, err
}

func (c *observedCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	begin := time.Now()
	err := c.base.Expire(ctx, key, ttl)
	c.observer.OnOperation(ctx, "expire", key, time.Since(begin), err)
	return err
}

func (c *observedCache) TTL(ctx context.Context, key string) (int, error) {
	begin := time.Now()
	val, err := c.base.TTL(ctx, key)
	c.observer.OnOperation(ctx, "ttl", key, time.Since(begin), err)
	return val, err
}

func (c *observedCache) Persist(ctx context.Context, key string) error {
	begin := time.Now()
	err := c.base.Persist(ctx, key)
	c.observer.OnOperation(ctx, "persist", key, time.Since(begin), err)
	return err
}

func (c *observedCache) HashIncrBy(ctx context.Context, key, field string, delta int64) (int64, error) {
	begin := time.Now()
	val, err := c.base.HashIncrBy(ctx, key, field, delta)
	c.observer.OnOperation(ctx, "hincrby", key, time.Since(begin), err)
	return val, err
}

func (c *observedCache) HashSet(ctx context.Context, key, field, value string) error {
	begin := time.Now()
	err := c.base.HashSet(ctx, key, field, value)
	c.observer.OnOperation(ctx, "hset", key, time.Since(begin), err)
	return err
}

func (c *observedCache) HashGet(ctx context.Context, key, field string) (string, bool, error) {
	begin := time.Now()
	val, ok, err := c.base.HashGet(ctx, key, field)
	c.observer.OnOperation(ctx, "hget", key, time.Since(begin), err)
	return val, ok, err
}

func (c *observedCache) HashGetAll(ctx context.Context, key string) (map[string]string, error) {
	begin := time.Now()
	vals, err := c.base.HashGetAll(ctx, key)
	c.observer.OnOperation(ctx, "hgetall", key, time.Since(begin), err)
	return vals, err
}

func (c *observedCache) ListPush(ctx context.Context, key, value string) error {
	begin := time.Now()
	err := c.base.ListPush(ctx, key, value)
	c.observer.OnOperation(ctx, "rpush", key, time.Since(begin), err)
	return err
}

func (c *observedCache) ListLen(ctx context.Context, key string) (int64, error) {
	begin := time.Now()
	val, err := c.base.ListLen(ctx, key)
	c.observer.OnOperation(ctx, "llen", key, time.Since(begin), err)
	return val, err
}

func (c *observedCache) ListRange(ctx context.Context, key string, start, end int64) ([]string, error) {
	begin := time.Now()
	vals, err := c.base.ListRange(ctx, key, start, end)
	c.observer.OnOperation(ctx, "lrange", key, time.Since(begin), err)
	return vals, err
}

func (c *observedCache) ListIndex(ctx context.Context, key string, index int64) (string, error) {
	begin := time.Now()
	val, err := c.base.ListIndex(ctx, key, index)
	c.observer.OnOperation(ctx, "lindex", key, time.Since(begin), err)
	return val, err
}

func (c *observedCache) ListSet(ctx context.Context, key string, index int64, value string) error {
	begin := time.Now()
	err := c.base.ListSet(ctx, key, index, value)
	c.observer.OnOperation(ctx, "lset", key, time.Since(begin), err)
	return err
}

func (c *observedCache) SetAdd(ctx context.Context, key string, members ...string) error {
	begin := time.Now()
	err := c.base.SetAdd(ctx, key, members...)
	c.observer.OnOperation(ctx, "sadd", key, time.Since(begin), err)
	return err
}

func (c *observedCache) SetIsMember(ctx context.Context, key, member string) (bool, error) {
	begin := time.Now()
	ok, err := c.base.SetIsMember(ctx, key, member)
	c.observer.OnOperation(ctx, "sismember", key, time.Since(begin), err)
	return ok, err
}

func (c *observedCache) SetMembers(ctx context.Context, key string) ([]string, error) {
	begin := time.Now()
	vals, err := c.base.SetMembers(ctx, key)
	c.observer.OnOperation(ctx, "smembers", key, time.Since(begin), err)
	return vals, err
}

func (c *observedCache) SetRemove(ctx context.Context, key string, members ...string) error {
	begin := time.Now()
	err := c.base.SetRemove(ctx, key, members...)
	c.observer.OnOperation(ctx, "srem", key, time.Since(begin), err)
	return err
}

func (c *observedCache) SortedSetAdd(ctx context.Context, key string, score float64, member string) error {
	begin := time.Now()
	err := c.base.SortedSetAdd(ctx, key, score, member)
	c.observer.OnOperation(ctx, "zadd", key, time.Since(begin), err)
	return err
}

func (c *observedCache) GeoAdd(ctx context.Context, key, member string, lng, lat float64) error {
	begin := time.Now()
	err := c.base.GeoAdd(ctx, key, member, lng, lat)
	c.observer.OnOperation(ctx, "geoadd", key, time.Since(begin), err)
	return err
}

func (c *observedCache) GeoRemove(ctx context.Context, key string, members ...string) error {
	begin := time.Now()
	err := c.base.GeoRemove(ctx, key, members...)
	c.observer.OnOperation(ctx, "georem", key, time.Since(begin), err)
	return err
}

func (c *observedCache) GeoSearchByRadiusWithDist(ctx context.Context, key string, lng, lat, radiusKm float64, count int) ([]GeoMemberDist, error) {
	begin := time.Now()
	vals, err := c.base.GeoSearchByRadiusWithDist(ctx, key, lng, lat, radiusKm, count)
	c.observer.OnOperation(ctx, "geosearch", key, time.Since(begin), err)
	return vals, err
}

func (c *observedCache) GeoPosBatch(ctx context.Context, key string, members []string) (map[string]struct{}, error) {
	begin := time.Now()
	vals, err := c.base.GeoPosBatch(ctx, key, members)
	c.observer.OnOperation(ctx, "geopos", key, time.Since(begin), err)
	return vals, err
}

func (c *observedCache) SortedSetRemove(ctx context.Context, key string, members ...string) error {
	begin := time.Now()
	err := c.base.SortedSetRemove(ctx, key, members...)
	c.observer.OnOperation(ctx, "zrem", key, time.Since(begin), err)
	return err
}

func (c *observedCache) SortedSetScores(ctx context.Context, key string, members []string) (map[string]float64, error) {
	begin := time.Now()
	vals, err := c.base.SortedSetScores(ctx, key, members)
	c.observer.OnOperation(ctx, "zscore", key, time.Since(begin), err)
	return vals, err
}

func (c *observedCache) SortedSetRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ZSetMemberScore, error) {
	begin := time.Now()
	vals, err := c.base.SortedSetRevRangeWithScores(ctx, key, start, stop)
	c.observer.OnOperation(ctx, "zrevrange", key, time.Since(begin), err)
	return vals, err
}

func (c *observedCache) SortedSetRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	begin := time.Now()
	vals, err := c.base.SortedSetRange(ctx, key, start, stop)
	c.observer.OnOperation(ctx, "zrange", key, time.Since(begin), err)
	return vals, err
}

func (c *observedCache) SortedSetCard(ctx context.Context, key string) (int64, error) {
	begin := time.Now()
	n, err := c.base.SortedSetCard(ctx, key)
	c.observer.OnOperation(ctx, "zcard", key, time.Since(begin), err)
	return n, err
}

func (c *observedCache) SetIsMemberBatch(ctx context.Context, key string, members []string) (map[string]bool, error) {
	begin := time.Now()
	vals, err := c.base.SetIsMemberBatch(ctx, key, members)
	c.observer.OnOperation(ctx, "sismember", key, time.Since(begin), err)
	return vals, err
}

func (c *observedCache) SetAddWithExpire(ctx context.Context, key string, ttl time.Duration, members ...string) error {
	begin := time.Now()
	err := c.base.SetAddWithExpire(ctx, key, ttl, members...)
	c.observer.OnOperation(ctx, "sadd", key, time.Since(begin), err)
	return err
}

func (c *observedCache) Eval(ctx context.Context, script string, keys []string, args []string) (string, error) {
	begin := time.Now()
	marker := "-"
	if len(keys) > 0 {
		marker = keys[0]
	}
	ret, err := c.base.Eval(ctx, script, keys, args)
	c.observer.OnOperation(ctx, "eval", marker, time.Since(begin), err)
	return ret, err
}

