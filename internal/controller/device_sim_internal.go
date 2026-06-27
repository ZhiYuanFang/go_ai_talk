package controller

import (
	"context"
	"errors"

	v1 "hello/api/v1"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// DeviceSimInternalCtrl 模拟用户 device 内部接口。
type DeviceSimInternalCtrl struct{}

func NewDeviceSimInternalCtrl() *DeviceSimInternalCtrl { return &DeviceSimInternalCtrl{} }

// SimUsernameRegister POST /device/internal/api/sim/username/register
func (c *DeviceSimInternalCtrl) SimUsernameRegister(ctx context.Context, req *v1.DeviceSimUsernameRegisterReq) (res *v1.DeviceSimUsernameRegisterRes, err error) {
	_ = c
	wxID, err := device.SimUsernameRegister(ctx, req.Account, req.Password)
	if err != nil {
		if errors.Is(err, device.ErrWxUsernameInvalid) || errors.Is(err, device.ErrWxPasswordInvalid) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		if errors.Is(err, device.ErrWxUsernameTaken) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceSimUsernameRegisterRes{WxId: wxID, Account: req.Account}, nil
}

// SimWxList GET /device/internal/api/sim/wx/list
func (c *DeviceSimInternalCtrl) SimWxList(ctx context.Context, req *v1.DeviceSimWxListReq) (res *v1.DeviceSimWxListRes, err error) {
	_ = c
	list, total, err := device.ListSimulatedWx(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	items := make([]v1.DeviceSimWxListItem, 0, len(list))
	for _, row := range list {
		items = append(items, v1.DeviceSimWxListItem{WxId: row.WxId, Account: row.Account})
	}
	return &v1.DeviceSimWxListRes{
		List:     items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// SimWxRandom GET /device/internal/api/sim/wx/random
func (c *DeviceSimInternalCtrl) SimWxRandom(ctx context.Context, req *v1.DeviceSimWxRandomReq) (res *v1.DeviceSimWxRandomRes, err error) {
	_ = c
	item, ok, err := device.PickRandomSimulatedWx(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &v1.DeviceSimWxRandomRes{}, nil
	}
	return &v1.DeviceSimWxRandomRes{WxId: item.WxId, Account: item.Account}, nil
}
