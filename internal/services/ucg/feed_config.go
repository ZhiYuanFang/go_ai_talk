package ucg

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// FeedConfig Feed 复合分与 Redis 读路径参数。
type FeedConfig struct {
	WDist              float64
	DistDecayKm        float64
	RadiusStepsKm      []float64
	CandidateBatchSize int
	SessionTTLMinutes  int
	SnapshotTTLDays    int
}

// LoadFeedConfig 读取 ucg.feed.* 配置并填充默认值。
func LoadFeedConfig(ctx context.Context) FeedConfig {
	cfg := FeedConfig{
		WDist:              g.Cfg().MustGet(ctx, "ucg.feed.wDist").Float64(),
		DistDecayKm:        g.Cfg().MustGet(ctx, "ucg.feed.distDecayKm").Float64(),
		CandidateBatchSize: g.Cfg().MustGet(ctx, "ucg.feed.candidateBatchSize").Int(),
		SessionTTLMinutes:  g.Cfg().MustGet(ctx, "ucg.feed.sessionTtlMinutes").Int(),
		SnapshotTTLDays:    g.Cfg().MustGet(ctx, "ucg.feed.snapshotTtlDays").Int(),
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
	return cfg
}

func (cfg FeedConfig) sessionTTL() time.Duration {
	return time.Duration(cfg.SessionTTLMinutes) * time.Minute
}

func (cfg FeedConfig) snapshotTTL() time.Duration {
	return time.Duration(cfg.SnapshotTTLDays) * 24 * time.Hour
}
