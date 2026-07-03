package ucg

import (
	"context"
	"strconv"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/frame/g"
)

// patchVoteRedisScript 原子更新帖级计数与用户立场；MySQL 成功后的 best-effort 读模型 patch。
// KEYS[1]=vote-counts, KEYS[2]=user-post-votes
// ARGV[1]=postId field, ARGV[2]=newSide, ARGV[3]=oldSide（新票为空串）
const patchVoteRedisScript = `
local countsKey = KEYS[1]
local userKey = KEYS[2]
local postField = ARGV[1]
local newSide = ARGV[2]
local oldSide = ARGV[3]
if oldSide ~= '' and oldSide ~= newSide then
  redis.call('HINCRBY', countsKey, oldSide, -1)
end
if oldSide == '' or oldSide ~= newSide then
  redis.call('HINCRBY', countsKey, newSide, 1)
end
redis.call('HSET', userKey, postField, newSide)
return 'OK'
`

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

// patchVoteRedis VotePost 写路径同步 patch Redis；失败仅打日志，读 miss 时 MySQL backfill。
func patchVoteRedis(ctx context.Context, wxID int64, postID uint64, oldSide, newSide string) error {
	if wxID <= 0 || postID == 0 || newSide == "" {
		return nil
	}
	if oldSide == newSide {
		return nil
	}
	keys := []string{
		cachekit.UCGPostVoteCountsKey(postID),
		cachekit.UCGUserPostVotesKey(wxID),
	}
	args := []string{
		strconv.FormatUint(postID, 10),
		newSide,
		oldSide,
	}
	_, err := ucgCache.Eval(ctx, patchVoteRedisScript, keys, args)
	return err
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
