package cachekit

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// 扩展 Cache 接口：Hash / List / Set / ZSet 与 clinic 等场景所需命令。
// 把一个整数减一，返回减一后的值。
func (c *RedisCache) Decr(ctx context.Context, key string) (int64, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return 0, ErrInvalidKey
	}
	ret, err := g.Redis().Do(ctx, "DECR", key)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return ret.Int64(), nil
}

// 返回 key 的剩余生存时间（单位：秒），-1 表示永久存在，-2 表示已过期。
func (c *RedisCache) TTL(ctx context.Context, key string) (int, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return 0, ErrInvalidKey
	}
	ret, err := g.Redis().Do(ctx, "TTL", key)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return ret.Int(), nil
}

// Set 写入无过期时间的字符串值；仅用于 clinic session 等固定 TTL 续写场景。
// 不设置过期时间,沿用之前的过期时间(如果没有设置过,则永久存在)
func (c *RedisCache) Set(ctx context.Context, key, value string) error {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" {
		return ErrInvalidKey
	}
	if value == "" {
		return ErrEmptyValue
	}
	if _, err := g.Redis().Do(ctx, "SET", key, value); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

// 永不过期：将 key 的生存时间设置为永久存在。
func (c *RedisCache) Persist(ctx context.Context, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidKey
	}
	if _, err := g.Redis().Do(ctx, "PERSIST", key); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

func (c *RedisCache) HashIncrBy(ctx context.Context, key, field string, delta int64) (int64, error) {
	key = strings.TrimSpace(key)
	field = strings.TrimSpace(field)
	if key == "" || field == "" {
		return 0, ErrInvalidKey
	}
	ret, err := g.Redis().Do(ctx, "HINCRBY", key, field, delta)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return ret.Int64(), nil
}

// 存一个对象(对象名:key,对象属性:field,对象值:value)
// eg. 存一个用户对象(用户id:123,用户名:张三,用户年龄:20)
// redis.HashSet(ctx, "user:123", "name", "张三")
// redis.HashSet(ctx, "user:123", "age", "20")
func (c *RedisCache) HashSet(ctx context.Context, key, field, value string) error {
	key = strings.TrimSpace(key)
	field = strings.TrimSpace(field)
	if key == "" || field == "" {
		return ErrInvalidKey
	}
	if _, err := g.Redis().Do(ctx, "HSET", key, field, value); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

// 获取一个对象的属性值(对象名:key,对象属性:field)
// eg. 获取用户对象的姓名(用户id:123,用户名:张三,用户年龄:20)
// redis.HashGet(ctx, "user:123", "name")
// 返回: 20, true, nil
func (c *RedisCache) HashGet(ctx context.Context, key, field string) (string, bool, error) {
	key = strings.TrimSpace(key)
	field = strings.TrimSpace(field)
	if key == "" || field == "" {
		return "", false, ErrInvalidKey
	}
	ret, err := g.Redis().Do(ctx, "HGET", key, field)
	if err != nil {
		return "", false, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if ret == nil || ret.IsNil() {
		return "", false, nil
	}
	v := strings.TrimSpace(ret.String())
	if v == "" {
		return "", false, nil
	}
	return v, true, nil
}

// 获取一个对象的所有属性值(对象名:key)
// eg. 获取用户对象的所有属性值(用户id:123,用户名:张三,用户年龄:20)
// redis.HashGetAll(ctx, "user:123")
// 返回: map[string]string{"name": "张三", "age": "20"}, nil
func (c *RedisCache) HashGetAll(ctx context.Context, key string) (map[string]string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidKey
	}
	ret, err := g.Redis().Do(ctx, "HGETALL", key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return parseHashGetAllResult(ret)
}

// 将一个值插入到列表的末尾(列表名:key,值:value)
// eg. 将一个值插入到列表的末尾(列表名:user:123,值:张三)
// redis.ListPush(ctx, "user:123", "张三")
// 返回: nil
func (c *RedisCache) ListPush(ctx context.Context, key, value string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidKey
	}
	if _, err := g.Redis().Do(ctx, "RPUSH", key, value); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

// 返回列表的长度(列表名:key)
// eg. 返回列表的长度(列表名:user:123)
// redis.ListLen(ctx, "user:123")
// 返回: 1, nil
func (c *RedisCache) ListLen(ctx context.Context, key string) (int64, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return 0, ErrInvalidKey
	}
	ret, err := g.Redis().Do(ctx, "LLEN", key)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return ret.Int64(), nil
}

// 返回列表指定范围内的元素(列表名:key,起始索引:start,结束索引:end)
// eg. 返回列表指定范围内的元素(列表名:user:123,起始索引:0,结束索引:1)
// redis.ListRange(ctx, "user:123", 0, 1)
// 返回: []string{"张三", "李四"}, nil
func (c *RedisCache) ListRange(ctx context.Context, key string, start, end int64) ([]string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidKey
	}
	ret, err := g.Redis().Do(ctx, "LRANGE", key, start, end)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	arr := ret.Array()
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if item == nil {
			continue
		}
		out = append(out, g.NewVar(item).String())
	}
	return out, nil
}

// 返回列表指定索引的元素(列表名:key,索引:index)
// eg. 返回列表指定索引的元素(列表名:user:123,索引:0)
// redis.ListIndex(ctx, "user:123", 0)
// 返回: "张三", nil
func (c *RedisCache) ListIndex(ctx context.Context, key string, index int64) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", ErrInvalidKey
	}
	ret, err := g.Redis().Do(ctx, "LINDEX", key, index)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if ret == nil || ret.IsNil() {
		return "", nil
	}
	return ret.String(), nil
}

// 将列表指定索引的元素设置为指定值(列表名:key,索引:index,值:value)
// eg. 将列表指定索引的元素设置为指定值(列表名:user:123,索引:0,值:张三)
// redis.ListSet(ctx, "user:123", 0, "张三")
// 返回: nil
func (c *RedisCache) ListSet(ctx context.Context, key string, index int64, value string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidKey
	}
	if _, err := g.Redis().Do(ctx, "LSET", key, index, value); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

// 将一个或多个成员元素去重加入到集合中(集合名:key,成员:members)
// eg. 将一个或多个成员元素去重加入到集合中(集合名:user:123,成员:张三,李四)
// redis.SetAdd(ctx, "user:123", "张三", "李四")
// 返回: nil
func (c *RedisCache) SetAdd(ctx context.Context, key string, members ...string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidKey
	}
	if len(members) == 0 {
		return ErrEmptyValue
	}
	args := make([]interface{}, 0, 1+len(members))
	args = append(args, key)
	for _, m := range members {
		m = strings.TrimSpace(m)
		if m == "" {
			return ErrEmptyValue
		}
		args = append(args, m)
	}
	if _, err := g.Redis().Do(ctx, "SADD", args...); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

// 判断一个成员元素是否是集合的成员(集合名:key,成员:member)
// eg. 判断一个成员元素是否是集合的成员(集合名:user:123,成员:张三)
// redis.SetIsMember(ctx, "user:123", "张三")
// 返回: true, nil
func (c *RedisCache) SetIsMember(ctx context.Context, key, member string) (bool, error) {
	key = strings.TrimSpace(key)
	member = strings.TrimSpace(member)
	if key == "" || member == "" {
		return false, ErrInvalidKey
	}
	ret, err := g.Redis().Do(ctx, "SISMEMBER", key, member)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return ret.Int() == 1, nil
}

// 返回集合的所有成员(集合名:key)
// eg. 返回集合的所有成员(集合名:user:123)
// redis.SetMembers(ctx, "user:123")
// 返回: []string{"张三", "李四"}, nil
func (c *RedisCache) SetMembers(ctx context.Context, key string) ([]string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidKey
	}
	ret, err := g.Redis().Do(ctx, "SMEMBERS", key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	arr := ret.Array()
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if item == nil {
			continue
		}
		out = append(out, strings.TrimSpace(g.NewVar(item).String()))
	}
	return out, nil
}

// 将一个或多个成员元素从集合中移除(集合名:key,成员:members)
// eg. 将一个或多个成员元素从集合中移除(集合名:user:123,成员:张三,李四)
// redis.SetRemove(ctx, "user:123", "张三", "李四")
// 返回: nil
func (c *RedisCache) SetRemove(ctx context.Context, key string, members ...string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidKey
	}
	if len(members) == 0 {
		return ErrEmptyValue
	}
	args := make([]interface{}, 0, 1+len(members))
	args = append(args, key)
	for _, m := range members {
		m = strings.TrimSpace(m)
		if m == "" {
			return ErrEmptyValue
		}
		args = append(args, m)
	}
	if _, err := g.Redis().Do(ctx, "SREM", args...); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

// 将一个成员元素及其 score 值加入到有序集合中(有序集合名:key,成员:member,score:score)
// eg. 将一个成员元素及其 score 值加入到有序集合中(有序集合名:user:123,成员:张三,score:100)
// redis.SortedSetAdd(ctx, "user:123", 100, "张三")
// 返回: nil
func (c *RedisCache) SortedSetAdd(ctx context.Context, key string, score float64, member string) error {
	key = strings.TrimSpace(key)
	member = strings.TrimSpace(member)
	if key == "" || member == "" {
		return ErrInvalidKey
	}
	if _, err := g.Redis().Do(ctx, "ZADD", key, score, member); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

// ParseInt64 辅助：从 Get 返回值解析整数，miss 时返回 0。
func ParseInt64(val string, ok bool) int64 {
	if !ok {
		return 0
	}
	n, _ := strconv.ParseInt(strings.TrimSpace(val), 10, 64)
	return n
}
