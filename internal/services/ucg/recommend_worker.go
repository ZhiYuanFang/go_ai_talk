package ucg

import (
	"context"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// ucgRecommendHotScanIntervalEnv 热区 reconciler tick 间隔（秒），优先于 yaml。
const ucgRecommendHotScanIntervalEnv = "UCG_RECOMMEND_HOT_SCAN_INTERVAL_SECONDS"

// RecommendConfig 推荐算法参数（见 design.md）。
type RecommendConfig struct {
	WNew                   float64
	TauHours               float64
	WLike                  float64
	WComment               float64
	HotWindowHours         float64
	HotScanPageSize        int
	HotScanIntervalSeconds int
	LikeThrottleMs         int
}

func LoadRecommendConfig(ctx context.Context) RecommendConfig {
	cfg := RecommendConfig{
		WNew:                   g.Cfg().MustGet(ctx, "ucg.recommend.wNew").Float64(),
		TauHours:               g.Cfg().MustGet(ctx, "ucg.recommend.tauHours").Float64(),
		WLike:                  g.Cfg().MustGet(ctx, "ucg.recommend.wLike").Float64(),
		WComment:               g.Cfg().MustGet(ctx, "ucg.recommend.wComment").Float64(),
		HotWindowHours:         g.Cfg().MustGet(ctx, "ucg.recommend.hotWindowHours").Float64(),
		HotScanPageSize:        g.Cfg().MustGet(ctx, "ucg.recommend.hotScanPageSize").Int(),
		HotScanIntervalSeconds: g.Cfg().MustGet(ctx, "ucg.recommend.hotScanIntervalSeconds").Int(),
		LikeThrottleMs:         g.Cfg().MustGet(ctx, "ucg.recommend.likeThrottleMs").Int(),
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
	if cfg.HotWindowHours <= 0 {
		cfg.HotWindowHours = cfg.TauHours
	}
	if cfg.HotScanPageSize <= 0 {
		cfg.HotScanPageSize = 200
	}
	if v := strings.TrimSpace(os.Getenv(ucgRecommendHotScanIntervalEnv)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.HotScanIntervalSeconds = n
		}
	}
	if cfg.HotScanIntervalSeconds <= 0 {
		cfg.HotScanIntervalSeconds = 3600
	}
	if cfg.LikeThrottleMs <= 0 {
		cfg.LikeThrottleMs = 500
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

// recommendHotZoneCutoffUnix 返回热区下界 published_at（与 reconciler round_hot_cutoff 语义一致）。
func recommendHotZoneCutoffUnix(cfg RecommendConfig, now time.Time) int64 {
	windowSec := int64(cfg.HotWindowHours * 3600)
	return now.Unix() - windowSec
}

// postPublishedAtForRecommend 取帖用于推荐分的时间戳（published_at 优先，否则 created_at）。
func postPublishedAtForRecommend(post entity.UcgPost) int64 {
	pub := post.PublishedAt
	if pub <= 0 {
		pub = post.CreatedAt
	}
	return pub
}

// isPostInRecommendHotZone 判断帖是否落在热区窗口内（published_at >= now - hotWindowHours）。
func isPostInRecommendHotZone(ctx context.Context, postID uint64) (bool, error) {
	if postID == 0 {
		return false, nil
	}
	row, err := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).One()
	if err != nil {
		return false, err
	}
	if row.IsEmpty() {
		return false, nil
	}
	var post entity.UcgPost
	if err = row.Struct(&post); err != nil {
		return false, err
	}
	cfg := LoadRecommendConfig(ctx)
	pub := postPublishedAtForRecommend(post)
	return pub >= recommendHotZoneCutoffUnix(cfg, time.Now()), nil
}

// RecomputeRecommendScore 读库重算单帖 score 并 UPSERT；非 published 或帖不存在时静默跳过。
func RecomputeRecommendScore(ctx context.Context, postID uint64) error {
	if postID == 0 {
		return nil
	}
	row, err := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).One()
	if err != nil {
		return err
	}
	if row.IsEmpty() {
		return nil
	}
	var post entity.UcgPost
	if err = row.Struct(&post); err != nil {
		return err
	}
	if post.Status != PostStatusPublished {
		return nil
	}
	cfg := LoadRecommendConfig(ctx)
	now := time.Now()
	score := computeRecommendScore(cfg, post, now)
	_, err = dao.UcgPostRecommend.Ctx(ctx).Data(g.Map{
		dao.UcgPostRecommend.Columns().PostId:     post.Id,
		dao.UcgPostRecommend.Columns().Score:      score,
		dao.UcgPostRecommend.Columns().ComputedAt: now.Unix(),
	}).Save()
	return err
}

// RemoveRecommendScore 下架/删帖时删除推荐行；0 行 affected 视为成功。
func RemoveRecommendScore(ctx context.Context, postID uint64) error {
	if postID == 0 {
		return nil
	}
	_, err := dao.UcgPostRecommend.Ctx(ctx).Where(dao.UcgPostRecommend.Columns().PostId, postID).Delete()
	return err
}
