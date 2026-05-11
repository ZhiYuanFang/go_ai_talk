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

const hdrAppUserInternalWxCode = "X-Internal-Wx-Code"

// DeviceAppUserCtrl 设备域 App「用户」子域 HTTP（画像 + wx 等；路径扁平在 /device/app/api/user/*，与网关 /device/app/api/login 区分）。
type DeviceAppUserCtrl struct{}

// NewDeviceAppUserCtrl 构造 App 用户域控制器。
func NewDeviceAppUserCtrl() *DeviceAppUserCtrl { return &DeviceAppUserCtrl{} }

// Get 按设备号查询生日与性别等画像字段。
func (c *DeviceAppUserCtrl) Get(ctx context.Context, req *v1.DeviceProfileGetReq) (res *v1.DeviceProfileGetRes, err error) {
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
func (c *DeviceAppUserCtrl) Save(ctx context.Context, req *v1.DeviceProfileSaveReq) (res *v1.DeviceProfileSaveRes, err error) {
	_ = c
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	sex := 0
	if req.Sex > 0 {
		sex = 1
	}
	if err := device.DeviceAdmin().SaveUserProfile(ctx, deviceNo, req.Birthday, sex); err != nil {
		return nil, err
	}
	return &v1.DeviceProfileSaveRes{}, nil
}

// BindWx POST /device/app/api/user/bindwx（wxCode 来自 Header X-Internal-Wx-Code）。
func (c *DeviceAppUserCtrl) BindWx(ctx context.Context, req *v1.DeviceProfileBindWxReq) (res *v1.DeviceProfileBindWxRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxCode := strings.TrimSpace(r.GetHeader(hdrAppUserInternalWxCode))
	if wxCode == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Code")
	}
	if err := device.WxBindDevice(ctx, wxCode, req.DeviceNo); err != nil {
		return nil, err
	}
	return &v1.DeviceProfileBindWxRes{}, nil
}

// AutoSave POST /device/app/api/user/auto_save
func (c *DeviceAppUserCtrl) AutoSave(ctx context.Context, req *v1.DeviceProfileAutoSaveReq) (res *v1.DeviceProfileAutoSaveRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxCode := strings.TrimSpace(r.GetHeader(hdrAppUserInternalWxCode))
	if wxCode == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Code")
	}
	dn, err := device.WxAutoSaveProfile(ctx, wxCode, req.Birthday, req.Sex)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceProfileAutoSaveRes{DeviceNo: dn}, nil
}

// Login POST /device/app/api/user/login（纯业务 wx 登录，无 JWT；非网关聚合 POST /device/app/api/login）。
func (c *DeviceAppUserCtrl) Login(ctx context.Context, req *v1.DeviceWxLoginReq) (res *v1.DeviceWxLoginRes, err error) {
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

// Detail GET /device/app/api/user/detail
func (c *DeviceAppUserCtrl) Detail(ctx context.Context, req *v1.DeviceWxDetailReq) (res *v1.DeviceWxDetailRes, err error) {
	_ = req
	r := ghttp.RequestFromCtx(ctx)
	wxCode := strings.TrimSpace(r.GetHeader(hdrAppUserInternalWxCode))
	if wxCode == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Code")
	}
	dn, err := device.WxDeviceNoByCode(ctx, wxCode)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceWxDetailRes{DeviceNo: dn}, nil
}

// InternalByID GET /device/app/api/user/internal/by-id — 仅允许携带正确网关密钥的调用方。
func (c *DeviceAppUserCtrl) InternalByID(ctx context.Context, req *v1.DeviceWxInternalByIDReq) (res *v1.DeviceWxInternalByIDRes, err error) {
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
