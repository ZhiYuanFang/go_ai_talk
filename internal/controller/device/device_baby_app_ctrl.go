package devicectrl

import (
	"context"
	"errors"

	v1 "hello/api/v1"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ListBabies GET /device/app/api/user/babies
func (c *DeviceAppUserCtrl) ListBabies(ctx context.Context, req *v1.DeviceBabyListReq) (*v1.DeviceBabyListRes, error) {
	_ = req
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	items, err := device.ListBabyMembers(ctx, wxID)
	if err != nil {
		return nil, err
	}
	out := make([]v1.DeviceBabyListItem, 0, len(items))
	for _, it := range items {
		out = append(out, v1.DeviceBabyListItem{
			DeviceNo: it.DeviceNo,
			BabyName: it.BabyName,
			Source:   it.Source,
			IsFocus:  it.IsFocus,
		})
	}
	return &v1.DeviceBabyListRes{Items: out}, nil
}

// FocusBaby POST /device/app/api/user/babies/focus
func (c *DeviceAppUserCtrl) FocusBaby(ctx context.Context, req *v1.DeviceBabyFocusReq) (*v1.DeviceBabyFocusRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	if err := device.SetFocusBaby(ctx, wxID, req.DeviceNo); err != nil {
		if errors.Is(err, device.ErrBabyMemberNotOwned) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceBabyFocusRes{DeviceNo: req.DeviceNo}, nil
}

// CreateBaby POST /device/app/api/user/babies/create
func (c *DeviceAppUserCtrl) CreateBaby(ctx context.Context, req *v1.DeviceBabyCreateReq) (*v1.DeviceBabyCreateRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	sex := 0
	if req.Sex > 0 {
		sex = 1
	}
	dn, err := device.CreateVirtualBabyForParent(ctx, wxID, req.BabyName, req.Birthday, sex)
	if err != nil {
		if errors.Is(err, device.ErrWxRoleParentOnly) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceBabyCreateRes{DeviceNo: dn}, nil
}

// UnbindBaby POST /device/app/api/user/babies/unbind — 恒拒绝
func (c *DeviceAppUserCtrl) UnbindBaby(ctx context.Context, req *v1.DeviceBabyUnbindReq) (*v1.DeviceBabyUnbindRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	if err := device.AppUnbindBabyRejected(ctx, wxID, req.DeviceNo); err != nil {
		return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	return &v1.DeviceBabyUnbindRes{}, nil
}
