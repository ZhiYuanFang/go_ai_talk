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

