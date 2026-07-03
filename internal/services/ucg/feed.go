package ucg

import (
	"context"
	"sort"
	"strconv"
	"strings"

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
	postType string,
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
	pageIDs, preloadedSnaps, err := filterPostIDsByTypeFromSnapshots(ctx, pageIDs, postType)
	if err != nil {
		return nil, err
	}

	hasMore := len(pageIDs) == p.PageSize && (len(candidates) > p.PageSize || nextCur.hasMoreWork(cfg))
	list, err := assembleFeedPosts(ctx, viewerWxID, viewer, hasViewer, pageIDs, candidates, preloadedSnaps)
	if err != nil {
		return nil, err
	}
	if err = enrichPostsWithVoteData(ctx, viewerWxID, list); err != nil {
		return nil, err
	}
	enrichAuthorForceOnPosts(ctx, list)
	// 推荐 Feed 响应：本人帖 omit distanceMeters；composite 排序 distanceTerm 不变。
	omitRecommendOwnPostDistance(viewerWxID, list)
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

		// 有 viewer 坐标时仅正半径走 GEO；radiusKm=0（unlimited）改 ZSET 全量扫（见 ucg-feed-geo-composite-score D3）。
		if hasViewer && radiusKm > 0 {
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
			// unlimited 步：ZSET 补全无 GEO 帖（及远距 GEO 帖）；pool/seen 去重。无 viewer 时各阶梯均走 ZSET。
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
				// 非 unlimited 且 viewer 有坐标时跳过已在 GEO 索引的帖（由 GEO 半径步覆盖）；unlimited 不 skip，保证无 lat/lng 帖可见。
				if hasViewer && radiusKm != 0 {
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

// omitRecommendOwnPostDistance 推荐 Feed 专用：viewer 为作者时不返回 distanceMeters。
func omitRecommendOwnPostDistance(viewerWxID int64, list []*PostDTO) {
	if viewerWxID <= 0 {
		return
	}
	viewer := uint64(viewerWxID)
	for _, item := range list {
		if item != nil && item.AuthorWxId == viewer {
			item.DistanceMeters = ""
		}
	}
}

func assembleFeedPosts(
	ctx context.Context,
	viewerWxID int64,
	viewer ViewerCoords,
	hasViewer bool,
	pageIDs []uint64,
	candidates []feedCandidate,
	preloadedSnaps map[uint64]PostSnapshot,
) ([]*PostDTO, error) {
	distByID := make(map[uint64]float64, len(candidates))
	for _, c := range candidates {
		distByID[c.postID] = c.distKm
	}

	snaps := preloadedSnaps
	if snaps == nil {
		var miss []uint64
		var err error
		snaps, miss, err = loadPostSnapshots(ctx, pageIDs)
		if err != nil {
			return nil, err
		}
		for _, id := range miss {
			snap, bfErr := backfillPostSnapshot(ctx, id)
			if bfErr == nil {
				snaps[id] = snap
			}
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
	postType string,
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
	if ft := normalizeFeedTypeFilter(postType); ft != "" {
		model = model.Where(dao.UcgPost.Columns().Type, ft)
	}
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
	list, err := assembleFeedPosts(ctx, wxID, viewer, hasViewer, pageIDs, candidates, nil)
	if err != nil {
		return nil, err
	}
	if err = enrichPostsWithVoteData(ctx, wxID, list); err != nil {
		return nil, err
	}
	enrichAuthorForceOnPosts(ctx, list)
	return &PageResult{List: list, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// normalizeFeedTypeFilter 广场 Feed 默认仅 debate。
func normalizeFeedTypeFilter(t string) string {
	t = strings.TrimSpace(t)
	if t == "" {
		return PostTypeDebate
	}
	if t == PostTypeDebate || t == PostTypeMoment {
		return t
	}
	return PostTypeDebate
}
