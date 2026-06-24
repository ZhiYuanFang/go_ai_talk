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

// ensureFeedIndexWarm 索引冷启动时在单次 Feed 请求内有界 warm Redis（ZSET/GEO/snapshot）。
// indexAutoWarmEnabled=false 时 no-op；未获锁方短退避后不阻塞，继续空 Feed 或读到他方 warm 结果。
func ensureFeedIndexWarm(ctx context.Context, cfg FeedConfig) error {
	if !cfg.IndexAutoWarmEnabled {
		return nil
	}
	cold, err := isFeedIndexCold(ctx)
	if err != nil || !cold {
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

	cold, err = isFeedIndexCold(ctx)
	if err != nil || !cold {
		return err
	}

	start := time.Now()
	g.Log().Infof(ctx, "[ucg-feed] feed_index_warm_start batch=%d cap=%d",
		cfg.IndexWarmBatchSize, cfg.IndexWarmMaxPosts)

	var okN, failN int64
	var lastID uint64
	processed := 0
	for processed < cfg.IndexWarmMaxPosts {
		batch := cfg.IndexWarmBatchSize
		if remaining := cfg.IndexWarmMaxPosts - processed; batch > remaining {
			batch = remaining
		}
		rows, err := dao.UcgPost.Ctx(ctx).
			Where(dao.UcgPost.Columns().Status, PostStatusPublished).
			WhereGT(dao.UcgPost.Columns().Id, lastID).
			OrderAsc(dao.UcgPost.Columns().Id).
			Limit(batch).
			All()
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			var post entity.UcgPost
			if err = row.Struct(&post); err != nil {
				failN++
				g.Log().Warningf(ctx, "[ucg-feed] feed_index_warm post struct fail id=? err=%v", err)
				continue
			}
			lastID = post.Id
			if err = syncPublishedPostRedis(ctx, post.Id); err != nil {
				failN++
				g.Log().Warningf(ctx, "[ucg-feed] feed_index_warm post fail id=%d err=%v", post.Id, err)
				continue
			}
			okN++
			processed++
			if processed >= cfg.IndexWarmMaxPosts {
				break
			}
		}
	}

	zcardAfter, zErr := recommendScoreIndexSize(ctx)
	if zErr != nil {
		g.Log().Warningf(ctx, "[ucg-feed] feed_index_warm zcard_after err=%v", zErr)
	}
	g.Log().Infof(ctx, "[ucg-feed] feed_index_warm_done posts_ok=%d posts_fail=%d duration_ms=%d zcard_after=%d",
		okN, failN, time.Since(start).Milliseconds(), zcardAfter)
	return nil
}

func isFeedIndexCold(ctx context.Context) (bool, error) {
	zcard, err := recommendScoreIndexSize(ctx)
	if err != nil {
		return false, err
	}
	if zcard > 0 {
		return false, nil
	}
	n, err := countPublishedPostsMySQL(ctx)
	if err != nil {
		return false, err
	}
	return n > 0, nil
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
