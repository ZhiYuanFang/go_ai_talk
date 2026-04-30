package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	contracts "hello/internal/services/contracts"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AdminCtrl 设备管理后台 API（Header: X-Admin-Password）。
type AdminCtrl struct {
	Admin contracts.DeviceAdminContract
}

// NewAdminCtrl 构造 AdminCtrl。
func NewAdminCtrl(admin contracts.DeviceAdminContract) *AdminCtrl {
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
		if err == device.ErrDeviceExists {
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
	needQuantity := 0
	if req.NeedQuantity > 0 {
		needQuantity = 1
	}
	err = c.Admin.AddEvent(ctx, name, needQuantity, req.ExtraNames)
	if err != nil {
		if err == device.ErrEventExists {
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
	needQuantity := 0
	if req.NeedQuantity > 0 {
		needQuantity = 1
	}
	err = c.Admin.UpdateEvent(ctx, req.Id, name, needQuantity, req.ExtraNames)
	if err != nil {
		switch err {
		case device.ErrEventExists:
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, err.Error())
		case device.ErrEventNotFound:
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		default:
			return nil, err
		}
	}
	return &v1.DeviceAdminEventUpdateRes{}, nil
}

// EventDelete 删除事件。
func (c *AdminCtrl) EventDelete(ctx context.Context, req *v1.DeviceAdminEventDeleteReq) (res *v1.DeviceAdminEventDeleteRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "事件ID无效")
	}
	if err := c.Admin.DeleteEvent(ctx, req.Id); err != nil {
		if err == device.ErrEventNotFound {
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceAdminEventDeleteRes{}, nil
}

// QaList 问答库列表（qa 表）。
func (c *AdminCtrl) QaList(ctx context.Context, req *v1.DeviceAdminQaListReq) (res *v1.DeviceAdminQaListRes, err error) {
	_ = req
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	items, err := c.Admin.ListQA(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminQaListRes{List: items}, nil
}

// ActionList 动作预设列表（action 表）。
func (c *AdminCtrl) ActionList(ctx context.Context, req *v1.DeviceAdminActionListReq) (res *v1.DeviceAdminActionListRes, err error) {
	_ = req
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	raw, err := c.Admin.ListActionsForAdmin(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]v1.DeviceAdminActionItem, len(raw))
	for i, a := range raw {
		list[i] = v1.DeviceAdminActionItem{
			Id:              a.Id,
			Name:            a.Name,
			TargetType:      a.TargetType,
			TargetTypeLabel: a.TargetTypeLabel,
		}
	}
	return &v1.DeviceAdminActionListRes{List: list}, nil
}

// ActionUpdate 更新动作预设。
func (c *AdminCtrl) ActionUpdate(ctx context.Context, req *v1.DeviceAdminActionUpdateReq) (res *v1.DeviceAdminActionUpdateRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "动作ID无效")
	}
	name := strings.TrimSpace(req.Name)
	targetType := strings.TrimSpace(req.TargetType)
	if name == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "动作名称不能为空")
	}
	if targetType == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "动作目标不能为空")
	}
	if err := c.Admin.UpdateAction(ctx, req.Id, name, targetType); err != nil {
		if err == device.ErrActionNotFound {
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceAdminActionUpdateRes{}, nil
}

// ActionDelete 删除动作预设。
func (c *AdminCtrl) ActionDelete(ctx context.Context, req *v1.DeviceAdminActionDeleteReq) (res *v1.DeviceAdminActionDeleteRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "动作ID无效")
	}
	if err := c.Admin.DeleteAction(ctx, req.Id); err != nil {
		if err == device.ErrActionNotFound {
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceAdminActionDeleteRes{}, nil
}
