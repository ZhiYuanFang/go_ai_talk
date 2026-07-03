package ucg

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

// CommentSnapshot Redis 已发布评论 JSON 快照（仅 published）。
type CommentSnapshot struct {
	ID         uint64 `json:"id"`
	PostID     uint64 `json:"postId"`
	AuthorWxID uint64 `json:"authorWxId"`
	Content    string `json:"content"`
	VoteSide   string `json:"voteSide,omitempty"`
	CreatedAt  int64  `json:"createdAt"`
}

// debateSideLabels 辩论帖正反方展示文案，供评论 voteSideLabel 计算。
type debateSideLabels struct {
	Left  string
	Right string
}

func debateLabelsMapFromPostSnaps(snaps map[uint64]PostSnapshot) map[uint64]debateSideLabels {
	out := make(map[uint64]debateSideLabels)
	for id, snap := range snaps {
		if normalizePostType(snap.Type) != PostTypeDebate {
			continue
		}
		out[id] = debateSideLabels{
			Left:  strings.TrimSpace(snap.DebateLeft),
			Right: strings.TrimSpace(snap.DebateRight),
		}
	}
	return out
}

func debateLabelsForPost(post entity.UcgPost) map[uint64]debateSideLabels {
	if normalizePostType(post.Type) != PostTypeDebate {
		return nil
	}
	return map[uint64]debateSideLabels{
		post.Id: {
			Left:  strings.TrimSpace(post.DebateLeftLabel),
			Right: strings.TrimSpace(post.DebateRightLabel),
		},
	}
}

func commentsPreviewMax(ctx context.Context) int {
	cfg := LoadFeedConfig(ctx)
	if cfg.CommentsPreviewMax <= 0 {
		return 6
	}
	return cfg.CommentsPreviewMax
}

func writeCommentSnapshot(ctx context.Context, snap CommentSnapshot) error {
	cfg := LoadFeedConfig(ctx)
	raw, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	return ucgCache.SetEX(ctx, cachekit.UCGCommentSnapshotKey(snap.ID), string(raw), cfg.snapshotTTL())
}

// appendCommentRedis 评论审核通过后写入 Redis 读模型。
func appendCommentRedis(ctx context.Context, c entity.UcgPostComment) error {
	if c.Id == 0 || c.PostId == 0 {
		return nil
	}
	snap := CommentSnapshot{
		ID: c.Id, PostID: c.PostId, AuthorWxID: c.AuthorWxId,
		Content: c.Content, VoteSide: strings.TrimSpace(c.DebateVoteSide), CreatedAt: c.CreatedAt,
	}
	member := strconv.FormatUint(c.Id, 10)
	if err := ucgCache.SortedSetAdd(ctx, cachekit.UCGPostCommentsKey(c.PostId), float64(c.CreatedAt), member); err != nil {
		return err
	}
	return writeCommentSnapshot(ctx, snap)
}

// removeCommentRedis 删除已发布评论时从 Redis 读模型移除。
func removeCommentRedis(ctx context.Context, postID, commentID uint64) error {
	if postID == 0 || commentID == 0 {
		return nil
	}
	member := strconv.FormatUint(commentID, 10)
	_ = ucgCache.SortedSetRemove(ctx, cachekit.UCGPostCommentsKey(postID), member)
	return ucgCache.Del(ctx, cachekit.UCGCommentSnapshotKey(commentID))
}

// deletePostCommentsRedis 删帖/下架时清理帖子全部评论 Redis 索引与快照。
func deletePostCommentsRedis(ctx context.Context, postID uint64) error {
	if postID == 0 {
		return nil
	}
	key := cachekit.UCGPostCommentsKey(postID)
	members, err := ucgCache.SortedSetRange(ctx, key, 0, -1)
	if err != nil {
		return err
	}
	for _, m := range members {
		id, pErr := strconv.ParseUint(strings.TrimSpace(m), 10, 64)
		if pErr != nil || id == 0 {
			continue
		}
		_ = ucgCache.Del(ctx, cachekit.UCGCommentSnapshotKey(id))
	}
	return ucgCache.Del(ctx, key)
}

func loadCommentSnapshots(ctx context.Context, commentIDs []uint64) (map[uint64]CommentSnapshot, []uint64, error) {
	hits := make(map[uint64]CommentSnapshot, len(commentIDs))
	miss := make([]uint64, 0)
	if len(commentIDs) == 0 {
		return hits, miss, nil
	}
	keys := make([]string, 0, len(commentIDs))
	idByKey := make(map[string]uint64, len(commentIDs))
	for _, id := range commentIDs {
		k := cachekit.UCGCommentSnapshotKey(id)
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
		var snap CommentSnapshot
		if err = json.Unmarshal([]byte(raw), &snap); err != nil {
			miss = append(miss, id)
			continue
		}
		hits[id] = snap
	}
	return hits, miss, nil
}

func backfillCommentSnapshot(ctx context.Context, commentID uint64) (CommentSnapshot, error) {
	row, err := dao.UcgPostComment.Ctx(ctx).Where(dao.UcgPostComment.Columns().Id, commentID).One()
	if err != nil {
		return CommentSnapshot{}, err
	}
	if row.IsEmpty() {
		return CommentSnapshot{}, gerror.NewCode(gcode.CodeNotFound, "评论不存在")
	}
	var c entity.UcgPostComment
	if err = row.Struct(&c); err != nil {
		return CommentSnapshot{}, err
	}
	if c.Status != CommentStatusPublished {
		return CommentSnapshot{}, gerror.NewCode(gcode.CodeNotFound, "评论不存在")
	}
	_ = appendCommentRedis(ctx, c)
	return CommentSnapshot{
		ID: c.Id, PostID: c.PostId, AuthorWxID: c.AuthorWxId,
		Content: c.Content, VoteSide: strings.TrimSpace(c.DebateVoteSide), CreatedAt: c.CreatedAt,
	}, nil
}

func commentSnapshotsToDTOs(ctx context.Context, snaps []CommentSnapshot, debateLabelsByPost map[uint64]debateSideLabels) ([]*CommentDTO, error) {
	if len(snaps) == 0 {
		return nil, nil
	}
	authorIDs := make([]uint64, 0, len(snaps))
	seen := make(map[uint64]struct{}, len(snaps))
	for _, s := range snaps {
		if s.AuthorWxID == 0 {
			continue
		}
		if _, ok := seen[s.AuthorWxID]; ok {
			continue
		}
		seen[s.AuthorWxID] = struct{}{}
		authorIDs = append(authorIDs, s.AuthorWxID)
	}
	profileSnaps, err := loadProfileSnapshots(ctx, authorIDs)
	if err != nil {
		return nil, err
	}
	out := make([]*CommentDTO, 0, len(snaps))
	for _, s := range snaps {
		dto := &CommentDTO{
			Id: s.ID, PostId: s.PostID, AuthorWxId: s.AuthorWxID,
			Content: s.Content, Status: CommentStatusPublished, CreatedAt: s.CreatedAt,
			VoteSide: strings.TrimSpace(s.VoteSide),
		}
		if labels, ok := debateLabelsByPost[s.PostID]; ok {
			dto.VoteSideLabel = debateVoteSideLabel(dto.VoteSide, labels.Left, labels.Right)
		}
		if profSnap, ok := profileSnaps[s.AuthorWxID]; ok {
			dto.Author = &ProfileDTO{
				WxId:               profSnap.WxID,
				Nickname:           profSnap.Nickname,
				Bio:                profSnap.Bio,
				AvatarUrl:          profSnap.AvatarUrl,
				AvatarThumbnailUrl: profSnap.AvatarThumbnailUrl,
			}
		}
		out = append(out, dto)
	}
	return out, nil
}

// batchLoadCommentsPreview 批量加载 Feed 评论预览（每帖最多 limit 条，按 created_at 升序）。
func batchLoadCommentsPreview(ctx context.Context, postIDs []uint64, limit int) (map[uint64][]*CommentDTO, error) {
	out := make(map[uint64][]*CommentDTO, len(postIDs))
	if len(postIDs) == 0 || limit <= 0 {
		return out, nil
	}
	stop := int64(limit - 1)
	type postComments struct {
		postID     uint64
		commentIDs []uint64
	}
	batches := make([]postComments, 0, len(postIDs))
	allCommentIDs := make([]uint64, 0)
	for _, postID := range postIDs {
		if postID == 0 {
			continue
		}
		members, err := ucgCache.SortedSetRange(ctx, cachekit.UCGPostCommentsKey(postID), 0, stop)
		if err != nil {
			return nil, err
		}
		if len(members) == 0 {
			continue
		}
		ids := make([]uint64, 0, len(members))
		for _, m := range members {
			id, pErr := strconv.ParseUint(strings.TrimSpace(m), 10, 64)
			if pErr != nil || id == 0 {
				continue
			}
			ids = append(ids, id)
			allCommentIDs = append(allCommentIDs, id)
		}
		if len(ids) == 0 {
			continue
		}
		batches = append(batches, postComments{postID: postID, commentIDs: ids})
	}
	if len(allCommentIDs) == 0 {
		return out, nil
	}
	postSnaps, _, pErr := loadPostSnapshots(ctx, postIDs)
	if pErr != nil {
		return nil, pErr
	}
	debateLabels := debateLabelsMapFromPostSnaps(postSnaps)
	snaps, miss, err := loadCommentSnapshots(ctx, allCommentIDs)
	if err != nil {
		return nil, err
	}
	for _, id := range miss {
		snap, bfErr := backfillCommentSnapshot(ctx, id)
		if bfErr == nil {
			snaps[id] = snap
		}
	}
	for _, batch := range batches {
		ordered := make([]CommentSnapshot, 0, len(batch.commentIDs))
		for _, id := range batch.commentIDs {
			if snap, ok := snaps[id]; ok {
				ordered = append(ordered, snap)
			}
		}
		dtos, dErr := commentSnapshotsToDTOs(ctx, ordered, debateLabels)
		if dErr != nil {
			return nil, dErr
		}
		out[batch.postID] = dtos
	}
	return out, nil
}

// rebuildPostCommentsRedis 从 MySQL 重建单帖全部已发布评论 Redis 读模型。
func rebuildPostCommentsRedis(ctx context.Context, postID uint64) error {
	if postID == 0 {
		return nil
	}
	_ = deletePostCommentsRedis(ctx, postID)
	rows, err := dao.UcgPostComment.Ctx(ctx).
		Where(dao.UcgPostComment.Columns().PostId, postID).
		Where(dao.UcgPostComment.Columns().Status, CommentStatusPublished).
		OrderAsc(dao.UcgPostComment.Columns().CreatedAt).
		All()
	if err != nil {
		return err
	}
	for _, row := range rows {
		var c entity.UcgPostComment
		if err = row.Struct(&c); err != nil {
			continue
		}
		if err = appendCommentRedis(ctx, c); err != nil {
			glog.Warningf(ctx, "[ucg-comment-redis] rebuild append fail postId=%d commentId=%d err=%v", postID, c.Id, err)
		}
	}
	return nil
}

// loadCommentsFullFromRedis 加载单帖全部已发布评论；索引 miss 时 MySQL 重建。
func loadCommentsFullFromRedis(ctx context.Context, postID uint64) ([]CommentSnapshot, error) {
	if postID == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "postId 无效")
	}
	key := cachekit.UCGPostCommentsKey(postID)
	n, err := ucgCache.SortedSetCard(ctx, key)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		exists, exErr := ucgCache.Exists(ctx, key)
		if exErr != nil {
			return nil, exErr
		}
		if !exists {
			if bfErr := rebuildPostCommentsRedis(ctx, postID); bfErr != nil {
				return nil, bfErr
			}
			n, err = ucgCache.SortedSetCard(ctx, key)
			if err != nil {
				return nil, err
			}
		}
	}
	members, err := ucgCache.SortedSetRange(ctx, key, 0, -1)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 && n == 0 {
		return nil, nil
	}
	commentIDs := make([]uint64, 0, len(members))
	for _, m := range members {
		id, pErr := strconv.ParseUint(strings.TrimSpace(m), 10, 64)
		if pErr != nil || id == 0 {
			continue
		}
		commentIDs = append(commentIDs, id)
	}
	snaps, miss, err := loadCommentSnapshots(ctx, commentIDs)
	if err != nil {
		return nil, err
	}
	for _, id := range miss {
		snap, bfErr := backfillCommentSnapshot(ctx, id)
		if bfErr == nil {
			snaps[id] = snap
		}
	}
	out := make([]CommentSnapshot, 0, len(commentIDs))
	for _, id := range commentIDs {
		if snap, ok := snaps[id]; ok {
			out = append(out, snap)
		}
	}
	return out, nil
}

// ListCommentsFromRedis v2 全量评论列表：Redis 读模型，miss 回源 MySQL 重建。
func ListCommentsFromRedis(ctx context.Context, postID uint64) (*CommentsListResult, error) {
	if postID == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "postId 无效")
	}
	post, err := loadPublishedPostMeta(ctx, postID)
	if err != nil {
		return nil, err
	}
	commentCount := int(post.CommentCount)
	cap := commentsListMax(ctx)

	snaps, err := loadCommentsFullFromRedis(ctx, postID)
	if err != nil {
		return nil, err
	}
	if cap > 0 && len(snaps) > cap {
		snaps = snaps[:cap]
	}
	dtos, err := commentSnapshotsToDTOs(ctx, snaps, debateLabelsForPost(post))
	if err != nil {
		return nil, err
	}
	total := commentCount
	if total == 0 {
		total = len(dtos)
	}
	truncated := cap > 0 && commentCount > cap
	return &CommentsListResult{List: dtos, Total: total, Truncated: truncated}, nil
}

// enrichPostsWithCommentsPreview 为 Feed 帖子填充评论预览（Redis，每帖最多 commentsPreviewMax 条）。
func enrichPostsWithCommentsPreview(ctx context.Context, posts []*PostDTO) error {
	if len(posts) == 0 {
		return nil
	}
	postIDs := make([]uint64, 0, len(posts))
	for _, p := range posts {
		if p == nil || p.Id == 0 {
			continue
		}
		postIDs = append(postIDs, p.Id)
	}
	if len(postIDs) == 0 {
		return nil
	}
	preview, err := batchLoadCommentsPreview(ctx, postIDs, commentsPreviewMax(ctx))
	if err != nil {
		return err
	}
	for _, p := range posts {
		if p == nil {
			continue
		}
		comments, ok := preview[p.Id]
		if !ok || len(comments) == 0 {
			continue
		}
		p.Comments = comments
		// 旧 Redis post snapshot 可能缺 commentCount，用 ZCARD 补展示计数避免 Feed 长期显示 0。
		if p.CommentCount == 0 {
			if n, cardErr := ucgCache.SortedSetCard(ctx, cachekit.UCGPostCommentsKey(p.Id)); cardErr == nil && n > 0 {
				p.CommentCount = uint(n)
			}
		}
	}
	return nil
}

// BackfillPostCommentsRedis 运维 backfill：单帖评论 ZSET + snapshot。
func BackfillPostCommentsRedis(ctx context.Context, postID uint64) error {
	return rebuildPostCommentsRedis(ctx, postID)
}

func loadPublishedPostMeta(ctx context.Context, postID uint64) (entity.UcgPost, error) {
	row, err := dao.UcgPost.Ctx(ctx).
		Where(dao.UcgPost.Columns().Id, postID).
		Where(dao.UcgPost.Columns().Status, PostStatusPublished).
		One()
	if err != nil {
		return entity.UcgPost{}, err
	}
	if row.IsEmpty() {
		return entity.UcgPost{}, gerror.NewCode(gcode.CodeNotFound, "帖子不存在")
	}
	var post entity.UcgPost
	if err = row.Struct(&post); err != nil {
		return entity.UcgPost{}, err
	}
	return post, nil
}
