package devicectrl

import (
	"hello/internal/platform/httpmeta"
	"context"
	"errors"
	v1 "hello/api/v1"
	device "hello/internal/services/device"
	"strings"

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
		BabyName: profile.BabyName,
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
	if err := device.DeviceAdmin().SaveUserProfile(ctx, deviceNo, strings.TrimSpace(req.BabyName), req.Birthday, sex); err != nil {
		return nil, err
	}
	return &v1.DeviceProfileSaveRes{}, nil
}

func wxIDFromAppUserHeader(r *ghttp.Request) (int64, error) {
	id, msg := httpmeta.RequireHeaderWxID(r.GetHeader(httpmeta.HeaderInternalWxId))
	if msg != "" {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, msg)
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
	dn, err := device.WxAutoSaveProfile(ctx, wxID, strings.TrimSpace(req.BabyName), req.Birthday, req.Sex)
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

// AppleLogin POST /device/app/api/user/apple/login（纯业务 Apple 登录，无 JWT）。
func (c *DeviceAppUserCtrl) AppleLogin(ctx context.Context, req *v1.DeviceAppleLoginReq) (res *v1.DeviceAppleLoginRes, err error) {
	out, err := device.WxAppleLogin(ctx, req.IdentityToken, req.Platform)
	if err != nil {
		if errors.Is(err, device.ErrAppleIdentityTokenEmpty) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		if errors.Is(err, device.ErrAppleIdentityTokenInvalid) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceAppleLoginRes{
		WxId:      out.WxId,
		DeviceNo:  out.DeviceNo,
		IsNewUser: out.IsNewUser,
	}, nil
}

// AppleBind POST /device/app/api/user/apple/bind（当前账号由 Header X-Internal-Wx-Id 指定）。
func (c *DeviceAppUserCtrl) AppleBind(ctx context.Context, req *v1.DeviceAppleBindReq) (res *v1.DeviceAppleBindRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	if err = device.WxBindApple(ctx, wxID, req.IdentityToken, req.Platform); err != nil {
		if errors.Is(err, device.ErrWxIDInvalid) || errors.Is(err, device.ErrAppleIdentityTokenEmpty) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		if errors.Is(err, device.ErrAppleIdentityTokenInvalid) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		if errors.Is(err, device.ErrAppleAlreadyBound) || errors.Is(err, device.ErrAppleSubTakenByOther) || errors.Is(err, device.ErrAccountMergeConflict) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceAppleBindRes{}, nil
}

// WxBindWx POST /device/app/api/user/wx/bindwx（当前账号由 Header X-Internal-Wx-Id 指定）。
func (c *DeviceAppUserCtrl) WxBindWx(ctx context.Context, req *v1.DeviceWxBindWxReq) (res *v1.DeviceWxBindWxRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	if err = device.WxBindWxByCode(ctx, wxID, req.JsCode, req.Platform); err != nil {
		if errors.Is(err, device.ErrWxIDInvalid) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		if errors.Is(err, device.ErrWxAlreadyBoundUnionID) || errors.Is(err, device.ErrWxUnionIDTakenByOther) || errors.Is(err, device.ErrAccountMergeConflict) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceWxBindWxRes{}, nil
}

// UsernameRegister POST /device/app/api/user/username/register（纯业务，无 JWT）。
func (c *DeviceAppUserCtrl) UsernameRegister(ctx context.Context, req *v1.DeviceUsernameRegisterReq) (res *v1.DeviceUsernameRegisterRes, err error) {
	_ = c
	wxID, err := device.WxUsernameRegister(ctx, req.Account, req.Password)
	if err != nil {
		if errors.Is(err, device.ErrWxUsernameInvalid) || errors.Is(err, device.ErrWxPasswordInvalid) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		if errors.Is(err, device.ErrWxUsernameTaken) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceUsernameRegisterRes{WxId: wxID}, nil
}

// UsernameLogin POST /device/app/api/user/username/login（纯业务，无 JWT）。
func (c *DeviceAppUserCtrl) UsernameLogin(ctx context.Context, req *v1.DeviceUsernameLoginReq) (res *v1.DeviceUsernameLoginRes, err error) {
	_ = c
	out, err := device.WxUsernameLogin(ctx, req.Account, req.Password)
	if err != nil {
		if errors.Is(err, device.ErrWxUsernameInvalid) || errors.Is(err, device.ErrWxPasswordInvalid) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		if errors.Is(err, device.ErrWxUsernameAuthFailed) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceUsernameLoginRes{WxId: out.WxId, DeviceNo: out.DeviceNo, IsNewUser: out.IsNewUser}, nil
}

// UsernameBindWx POST /device/app/api/user/username/bindwx（当前账号由 Header X-Internal-Wx-Id 指定）。
func (c *DeviceAppUserCtrl) UsernameBindWx(ctx context.Context, req *v1.DeviceUsernameBindWxReq) (res *v1.DeviceUsernameBindWxRes, err error) {
	_ = c
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	if err = device.WxUsernameBindWxByCode(ctx, wxID, req.JsCode, req.Platform); err != nil {
		if errors.Is(err, device.ErrWxUsernameNotSet) || errors.Is(err, device.ErrWxIDInvalid) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		if errors.Is(err, device.ErrWxAlreadyBoundUnionID) || errors.Is(err, device.ErrWxUnionIDTakenByOther) || errors.Is(err, device.ErrAccountMergeConflict) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceUsernameBindWxRes{}, nil
}

// UsernameBindDevice POST /device/app/api/user/username/bind_device（当前账号由 Header X-Internal-Wx-Id 指定）。
func (c *DeviceAppUserCtrl) UsernameBindDevice(ctx context.Context, req *v1.DeviceUsernameBindDeviceReq) (res *v1.DeviceUsernameBindDeviceRes, err error) {
	_ = c
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	if err = device.WxUsernameBindDevice(ctx, wxID, req.DeviceNo); err != nil {
		if errors.Is(err, device.ErrWxUsernameNotSet) || errors.Is(err, device.ErrWxIDInvalid) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		if errors.Is(err, device.ErrDeviceNotRegistered) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceUsernameBindDeviceRes{}, nil
}

// UsernameChangePassword POST /device/app/api/user/username/change_password（当前账号由 Header X-Internal-Wx-Id 指定）。
func (c *DeviceAppUserCtrl) UsernameChangePassword(ctx context.Context, req *v1.DeviceUsernameChangePasswordReq) (res *v1.DeviceUsernameChangePasswordRes, err error) {
	_ = c
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	if err = device.WxUsernameChangePassword(ctx, wxID, req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, device.ErrWxPasswordInvalid) || errors.Is(err, device.ErrWxIDInvalid) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		if errors.Is(err, device.ErrWxUsernameAuthFailed) || errors.Is(err, device.ErrWxUsernameNotSet) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceUsernameChangePasswordRes{}, nil
}

// WxCreateUsername POST /device/app/api/user/wx/create_username（当前账号由 Header X-Internal-Wx-Id 指定）。
func (c *DeviceAppUserCtrl) WxCreateUsername(ctx context.Context, req *v1.DeviceWxCreateUsernameReq) (res *v1.DeviceWxCreateUsernameRes, err error) {
	_ = c
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	if err = device.WxCreateUsernamePassword(ctx, wxID, req.Account, req.Password); err != nil {
		if errors.Is(err, device.ErrWxUsernameInvalid) || errors.Is(err, device.ErrWxPasswordInvalid) || errors.Is(err, device.ErrWxIDInvalid) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		if errors.Is(err, device.ErrWxUsernameTaken) || errors.Is(err, device.ErrWxUsernameAlreadySet) || errors.Is(err, device.ErrWxUnionIDRequired) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceWxCreateUsernameRes{}, nil
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

// Profile GET /device/app/api/user/profile（wx 主键来自 Header X-Internal-Wx-Id）。
func (c *DeviceAppUserCtrl) Profile(ctx context.Context, req *v1.DeviceUserProfileReq) (res *v1.DeviceUserProfileRes, err error) {
	_ = req
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	out, err := device.WxUserProfileByWxID(ctx, wxID)
	if err != nil {
		if errors.Is(err, device.ErrWxDeactivateWxIDInvalid) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		if errors.Is(err, device.ErrWxDeactivateNotFound) {
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceUserProfileRes{
		IsWxBound:     out.IsWxBound,
		IsAppleBound:  out.IsAppleBound,
		AuthProviders: out.AuthProviders,
		Account:       out.Account,
		DeviceNo:      out.DeviceNo,
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

// Deactivate POST /device/app/api/user/deactivate
// 仅删除 wx 表当前主键记录，不联动删除其他域数据。
func (c *DeviceAppUserCtrl) Deactivate(ctx context.Context, req *v1.DeviceUserDeactivateReq) (res *v1.DeviceUserDeactivateRes, err error) {
	_ = req
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	if err = device.WxDeactivateByID(ctx, wxID); err != nil {
		if errors.Is(err, device.ErrWxDeactivateWxIDInvalid) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		if errors.Is(err, device.ErrWxDeactivateNotFound) {
			return nil, gerror.NewCode(gcode.CodeNotFound, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceUserDeactivateRes{}, nil
}

// InternalByID GET /device/app/api/user/internal/by-id — 仅允许携带正确网关密钥的调用方。
func (c *DeviceAppUserCtrl) InternalByID(ctx context.Context, req *v1.DeviceWxInternalByIDReq) (res *v1.DeviceWxInternalByIDRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	// 兼容新旧头名：优先 X-Device-Gateway-Internal-Secret，回退 X-Gateway-Internal-Secret。
	// voice-service 等内部调用方统一发送新头名，旧头名仅用于历史调用方兼容。
	if !device.ValidateGatewayInternalSecret(device.GatewayInternalSecretHeaderFromRequest(r)) {
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
	// 兼容新旧头名：优先 X-Device-Gateway-Internal-Secret，回退 X-Gateway-Internal-Secret。
	// voice-service 等内部调用方统一发送新头名，旧头名仅用于历史调用方兼容。
	if !device.ValidateGatewayInternalSecret(device.GatewayInternalSecretHeaderFromRequest(r)) {
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

// InternalWxIDByDeviceNo GET /device/app/api/user/internal/wx-id-by-device-no — voice 反查登录态。
func (c *DeviceAppUserCtrl) InternalWxIDByDeviceNo(ctx context.Context, req *v1.DeviceWxInternalWxIDByDeviceNoReq) (res *v1.DeviceWxInternalWxIDByDeviceNoRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	// 兼容新旧头名：优先 X-Device-Gateway-Internal-Secret，回退 X-Gateway-Internal-Secret。
	// voice-service 等内部调用方统一发送新头名，旧头名仅用于历史调用方兼容。
	if !device.ValidateGatewayInternalSecret(device.GatewayInternalSecretHeaderFromRequest(r)) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "内部接口未授权")
	}
	wxID, err := device.WxIDByDeviceNo(ctx, req.DeviceNo)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceWxInternalWxIDByDeviceNoRes{WxId: wxID}, nil
}
