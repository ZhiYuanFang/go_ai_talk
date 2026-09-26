package devicectrl

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/net/ghttp"
)

// DeviceAppAppointmentCtrl App 预约下次约定 API（JWT + X-Internal-Wx-Id；须绑机一致）。
// 业务：仅持久化客户端记忆；与 history / predict-imminent 解耦。
type DeviceAppAppointmentCtrl struct{}

// NewDeviceAppAppointmentCtrl 构造预约 App 控制器。
func NewDeviceAppAppointmentCtrl() *DeviceAppAppointmentCtrl {
	return &DeviceAppAppointmentCtrl{}
}

// NextGet GET /device/app/api/appointment/next
func (c *DeviceAppAppointmentCtrl) NextGet(ctx context.Context, req *v1.DeviceAppAppointmentNextGetReq) (res *v1.DeviceAppAppointmentNextGetRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if err := device.RequireWxBoundDeviceNo(ctx, wxID, deviceNo); err != nil {
		return nil, err
	}
	nextAt, err := device.GetAppointmentNextAt(ctx, deviceNo, req.EventId)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceAppAppointmentNextGetRes{NextAt: nextAt}, nil
}

// NextPut PUT /device/app/api/appointment/next
func (c *DeviceAppAppointmentCtrl) NextPut(ctx context.Context, req *v1.DeviceAppAppointmentNextPutReq) (res *v1.DeviceAppAppointmentNextPutRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if err := device.RequireWxBoundDeviceNo(ctx, wxID, deviceNo); err != nil {
		return nil, err
	}
	if err := device.PutAppointmentNextAt(ctx, deviceNo, req.EventId, req.NextAt); err != nil {
		return nil, err
	}
	return &v1.DeviceAppAppointmentNextPutRes{}, nil
}
