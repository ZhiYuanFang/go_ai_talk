package devicectrl

import (
	"context"

	v1 "hello/api/v1"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// PredictProIdeasList GET /device/admin/api/predict-pro-ideas/list
// 业务说明：经 DeviceAdmin() 调用，避免扩大 DeviceAdminContract 面。
func (c *AdminCtrl) PredictProIdeasList(ctx context.Context, req *v1.DeviceAdminPredictProIdeaListReq) (*v1.DeviceAdminPredictProIdeaListRes, error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	page, err := device.DeviceAdmin().ListPredictProIdeasAdmin(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	list := make([]v1.DeviceAdminPredictProIdeaItem, 0, len(page.List))
	for _, row := range page.List {
		list = append(list, v1.DeviceAdminPredictProIdeaItem{
			Id: row.Id, WxId: row.WxId, BabyName: row.BabyName,
			Content: row.Content, DisplayText: row.DisplayText,
			LikeCount: row.LikeCount, CreatedAt: row.CreatedAt,
		})
	}
	return &v1.DeviceAdminPredictProIdeaListRes{
		List: list, Total: page.Total, Page: page.Page, PageSize: page.PageSize,
	}, nil
}

// PredictProIdeasUpdate POST /device/admin/api/predict-pro-ideas/update
func (c *AdminCtrl) PredictProIdeasUpdate(ctx context.Context, req *v1.DeviceAdminPredictProIdeaUpdateReq) (*v1.DeviceAdminPredictProIdeaUpdateRes, error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id 无效")
	}
	if err := device.DeviceAdmin().UpdatePredictProIdeaDisplayText(ctx, req.Id, req.DisplayText); err != nil {
		return nil, mapPredictProIdeaBizErr(err)
	}
	return &v1.DeviceAdminPredictProIdeaUpdateRes{}, nil
}

// PredictProIdeasDelete POST /device/admin/api/predict-pro-ideas/delete
func (c *AdminCtrl) PredictProIdeasDelete(ctx context.Context, req *v1.DeviceAdminPredictProIdeaDeleteReq) (*v1.DeviceAdminPredictProIdeaDeleteRes, error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id 无效")
	}
	if err := device.DeviceAdmin().SoftDeletePredictProIdea(ctx, req.Id); err != nil {
		return nil, mapPredictProIdeaBizErr(err)
	}
	return &v1.DeviceAdminPredictProIdeaDeleteRes{}, nil
}
