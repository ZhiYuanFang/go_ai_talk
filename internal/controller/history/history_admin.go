package historyctrl

import (
	deviceclient "hello/internal/clients/device"
	"context"
	"strings"

	v1 "hello/api/v1"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// HistoryAdminCtrl 运维 Hub 设备历史 API（Header: X-Admin-Password，由网关注入）。
type HistoryAdminCtrl struct {
	inner *HistoryCtrl
}

// NewHistoryAdminCtrl 构造 HistoryAdminCtrl。
func NewHistoryAdminCtrl(inner *HistoryCtrl) *HistoryAdminCtrl {
	return &HistoryAdminCtrl{inner: inner}
}

func (c *HistoryAdminCtrl) requireAdmin(ctx context.Context) error {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return gerror.NewCode(gcode.CodeNotAuthorized, "口令错误")
	}
	if !deviceclient.HTTPDeviceAdmin().VerifyPassword(strings.TrimSpace(r.GetHeader("X-Admin-Password"))) {
		return gerror.NewCode(gcode.CodeNotAuthorized, "口令错误")
	}
	return nil
}

// List GET /device/admin/api/history/list
func (c *HistoryAdminCtrl) List(ctx context.Context, req *v1.DeviceAdminHistoryListReq) (res *v1.DeviceAdminHistoryListRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	out, err := c.inner.List(ctx, &v1.DeviceHistoryListReq{
		DeviceNo: req.DeviceNo,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminHistoryListRes{
		List:     out.List,
		Total:    out.Total,
		Page:     out.Page,
		PageSize: out.PageSize,
	}, nil
}

// Suggest GET /device/admin/api/history/suggest
func (c *HistoryAdminCtrl) Suggest(ctx context.Context, req *v1.DeviceAdminHistorySuggestReq) (res *v1.DeviceAdminHistorySuggestRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	out, err := c.inner.Suggest(ctx, &v1.DeviceHistorySuggestReq{DeviceNo: req.DeviceNo})
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminHistorySuggestRes{List: out.List}, nil
}

// Birthday GET /device/admin/api/history/birthday
func (c *HistoryAdminCtrl) Birthday(ctx context.Context, req *v1.DeviceAdminHistoryBirthdayGetReq) (res *v1.DeviceAdminHistoryBirthdayGetRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	out, err := c.inner.Birthday(ctx, &v1.DeviceHistoryBirthdayGetReq{DeviceNo: req.DeviceNo})
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminHistoryBirthdayGetRes{
		BabyName: out.BabyName,
		Birthday: out.Birthday,
		Sex:      out.Sex,
	}, nil
}

// BirthdaySave POST /device/admin/api/history/birthday/save
func (c *HistoryAdminCtrl) BirthdaySave(ctx context.Context, req *v1.DeviceAdminHistoryBirthdaySaveReq) (res *v1.DeviceAdminHistoryBirthdaySaveRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	_, err = c.inner.BirthdaySave(ctx, &v1.DeviceHistoryBirthdaySaveReq{
		DeviceNo: req.DeviceNo,
		BabyName: req.BabyName,
		Birthday: req.Birthday,
		Sex:      req.Sex,
	})
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminHistoryBirthdaySaveRes{}, nil
}

// EventAdd POST /device/admin/api/history/event/add
func (c *HistoryAdminCtrl) EventAdd(ctx context.Context, req *v1.DeviceAdminHistoryEventAddReq) (res *v1.DeviceAdminHistoryEventAddRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	out, err := c.inner.EventAdd(ctx, &v1.DeviceHistoryEventAddReq{
		DeviceNo:    req.DeviceNo,
		EventId:     req.EventId,
		EventName:   req.EventName,
		EventUnit:   req.EventUnit,
		EventNumber: req.EventNumber,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Remark:      req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminHistoryEventAddRes{Id: out.Id}, nil
}

// EventUpdate POST /device/admin/api/history/event/update
func (c *HistoryAdminCtrl) EventUpdate(ctx context.Context, req *v1.DeviceAdminHistoryEventUpdateReq) (res *v1.DeviceAdminHistoryEventUpdateRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	_, err = c.inner.EventUpdate(ctx, &v1.DeviceHistoryEventUpdateReq{
		Id:          req.Id,
		DeviceNo:    req.DeviceNo,
		EventId:     req.EventId,
		EventName:   req.EventName,
		EventUnit:   req.EventUnit,
		EventNumber: req.EventNumber,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Remark:      req.Remark,
		PostId:      req.PostId,
		MediaType:   req.MediaType,
		ImageKeys:   req.ImageKeys,
		VideoKey:    req.VideoKey,
	})
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminHistoryEventUpdateRes{}, nil
}

// EventDelete POST /device/admin/api/history/event/delete
func (c *HistoryAdminCtrl) EventDelete(ctx context.Context, req *v1.DeviceAdminHistoryEventDeleteReq) (res *v1.DeviceAdminHistoryEventDeleteRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	_, err = c.inner.EventDelete(ctx, &v1.DeviceHistoryEventDeleteReq{
		Id:       req.Id,
		DeviceNo: req.DeviceNo,
	})
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAdminHistoryEventDeleteRes{}, nil
}
