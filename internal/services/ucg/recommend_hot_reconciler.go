package ucg

import (
	"context"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const recommendHotScanTable = "ucg_recommend_hot_scan_state"

type hotScanState struct {
	LastPostID      uint64
	RoundHotCutoff  int64
}

// StartRecommendHotReconciler 热区分页 reconciler：轮首固定 hotCutoff，无互动也 Recompute。
func StartRecommendHotReconciler(ctx context.Context) {
	cfg := LoadRecommendConfig(ctx)
	interval := time.Duration(cfg.HotScanIntervalSeconds) * time.Second
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := tickRecommendHotReconciler(ctx); err != nil {
					glog.Warningf(ctx, "[ucg-recommend-hot] tick failed: %v", err)
				}
			}
		}
	}()
	glog.Infof(ctx, "[ucg-recommend-hot] started interval=%s pageSize=%d hotWindowHours=%.0f",
		interval, cfg.HotScanPageSize, cfg.HotWindowHours)
}

func tickRecommendHotReconciler(ctx context.Context) error {
	cfg := LoadRecommendConfig(ctx)
	state, err := loadHotScanState(ctx)
	if err != nil {
		return err
	}
	if state.LastPostID == 0 {
		windowSec := int64(cfg.HotWindowHours * 3600)
		state.RoundHotCutoff = time.Now().Unix() - windowSec
		if err = saveHotScanRoundCutoff(ctx, state.RoundHotCutoff); err != nil {
			return err
		}
	}

	pageSize := cfg.HotScanPageSize
	rows, err := dao.UcgPost.Ctx(ctx).
		Where(dao.UcgPost.Columns().Status, PostStatusPublished).
		WhereGTE(dao.UcgPost.Columns().PublishedAt, state.RoundHotCutoff).
		WhereGT(dao.UcgPost.Columns().Id, state.LastPostID).
		OrderAsc(dao.UcgPost.Columns().Id).
		Limit(pageSize).
		All()
	if err != nil {
		return err
	}

	var maxID uint64
	for _, row := range rows {
		var post entity.UcgPost
		if err = row.Struct(&post); err != nil {
			return err
		}
		if post.Id > maxID {
			maxID = post.Id
		}
		if err = RecomputeRecommendScore(ctx, post.Id); err != nil {
			return err
		}
	}

	if len(rows) < pageSize {
		return saveHotScanCursor(ctx, 0, state.RoundHotCutoff)
	}
	return saveHotScanCursor(ctx, maxID, state.RoundHotCutoff)
}

func loadHotScanState(ctx context.Context) (hotScanState, error) {
	var out hotScanState
	row, err := g.DB().Model(recommendHotScanTable).Ctx(ctx).Where("id", 1).One()
	if err != nil {
		return out, err
	}
	if row.IsEmpty() {
		return out, nil
	}
	out.LastPostID = row["last_post_id"].Uint64()
	out.RoundHotCutoff = row["round_hot_cutoff"].Int64()
	return out, nil
}

func saveHotScanRoundCutoff(ctx context.Context, cutoff int64) error {
	now := time.Now().Unix()
	_, err := g.DB().Model(recommendHotScanTable).Ctx(ctx).Data(g.Map{
		"round_hot_cutoff": cutoff,
		"updated_at":       now,
	}).Where("id", 1).Update()
	return err
}

func saveHotScanCursor(ctx context.Context, lastPostID uint64, roundCutoff int64) error {
	now := time.Now().Unix()
	_, err := g.DB().Model(recommendHotScanTable).Ctx(ctx).Data(g.Map{
		"last_post_id":     lastPostID,
		"round_hot_cutoff": roundCutoff,
		"updated_at":       now,
	}).Where("id", 1).Update()
	return err
}
