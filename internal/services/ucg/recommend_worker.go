package ucg

import (
	"context"
	"math"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// RecommendConfig 推荐算法参数（见 design.md）。
type RecommendConfig struct {
	WNew                 float64
	TauHours             float64
	WLike                float64
	WComment             float64
	RefreshIntervalSeconds int
}

func LoadRecommendConfig(ctx context.Context) RecommendConfig {
	cfg := RecommendConfig{
		WNew:                   g.Cfg().MustGet(ctx, "ucg.recommend.wNew").Float64(),
		TauHours:               g.Cfg().MustGet(ctx, "ucg.recommend.tauHours").Float64(),
		WLike:                  g.Cfg().MustGet(ctx, "ucg.recommend.wLike").Float64(),
		WComment:               g.Cfg().MustGet(ctx, "ucg.recommend.wComment").Float64(),
		RefreshIntervalSeconds: g.Cfg().MustGet(ctx, "ucg.recommend.refreshIntervalSeconds").Int(),
	}
	if cfg.WNew <= 0 {
		cfg.WNew = 1.0
	}
	if cfg.TauHours <= 0 {
		cfg.TauHours = 72
	}
	if cfg.WLike <= 0 {
		cfg.WLike = 0.3
	}
	if cfg.WComment <= 0 {
		cfg.WComment = 0.5
	}
	if cfg.RefreshIntervalSeconds <= 0 {
		cfg.RefreshIntervalSeconds = 300
	}
	return cfg
}

func computeRecommendScore(cfg RecommendConfig, post entity.UcgPost, now time.Time) float64 {
	pub := post.PublishedAt
	if pub <= 0 {
		pub = post.CreatedAt
	}
	ageHours := now.Sub(time.Unix(pub, 0)).Hours()
	if ageHours < 0 {
		ageHours = 0
	}
	newScore := cfg.WNew * math.Exp(-ageHours/cfg.TauHours)
	engagement := cfg.WLike*math.Log1p(float64(post.LikeCount)) + cfg.WComment*math.Log1p(float64(post.CommentCount))
	return newScore + engagement
}

// RefreshRecommendScores 重算 published 帖 score 写入 ucg_post_recommend。
func RefreshRecommendScores(ctx context.Context) error {
	cfg := LoadRecommendConfig(ctx)
	now := time.Now()
	rows, err := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Status, PostStatusPublished).All()
	if err != nil {
		return err
	}
	for _, row := range rows {
		var post entity.UcgPost
		if err = row.Struct(&post); err != nil {
			return err
		}
		score := computeRecommendScore(cfg, post, now)
		_, err = dao.UcgPostRecommend.Ctx(ctx).Data(g.Map{
			dao.UcgPostRecommend.Columns().PostId:     post.Id,
			dao.UcgPostRecommend.Columns().Score:      score,
			dao.UcgPostRecommend.Columns().ComputedAt: now.Unix(),
		}).Save()
		if err != nil {
			return err
		}
	}
	return nil
}

// StartRecommendWorker 后台刷新推荐 score。
func StartRecommendWorker(ctx context.Context) {
	cfg := LoadRecommendConfig(ctx)
	interval := time.Duration(cfg.RefreshIntervalSeconds) * time.Second
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := RefreshRecommendScores(ctx); err != nil {
					glog.Warningf(ctx, "ucg recommend refresh failed: %v", err)
				}
			}
		}
	}()
	glog.Infof(ctx, "ucg recommend worker started, interval=%s", interval)
}
