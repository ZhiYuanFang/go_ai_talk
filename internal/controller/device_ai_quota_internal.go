package controller

import (
	"context"

	v1 "hello/api/v1"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// DeviceAIQuotaInternalCtrl device 域 AI 额度 internal API（须携带网关共享密钥）。
type DeviceAIQuotaInternalCtrl struct{}

func NewDeviceAIQuotaInternalCtrl() *DeviceAIQuotaInternalCtrl {
	return &DeviceAIQuotaInternalCtrl{}
}

func mapQuotaErr(err error) error {
	if err == nil {
		return nil
	}
	switch err {
	case device.ErrAINotLoggedIn:
		return gerror.NewCode(device.GCodeAINotLoggedIn(), err.Error())
	case device.ErrAIQuotaExhausted:
		return gerror.NewCode(device.GCodeAIQuotaExhausted(), err.Error())
	default:
		return err
	}
}

func parseFeature(s string) (device.AIQuotaFeature, error) {
	switch device.AIQuotaFeature(s) {
	case device.AIQuotaPolish, device.AIQuotaVoiceAI:
		return device.AIQuotaFeature(s), nil
	default:
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "feature 无效")
	}
}

func (c *DeviceAIQuotaInternalCtrl) Check(ctx context.Context, req *v1.DeviceInternalAIQuotaCheckReq) (res *v1.DeviceInternalAIQuotaCheckRes, err error) {
	_ = c
	feature, err := parseFeature(req.Feature)
	if err != nil {
		return nil, err
	}
	snap, err := device.CheckAIQuota(ctx, req.WxId, feature)
	if err != nil {
		return nil, mapQuotaErr(err)
	}
	return &v1.DeviceInternalAIQuotaCheckRes{
		Allowed: snap.Allowed,
		Used:    snap.Used,
		Limit:   snap.Limit,
	}, nil
}

func (c *DeviceAIQuotaInternalCtrl) Consume(ctx context.Context, req *v1.DeviceInternalAIQuotaConsumeReq) (res *v1.DeviceInternalAIQuotaConsumeRes, err error) {
	_ = c
	feature, err := parseFeature(req.Feature)
	if err != nil {
		return nil, err
	}
	snap, err := device.ConsumeAIQuota(ctx, req.WxId, feature)
	if err != nil {
		return nil, mapQuotaErr(err)
	}
	return &v1.DeviceInternalAIQuotaConsumeRes{Used: snap.Used, Limit: snap.Limit}, nil
}

func (c *DeviceAIQuotaInternalCtrl) DefaultGet(ctx context.Context, req *v1.DeviceInternalAIQuotaDefaultGetReq) (res *v1.DeviceInternalAIQuotaDefaultGetRes, err error) {
	_ = c
	_ = req
	dto, err := device.GetAIQuotaDefaultForAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceInternalAIQuotaDefaultGetRes{
		PolishMonthlyLimit:  dto.PolishMonthlyLimit,
		VoiceAiMonthlyLimit: dto.VoiceAiMonthlyLimit,
		UpdatedAt:           dto.UpdatedAt,
	}, nil
}

func (c *DeviceAIQuotaInternalCtrl) DefaultPut(ctx context.Context, req *v1.DeviceInternalAIQuotaDefaultPutReq) (res *v1.DeviceInternalAIQuotaDefaultPutRes, err error) {
	_ = c
	dto, err := device.UpdateAIQuotaDefaultForAdmin(ctx, req.PolishMonthlyLimit, req.VoiceAiMonthlyLimit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
	}
	return &v1.DeviceInternalAIQuotaDefaultPutRes{
		PolishMonthlyLimit:  dto.PolishMonthlyLimit,
		VoiceAiMonthlyLimit: dto.VoiceAiMonthlyLimit,
		UpdatedAt:           dto.UpdatedAt,
	}, nil
}

func (c *DeviceAIQuotaInternalCtrl) UserGet(ctx context.Context, req *v1.DeviceInternalAIQuotaUserGetReq) (res *v1.DeviceInternalAIQuotaUserGetRes, err error) {
	_ = c
	dto, err := device.GetAIQuotaUserOverrideForAdmin(ctx, req.WxId)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceInternalAIQuotaUserGetRes{
		WxId:                dto.WxId,
		PolishMonthlyLimit:  dto.PolishMonthlyLimit,
		VoiceAiMonthlyLimit: dto.VoiceAiMonthlyLimit,
		UpdatedAt:           dto.UpdatedAt,
	}, nil
}

func (c *DeviceAIQuotaInternalCtrl) UserPut(ctx context.Context, req *v1.DeviceInternalAIQuotaUserPutReq) (res *v1.DeviceInternalAIQuotaUserPutRes, err error) {
	_ = c
	polish := req.PolishMonthlyLimit
	voice := req.VoiceAiMonthlyLimit
	if req.ClearPolish {
		polish = nil
	}
	if req.ClearVoiceAi {
		voice = nil
	}
	dto, err := device.UpdateAIQuotaUserOverrideForAdmin(ctx, req.WxId, polish, voice)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
	}
	return &v1.DeviceInternalAIQuotaUserPutRes{
		WxId:                dto.WxId,
		PolishMonthlyLimit:  dto.PolishMonthlyLimit,
		VoiceAiMonthlyLimit: dto.VoiceAiMonthlyLimit,
		UpdatedAt:           dto.UpdatedAt,
	}, nil
}

func (c *DeviceAIQuotaInternalCtrl) WxIdByDeviceNo(ctx context.Context, req *v1.DeviceInternalWxIdByDeviceNoReq) (res *v1.DeviceInternalWxIdByDeviceNoRes, err error) {
	_ = c
	wxID, err := device.WxIDByDeviceNo(ctx, req.DeviceNo)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceInternalWxIdByDeviceNoRes{WxId: wxID}, nil
}
