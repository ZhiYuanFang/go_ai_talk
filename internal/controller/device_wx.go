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

const headerInternalWxCode = "X-Internal-Wx-Code"

// DeviceWxCtrl 微信 wx 表相关 HTTP。
type DeviceWxCtrl struct{}

func NewDeviceWxCtrl() *DeviceWxCtrl { return &DeviceWxCtrl{} }

// Login POST /device/wx/api/login
func (c *DeviceWxCtrl) Login(ctx context.Context, req *v1.DeviceWxLoginReq) (res *v1.DeviceWxLoginRes, err error) {
	out, err := device.WxLogin(ctx, req.WxCode, req.Platform)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceWxLoginRes{
		WxId:      out.WxId,
		WxCode:    out.WxCode,
		DeviceNo:  out.DeviceNo,
		IsNewUser: out.IsNewUser,
	}, nil
}

// Detail GET /device/wx/api/detail
func (c *DeviceWxCtrl) Detail(ctx context.Context, req *v1.DeviceWxDetailReq) (res *v1.DeviceWxDetailRes, err error) {
	_ = req
	r := ghttp.RequestFromCtx(ctx)
	wxCode := strings.TrimSpace(r.GetHeader(headerInternalWxCode))
	if wxCode == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Code")
	}
	dn, err := device.WxDeviceNoByCode(ctx, wxCode)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceWxDetailRes{DeviceNo: dn}, nil
}

// InternalByID GET /device/wx/api/internal/by-id — 仅允许携带正确网关密钥的调用方。
func (c *DeviceWxCtrl) InternalByID(ctx context.Context, req *v1.DeviceWxInternalByIDReq) (res *v1.DeviceWxInternalByIDRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	if !device.ValidateGatewayInternalSecret(r.GetHeader("X-Gateway-Internal-Secret")) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "内部接口未授权")
	}
	if req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id 无效")
	}
	code, err := device.WxCodeByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceWxInternalByIDRes{WxCode: code}, nil
}
