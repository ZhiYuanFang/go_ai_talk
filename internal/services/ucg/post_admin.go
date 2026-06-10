package ucg

import (
	"context"
	"strings"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

const (
	adminMaxPageSize    = 100
	adminMaxBatchReject = 100
)

// AdminBatchRejectResult 管理端批量驳回结果（部分成功）。
type AdminBatchRejectResult struct {
	Rejected []uint64 `json:"rejected"`
	Skipped  []uint64 `json:"skipped"`
	Failed   []uint64 `json:"failed"`
}

// NormalizeAdminPage 管理端分页；pageSize 上限 100。
func NormalizeAdminPage(page, pageSize int) PageParams {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > adminMaxPageSize {
		pageSize = adminMaxPageSize
	}
	return PageParams{Page: page, PageSize: pageSize}
}

// ListPostsForAdmin 管理端分页列出动态；statusFilter 为 nil 时不按状态过滤。
func ListPostsForAdmin(ctx context.Context, page, pageSize int, statusFilter *int) (*PageResult, error) {
	p := NormalizeAdminPage(page, pageSize)
	model := dao.UcgPost.Ctx(ctx)
	if statusFilter != nil {
		model = model.Where(dao.UcgPost.Columns().Status, *statusFilter)
	}
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows, err := model.
		OrderDesc(dao.UcgPost.Columns().UpdatedAt).
		Limit(p.PageSize).
		Offset(pageOffset(p)).
		All()
	if err != nil {
		return nil, err
	}
	// viewerWxID=0：管理端不需 likedByMe
	list, err := postsFromResult(ctx, rows, 0)
	if err != nil {
		return nil, err
	}
	return &PageResult{List: list, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// AdminBatchRejectPosts 批量人工驳回；不通知作者。已是 rejected 的 id 计入 skipped。
func AdminBatchRejectPosts(ctx context.Context, postIDs []uint64, reason string) (*AdminBatchRejectResult, error) {
	if len(postIDs) == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "postIds 不能为空")
	}
	if len(postIDs) > adminMaxBatchReject {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "单次最多驳回 100 条")
	}
	reason = strings.TrimSpace(reason)
	out := &AdminBatchRejectResult{
		Rejected: make([]uint64, 0, len(postIDs)),
		Skipped:  make([]uint64, 0),
		Failed:   make([]uint64, 0),
	}
	seen := make(map[uint64]struct{}, len(postIDs))
	for _, id := range postIDs {
		if id == 0 {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}

		row, err := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, id).One()
		if err != nil {
			out.Failed = append(out.Failed, id)
			continue
		}
		if row.IsEmpty() {
			out.Failed = append(out.Failed, id)
			continue
		}
		var post entity.UcgPost
		if err = row.Struct(&post); err != nil {
			out.Failed = append(out.Failed, id)
			continue
		}
		if post.Status == PostStatusRejected {
			out.Skipped = append(out.Skipped, id)
			continue
		}
		if err = rejectPostByID(ctx, id, reason); err != nil {
			out.Failed = append(out.Failed, id)
			continue
		}
		out.Rejected = append(out.Rejected, id)
	}
	return out, nil
}
