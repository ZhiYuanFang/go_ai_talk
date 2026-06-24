package cachekit

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// GeoMemberDist GEO 半径查询结果：member 与距圆心距离（km）。
type GeoMemberDist struct {
	Member string
	DistKm float64
}

// ZSetMemberScore ZSET 成员与 score。
type ZSetMemberScore struct {
	Member string
	Score  float64
}

// GeoAdd 向 GEO 索引写入 member；Redis 坐标顺序为 (lng, lat)。
func (c *RedisCache) GeoAdd(ctx context.Context, key, member string, lng, lat float64) error {
	key = strings.TrimSpace(key)
	member = strings.TrimSpace(member)
	if key == "" || member == "" {
		return ErrInvalidKey
	}
	if _, err := g.Redis().Do(ctx, "GEOADD", key, lng, lat, member); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

// GeoRemove 从 GEO 索引移除 member（底层 ZREM）。
func (c *RedisCache) GeoRemove(ctx context.Context, key string, members ...string) error {
	return c.SortedSetRemove(ctx, key, members...)
}

// GeoSearchByRadiusWithDist 以 FROMLONLAT + BYRADIUS 查询 GEO 成员及距离（km）。
// radiusKm<=0 时视为不限半径（使用极大半径 20000km 近似全球）。
func (c *RedisCache) GeoSearchByRadiusWithDist(
	ctx context.Context,
	key string,
	lng, lat, radiusKm float64,
	count int,
) ([]GeoMemberDist, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidKey
	}
	if count <= 0 {
		count = 200
	}
	radius := radiusKm
	if radius <= 0 {
		radius = 20000
	}
	ret, err := g.Redis().Do(ctx, "GEOSEARCH", key,
		"FROMLONLAT", lng, lat,
		"BYRADIUS", radius, "km",
		"WITHDIST",
		"ASC",
		"COUNT", count,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return parseGeoSearchWithDist(ret), nil
}

func parseGeoSearchWithDist(ret interface{}) []GeoMemberDist {
	arr := g.NewVar(ret).Array()
	if len(arr) == 0 {
		return nil
	}
	out := make([]GeoMemberDist, 0, len(arr))
	for _, item := range arr {
		row := g.NewVar(item).Array()
		if len(row) < 2 {
			continue
		}
		member := strings.TrimSpace(g.NewVar(row[0]).String())
		distKm, _ := strconv.ParseFloat(strings.TrimSpace(g.NewVar(row[1]).String()), 64)
		if member == "" {
			continue
		}
		out = append(out, GeoMemberDist{Member: member, DistKm: distKm})
	}
	return out
}

// SortedSetRemove 批量 ZREM。
func (c *RedisCache) SortedSetRemove(ctx context.Context, key string, members ...string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidKey
	}
	if len(members) == 0 {
		return nil
	}
	args := make([]interface{}, 0, 1+len(members))
	args = append(args, key)
	for _, m := range members {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		args = append(args, m)
	}
	if len(args) <= 1 {
		return nil
	}
	if _, err := g.Redis().Do(ctx, "ZREM", args...); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

// SortedSetScores 批量 ZSCORE；缺失 member 不出现在 map 中。
func (c *RedisCache) SortedSetScores(ctx context.Context, key string, members []string) (map[string]float64, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidKey
	}
	out := make(map[string]float64, len(members))
	if len(members) == 0 {
		return out, nil
	}
	for _, m := range members {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		ret, err := g.Redis().Do(ctx, "ZSCORE", key, m)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
		}
		if ret == nil || ret.IsNil() {
			continue
		}
		score, _ := strconv.ParseFloat(strings.TrimSpace(ret.String()), 64)
		out[m] = score
	}
	return out, nil
}

// SortedSetRevRangeWithScores ZREVRANGE WITHSCORES，按 score 降序。
func (c *RedisCache) SortedSetRevRangeWithScores(
	ctx context.Context,
	key string,
	start, stop int64,
) ([]ZSetMemberScore, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidKey
	}
	ret, err := g.Redis().Do(ctx, "ZREVRANGE", key, start, stop, "WITHSCORES")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	arr := g.NewVar(ret).Array()
	out := make([]ZSetMemberScore, 0, len(arr)/2)
	for i := 0; i+1 < len(arr); i += 2 {
		member := strings.TrimSpace(g.NewVar(arr[i]).String())
		score, _ := strconv.ParseFloat(strings.TrimSpace(g.NewVar(arr[i+1]).String()), 64)
		if member == "" {
			continue
		}
		out = append(out, ZSetMemberScore{Member: member, Score: score})
	}
	return out, nil
}

// SetIsMemberBatch pipeline SISMEMBER，返回 member→是否成员。
func (c *RedisCache) SetIsMemberBatch(ctx context.Context, key string, members []string) (map[string]bool, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidKey
	}
	out := make(map[string]bool, len(members))
	if len(members) == 0 {
		return out, nil
	}
	for _, m := range members {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		ok, err := c.SetIsMember(ctx, key, m)
		if err != nil {
			return nil, err
		}
		out[m] = ok
	}
	return out, nil
}

// SetAddWithExpire SADD 成员并刷新 key TTL（Feed session 用）。
func (c *RedisCache) SetAddWithExpire(ctx context.Context, key string, ttl time.Duration, members ...string) error {
	if err := c.SetAdd(ctx, key, members...); err != nil {
		return err
	}
	if ttl <= 0 {
		return ErrInvalidTTL
	}
	return c.Expire(ctx, key, ttl)
}

// GeoPosBatch 批量 GEOPOS；在 GEO 索引中的 member 写入 map。
func (c *RedisCache) GeoPosBatch(ctx context.Context, key string, members []string) (map[string]struct{}, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidKey
	}
	out := make(map[string]struct{}, len(members))
	if len(members) == 0 {
		return out, nil
	}
	args := make([]interface{}, 0, 2+len(members))
	args = append(args, key)
	for _, m := range members {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		args = append(args, m)
	}
	if len(args) <= 1 {
		return out, nil
	}
	ret, err := g.Redis().Do(ctx, "GEOPOS", args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	arr := g.NewVar(ret).Array()
	idx := 0
	for _, m := range members {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		if idx >= len(arr) {
			break
		}
		pos := g.NewVar(arr[idx]).Array()
		idx++
		if len(pos) >= 2 && !g.NewVar(pos[0]).IsNil() && !g.NewVar(pos[1]).IsNil() {
			out[m] = struct{}{}
		}
	}
	return out, nil
}
