package devicectrl

import (
	"context"
	"errors"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/platform/httpmeta"
	device "hello/internal/services/device"
	cashclient "hello/internal/clients/cash"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// DeviceMerchantCtrl 商户后台 API（鉴权：网关注入 X-Internal-Merchant-Id）。
type DeviceMerchantCtrl struct{}

func NewDeviceMerchantCtrl() *DeviceMerchantCtrl { return &DeviceMerchantCtrl{} }

func merchantIDFromHeader(r *ghttp.Request) (int64, error) {
	raw := strings.TrimSpace(r.GetHeader(httpmeta.HeaderInternalMerchantId))
	if raw == "" {
		return 0, gerror.NewCode(gcode.CodeNotAuthorized, "缺少商户身份")
	}
	var id int64
	for _, c := range raw {
		if c < '0' || c > '9' {
			return 0, gerror.NewCode(gcode.CodeNotAuthorized, "商户身份无效")
		}
		id = id*10 + int64(c-'0')
	}
	if id <= 0 {
		return 0, gerror.NewCode(gcode.CodeNotAuthorized, "商户身份无效")
	}
	return id, nil
}

func mapMerchantErr(err error) error {
	switch {
	case errors.Is(err, device.ErrMerchantStoreNameTaken),
		errors.Is(err, device.ErrMerchantStoreNameInvalid),
		errors.Is(err, device.ErrMerchantPasswordInvalid),
		errors.Is(err, device.ErrMerchantAuthFailed),
		errors.Is(err, device.ErrMerchantBabyQuota),
		errors.Is(err, device.ErrMerchantBabyNotOwned),
		errors.Is(err, device.ErrAssignStaffRoleRequired),
		errors.Is(err, device.ErrBabyMemberNotFound),
		errors.Is(err, device.ErrWxUsernameTaken),
		errors.Is(err, device.ErrWxUsernameInvalid),
		errors.Is(err, device.ErrWxPasswordInvalid):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	default:
		return err
	}
}

// Register POST /device/merchant/api/register
func (c *DeviceMerchantCtrl) Register(ctx context.Context, req *v1.MerchantRegisterReq) (*v1.MerchantRegisterRes, error) {
	_ = c
	id, err := device.MerchantRegister(ctx, req.StoreName, req.Password)
	if err != nil {
		return nil, mapMerchantErr(err)
	}
	row, _ := device.MerchantByID(ctx, id)
	name := req.StoreName
	if row != nil {
		name = row.StoreName
	}
	return &v1.MerchantRegisterRes{MerchantId: id, StoreName: name}, nil
}

// Login POST /device/merchant/api/login
func (c *DeviceMerchantCtrl) Login(ctx context.Context, req *v1.MerchantLoginReq) (*v1.MerchantLoginRes, error) {
	_ = c
	row, err := device.MerchantLogin(ctx, req.StoreName, req.Password)
	if err != nil {
		return nil, mapMerchantErr(err)
	}
	return &v1.MerchantLoginRes{MerchantId: row.ID, StoreName: row.StoreName}, nil
}

// CreateStaff POST /device/merchant/api/staff/create
func (c *DeviceMerchantCtrl) CreateStaff(ctx context.Context, req *v1.MerchantCreateStaffReq) (*v1.MerchantCreateStaffRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	mid, err := merchantIDFromHeader(r)
	if err != nil {
		return nil, err
	}
	wxID, err := device.MerchantCreateStaff(ctx, mid, req.Account, req.Password)
	if err != nil {
		return nil, mapMerchantErr(err)
	}
	return &v1.MerchantCreateStaffRes{WxId: wxID}, nil
}

// CreateBaby POST /device/merchant/api/baby/create
func (c *DeviceMerchantCtrl) CreateBaby(ctx context.Context, req *v1.MerchantCreateBabyReq) (*v1.MerchantCreateBabyRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	mid, err := merchantIDFromHeader(r)
	if err != nil {
		return nil, err
	}
	entitled, e2 := cashclient.RemoteIsMerchantEntitled(ctx, mid)
	if e2 != nil {
		// 权益探活失败时按未开通额度处理（保守）
		entitled = false
	}
	sex := 0
	if req.Sex > 0 {
		sex = 1
	}
	dn, err := device.MerchantCreateBaby(ctx, mid, req.BabyName, req.Birthday, sex, entitled)
	if err != nil {
		return nil, mapMerchantErr(err)
	}
	return &v1.MerchantCreateBabyRes{DeviceNo: dn}, nil
}

// ListBabies GET /device/merchant/api/baby/list
func (c *DeviceMerchantCtrl) ListBabies(ctx context.Context, req *v1.MerchantListBabiesReq) (*v1.MerchantListBabiesRes, error) {
	_ = req
	r := ghttp.RequestFromCtx(ctx)
	mid, err := merchantIDFromHeader(r)
	if err != nil {
		return nil, err
	}
	list, err := device.MerchantListBabies(ctx, mid)
	if err != nil {
		return nil, err
	}
	return &v1.MerchantListBabiesRes{DeviceNos: list}, nil
}

// AssignBaby POST /device/merchant/api/baby/assign
func (c *DeviceMerchantCtrl) AssignBaby(ctx context.Context, req *v1.MerchantAssignBabyReq) (*v1.MerchantAssignBabyRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	mid, err := merchantIDFromHeader(r)
	if err != nil {
		return nil, err
	}
	if err := device.MerchantAssignBabyToStaff(ctx, mid, req.WxId, req.DeviceNo, req.SetFocus); err != nil {
		return nil, mapMerchantErr(err)
	}
	return &v1.MerchantAssignBabyRes{}, nil
}

// UnassignBaby POST /device/merchant/api/baby/unassign
func (c *DeviceMerchantCtrl) UnassignBaby(ctx context.Context, req *v1.MerchantUnassignBabyReq) (*v1.MerchantUnassignBabyRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	mid, err := merchantIDFromHeader(r)
	if err != nil {
		return nil, err
	}
	if err := device.MerchantUnassignBabyFromStaff(ctx, mid, req.WxId, req.DeviceNo); err != nil {
		return nil, mapMerchantErr(err)
	}
	return &v1.MerchantUnassignBabyRes{}, nil
}

// UpdateBabyProfile POST /device/merchant/api/baby/profile
func (c *DeviceMerchantCtrl) UpdateBabyProfile(ctx context.Context, req *v1.MerchantUpdateBabyProfileReq) (*v1.MerchantUpdateBabyProfileRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	mid, err := merchantIDFromHeader(r)
	if err != nil {
		return nil, err
	}
	sex := 0
	if req.Sex > 0 {
		sex = 1
	}
	if err := device.MerchantUpdateBabyProfile(ctx, mid, req.DeviceNo, req.BabyName, req.Birthday, sex); err != nil {
		return nil, mapMerchantErr(err)
	}
	return &v1.MerchantUpdateBabyProfileRes{}, nil
}
