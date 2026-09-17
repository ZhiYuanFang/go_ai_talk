package pushctrl

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	pushsvc "hello/internal/services/push"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AdminPushCtrl 推送设备运维 API（宿主 push-service）。
type AdminPushCtrl struct{}

// requirePushAdmin 校验网关注入的 X-Admin-Password。
func requirePushAdmin(ctx context.Context) error {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return gerror.NewCode(gcode.CodeNotAuthorized, "缺少请求上下文")
	}
	if !pushsvc.VerifyPushAdminPassword(strings.TrimSpace(r.GetHeader(pushsvc.HeaderAdminPassword))) {
		return gerror.NewCode(gcode.CodeNotAuthorized, "管理口令无效")
	}
	return nil
}

// DevicesGet GET /push/admin/api/devices — 分页列表。
func (c *AdminPushCtrl) DevicesGet(ctx context.Context, req *v1.PushAdminDevicesListReq) (res *v1.PushAdminDevicesListRes, err error) {
	if err = requirePushAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := pushsvc.ListPushDevicesForAdmin(ctx, pushsvc.AdminDeviceListFilter{
		Page:        req.Page,
		PageSize:    req.PageSize,
		WxID:        req.WxId,
		Channel:     req.Channel,
		TokenPrefix: req.TokenPrefix,
	})
	if err != nil {
		return nil, err
	}
	out := &v1.PushAdminDevicesListRes{
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
		List:     make([]v1.PushAdminDeviceItem, 0, len(result.List)),
	}
	for _, row := range result.List {
		out.List = append(out.List, v1.PushAdminDeviceItem{
			Id:        row.ID,
			WxId:      row.WxID,
			Channel:   row.Channel,
			DeviceKey: row.DeviceKey,
			Token:     row.Token,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

// DeviceDelete DELETE /push/admin/api/devices/{id} — 按主键删除。
func (c *AdminPushCtrl) DeviceDelete(ctx context.Context, req *v1.PushAdminDeviceDeleteReq) (res *v1.PushAdminDeviceDeleteRes, err error) {
	if err = requirePushAdmin(ctx); err != nil {
		return nil, err
	}
	if req.Id == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id 无效")
	}
	if err = pushsvc.DeletePushDeviceByID(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.PushAdminDeviceDeleteRes{}, nil
}
