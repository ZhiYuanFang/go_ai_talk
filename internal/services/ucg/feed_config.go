package ucg

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	feedIndexAutoWarmEnabledEnv  = "UCG_FEED_INDEX_AUTO_WARM_ENABLED"
	feedIndexWarmBatchSizeEnv    = "UCG_FEED_INDEX_WARM_BATCH_SIZE"
	feedIndexWarmMaxPostsEnv     = "UCG_FEED_INDEX_WARM_MAX_POSTS"
	feedIndexWarmLockSecondsEnv  = "UCG_FEED_INDEX_WARM_LOCK_SECONDS"
	feedIndexAutoWarmEnabledDef  = true
	feedIndexWarmBatchSizeDef    = 200
	feedIndexWarmMaxPostsDef     = 2000
	feedIndexWarmLockSecondsDef  = 60
)

// FeedConfig Feed 复合分与 Redis 读路径参数。
type FeedConfig struct {
	WDist                 float64
	DistDecayKm           float64
	RadiusStepsKm         []float64
	CandidateBatchSize    int
	SessionTTLMinutes     int
	SnapshotTTLDays       int
	IndexAutoWarmEnabled    bool
	IndexWarmBatchSize      int
	IndexWarmMaxPosts       int
	IndexWarmLockSeconds    int
	CommentsPreviewMax      int
}

// LoadFeedConfig 读取 ucg.feed.* 配置并填充默认值。
func LoadFeedConfig(ctx context.Context) FeedConfig {
	cfg := FeedConfig{
		WDist:                g.Cfg().MustGet(ctx, "ucg.feed.wDist").Float64(),
		DistDecayKm:          g.Cfg().MustGet(ctx, "ucg.feed.distDecayKm").Float64(),
		CandidateBatchSize:   g.Cfg().MustGet(ctx, "ucg.feed.candidateBatchSize").Int(),
		SessionTTLMinutes:    g.Cfg().MustGet(ctx, "ucg.feed.sessionTtlMinutes").Int(),
		SnapshotTTLDays:      g.Cfg().MustGet(ctx, "ucg.feed.snapshotTtlDays").Int(),
		IndexAutoWarmEnabled: g.Cfg().MustGet(ctx, "ucg.feed.indexAutoWarmEnabled", feedIndexAutoWarmEnabledDef).Bool(),
		IndexWarmBatchSize:   g.Cfg().MustGet(ctx, "ucg.feed.indexWarmBatchSize", feedIndexWarmBatchSizeDef).Int(),
		IndexWarmMaxPosts:    g.Cfg().MustGet(ctx, "ucg.feed.indexWarmMaxPosts", feedIndexWarmMaxPostsDef).Int(),
		IndexWarmLockSeconds: g.Cfg().MustGet(ctx, "ucg.feed.indexWarmLockSeconds", feedIndexWarmLockSecondsDef).Int(),
		CommentsPreviewMax:   g.Cfg().MustGet(ctx, "ucg.feed.commentsPreviewMax", 6).Int(),
	}
	steps := g.Cfg().MustGet(ctx, "ucg.feed.radiusStepsKm").Interfaces()
	for _, v := range steps {
		cfg.RadiusStepsKm = append(cfg.RadiusStepsKm, g.NewVar(v).Float64())
	}
	if cfg.WDist <= 0 {
		cfg.WDist = 0.5
	}
	if cfg.DistDecayKm <= 0 {
		cfg.DistDecayKm = 10
	}
	if len(cfg.RadiusStepsKm) == 0 {
		cfg.RadiusStepsKm = []float64{50, 100, 200, 500, 0}
	}
	if cfg.CandidateBatchSize <= 0 {
		cfg.CandidateBatchSize = 200
	}
	if cfg.SessionTTLMinutes <= 0 {
		cfg.SessionTTLMinutes = 30
	}
	if cfg.SnapshotTTLDays <= 0 {
		cfg.SnapshotTTLDays = 7
	}
	if v := strings.TrimSpace(os.Getenv(feedIndexAutoWarmEnabledEnv)); v != "" {
		cfg.IndexAutoWarmEnabled = v == "1" || strings.EqualFold(v, "true")
	}
	if v := strings.TrimSpace(os.Getenv(feedIndexWarmBatchSizeEnv)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.IndexWarmBatchSize = n
		}
	}
	if v := strings.TrimSpace(os.Getenv(feedIndexWarmMaxPostsEnv)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.IndexWarmMaxPosts = n
		}
	}
	if v := strings.TrimSpace(os.Getenv(feedIndexWarmLockSecondsEnv)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.IndexWarmLockSeconds = n
		}
	}
	if cfg.IndexWarmBatchSize <= 0 {
		cfg.IndexWarmBatchSize = feedIndexWarmBatchSizeDef
	}
	if cfg.IndexWarmMaxPosts <= 0 {
		cfg.IndexWarmMaxPosts = feedIndexWarmMaxPostsDef
	}
	if cfg.IndexWarmLockSeconds <= 0 {
		cfg.IndexWarmLockSeconds = feedIndexWarmLockSecondsDef
	}
	return cfg
}

func (cfg FeedConfig) sessionTTL() time.Duration {
	return time.Duration(cfg.SessionTTLMinutes) * time.Minute
}

func (cfg FeedConfig) snapshotTTL() time.Duration {
	return time.Duration(cfg.SnapshotTTLDays) * 24 * time.Hour
}
