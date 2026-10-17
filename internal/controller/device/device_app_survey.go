package devicectrl

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/platform/httpmeta"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/net/ghttp"
)

// DeviceAppSurveyCtrl App 满意度问卷（须登录；身份只取网关注入头）。
type DeviceAppSurveyCtrl struct{}

// NewDeviceAppSurveyCtrl 构造问卷 App 控制器。
func NewDeviceAppSurveyCtrl() *DeviceAppSurveyCtrl {
	return &DeviceAppSurveyCtrl{}
}

// Current GET /device/app/api/survey/current
// 业务：返回最新活动是否对该 wx 展示。
func (c *DeviceAppSurveyCtrl) Current(ctx context.Context, _ *v1.DeviceAppSurveyCurrentReq) (res *v1.DeviceAppSurveyCurrentRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	cur, err := device.SurveyCurrentForWx(ctx, wxID)
	if err != nil {
		return nil, err
	}
	opts := make([]v1.DeviceAppSurveyOption, 0, len(cur.Options))
	for _, opt := range cur.Options {
		opts = append(opts, v1.DeviceAppSurveyOption{Id: opt.Id, Text: opt.Text})
	}
	return &v1.DeviceAppSurveyCurrentRes{
		Show: cur.Show, CampaignId: cur.CampaignId, Intro: cur.Intro, Options: opts,
	}, nil
}

// Submit POST /device/app/api/survey/submit
// 业务：双空拒绝；成功后该活动不再对该用户展示。设备号只信内部头。
func (c *DeviceAppSurveyCtrl) Submit(ctx context.Context, req *v1.DeviceAppSurveySubmitReq) (res *v1.DeviceAppSurveySubmitRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	deviceNo := ""
	if r != nil {
		deviceNo = strings.TrimSpace(r.GetHeader(httpmeta.HeaderInternalDeviceNo))
	}
	if err := device.SurveySubmit(ctx, wxID, deviceNo, req.CampaignId, req.OptionResults, req.Comment); err != nil {
		return nil, err
	}
	return &v1.DeviceAppSurveySubmitRes{}, nil
}
