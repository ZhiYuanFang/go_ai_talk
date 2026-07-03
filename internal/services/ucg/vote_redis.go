package ucg

import (
	"context"
	"strconv"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/frame/g"
)

// initPostVoteCountsRedis 辩论帖发布时初始化计数 Hash，避免读路径把「零票」误判为 cache miss。
func initPostVoteCountsRedis(ctx context.Context, postID uint64) error {
	if postID == 0 {
		return nil
	}
	key := cachekit.UCGPostVoteCountsKey(postID)
	exists, err := ucgCache.Exists(ctx, key)
	if err != nil || exists {
		return err
	}
	if err = ucgCache.HashSet(ctx, key, VoteSideLeft, "0"); err != nil {
		return err
	}
	return ucgCache.HashSet(ctx, key, VoteSideRight, "0")
}

// patchVoteRedis VotePost 写路径同步 patch Redis（分 key 写入，兼容 Redis Cluster，避免跨 slot Lua）。
func patchVoteRedis(ctx context.Context, wxID int64, postID uint64, oldSide, newSide string) error {
	if wxID <= 0 || postID == 0 || newSide == "" {
		return nil
	}
	if oldSide == newSide {
		return nil
	}
	countsKey := cachekit.UCGPostVoteCountsKey(postID)
	if oldSide != "" && oldSide != newSide {
		if _, err := ucgCache.HashIncrBy(ctx, countsKey, oldSide, -1); err != nil {
			return err
		}
	}
	if oldSide == "" || oldSide != newSide {
		if _, err := ucgCache.HashIncrBy(ctx, countsKey, newSide, 1); err != nil {
			return err
		}
	}
	postField := strconv.FormatUint(postID, 10)
	return ucgCache.HashSet(ctx, cachekit.UCGUserPostVotesKey(wxID), postField, newSide)
}

func parseVoteCountsHash(fields map[string]string) voteCounts {
	var c voteCounts
	if fields == nil {
		return c
	}
	if v, ok := fields[VoteSideLeft]; ok {
		n, _ := strconv.ParseUint(v, 10, 64)
		c.left = uint(n)
	}
	if v, ok := fields[VoteSideRight]; ok {
		n, _ := strconv.ParseUint(v, 10, 64)
		c.right = uint(n)
	}
	return c
}

func voteCountsTotal(c voteCounts) uint {
	return c.left + c.right
}

// loadVoteCountsFromRedis 批量读帖级 left/right 计数；key 不存在视为 miss（需 MySQL backfill）。
func loadVoteCountsFromRedis(ctx context.Context, postIDs []uint64) (map[uint64]voteCounts, []uint64, error) {
	out := make(map[uint64]voteCounts, len(postIDs))
	miss := make([]uint64, 0)
	if len(postIDs) == 0 {
		return out, miss, nil
	}
	for _, id := range postIDs {
		key := cachekit.UCGPostVoteCountsKey(id)
		exists, err := ucgCache.Exists(ctx, key)
		if err != nil {
			return nil, nil, err
		}
		if !exists {
			miss = append(miss, id)
			continue
		}
		fields, err := ucgCache.HashGetAll(ctx, key)
		if err != nil {
			return nil, nil, err
		}
		out[id] = parseVoteCountsHash(fields)
	}
	return out, miss, nil
}

// reconcileVoteCountsWithMySQL 对 Redis miss 或计数全 0 的帖，批量与 MySQL 对齐并回写 Hash（修复 patch 失败/Cluster 脏 0）。
func reconcileVoteCountsWithMySQL(ctx context.Context, counts map[uint64]voteCounts, miss []uint64) (map[uint64]voteCounts, error) {
	if counts == nil {
		counts = make(map[uint64]voteCounts)
	}
	need := make([]uint64, 0, len(miss))
	seen := make(map[uint64]struct{}, len(miss))
	for _, id := range miss {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		need = append(need, id)
	}
	for id, c := range counts {
		if id == 0 {
			continue
		}
		if voteCountsTotal(c) > 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		need = append(need, id)
	}
	if len(need) == 0 {
		return counts, nil
	}
	mysql, err := aggregateVoteCounts(ctx, need)
	if err != nil {
		return counts, err
	}
	for _, id := range need {
		c := mysql[id]
		counts[id] = c
		if wErr := writeVoteCountsRedis(ctx, id, c); wErr != nil {
			g.Log().Warningf(ctx, "[ucg-vote] reconcile vote counts redis failed post=%d err=%v", id, wErr)
		}
	}
	return counts, nil
}

// loadMyVoteSidesFromRedis 批量读 viewer 在各帖的投票立场；field miss 表示未投票。
func loadMyVoteSidesFromRedis(ctx context.Context, viewerWxID int64, postIDs []uint64) (map[uint64]string, error) {
	out := make(map[uint64]string)
	if viewerWxID <= 0 || len(postIDs) == 0 {
		return out, nil
	}
	userKey := cachekit.UCGUserPostVotesKey(viewerWxID)
	for _, id := range postIDs {
		field := strconv.FormatUint(id, 10)
		side, ok, err := ucgCache.HashGet(ctx, userKey, field)
		if err != nil {
			return nil, err
		}
		if ok && side != "" {
			out[id] = side
		}
	}
	return out, nil
}

// writeVoteCountsRedis 将聚合结果写入 Redis（backfill 或灌库后调用）。
func writeVoteCountsRedis(ctx context.Context, postID uint64, c voteCounts) error {
	if postID == 0 {
		return nil
	}
	key := cachekit.UCGPostVoteCountsKey(postID)
	if err := ucgCache.HashSet(ctx, key, VoteSideLeft, strconv.FormatUint(uint64(c.left), 10)); err != nil {
		return err
	}
	return ucgCache.HashSet(ctx, key, VoteSideRight, strconv.FormatUint(uint64(c.right), 10))
}

// backfillVoteCountsFromMySQL cache miss 时从 MySQL 聚合单帖计数并写入 Redis。
func backfillVoteCountsFromMySQL(ctx context.Context, postID uint64) (voteCounts, error) {
	counts, err := aggregateVoteCounts(ctx, []uint64{postID})
	if err != nil {
		return voteCounts{}, err
	}
	c := counts[postID]
	if wErr := writeVoteCountsRedis(ctx, postID, c); wErr != nil {
		g.Log().Warningf(ctx, "[ucg-vote] backfill vote counts redis failed post=%d err=%v", postID, wErr)
	}
	return c, nil
}

func deletePostVoteCountsRedis(ctx context.Context, postID uint64) error {
	if postID == 0 {
		return nil
	}
	return ucgCache.Del(ctx, cachekit.UCGPostVoteCountsKey(postID))
}
