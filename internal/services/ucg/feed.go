package ucg

import (
	"context"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// ListRecommendFeed 推荐 Feed：尚无 recommend 行的帖置顶，已算分帖按 score 排序。
// viewerWxID>0 时填充 likedByMe。
func ListRecommendFeed(
	ctx context.Context,
	viewerWxID int64,
	page, pageSize int,
) (*PageResult, error) {

	// 根据入参坐标,从redis中按距离排序的动态id列表(在动态发布/编辑时写入redis id+坐标)
	// 将上一步得到的id列表,根据redis中缓存的热区动态分值排序（在冷热重算推荐分值写入redis）
	//  todo 内存分页
	//  todo 再试图从redis中获取每个动态的详细信息（在动态发布/编辑时写入redis，可能缓存过期）
	//  todo 如果缓存过期，则从db中获取每个动态的详细信息
	//  todo 最后返回PageResult

	// ==========================
	// 1️⃣ 规范化分页参数
	// 防止 page <= 0 / pageSize 过大
	// ==========================
	p := NormalizePage(page, pageSize)

	// ==========================
	// 2️⃣ 统计总数（只统计主表）
	// ⚠️ 不 JOIN 推荐表，避免统计放大
	// ==========================
	total, err := dao.UcgPost.Ctx(ctx).
		Where(dao.UcgPost.Columns().Status, PostStatusPublished).
		Count()
	if err != nil {
		return nil, err
	}

	// ==========================
	// 3️⃣ 构造推荐流查询
	// 推荐分来自 ucg_post_recommend
	// 排序由推荐分主导，时间为辅
	// ==========================
	model := dao.UcgPost.Ctx(ctx).As("p").

		// ✅ LEFT JOIN 推荐表
		// 目的：用推荐分参与排序，不是筛选
		LeftJoin("ucg_post_recommend r", "r.post_id = p.id").

		// ✅ 只查已发布帖子
		Where("p."+dao.UcgPost.Columns().Status, PostStatusPublished).

		// ✅ 只返回帖子字段，减少 IO
		Fields("p.*")

	// ==========================
	// 4️⃣ 排序规则（核心）
	// 推荐分优先，时间兜底
	// ==========================
	model = model.
		// ① 有推荐的帖子永远在前（NULL 在后）
		OrderDesc("(r.post_id IS NULL)").

		// ② 推荐分高的在前
		OrderDesc("r.score").

		// ③ 发布时间新的在前（时间衰减已在推荐分算法中体现）
		OrderDesc("p." + dao.UcgPost.Columns().PublishedAt).

		// ④ 最终按 ID 兜底，保证排序稳定
		OrderDesc("p." + dao.UcgPost.Columns().Id)

	// ==========================
	// 5️⃣ 分页查询
	// 永远是小结果集
	// ==========================
	rows, err := model.
		Limit(p.PageSize).
		Offset(pageOffset(p)).
		All()
	if err != nil {
		return nil, err
	}

	// ==========================
	// 6️⃣ 结果转换
	// 补充点赞、收藏、作者等附属信息
	// ==========================
	list, err := postsFromResult(ctx, rows, viewerWxID)
	if err != nil {
		return nil, err
	}

	// ==========================
	// 7️⃣ 返回分页结果
	// ==========================
	return &PageResult{
		List:     list,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, nil
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
