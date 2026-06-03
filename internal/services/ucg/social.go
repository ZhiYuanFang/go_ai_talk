package ucg

import (
	"context"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

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
	if err := ensurePublishedPost(ctx, postID); err != nil {
		return err
	}
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
	return err
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
	}
	return err
}

// CommentDTO 评论视图。
type CommentDTO struct {
	Id         uint64       `json:"id"`
	PostId     uint64       `json:"postId"`
	AuthorWxId uint64       `json:"authorWxId"`
	Content    string       `json:"content"`
	CreatedAt  int64        `json:"createdAt"`
	Author     *ProfileDTO  `json:"author,omitempty"`
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
	res, err := dao.UcgPostComment.Ctx(ctx).Data(g.Map{
		dao.UcgPostComment.Columns().PostId:     postID,
		dao.UcgPostComment.Columns().AuthorWxId: wxID,
		dao.UcgPostComment.Columns().Content:    content,
		dao.UcgPostComment.Columns().CreatedAt:  now,
	}).Insert()
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	_, _ = dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).Increment(dao.UcgPost.Columns().CommentCount, 1)
	dto := &CommentDTO{
		Id:         uint64(id),
		PostId:     postID,
		AuthorWxId: uint64(wxID),
		Content:    content,
		CreatedAt:  now,
	}
	if prof, pErr := GetPublicProfile(ctx, uint64(wxID)); pErr == nil {
		dto.Author = prof
	}
	return dto, nil
}

// ListComments 评论分页。
func ListComments(ctx context.Context, postID uint64, page, pageSize int) (*PageResult, error) {
	if postID == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "postId 无效")
	}
	p := NormalizePage(page, pageSize)
	model := dao.UcgPostComment.Ctx(ctx).Where(dao.UcgPostComment.Columns().PostId, postID)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows, err := model.OrderAsc(dao.UcgPostComment.Columns().CreatedAt).Limit(p.PageSize).Offset(pageOffset(p)).All()
	if err != nil {
		return nil, err
	}
	list := make([]*CommentDTO, 0, len(rows))
	for _, row := range rows {
		var c entity.UcgPostComment
		if err = row.Struct(&c); err != nil {
			return nil, err
		}
		dto := &CommentDTO{
			Id:         c.Id,
			PostId:     c.PostId,
			AuthorWxId: c.AuthorWxId,
			Content:    c.Content,
			CreatedAt:  c.CreatedAt,
		}
		if prof, pErr := GetPublicProfile(ctx, c.AuthorWxId); pErr == nil {
			dto.Author = prof
		}
		list = append(list, dto)
	}
	return &PageResult{List: list, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
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
	_, err = dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, c.PostId).
		WhereGT(dao.UcgPost.Columns().CommentCount, 0).
		Decrement(dao.UcgPost.Columns().CommentCount, 1)
	return err
}

func ensurePublishedPost(ctx context.Context, postID uint64) error {
	row, err := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).One()
	if err != nil {
		return err
	}
	if row.IsEmpty() {
		return gerror.NewCode(gcode.CodeNotFound, "帖子不存在")
	}
	var post entity.UcgPost
	if err = row.Struct(&post); err != nil {
		return err
	}
	if post.Status != PostStatusPublished {
		return gerror.NewCode(gcode.CodeInvalidParameter, "帖子未发布")
	}
	return nil
}
