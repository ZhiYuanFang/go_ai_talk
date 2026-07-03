package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// DeviceUcgInternalCtrl device 域 UCG 专用内部接口（须携带网关共享密钥）。
type DeviceUcgInternalCtrl struct{}

func NewDeviceUcgInternalCtrl() *DeviceUcgInternalCtrl { return &DeviceUcgInternalCtrl{} }

func (c *DeviceUcgInternalCtrl) WxValidate(ctx context.Context, req *v1.DeviceUcgWxValidateReq) (res *v1.DeviceUcgWxValidateRes, err error) {
	_ = c
	item, err := device.UcgWxValidate(ctx, req.WxId)
	if err != nil {
		if err == device.ErrUcgWxIDInvalid {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceUcgWxValidateRes{
		WxId:     item.WxId,
		Exists:   item.Exists,
		DeviceNo: item.DeviceNo,
		BabyName: item.BabyName,
	}, nil
}

func (c *DeviceUcgInternalCtrl) WxBatch(ctx context.Context, req *v1.DeviceUcgWxBatchReq) (res *v1.DeviceUcgWxBatchRes, err error) {
	_ = c
	list, err := device.UcgWxBatch(ctx, req.WxIds)
	if err != nil {
		return nil, err
	}
	items := make([]v1.DeviceUcgWxBatchItem, 0, len(list))
	for _, row := range list {
		items = append(items, v1.DeviceUcgWxBatchItem{
			WxId:        row.WxId,
			Exists:      row.Exists,
			DeviceNo:    row.DeviceNo,
			BabyName:    row.BabyName,
			IpLocation:  row.IpLocation,
			IsSimulated: row.IsSimulated,
			ForceValue:  row.ForceValue,
		})
	}
	return &v1.DeviceUcgWxBatchRes{List: items}, nil
}

func (c *DeviceUcgInternalCtrl) WxIpLocationPut(ctx context.Context, req *v1.DeviceUcgWxIpLocationPutReq) (res *v1.DeviceUcgWxIpLocationPutRes, err error) {
	_ = c
	if err = device.UcgWxUpdateIpLocation(ctx, req.WxId, req.IpLocation); err != nil {
		if err == device.ErrUcgWxIDInvalid {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceUcgWxIpLocationPutRes{}, nil
}

func (c *DeviceUcgInternalCtrl) WxBabyName(ctx context.Context, req *v1.DeviceUcgWxBabyNameReq) (res *v1.DeviceUcgWxBabyNameRes, err error) {
	_ = c
	name, err := device.UcgWxBabyName(ctx, req.WxId)
	if err != nil {
		if err == device.ErrUcgWxIDInvalid {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceUcgWxBabyNameRes{BabyName: name}, nil
}

func (c *DeviceUcgInternalCtrl) WxForceIncrement(ctx context.Context, req *v1.DeviceUcgWxForceIncrementReq) (res *v1.DeviceUcgWxForceIncrementRes, err error) {
	_ = c
	if err = device.UcgWxIncrementForceValue(ctx, req.WxId); err != nil {
		if err == device.ErrUcgWxIDInvalid {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
		}
		return nil, err
	}
	return &v1.DeviceUcgWxForceIncrementRes{}, nil
}

// deviceUcgInternalSecretMiddleware 校验 X-Device-Gateway-Internal-Secret（兼容 X-Gateway-Internal-Secret）。
func deviceUcgInternalSecretMiddleware(r *ghttp.Request) {
	secret := strings.TrimSpace(r.GetHeader(device.HeaderDeviceGatewayInternalSecret))
	if secret == "" {
		secret = strings.TrimSpace(r.GetHeader("X-Gateway-Internal-Secret"))
	}
	if !device.ValidateGatewayInternalSecretHeader(secret) {
		r.Response.Status = 403
		r.Response.WriteJson(g.Map{"code": 403, "message": "内部接口未授权"})
		r.ExitAll()
		return
	}
	r.Middleware.Next()
}
