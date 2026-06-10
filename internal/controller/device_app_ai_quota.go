package controller

import (
	"context"

	v1 "hello/api/v1"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/net/ghttp"
)

// DeviceAppAIQuotaCtrl App AI 额度读 API。
type DeviceAppAIQuotaCtrl struct{}

func NewDeviceAppAIQuotaCtrl() *DeviceAppAIQuotaCtrl { return &DeviceAppAIQuotaCtrl{} }

func (c *DeviceAppAIQuotaCtrl) Get(ctx context.Context, req *v1.DeviceAppAIQuotaGetReq) (res *v1.DeviceAppAIQuotaGetRes, err error) {
	_ = c
	_ = req
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	status, err := device.GetAIQuotaAppStatus(ctx, wxID)
	if err != nil {
		return nil, mapQuotaErr(err)
	}
	return &v1.DeviceAppAIQuotaGetRes{
		Polish:  v1.DeviceAppAIQuotaFeatureStatus{Used: status.Polish.Used, Limit: status.Polish.Limit},
		VoiceAi: v1.DeviceAppAIQuotaFeatureStatus{Used: status.VoiceAi.Used, Limit: status.VoiceAi.Limit},
	}, nil
}
