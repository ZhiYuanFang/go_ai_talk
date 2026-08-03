package ucg

import (
	"context"
	"time"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// StartFeedIndexStartupHeal ucg-service 启动期 Feed 索引自检 + 可选异步自动补齐（任务 FeedIndexStartupHeal）。
// 业务逻辑：非 ticker；不阻塞调用方；check 关则直接返回；检出缺口打 ERROR；heal 开则后台争锁灌库。
// Side Effects: 可能异步写 Redis 推荐索引/snapshot；争用 startup-heal 锁。
func StartFeedIndexStartupHeal(parent context.Context) {
	_ = parent
	cfg := LoadFeedConfig(gctx.New())
	if !cfg.IndexStartupCheckEnabled {
		g.Log().Infof(gctx.New(), "[ucg-feed] feed_index_startup_check skipped (disabled)")
		return
	}
	// 异步执行，避免拖慢 HTTP Listen。
	go runFeedIndexStartupHeal(cfg)
}

// runFeedIndexStartupHeal 自检缺口；需要时异步 heal（独立锁，与请求 warm 锁分离）。
func runFeedIndexStartupHeal(cfg FeedConfig) {
	ctx := gctx.New()
	snap, err := loadFeedIndexGap(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "[ucg-feed] feed_index_startup_check failed err=%v", err)
		return
	}
	if !feedIndexNeedsWarm(snap, cfg.IndexHealGapThreshold) {
		g.Log().Infof(ctx, "[ucg-feed] feed_index_startup_check ok published=%d zcard=%d gap=%d",
			snap.Published, snap.ZCard, snap.Gap)
		return
	}

	g.Log().Errorf(ctx, "[ucg-feed] feed_index_startup_gap published=%d zcard=%d gap=%d threshold=%d heal_enabled=%v",
		snap.Published, snap.ZCard, snap.Gap, cfg.IndexHealGapThreshold, cfg.IndexStartupHealEnabled)

	if !cfg.IndexStartupHealEnabled {
		g.Log().Warningf(ctx, "[ucg-feed] feed_index_startup_heal skipped (heal disabled); run cmd/ucg-feed-backfill --posts-only")
		return
	}

	lockKey := cachekit.UCGFeedIndexStartupHealLockKey()
	lockTTL := time.Duration(cfg.IndexStartupHealLockSec) * time.Second
	got, err := ucgCache.SetNXEX(ctx, lockKey, "1", lockTTL)
	if err != nil {
		g.Log().Warningf(ctx, "[ucg-feed] feed_index_startup_heal lock err=%v", err)
		return
	}
	if !got {
		// 他副本已在补齐：跳过，避免双全量狂写。
		g.Log().Warningf(ctx, "[ucg-feed] feed_index_startup_heal skipped (lock held by another replica)")
		return
	}
	defer func() { _ = ucgCache.Del(ctx, lockKey) }()

	// 获锁后再读一次，避免他方已补齐仍全量重跑。
	snap, err = loadFeedIndexGap(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "[ucg-feed] feed_index_startup_heal recheck err=%v", err)
		return
	}
	if !feedIndexNeedsWarm(snap, cfg.IndexHealGapThreshold) {
		g.Log().Infof(ctx, "[ucg-feed] feed_index_startup_heal abort after lock: gap closed published=%d zcard=%d",
			snap.Published, snap.ZCard)
		return
	}

	start := time.Now()
	g.Log().Infof(ctx, "[ucg-feed] feed_index_startup_heal_start published=%d zcard=%d gap=%d batch=%d cap=%d",
		snap.Published, snap.ZCard, snap.Gap, cfg.IndexWarmBatchSize, cfg.IndexStartupHealMaxPosts)

	okN, failN, wErr := warmPublishedPostsBounded(ctx, cfg.IndexWarmBatchSize, cfg.IndexStartupHealMaxPosts)
	if wErr != nil {
		g.Log().Warningf(ctx, "[ucg-feed] feed_index_startup_heal scan err=%v posts_ok=%d posts_fail=%d", wErr, okN, failN)
	}
	zcardAfter, zErr := recommendScoreIndexSize(ctx)
	if zErr != nil {
		g.Log().Warningf(ctx, "[ucg-feed] feed_index_startup_heal zcard_after err=%v", zErr)
	}
	g.Log().Infof(ctx, "[ucg-feed] feed_index_startup_heal_done posts_ok=%d posts_fail=%d duration_ms=%d zcard_after=%d",
		okN, failN, time.Since(start).Milliseconds(), zcardAfter)
}
