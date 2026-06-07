package ucg

import (
	"context"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// ListRecommendFeed 推荐 Feed：优先按 ucg_post_recommend.score 排序，无 score 时回退 published_at。
// viewerWxID>0 时填充 likedByMe。
func ListRecommendFeed(ctx context.Context, viewerWxID int64, page, pageSize int) (*PageResult, error) {
	p := NormalizePage(page, pageSize)
	// 联结推荐分；仅 published
	model := dao.UcgPost.Ctx(ctx).As("p").
		LeftJoin("ucg_post_recommend r", "r.post_id=p.id").
		Where("p."+dao.UcgPost.Columns().Status, PostStatusPublished)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows, err := model.
		Fields("p.*").
		OrderDesc("r.score").
		OrderDesc("p." + dao.UcgPost.Columns().PublishedAt).
		OrderDesc("p." + dao.UcgPost.Columns().Id).
		Limit(p.PageSize).Offset(pageOffset(p)).All()
	if err != nil {
		return nil, err
	}
	list, err := postsFromResult(ctx, rows, viewerWxID)
	if err != nil {
		return nil, err
	}
	return &PageResult{List: list, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// ListFollowingFeed 关注 Feed：需登录 wxId，仅 followee 的 published 帖。
func ListFollowingFeed(ctx context.Context, wxID int64, page, pageSize int) (*PageResult, error) {
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
	model := dao.UcgPost.Ctx(ctx).
		Where(dao.UcgPost.Columns().Status, PostStatusPublished).
		WhereIn(dao.UcgPost.Columns().AuthorWxId, ids)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows, err := model.OrderDesc(dao.UcgPost.Columns().PublishedAt).Limit(p.PageSize).Offset(pageOffset(p)).All()
	if err != nil {
		return nil, err
	}
	list, err := postsFromResult(ctx, rows, wxID)
	if err != nil {
		return nil, err
	}
	return &PageResult{List: list, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}
