package devicectrl

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// DeviceAppointmentInternalCtrl 预约下次约定内部读（供 Python；须 InternalSecretMiddleware）。
// 业务：只读 appointment_next，不触碰 predict Redis。
type DeviceAppointmentInternalCtrl struct{}

// NewDeviceAppointmentInternalCtrl 构造内部读控制器。
func NewDeviceAppointmentInternalCtrl() *DeviceAppointmentInternalCtrl {
	return &DeviceAppointmentInternalCtrl{}
}

// NextGet GET /device/internal/api/appointment/next?deviceNo=&eventId=
// 无行返回 nextAt=0；不校验微信绑机（靠内部密钥）。
func (c *DeviceAppointmentInternalCtrl) NextGet(ctx context.Context, req *v1.DeviceInternalAppointmentNextGetReq) (res *v1.DeviceInternalAppointmentNextGetRes, err error) {
	_ = c
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if req.EventId <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "eventId 无效")
	}
	nextAt, err := device.GetAppointmentNextAt(ctx, deviceNo, req.EventId)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceInternalAppointmentNextGetRes{NextAt: nextAt}, nil
}
