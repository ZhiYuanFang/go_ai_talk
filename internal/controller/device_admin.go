package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/service"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AdminCtrl 设备管理后台 API（Header: X-Admin-Password）。
type AdminCtrl struct {
	Admin service.DeviceAdminContract
}

// NewAdminCtrl 构造 AdminCtrl。
func NewAdminCtrl(admin service.DeviceAdminContract) *AdminCtrl {
	return &AdminCtrl{Admin: admin}
}

func adminPassword(ctx context.Context) string {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return ""
	}
	return strings.TrimSpace(r.GetHeader("X-Admin-Password"))
}

func (c *AdminCtrl) requireAdmin(ctx context.Context) error {
	if !c.Admin.VerifyPassword(adminPassword(ctx)) {
		return gerror.NewCode(gcode.CodeNotAuthorized, "口令错误")
	}
	return nil
}

// Register 注册设备。
func (c *AdminCtrl) Register(ctx context.Context, req *v1.DeviceAdminRegisterReq) (res *v1.DeviceAdminRegisterRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	activeTime, err := c.Admin.Register(ctx, deviceNo)
	if err != nil {
		if err == service.ErrDeviceExists {
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceAdminRegisterRes{DeviceNo: deviceNo, ActiveTime: activeTime}, nil
}

// List 设备列表。
func (c *AdminCtrl) List(ctx context.Context, req *v1.DeviceAdminListReq) (res *v1.DeviceAdminListRes, err error) {
	_ = req
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	items, err := c.Admin.List(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminListRes{List: items}, nil
}

// EventList 事件字典列表。
func (c *AdminCtrl) EventList(ctx context.Context, req *v1.DeviceAdminEventListReq) (res *v1.DeviceAdminEventListRes, err error) {
	_ = req
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	items, err := c.Admin.ListEvents(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminEventListRes{List: items}, nil
}

// EventAdd 新增事件。
func (c *AdminCtrl) EventAdd(ctx context.Context, req *v1.DeviceAdminEventAddReq) (res *v1.DeviceAdminEventAddRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "事件名称不能为空")
	}
	needTime := 0
	if req.NeedTime > 0 {
		needTime = 1
	}
	needQuantity := 0
	if req.NeedQuantity > 0 {
		needQuantity = 1
	}
	err = c.Admin.AddEvent(ctx, name, needTime, needQuantity)
	if err != nil {
		if err == service.ErrEventExists {
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceAdminEventAddRes{}, nil
}

// EventUpdate 更新事件名称。
func (c *AdminCtrl) EventUpdate(ctx context.Context, req *v1.DeviceAdminEventUpdateReq) (res *v1.DeviceAdminEventUpdateRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "事件ID无效")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "事件名称不能为空")
	}
	needTime := 0
	if req.NeedTime > 0 {
		needTime = 1
	}
	needQuantity := 0
	if req.NeedQuantity > 0 {
		needQuantity = 1
	}
	err = c.Admin.UpdateEvent(ctx, req.Id, name, needTime, needQuantity)
	if err != nil {
		switch err {
		case service.ErrEventExists:
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, err.Error())
		case service.ErrEventNotFound:
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		default:
			return nil, err
		}
	}
	return &v1.DeviceAdminEventUpdateRes{}, nil
}

// IntentionList 意图列表。
func (c *AdminCtrl) IntentionList(ctx context.Context, req *v1.DeviceAdminIntentionListReq) (res *v1.DeviceAdminIntentionListRes, err error) {
	_ = req
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	items, err := c.Admin.ListIntentions(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminIntentionListRes{List: items}, nil
}

// IntentionUpdate 更新意图动态历史上限。
func (c *AdminCtrl) IntentionUpdate(ctx context.Context, req *v1.DeviceAdminIntentionUpdateReq) (res *v1.DeviceAdminIntentionUpdateRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "意图ID无效")
	}
	if req.UpperLimit < 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "upperLimit 不能小于0")
	}
	if err := c.Admin.UpdateIntentionUpperLimit(ctx, req.Id, req.UpperLimit); err != nil {
		if err == service.ErrIntentionNotFound {
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceAdminIntentionUpdateRes{}, nil
}
