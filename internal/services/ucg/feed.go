package ucg

import (
	"context"
	"sort"
	"strconv"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// FeedRecommendResult 推荐 Feed cursor 分页结果（无 total）。
type FeedRecommendResult struct {
	List       []*PostDTO
	HasMore    bool
	NextCursor string
}

type feedCandidate struct {
	postID     uint64
	baseScore  float64
	finalScore float64
	distKm     float64
}

// ListRecommendFeed Redis 复合分 Feed：GEO 半径 + ZSET + session 去重 + cursor。
func ListRecommendFeed(
	ctx context.Context,
	viewerWxID int64,
	viewerLat, viewerLng *float64,
	cursor string,
	pageSize int,
) (*FeedRecommendResult, error) {
	p := NormalizePage(1, pageSize)
	cfg := LoadFeedConfig(ctx)

	var cur feedCursor
	if decoded, ok := decodeFeedCursor(cursor); ok {
		cur = decoded
	} else {
		cur.SessionID = newFeedSessionID()
		if vc, ok := ValidViewerCoords(viewerLat, viewerLng); ok {
			cur.Lat = vc.Lat
			cur.Lng = vc.Lng
			cur.HasViewer = true
		}
	}

	seen, err := loadSessionSeen(ctx, cur.SessionID)
	if err != nil {
		return nil, err
	}

	if err = ensureFeedIndexWarm(ctx, cfg); err != nil {
		return nil, err
	}

	viewer, hasViewer := cursorViewer(cur)
	candidates, nextCur, err := collectFeedCandidates(ctx, cfg, cur, seen, viewer, hasViewer, p.PageSize*3)
	if err != nil {
		return nil, err
	}

	pageIDs := make([]uint64, 0, p.PageSize)
	for _, c := range candidates {
		if len(pageIDs) >= p.PageSize {
			break
		}
		pageIDs = append(pageIDs, c.postID)
		if len(pageIDs) == p.PageSize && len(candidates) > p.PageSize {
			nextCur.LastFinalScore = c.finalScore
			nextCur.LastPostID = c.postID
		}
	}

	hasMore := len(pageIDs) == p.PageSize && (len(candidates) > p.PageSize || nextCur.hasMoreWork(cfg))
	list, err := assembleFeedPosts(ctx, viewerWxID, viewer, hasViewer, pageIDs, candidates)
	if err != nil {
		return nil, err
	}
	_ = markSessionSeen(ctx, cfg, cur.SessionID, pageIDs)

	nextCursor := ""
	if hasMore {
		nextCursor = encodeFeedCursor(nextCur)
	}
	return &FeedRecommendResult{List: list, HasMore: hasMore, NextCursor: nextCursor}, nil
}

func (c feedCursor) hasMoreWork(cfg FeedConfig) bool {
	if c.RadiusIdx < len(cfg.RadiusStepsKm)-1 {
		return true
	}
	return c.ZsetOffset < 100000
}

func collectFeedCandidates(
	ctx context.Context,
	cfg FeedConfig,
	cur feedCursor,
	seen map[uint64]struct{},
	viewer ViewerCoords,
	hasViewer bool,
	need int,
) ([]feedCandidate, feedCursor, error) {
	pool := make(map[uint64]feedCandidate)
	next := cur

	for len(pool) < need && next.RadiusIdx < len(cfg.RadiusStepsKm) {
		radiusKm := cfg.RadiusStepsKm[next.RadiusIdx]

		if hasViewer && radiusKm >= 0 {
			geoRows, err := ucgCache.GeoSearchByRadiusWithDist(
				ctx, cachekit.UCGFeedGeoKey(), viewer.Lng, viewer.Lat, radiusKm, cfg.CandidateBatchSize,
			)
			if err != nil {
				return nil, next, err
			}
			if next.GeoOffset > 0 && next.GeoOffset < len(geoRows) {
				geoRows = geoRows[next.GeoOffset:]
			}
			for _, row := range geoRows {
				id, _ := strconv.ParseUint(row.Member, 10, 64)
				if id == 0 {
					continue
				}
				if _, ok := seen[id]; ok {
					continue
				}
				if _, ok := pool[id]; ok {
					continue
				}
				pool[id] = feedCandidate{postID: id, distKm: row.DistKm}
				if len(pool) >= need {
					next.GeoOffset += len(geoRows)
					break
				}
			}
			if len(pool) < need {
				next.GeoOffset = 0
				next.RadiusIdx++
				continue
			}
		} else if !hasViewer || radiusKm == 0 {
			zrows, err := ucgCache.SortedSetRevRangeWithScores(
				ctx, cachekit.UCGRecommendScoreKey(), int64(next.ZsetOffset), int64(next.ZsetOffset+cfg.CandidateBatchSize-1),
			)
			if err != nil {
				return nil, next, err
			}
			if len(zrows) == 0 {
				next.RadiusIdx++
				continue
			}
			members := make([]string, 0, len(zrows))
			for _, z := range zrows {
				members = append(members, z.Member)
			}
			inGeo, err := ucgCache.GeoPosBatch(ctx, cachekit.UCGFeedGeoKey(), members)
			if err != nil {
				return nil, next, err
			}
			added := 0
			for _, z := range zrows {
				id, _ := strconv.ParseUint(z.Member, 10, 64)
				if id == 0 {
					continue
				}
				if _, ok := seen[id]; ok {
					continue
				}
				if hasViewer {
					if _, ok := inGeo[z.Member]; ok {
						continue
					}
				}
				if _, ok := pool[id]; ok {
					continue
				}
				pool[id] = feedCandidate{postID: id, baseScore: z.Score}
				added++
				if len(pool) >= need {
					break
				}
			}
			next.ZsetOffset += len(zrows)
			if added == 0 && len(zrows) > 0 {
				continue
			}
			if len(pool) < need {
				next.RadiusIdx++
			}
		} else {
			next.RadiusIdx++
		}
	}

	ids := make([]uint64, 0, len(pool))
	members := make([]string, 0, len(pool))
	for id := range pool {
		ids = append(ids, id)
		members = append(members, strconv.FormatUint(id, 10))
	}
	scores, err := ucgCache.SortedSetScores(ctx, cachekit.UCGRecommendScoreKey(), members)
	if err != nil {
		return nil, next, err
	}
	snaps, miss, err := loadPostSnapshots(ctx, ids)
	if err != nil {
		return nil, next, err
	}
	for _, id := range miss {
		snap, bfErr := backfillPostSnapshot(ctx, id)
		if bfErr == nil {
			snaps[id] = snap
		}
	}

	out := make([]feedCandidate, 0, len(pool))
	for id, c := range pool {
		base := scores[strconv.FormatUint(id, 10)]
		if base == 0 && c.baseScore > 0 {
			base = c.baseScore
		}
		snap := snaps[id]
		var postLat, postLng float64
		if snap.hasCoords() {
			postLat, postLng = *snap.Lat, *snap.Lng
		}
		final := computeFinalScore(cfg, base, viewer, hasViewer, postLat, postLng, c.distKm)
		if !afterCursor(cur.LastFinalScore, cur.LastPostID, final, id) {
			continue
		}
		c.baseScore = base
		c.finalScore = final
		c.postID = id
		out = append(out, c)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].finalScore != out[j].finalScore {
			return out[i].finalScore > out[j].finalScore
		}
		return out[i].postID > out[j].postID
	})
	return out, next, nil
}

func assembleFeedPosts(
	ctx context.Context,
	viewerWxID int64,
	viewer ViewerCoords,
	hasViewer bool,
	pageIDs []uint64,
	candidates []feedCandidate,
) ([]*PostDTO, error) {
	distByID := make(map[uint64]float64, len(candidates))
	for _, c := range candidates {
		distByID[c.postID] = c.distKm
	}

	snaps, miss, err := loadPostSnapshots(ctx, pageIDs)
	if err != nil {
		return nil, err
	}
	for _, id := range miss {
		snap, bfErr := backfillPostSnapshot(ctx, id)
		if bfErr == nil {
			snaps[id] = snap
		}
	}

	authorIDs := make([]uint64, 0, len(pageIDs))
	for _, id := range pageIDs {
		if snap, ok := snaps[id]; ok {
			authorIDs = append(authorIDs, snap.AuthorWxID)
		}
	}
	profiles, err := loadProfileSnapshots(ctx, authorIDs)
	if err != nil {
		return nil, err
	}

	liked, err := likedPostIDsFromRedis(ctx, viewerWxID, pageIDs)
	if err != nil {
		return nil, err
	}

	out := make([]*PostDTO, 0, len(pageIDs))
	for _, id := range pageIDs {
		snap, ok := snaps[id]
		if !ok {
			continue
		}
		var prof *ProfileSnapshot
		if p, ok := profiles[snap.AuthorWxID]; ok {
			pp := p
			prof = &pp
		}
		distMeters := ""
		if hasViewer && snap.hasCoords() {
			distKm := distByID[id]
			if distKm <= 0 {
				distKm = haversineKm(viewer.Lat, viewer.Lng, *snap.Lat, *snap.Lng)
			}
			distMeters = formatDistanceMeters(distKm)
		}
		out = append(out, snapshotToPostDTO(snap, prof, liked[id], distMeters))
	}
	return out, nil
}

// DistanceMetersForPost 单帖详情可选 viewer 坐标距离。
func DistanceMetersForPost(ctx context.Context, postID uint64, viewerLat, viewerLng *float64) (string, error) {
	viewer, ok := ValidViewerCoords(viewerLat, viewerLng)
	if !ok {
		return "", nil
	}
	snaps, miss, err := loadPostSnapshots(ctx, []uint64{postID})
	if err != nil {
		return "", err
	}
	snap, ok := snaps[postID]
	if !ok {
		for _, id := range miss {
			if id == postID {
				snap, err = backfillPostSnapshot(ctx, postID)
				ok = err == nil
				break
			}
		}
	}
	if !ok || !snap.hasCoords() {
		row, rErr := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).One()
		if rErr != nil {
			return "", rErr
		}
		if row.IsEmpty() {
			return "", nil
		}
		var post entity.UcgPost
		if rErr = row.Struct(&post); rErr != nil {
			return "", rErr
		}
		if !postEntityHasCoords(post) {
			return "", nil
		}
		distKm := haversineKm(viewer.Lat, viewer.Lng, *post.Lat, *post.Lng)
		return formatDistanceMeters(distKm), nil
	}
	distKm := haversineKm(viewer.Lat, viewer.Lng, *snap.Lat, *snap.Lng)
	return formatDistanceMeters(distKm), nil
}

// ListFollowingFeed 关注 Feed：MySQL 负责 followee 集合与 published_at 分页；Redis snapshot 组装展示字段。
func ListFollowingFeed(
	ctx context.Context,
	wxID int64,
	page, pageSize int,
	viewerLat, viewerLng *float64,
) (*PageResult, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "缺少 X-Internal-Wx-Id")
	}
	p := NormalizePage(page, pageSize)
	followRows, err := dao.UcgFollow.Ctx(ctx).
		Where(dao.UcgFollow.Columns().FollowerWxId, wxID).
		Fields(dao.UcgFollow.Columns().FolloweeWxId).
		All()
	if err != nil {
		return nil, err
	}
	if len(followRows) == 0 {
		return &PageResult{List: []*PostDTO{}, Total: 0, Page: p.Page, PageSize: p.PageSize}, nil
	}
	ids := make([]uint64, 0, len(followRows))
	for _, row := range followRows {
		var f entity.UcgFollow
		if err = row.Struct(&f); err != nil {
			return nil, err
		}
		ids = append(ids, f.FolloweeWxId)
	}
	// MySQL 阶段：仅分页取 postId 与 total；不 JOIN profile/like/媒体表。
	model := dao.UcgPost.Ctx(ctx).
		Where(dao.UcgPost.Columns().Status, PostStatusPublished).
		WhereIn(dao.UcgPost.Columns().AuthorWxId, ids)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows, err := model.
		Fields(dao.UcgPost.Columns().Id).
		OrderDesc(dao.UcgPost.Columns().PublishedAt).
		Limit(p.PageSize).
		Offset(pageOffset(p)).
		All()
	if err != nil {
		return nil, err
	}
	pageIDs := make([]uint64, 0, len(rows))
	for _, row := range rows {
		var post entity.UcgPost
		if err = row.Struct(&post); err != nil {
			return nil, err
		}
		pageIDs = append(pageIDs, post.Id)
	}

	// Redis 组装阶段：snapshot / liked SET / distanceMeters；miss 时 backfillPostSnapshot best-effort 回源。
	viewer, hasViewer := ValidViewerCoords(viewerLat, viewerLng)
	candidates := make([]feedCandidate, len(pageIDs))
	for i, id := range pageIDs {
		candidates[i] = feedCandidate{postID: id} // 无 GEO 预计算，assembleFeedPosts 内 haversine 重算距离
	}
	list, err := assembleFeedPosts(ctx, wxID, viewer, hasViewer, pageIDs, candidates)
	if err != nil {
		return nil, err
	}
	return &PageResult{List: list, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}
