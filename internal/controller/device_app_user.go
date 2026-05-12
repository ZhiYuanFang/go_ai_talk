package controller

import (
	"context"
	"errors"
	"strconv"
	"strings"

	v1 "hello/api/v1"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

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

func wxIDFromAppUserHeader(r *ghttp.Request) (int64, error) {
	s := strings.TrimSpace(r.GetHeader(device.HeaderInternalWxId()))
	if s == "" {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "X-Internal-Wx-Id 无效")
	}
	return id, nil
}

// BindWx POST /device/app/api/user/bindwx（wx 主键来自 Header X-Internal-Wx-Id）。
func (c *DeviceAppUserCtrl) BindWx(ctx context.Context, req *v1.DeviceProfileBindWxReq) (res *v1.DeviceProfileBindWxRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	if err := device.WxBindDevice(ctx, wxID, req.DeviceNo); err != nil {
		return nil, err
	}
	return &v1.DeviceProfileBindWxRes{}, nil
}

// AutoSave POST /device/app/api/user/auto_save
func (c *DeviceAppUserCtrl) AutoSave(ctx context.Context, req *v1.DeviceProfileAutoSaveReq) (res *v1.DeviceProfileAutoSaveRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	dn, err := device.WxAutoSaveProfile(ctx, wxID, req.Birthday, req.Sex)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceProfileAutoSaveRes{DeviceNo: dn}, nil
}

// Login POST /device/app/api/user/login（纯业务 wx 登录，无 JWT；非网关聚合 POST /device/app/api/login）。
func (c *DeviceAppUserCtrl) Login(ctx context.Context, req *v1.DeviceWxLoginReq) (res *v1.DeviceWxLoginRes, err error) {
	out, err := device.WxLogin(ctx, req.JsCode, req.Platform)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceWxLoginRes{
		WxId:      out.WxId,
		DeviceNo:  out.DeviceNo,
		IsNewUser: out.IsNewUser,
	}, nil
}

// DeviceLogin POST /device/app/api/user/device_login（纯业务，无 JWT）。
func (c *DeviceAppUserCtrl) DeviceLogin(ctx context.Context, req *v1.DeviceWxDeviceLoginReq) (res *v1.DeviceWxDeviceLoginRes, err error) {
	out, err := device.WxDeviceLoginByDeviceNo(ctx, req.DeviceNo)
	if err != nil {
		if errors.Is(err, device.ErrWxDeviceLoginRejected) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		if errors.Is(err, device.ErrWxDeviceLoginDeviceNoEmpty) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceWxDeviceLoginRes{
		WxId:      out.WxId,
		DeviceNo:  out.DeviceNo,
		IsNewUser: out.IsNewUser,
	}, nil
}

// Detail GET /device/app/api/user/detail
func (c *DeviceAppUserCtrl) Detail(ctx context.Context, req *v1.DeviceWxDetailReq) (res *v1.DeviceWxDetailRes, err error) {
	_ = req
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	dn, err := device.WxDeviceNoByWxID(ctx, wxID)
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
	u, err := device.WxUnionIDByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceWxInternalByIDRes{UnionId: u}, nil
}

// InternalDeviceNoByWxID GET /device/app/api/user/internal/device-no-by-wx-id — 网关刷新 access 时拉取 device_no。
func (c *DeviceAppUserCtrl) InternalDeviceNoByWxID(ctx context.Context, req *v1.DeviceWxInternalDeviceNoByWxIDReq) (res *v1.DeviceWxInternalDeviceNoByWxIDRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	if !device.ValidateGatewayInternalSecret(r.GetHeader("X-Gateway-Internal-Secret")) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "内部接口未授权")
	}
	if req.WxId <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	dn, err := device.WxDeviceNoByWxID(ctx, req.WxId)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceWxInternalDeviceNoByWxIDRes{DeviceNo: dn}, nil
}
