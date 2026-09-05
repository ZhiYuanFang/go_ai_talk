package devicectrl

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	contracts "hello/internal/services/contracts"
	device "hello/internal/services/device"
	"hello/internal/shared/eventlogo"

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

// UserList 设备记录分页列表。
func (c *AdminCtrl) UserList(ctx context.Context, req *v1.DeviceAdminUserListReq) (res *v1.DeviceAdminUserListRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := c.Admin.ListUsersPage(ctx, req.Page, req.PageSize, req.Q)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminUserListRes{
		List:     result.List,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}

// WxList wx 账号分页列表。
func (c *AdminCtrl) WxList(ctx context.Context, req *v1.DeviceAdminWxListReq) (res *v1.DeviceAdminWxListRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := c.Admin.ListWxPage(ctx, req.Page, req.PageSize, req.Q)
	if err != nil {
		return nil, err
	}
	list := make([]v1.DeviceAdminWxListItem, 0, len(result.List))
	for _, it := range result.List {
		list = append(list, v1.DeviceAdminWxListItem{
			Id:        it.Id,
			DeviceNo:  it.DeviceNo,
			Unionid:   it.Unionid,
			Platform:  it.Platform,
			Account:   it.Account,
			CreatedAt: it.CreatedAt,
			BabyName:  it.BabyName,
		})
	}
	return &v1.DeviceAdminWxListRes{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
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
	return &v1.DeviceAdminEventListRes{List: eventlogo.MapEventsLogoCdn(ctx, items)}, nil
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
		switch err {
		case device.ErrEventNotFound:
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		case device.ErrEventHasChildren:
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, err.Error())
		default:
			return nil, err
		}
	}
	return &v1.DeviceAdminEventDeleteRes{}, nil
}

// QaList 问答库分页列表（qa 表，id 倒序）。
func (c *AdminCtrl) QaList(ctx context.Context, req *v1.DeviceAdminQaListReq) (res *v1.DeviceAdminQaListRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := c.Admin.ListQAPage(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminQaListRes{
		List:     result.List,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}

// QaDelete 删除问答库行。
func (c *AdminCtrl) QaDelete(ctx context.Context, req *v1.DeviceAdminQaDeleteReq) (res *v1.DeviceAdminQaDeleteRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id 无效")
	}
	if err := c.Admin.DeleteQA(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeviceAdminQaDeleteRes{}, nil
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

// FeedbackList 用户反馈分页列表。
func (c *AdminCtrl) FeedbackList(ctx context.Context, req *v1.DeviceAdminFeedbackListReq) (res *v1.DeviceAdminFeedbackListRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := c.Admin.ListFeedbackPage(ctx, req.Page, req.PageSize, req.UnrepliedOnly)
	if err != nil {
		return nil, err
	}
	list := make([]v1.DeviceAdminFeedbackItem, 0, len(result.List))
	for _, row := range result.List {
		item := v1.DeviceAdminFeedbackItem{
			Id:        row.Id,
			WxId:      row.WxId,
			Question:  row.Question,
			Status:    row.Status,
			CreatedAt: feedbackTimeUnix(row.CreatedAt),
		}
		if row.OfficialReply != "" {
			item.OfficialReply = row.OfficialReply
		}
		item.RepliedAt = feedbackTimeUnixPtr(row.RepliedAt)
		list = append(list, item)
	}
	return &v1.DeviceAdminFeedbackListRes{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}

// FeedbackReply 官方回复（每条仅一次）。
func (c *AdminCtrl) FeedbackReply(ctx context.Context, req *v1.DeviceAdminFeedbackReplyReq) (res *v1.DeviceAdminFeedbackReplyRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id 无效")
	}
	if err := c.Admin.ReplyFeedback(ctx, req.Id, req.OfficialReply); err != nil {
		switch err {
		case device.ErrFeedbackNotFound:
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		case device.ErrFeedbackAlreadyReplied:
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, err.Error())
		default:
			if gerror.Code(err) == gcode.CodeInvalidParameter {
				return nil, err
			}
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
	}
	return &v1.DeviceAdminFeedbackReplyRes{}, nil
}
