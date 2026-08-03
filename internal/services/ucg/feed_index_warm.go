package ucg

import (
	"context"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/frame/g"
)

const feedIndexWarmLockBackoff = 200 * time.Millisecond

// feedIndexGapSnapshot 推荐索引相对 MySQL published 的缺口快照。
type feedIndexGapSnapshot struct {
	Published int64
	ZCard     int64
	Gap       int64 // max(0, published-zcard)
}

// loadFeedIndexGap 读取 published 计数与 ZCARD，计算缺口。
func loadFeedIndexGap(ctx context.Context) (feedIndexGapSnapshot, error) {
	zcard, err := recommendScoreIndexSize(ctx)
	if err != nil {
		return feedIndexGapSnapshot{}, err
	}
	published, err := countPublishedPostsMySQL(ctx)
	if err != nil {
		return feedIndexGapSnapshot{}, err
	}
	gap := published - zcard
	if gap < 0 {
		gap = 0
	}
	return feedIndexGapSnapshot{Published: published, ZCard: zcard, Gap: gap}, nil
}

// feedIndexNeedsWarm 是否需要有界 warm/heal：published>0 且（空索引或 gap≥阈值）。
// Args: snap 缺口快照；threshold 来自 IndexHealGapThreshold。
func feedIndexNeedsWarm(snap feedIndexGapSnapshot, threshold int) bool {
	if snap.Published <= 0 {
		return false
	}
	if threshold <= 0 {
		threshold = feedIndexHealGapThresholdDef
	}
	if snap.ZCard == 0 {
		return true
	}
	return snap.Gap >= int64(threshold)
}

// ensureFeedIndexWarm 索引空或相对 published 明显短缺时，在单次 Feed 请求内有界 warm Redis。
// indexAutoWarmEnabled=false 时 no-op；未获锁方短退避后不阻塞。
func ensureFeedIndexWarm(ctx context.Context, cfg FeedConfig) error {
	if !cfg.IndexAutoWarmEnabled {
		return nil
	}
	snap, err := loadFeedIndexGap(ctx)
	if err != nil || !feedIndexNeedsWarm(snap, cfg.IndexHealGapThreshold) {
		return err
	}

	lockKey := cachekit.UCGFeedIndexWarmLockKey()
	lockTTL := time.Duration(cfg.IndexWarmLockSeconds) * time.Second
	got, err := ucgCache.SetNXEX(ctx, lockKey, "1", lockTTL)
	if err != nil {
		return err
	}
	if !got {
		time.Sleep(feedIndexWarmLockBackoff)
		return nil
	}
	defer func() { _ = ucgCache.Del(ctx, lockKey) }()

	snap, err = loadFeedIndexGap(ctx)
	if err != nil || !feedIndexNeedsWarm(snap, cfg.IndexHealGapThreshold) {
		return err
	}

	mode := "shortage"
	if snap.ZCard == 0 {
		mode = "empty"
	}
	start := time.Now()
	g.Log().Infof(ctx, "[ucg-feed] feed_index_warm_start mode=%s published=%d zcard=%d gap=%d batch=%d cap=%d",
		mode, snap.Published, snap.ZCard, snap.Gap, cfg.IndexWarmBatchSize, cfg.IndexWarmMaxPosts)

	okN, failN, err := warmPublishedPostsBounded(ctx, cfg.IndexWarmBatchSize, cfg.IndexWarmMaxPosts)
	if err != nil {
		return err
	}

	zcardAfter, zErr := recommendScoreIndexSize(ctx)
	if zErr != nil {
		g.Log().Warningf(ctx, "[ucg-feed] feed_index_warm zcard_after err=%v", zErr)
	}
	g.Log().Infof(ctx, "[ucg-feed] feed_index_warm_done mode=%s posts_ok=%d posts_fail=%d duration_ms=%d zcard_after=%d",
		mode, okN, failN, time.Since(start).Milliseconds(), zcardAfter)
	return nil
}

// warmPublishedPostsBounded 按 id 升序分页 syncPublishedPostRedis，最多 maxPosts 条。
// Returns: 成功数、失败数；扫描错误时返回 err（已处理部分仍计入计数）。
func warmPublishedPostsBounded(ctx context.Context, batchSize, maxPosts int) (okN, failN int64, err error) {
	if batchSize <= 0 {
		batchSize = feedIndexWarmBatchSizeDef
	}
	if maxPosts <= 0 {
		return 0, 0, nil
	}
	var lastID uint64
	processed := 0
	for processed < maxPosts {
		batch := batchSize
		if remaining := maxPosts - processed; batch > remaining {
			batch = remaining
		}
		rows, qErr := dao.UcgPost.Ctx(ctx).
			Where(dao.UcgPost.Columns().Status, PostStatusPublished).
			WhereGT(dao.UcgPost.Columns().Id, lastID).
			OrderAsc(dao.UcgPost.Columns().Id).
			Limit(batch).
			All()
		if qErr != nil {
			return okN, failN, qErr
		}
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			var post entity.UcgPost
			if sErr := row.Struct(&post); sErr != nil {
				failN++
				g.Log().Warningf(ctx, "[ucg-feed] feed_index_warm post struct fail err=%v", sErr)
				continue
			}
			lastID = post.Id
			if sErr := syncPublishedPostRedis(ctx, post.Id); sErr != nil {
				failN++
				g.Log().Warningf(ctx, "[ucg-feed] feed_index_warm post fail id=%d err=%v", post.Id, sErr)
				continue
			}
			okN++
			processed++
			if processed >= maxPosts {
				break
			}
		}
	}
	return okN, failN, nil
}

func recommendScoreIndexSize(ctx context.Context) (int64, error) {
	return ucgCache.SortedSetCard(ctx, cachekit.UCGRecommendScoreKey())
}

func countPublishedPostsMySQL(ctx context.Context) (int64, error) {
	n, err := dao.UcgPost.Ctx(ctx).
		Where(dao.UcgPost.Columns().Status, PostStatusPublished).
		Count()
	return int64(n), err
}
