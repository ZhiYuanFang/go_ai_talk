package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// DeviceProfileCtrl 设备画像内部查询接口（供 voice 等服务调用）。
type DeviceProfileCtrl struct{}

// NewDeviceProfileCtrl 构造设备画像控制器。
func NewDeviceProfileCtrl() *DeviceProfileCtrl {
	return &DeviceProfileCtrl{}
}

// Get 按设备号查询生日与性别等画像字段。
func (c *DeviceProfileCtrl) Get(ctx context.Context, req *v1.DeviceProfileGetReq) (res *v1.DeviceProfileGetRes, err error) {
	_ = c
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	profile, err := device.DeviceProfile().GetProfile(ctx, deviceNo)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceProfileGetRes{
		DeviceNo: profile.DeviceNo,
		Birthday: profile.Birthday,
		Sex:      profile.Sex,
	}, nil
}

// Save 保存设备画像（生日/性别），由 device 库权威落库。
func (c *DeviceProfileCtrl) Save(ctx context.Context, req *v1.DeviceProfileSaveReq) (res *v1.DeviceProfileSaveRes, err error) {
	_ = c
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	sex := 0
	if req.Sex > 0 {
		sex = 1
	}
	if err := device.DeviceAdmin().SaveUserProfile(ctx, deviceNo, strings.TrimSpace(req.Birthday), sex); err != nil {
		return nil, err
	}
	return &v1.DeviceProfileSaveRes{}, nil
}

// BindWx POST /device/profile/api/bindwx（wxCode 来自 Header X-Internal-Wx-Code）。
func (c *DeviceProfileCtrl) BindWx(ctx context.Context, req *v1.DeviceProfileBindWxReq) (res *v1.DeviceProfileBindWxRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxCode := strings.TrimSpace(r.GetHeader(headerInternalWxCode))
	if wxCode == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Code")
	}
	if err := device.WxBindDevice(ctx, wxCode, req.DeviceNo); err != nil {
		return nil, err
	}
	return &v1.DeviceProfileBindWxRes{}, nil
}

// AutoSave POST /device/profile/api/auto_save
func (c *DeviceProfileCtrl) AutoSave(ctx context.Context, req *v1.DeviceProfileAutoSaveReq) (res *v1.DeviceProfileAutoSaveRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxCode := strings.TrimSpace(r.GetHeader(headerInternalWxCode))
	if wxCode == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Code")
	}
	dn, err := device.WxAutoSaveProfile(ctx, wxCode, strings.TrimSpace(req.Birthday), req.Sex)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceProfileAutoSaveRes{DeviceNo: dn}, nil
}
