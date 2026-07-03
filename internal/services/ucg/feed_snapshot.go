package ucg

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// PostSnapshot Redis 帖子快照（lat/lng 仅服务端用于距离计算）。
type PostSnapshot struct {
	ID           uint64         `json:"id"`
	Type         string         `json:"type"`
	Content      string         `json:"content"`
	DebateLeft   string         `json:"debateLeft"`
	DebateRight  string         `json:"debateRight"`
	Media        []PostMediaDTO `json:"media"`
	AuthorWxID    uint64         `json:"authorWxId"`
	LikeCount      uint           `json:"likeCount"`
	CommentCount   uint           `json:"commentCount"`
	LeftVoteCount  uint           `json:"leftVoteCount"`
	RightVoteCount uint           `json:"rightVoteCount"`
	IpLocation     string         `json:"ipLocation"`
	PublishedAt int64          `json:"publishedAt"`
	MediaType   int            `json:"mediaType"`
	Lat         *float64       `json:"lat,omitempty"`
	Lng         *float64       `json:"lng,omitempty"`
}

// ProfileSnapshot Redis 作者公开 profile 快照。
type ProfileSnapshot struct {
	WxID               uint64 `json:"wxId"`
	Nickname           string `json:"nickname"`
	Bio                string `json:"bio"`
	AvatarUrl          string `json:"avatarUrl"`
	AvatarThumbnailUrl string `json:"avatarThumbnailUrl"`
}

func (s PostSnapshot) hasCoords() bool {
	if s.Lat == nil || s.Lng == nil {
		return false
	}
	return validCoord(*s.Lat, *s.Lng)
}

func writePostSnapshot(ctx context.Context, post entity.UcgPost) error {
	cfg := LoadFeedConfig(ctx)
	media, err := loadPostMedia(ctx, post.Id)
	if err != nil {
		return err
	}
	snap := PostSnapshot{
		ID:          post.Id,
		Type:        normalizePostType(post.Type),
		Content:     post.Content,
		DebateLeft:  strings.TrimSpace(post.DebateLeftLabel),
		DebateRight: strings.TrimSpace(post.DebateRightLabel),
		Media:       media,
		AuthorWxID:   post.AuthorWxId,
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		IpLocation:   strings.TrimSpace(post.IpLocation),
		PublishedAt: post.PublishedAt,
		MediaType:   post.MediaType,
		Lat:         post.Lat,
		Lng:         post.Lng,
	}
	if normalizePostType(post.Type) == PostTypeDebate {
		vc := resolveVoteCountsForPost(ctx, post.Id)
		snap.LeftVoteCount = vc.left
		snap.RightVoteCount = vc.right
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	return ucgCache.SetEX(ctx, cachekit.UCGPostSnapshotKey(post.Id), string(raw), cfg.snapshotTTL())
}

func writeProfileSnapshot(ctx context.Context, wxID uint64) error {
	cfg := LoadFeedConfig(ctx)
	prof, err := GetPublicProfile(ctx, wxID)
	if err != nil || prof == nil {
		return err
	}
	snap := ProfileSnapshot{
		WxID:               prof.WxId,
		Nickname:           prof.Nickname,
		Bio:                prof.Bio,
		AvatarUrl:          prof.AvatarUrl,
		AvatarThumbnailUrl: prof.AvatarThumbnailUrl,
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	return ucgCache.SetEX(ctx, cachekit.UCGProfileSnapshotKey(wxID), string(raw), cfg.snapshotTTL())
}

func deletePostSnapshot(ctx context.Context, postID uint64) error {
	return ucgCache.Del(ctx, cachekit.UCGPostSnapshotKey(postID))
}

// refreshPostSnapshotFromDB 从 MySQL 已发布帖重写 Redis 帖子快照（评论/点赞计数变更后保持 Feed 计数准确）。
func refreshPostSnapshotFromDB(ctx context.Context, postID uint64) {
	if postID == 0 {
		return
	}
	row, err := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).One()
	if err != nil || row.IsEmpty() {
		return
	}
	var post entity.UcgPost
	if err = row.Struct(&post); err != nil || post.Status != PostStatusPublished {
		return
	}
	_ = writePostSnapshot(ctx, post)
}

func loadPostSnapshots(ctx context.Context, postIDs []uint64) (map[uint64]PostSnapshot, []uint64, error) {
	hits := make(map[uint64]PostSnapshot, len(postIDs))
	miss := make([]uint64, 0)
	if len(postIDs) == 0 {
		return hits, miss, nil
	}
	keys := make([]string, 0, len(postIDs))
	idByKey := make(map[string]uint64, len(postIDs))
	for _, id := range postIDs {
		k := cachekit.UCGPostSnapshotKey(id)
		keys = append(keys, k)
		idByKey[k] = id
	}
	vals, err := ucgCache.MGet(ctx, keys)
	if err != nil {
		return nil, nil, err
	}
	for k, id := range idByKey {
		raw, ok := vals[k]
		if !ok || strings.TrimSpace(raw) == "" {
			miss = append(miss, id)
			continue
		}
		var snap PostSnapshot
		if err = json.Unmarshal([]byte(raw), &snap); err != nil {
			miss = append(miss, id)
			continue
		}
		hits[id] = snap
	}
	return hits, miss, nil
}

// filterPostIDsByTypeFromSnapshots 按 Redis snapshot.type 过滤 postId（保留原顺序），供 Feed 避免 MySQL。
// 返回的 snapshot map 可传入 assembleFeedPosts 复用，避免二次 MGET。
func filterPostIDsByTypeFromSnapshots(ctx context.Context, ids []uint64, postType string) ([]uint64, map[uint64]PostSnapshot, error) {
	want := normalizeFeedTypeFilter(postType)
	if want == "" || len(ids) == 0 {
		return ids, nil, nil
	}
	snaps, miss, err := loadPostSnapshots(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	for _, id := range miss {
		snap, bfErr := backfillPostSnapshot(ctx, id)
		if bfErr == nil {
			snaps[id] = snap
		}
	}
	out := make([]uint64, 0, len(ids))
	filtered := make(map[uint64]PostSnapshot, len(ids))
	for _, id := range ids {
		snap, ok := snaps[id]
		if !ok {
			continue
		}
		if normalizePostType(snap.Type) != want {
			continue
		}
		out = append(out, id)
		filtered[id] = snap
	}
	return out, filtered, nil
}

func backfillPostSnapshot(ctx context.Context, postID uint64) (PostSnapshot, error) {
	row, err := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).One()
	if err != nil {
		return PostSnapshot{}, err
	}
	if row.IsEmpty() {
		return PostSnapshot{}, gerror.NewCode(gcode.CodeNotFound, "帖子不存在")
	}
	var post entity.UcgPost
	if err = row.Struct(&post); err != nil {
		return PostSnapshot{}, err
	}
	if post.Status != PostStatusPublished {
		return PostSnapshot{}, gerror.NewCode(gcode.CodeNotFound, "帖子不存在")
	}
	_ = writePostSnapshot(ctx, post)
	_ = writeProfileSnapshot(ctx, post.AuthorWxId)
	media, err := loadPostMedia(ctx, post.Id)
	if err != nil {
		return PostSnapshot{}, err
	}
	snap := PostSnapshot{
		ID:             post.Id,
		Type:           normalizePostType(post.Type),
		Content:        post.Content,
		DebateLeft:     strings.TrimSpace(post.DebateLeftLabel),
		DebateRight:    strings.TrimSpace(post.DebateRightLabel),
		Media:          media,
		AuthorWxID:     post.AuthorWxId,
		LikeCount:      post.LikeCount,
		CommentCount:   post.CommentCount,
		IpLocation:     strings.TrimSpace(post.IpLocation),
		PublishedAt:    post.PublishedAt,
		MediaType:      post.MediaType,
		Lat:            post.Lat,
		Lng:            post.Lng,
	}
	if normalizePostType(post.Type) == PostTypeDebate {
		vc := resolveVoteCountsForPost(ctx, post.Id)
		snap.LeftVoteCount = vc.left
		snap.RightVoteCount = vc.right
	}
	return snap, nil
}

func loadProfileSnapshots(ctx context.Context, wxIDs []uint64) (map[uint64]ProfileSnapshot, error) {
	out := make(map[uint64]ProfileSnapshot, len(wxIDs))
	if len(wxIDs) == 0 {
		return out, nil
	}
	keys := make([]string, 0, len(wxIDs))
	idByKey := make(map[string]uint64, len(wxIDs))
	for _, id := range wxIDs {
		k := cachekit.UCGProfileSnapshotKey(id)
		keys = append(keys, k)
		idByKey[k] = id
	}
	vals, err := ucgCache.MGet(ctx, keys)
	if err != nil {
		return nil, err
	}
	for k, id := range idByKey {
		raw, ok := vals[k]
		if !ok || strings.TrimSpace(raw) == "" {
			continue
		}
		var snap ProfileSnapshot
		if err = json.Unmarshal([]byte(raw), &snap); err != nil {
			continue
		}
		// 旧 Redis profile snapshot 可能缺头像 URL，视为 stale 回源刷新（与 feed 帖 content 回填一致）。
		if strings.TrimSpace(snap.AvatarUrl) == "" && strings.TrimSpace(snap.AvatarThumbnailUrl) == "" {
			_ = writeProfileSnapshot(ctx, id)
			if prof, pErr := GetPublicProfile(ctx, id); pErr == nil && prof != nil {
				out[id] = ProfileSnapshot{
					WxID:               prof.WxId,
					Nickname:           prof.Nickname,
					Bio:                prof.Bio,
					AvatarUrl:          prof.AvatarUrl,
					AvatarThumbnailUrl: prof.AvatarThumbnailUrl,
				}
				continue
			}
		}
		out[id] = snap
	}
	for _, id := range wxIDs {
		if _, ok := out[id]; ok {
			continue
		}
		_ = writeProfileSnapshot(ctx, id)
		if prof, pErr := GetPublicProfile(ctx, id); pErr == nil && prof != nil {
			out[id] = ProfileSnapshot{
				WxID:               prof.WxId,
				Nickname:           prof.Nickname,
				Bio:                prof.Bio,
				AvatarUrl:          prof.AvatarUrl,
				AvatarThumbnailUrl: prof.AvatarThumbnailUrl,
			}
		}
	}
	return out, nil
}

func snapshotToPostDTO(
	snap PostSnapshot,
	prof *ProfileSnapshot,
	likedByMe bool,
	distanceMeters string,
) *PostDTO {
	var author *ProfileDTO
	if prof != nil {
		author = &ProfileDTO{
			WxId:               prof.WxID,
			Nickname:           prof.Nickname,
			Bio:                prof.Bio,
			AvatarUrl:          prof.AvatarUrl,
			AvatarThumbnailUrl: prof.AvatarThumbnailUrl,
		}
		ensureAuthorBio(author)
	}
	return &PostDTO{
		Id:             snap.ID,
		AuthorWxId:     snap.AuthorWxID,
		Type:           normalizePostType(snap.Type),
		Content:        snap.Content,
		DebateLeft:     snap.DebateLeft,
		DebateRight:    snap.DebateRight,
		Status:         PostStatusPublished,
		MediaType:      snap.MediaType,
		LikeCount:      snap.LikeCount,
		CommentCount:   snap.CommentCount,
		LeftVoteCount:  snap.LeftVoteCount,
		RightVoteCount: snap.RightVoteCount,
		LikedByMe:      likedByMe,
		PublishedAt:    snap.PublishedAt,
		IpLocation:     snap.IpLocation,
		Media:          snap.Media,
		Author:         author,
		DistanceMeters: distanceMeters,
		CreatedAt:      snap.PublishedAt,
		UpdatedAt:      snap.PublishedAt,
	}
}

func likedPostIDsFromRedis(ctx context.Context, viewerWxID int64, postIDs []uint64) (map[uint64]bool, error) {
	out := make(map[uint64]bool, len(postIDs))
	if viewerWxID <= 0 || len(postIDs) == 0 {
		return out, nil
	}
	members := make([]string, 0, len(postIDs))
	for _, id := range postIDs {
		members = append(members, strconv.FormatUint(id, 10))
	}
	hits, err := ucgCache.SetIsMemberBatch(ctx, cachekit.UCGUserLikedPostsKey(viewerWxID), members)
	if err != nil {
		return nil, err
	}
	for _, id := range postIDs {
		out[id] = hits[strconv.FormatUint(id, 10)]
	}
	return out, nil
}

func saddUserLikedPost(ctx context.Context, wxID int64, postID uint64) error {
	return ucgCache.SetAdd(ctx, cachekit.UCGUserLikedPostsKey(wxID), strconv.FormatUint(postID, 10))
}

func sremUserLikedPost(ctx context.Context, wxID int64, postID uint64) error {
	return ucgCache.SetRemove(ctx, cachekit.UCGUserLikedPostsKey(wxID), strconv.FormatUint(postID, 10))
}

func syncPublishedPostRedis(ctx context.Context, postID uint64) error {
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
	score := computeRecommendScore(cfg, post, time.Now())
	member := strconv.FormatUint(post.Id, 10)
	if err = ucgCache.SortedSetAdd(ctx, cachekit.UCGRecommendScoreKey(), score, member); err != nil {
		return err
	}
	if postEntityHasCoords(post) {
		if err = ucgCache.GeoAdd(ctx, cachekit.UCGFeedGeoKey(), member, *post.Lng, *post.Lat); err != nil {
			return err
		}
	} else {
		_ = ucgCache.GeoRemove(ctx, cachekit.UCGFeedGeoKey(), member)
	}
	if err = writePostSnapshot(ctx, post); err != nil {
		return err
	}
	if normalizePostType(post.Type) == PostTypeDebate {
		_ = initPostVoteCountsRedis(ctx, post.Id)
	}
	return writeProfileSnapshot(ctx, post.AuthorWxId)
}

func removePostFromRecommendRedis(ctx context.Context, postID uint64) error {
	if postID == 0 {
		return nil
	}
	member := strconv.FormatUint(postID, 10)
	_ = ucgCache.SortedSetRemove(ctx, cachekit.UCGRecommendScoreKey(), member)
	_ = ucgCache.GeoRemove(ctx, cachekit.UCGFeedGeoKey(), member)
	_ = deletePostVoteCountsRedis(ctx, postID)
	_ = deletePostCommentsRedis(ctx, postID)
	return deletePostSnapshot(ctx, postID)
}
