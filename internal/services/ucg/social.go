package ucg

import (
	"context"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/eventkit"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Follow 关注用户。
func Follow(ctx context.Context, followerWxID, followeeWxID int64) error {
	if followerWxID <= 0 || followeeWxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	if followerWxID == followeeWxID {
		return gerror.NewCode(gcode.CodeInvalidParameter, "不能关注自己")
	}
	exists, _, err := Device().ValidateWx(ctx, followeeWxID)
	if err != nil {
		return err
	}
	if !exists {
		return gerror.NewCode(gcode.CodeNotFound, "用户不存在")
	}
	cnt, err := dao.UcgFollow.Ctx(ctx).
		Where(dao.UcgFollow.Columns().FollowerWxId, followerWxID).
		Where(dao.UcgFollow.Columns().FolloweeWxId, followeeWxID).
		Count()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	_, err = dao.UcgFollow.Ctx(ctx).Data(g.Map{
		dao.UcgFollow.Columns().FollowerWxId: followerWxID,
		dao.UcgFollow.Columns().FolloweeWxId: followeeWxID,
		dao.UcgFollow.Columns().CreatedAt:    time.Now().Unix(),
	}).Insert()
	return err
}

// IsFollowing 当前用户是否已关注目标用户。
func IsFollowing(ctx context.Context, followerWxID, followeeWxID int64) (bool, error) {
	if followerWxID <= 0 || followeeWxID <= 0 {
		return false, nil
	}
	cnt, err := dao.UcgFollow.Ctx(ctx).
		Where(dao.UcgFollow.Columns().FollowerWxId, followerWxID).
		Where(dao.UcgFollow.Columns().FolloweeWxId, followeeWxID).
		Count()
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// Unfollow 取消关注。
func Unfollow(ctx context.Context, followerWxID, followeeWxID int64) error {
	if followerWxID <= 0 || followeeWxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	_, err := dao.UcgFollow.Ctx(ctx).
		Where(dao.UcgFollow.Columns().FollowerWxId, followerWxID).
		Where(dao.UcgFollow.Columns().FolloweeWxId, followeeWxID).
		Delete()
	return err
}

// ListFollowing 我关注的人 wxId 列表（分页）。
func ListFollowing(ctx context.Context, followerWxID int64, page, pageSize int) (*PageResult, error) {
	p := NormalizePage(page, pageSize)
	model := dao.UcgFollow.Ctx(ctx).Where(dao.UcgFollow.Columns().FollowerWxId, followerWxID)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows, err := model.OrderDesc(dao.UcgFollow.Columns().CreatedAt).Limit(p.PageSize).Offset(pageOffset(p)).All()
	if err != nil {
		return nil, err
	}
	list := make([]uint64, 0, len(rows))
	for _, row := range rows {
		var f entity.UcgFollow
		if err = row.Struct(&f); err != nil {
			return nil, err
		}
		list = append(list, f.FolloweeWxId)
	}
	return &PageResult{List: list, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// LikePost 点赞。
func LikePost(ctx context.Context, wxID int64, postID uint64) error {
	if wxID <= 0 || postID == 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "参数无效")
	}
	post, err := loadPublishedPost(ctx, postID)
	if err != nil {
		return err
	}
	if normalizePostType(post.Type) == PostTypeDebate {
		return gerror.NewCode(gcode.CodeInvalidParameter, "辩论帖不支持点赞")
	}
	_ = post
	cnt, err := dao.UcgPostLike.Ctx(ctx).
		Where(dao.UcgPostLike.Columns().PostId, postID).
		Where(dao.UcgPostLike.Columns().WxId, wxID).
		Count()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	if _, err = dao.UcgPostLike.Ctx(ctx).Data(g.Map{
		dao.UcgPostLike.Columns().PostId:    postID,
		dao.UcgPostLike.Columns().WxId:      wxID,
		dao.UcgPostLike.Columns().CreatedAt: time.Now().Unix(),
	}).Insert(); err != nil {
		return err
	}
	_, err = dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).Increment(dao.UcgPost.Columns().LikeCount, 1)
	if err != nil {
		return err
	}
	_ = saddUserLikedPost(ctx, wxID, postID)
	PublishPostLiked(ctx, postID)
	return nil
}

// LikerDTO 点赞用户视图（昵称与头像经 GetPublicProfile 填充）。
type LikerDTO struct {
	WxId               uint64 `json:"wxId"`
	Nickname           string `json:"nickname"`
	AvatarKey          string `json:"avatarKey,omitempty"`
	AvatarUrl          string `json:"avatarUrl,omitempty"`
	AvatarThumbnailUrl string `json:"avatarThumbnailUrl,omitempty"`
}

// ListPostLikes 帖子点赞用户分页（按点赞时间升序）。
func ListPostLikes(ctx context.Context, postID uint64, page, pageSize int) (*PageResult, error) {
	if postID == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "postId 无效")
	}
	if err := ensurePublishedPost(ctx, postID); err != nil {
		return nil, err
	}
	p := NormalizePage(page, pageSize)
	model := dao.UcgPostLike.Ctx(ctx).Where(dao.UcgPostLike.Columns().PostId, postID)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows, err := model.OrderAsc(dao.UcgPostLike.Columns().CreatedAt).Limit(p.PageSize).Offset(pageOffset(p)).All()
	if err != nil {
		return nil, err
	}
	list := make([]*LikerDTO, 0, len(rows))
	for _, row := range rows {
		var like entity.UcgPostLike
		if err = row.Struct(&like); err != nil {
			return nil, err
		}
		dto := &LikerDTO{WxId: like.WxId}
		if prof, pErr := GetPublicProfile(ctx, like.WxId); pErr == nil && prof != nil {
			dto.Nickname = prof.Nickname
			dto.AvatarKey = prof.AvatarKey
			dto.AvatarUrl = prof.AvatarUrl
			dto.AvatarThumbnailUrl = prof.AvatarThumbnailUrl
		}
		list = append(list, dto)
	}
	return &PageResult{List: list, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// UnlikePost 取消赞。
func UnlikePost(ctx context.Context, wxID int64, postID uint64) error {
	if wxID <= 0 || postID == 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "参数无效")
	}
	res, err := dao.UcgPostLike.Ctx(ctx).
		Where(dao.UcgPostLike.Columns().PostId, postID).
		Where(dao.UcgPostLike.Columns().WxId, wxID).
		Delete()
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		_, err = dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).
			WhereGT(dao.UcgPost.Columns().LikeCount, 0).
			Decrement(dao.UcgPost.Columns().LikeCount, 1)
		if err != nil {
			return err
		}
		_ = sremUserLikedPost(ctx, wxID, postID)
		PublishPostUnliked(ctx, postID)
	}
	return err
}

// CommentDTO 评论视图。
type CommentDTO struct {
	Id           uint64      `json:"id"`
	PostId       uint64      `json:"postId"`
	AuthorWxId   uint64      `json:"authorWxId"`
	Content      string      `json:"content"`
	Status       int         `json:"status,omitempty"`
	RejectReason string      `json:"rejectReason,omitempty"`
	AuditVersion int         `json:"auditVersion,omitempty"`
	CreatedAt    int64       `json:"createdAt"`
	Author       *ProfileDTO `json:"author,omitempty"`
}

// AddComment 发表评论（published 帖）。
func AddComment(ctx context.Context, wxID int64, postID uint64, content string) (*CommentDTO, error) {
	if wxID <= 0 || postID == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "参数无效")
	}
	content = strings.TrimSpace(content)
	if len([]rune(content)) > 1024 {
		content = string([]rune(content)[:1024])
	}
	if content == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "评论不能为空")
	}
	if err := ensurePublishedPost(ctx, postID); err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	auditVersion := 1
	var commentID uint64
	var outboxID uint64
	err := dao.UcgPostComment.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		res, insErr := tx.Model(dao.UcgPostComment.Table()).Ctx(ctx).Data(g.Map{
			dao.UcgPostComment.Columns().PostId:       postID,
			dao.UcgPostComment.Columns().AuthorWxId:   wxID,
			dao.UcgPostComment.Columns().Content:      content,
			dao.UcgPostComment.Columns().Status:       CommentStatusPendingAudit,
			dao.UcgPostComment.Columns().AuditVersion: auditVersion,
			dao.UcgPostComment.Columns().CreatedAt:    now,
		}).Insert()
		if insErr != nil {
			return insErr
		}
		id, _ := res.LastInsertId()
		commentID = uint64(id)
		outboxID, insErr = enqueueAuditPublishOutboxTx(ctx, tx, eventkit.RoutingUcgCommentCreated.String(),
			auditPublishCommentPayload(commentID, auditVersion))
		return insErr
	})
	if err != nil {
		return nil, err
	}
	scheduleAuditPublishAfterCommit(ctx, outboxID)
	dto := &CommentDTO{
		Id:           commentID,
		PostId:       postID,
		AuthorWxId:   uint64(wxID),
		Content:      content,
		Status:       CommentStatusPendingAudit,
		AuditVersion: auditVersion,
		CreatedAt:    now,
	}
	if prof, pErr := GetPublicProfile(ctx, uint64(wxID)); pErr == nil {
		dto.Author = prof
	}
	return dto, nil
}

// ListComments 按 viewer 过滤审态：公众仅见 published；作者可见 pending/rejected+reason。
func ListComments(ctx context.Context, postID uint64, viewerWxID int64) (*CommentsListResult, error) {
	if postID == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "postId 无效")
	}
	post, err := loadPublishedPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	commentCount := int(post.CommentCount)
	cap := commentsListMax(ctx)

	model := dao.UcgPostComment.Ctx(ctx).
		Where(dao.UcgPostComment.Columns().PostId, postID).
		OrderAsc(dao.UcgPostComment.Columns().CreatedAt)
	if viewerWxID <= 0 {
		model = model.Where(dao.UcgPostComment.Columns().Status, CommentStatusPublished)
	}
	if cap > 0 {
		model = model.Limit(cap)
	}
	rows, err := model.All()
	if err != nil {
		return nil, err
	}

	authorIDs := make([]uint64, 0, len(rows))
	seenAuthor := make(map[uint64]struct{}, len(rows))
	comments := make([]entity.UcgPostComment, 0, len(rows))
	for _, row := range rows {
		var c entity.UcgPostComment
		if err = row.Struct(&c); err != nil {
			return nil, err
		}
		comments = append(comments, c)
		if _, ok := seenAuthor[c.AuthorWxId]; !ok {
			seenAuthor[c.AuthorWxId] = struct{}{}
			authorIDs = append(authorIDs, c.AuthorWxId)
		}
	}

	profileMap, err := GetPublicProfilesByWxIDs(ctx, authorIDs)
	if err != nil {
		return nil, err
	}

	list := make([]*CommentDTO, 0, len(comments))
	for _, c := range comments {
		if !commentVisibleToViewer(c, viewerWxID) {
			continue
		}
		dto := &CommentDTO{
			Id:           c.Id,
			PostId:       c.PostId,
			AuthorWxId:   c.AuthorWxId,
			Content:      c.Content,
			Status:       c.Status,
			RejectReason: c.RejectReason,
			AuditVersion: c.AuditVersion,
			CreatedAt:    c.CreatedAt,
		}
		if prof := profileMap[c.AuthorWxId]; prof != nil {
			dto.Author = prof
		}
		list = append(list, dto)
	}

	total := commentCount
	if total <= 0 {
		total = len(list)
	}
	truncated := cap > 0 && commentCount > cap
	return &CommentsListResult{List: list, Total: total, Truncated: truncated}, nil
}

// DeleteComment 删除自己的评论。
func DeleteComment(ctx context.Context, wxID int64, commentID uint64) error {
	row, err := dao.UcgPostComment.Ctx(ctx).Where(dao.UcgPostComment.Columns().Id, commentID).One()
	if err != nil {
		return err
	}
	if row.IsEmpty() {
		return gerror.NewCode(gcode.CodeNotFound, "评论不存在")
	}
	var c entity.UcgPostComment
	if err = row.Struct(&c); err != nil {
		return err
	}
	if int64(c.AuthorWxId) != wxID {
		return gerror.NewCode(gcode.CodeNotAuthorized, "无权删除该评论")
	}
	if _, err = dao.UcgPostComment.Ctx(ctx).Where(dao.UcgPostComment.Columns().Id, commentID).Delete(); err != nil {
		return err
	}
	if c.Status == CommentStatusPublished {
		_, err = dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, c.PostId).
			WhereGT(dao.UcgPost.Columns().CommentCount, 0).
			Decrement(dao.UcgPost.Columns().CommentCount, 1)
		if err != nil {
			return err
		}
		PublishCommentRemoved(ctx, c.PostId, commentID)
	}
	return err
}

func ensurePublishedPost(ctx context.Context, postID uint64) error {
	_, err := loadPublishedPost(ctx, postID)
	return err
}

// loadPublishedPost 读取已发布帖子行（含 comment_count，供评论列表避免额外 COUNT）。
func loadPublishedPost(ctx context.Context, postID uint64) (*entity.UcgPost, error) {
	row, err := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).One()
	if err != nil {
		return nil, err
	}
	if row.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotFound, "帖子不存在")
	}
	var post entity.UcgPost
	if err = row.Struct(&post); err != nil {
		return nil, err
	}
	if post.Status != PostStatusPublished {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "帖子未发布")
	}
	return &post, nil
}

// commentVisibleToViewer 公众仅见 published；评论作者可见自己的 pending/rejected。
func commentVisibleToViewer(c entity.UcgPostComment, viewerWxID int64) bool {
	if c.Status == CommentStatusPublished {
		return true
	}
	return viewerWxID > 0 && int64(c.AuthorWxId) == viewerWxID
}

// commentsListMax 评论列表硬上限；配置 ucg.comments.listMax，默认 500，0 表示不限制。
func commentsListMax(ctx context.Context) int {
	v := g.Cfg().MustGet(ctx, "ucg.comments.listMax")
	if v.IsNil() || v.IsEmpty() {
		return 500
	}
	return v.Int()
}
