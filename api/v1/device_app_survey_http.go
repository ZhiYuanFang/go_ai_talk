package v1

import "github.com/gogf/gf/v2/frame/g"

// DeviceAppSurveyCurrentReq 读取当前满意度问卷是否需要展示。
// 业务：服务端判断开关、有效期与是否已提交；客户端另有喂养缓存门闩。
type DeviceAppSurveyCurrentReq struct {
	g.Meta `path:"/device/app/api/survey/current" method:"get" tags:"device" summary:"读取当前满意度问卷"`
}

// DeviceAppSurveyOption 问卷选项。
type DeviceAppSurveyOption struct {
	Id   int64  `json:"id" dc:"选项 id"`
	Text string `json:"text" dc:"选项文案"`
}

// DeviceAppSurveyCurrentRes 当前问卷。show=false 时其余字段可为空。
type DeviceAppSurveyCurrentRes struct {
	Show       bool                    `json:"show" dc:"是否需要弹出"`
	CampaignId int64                   `json:"campaignId" dc:"活动 id"`
	Intro      string                  `json:"intro" dc:"说明文案"`
	Options    []DeviceAppSurveyOption `json:"options" dc:"启用中的选项"`
}

// DeviceAppSurveySubmitReq 提交满意度问卷。
// optionResults 为「启用选项原文 + 未使用/不好用/好用」，可空；空时必须有意见。
type DeviceAppSurveySubmitReq struct {
	g.Meta        `path:"/device/app/api/survey/submit" method:"post" tags:"device" summary:"提交满意度问卷"`
	CampaignId    int64    `json:"campaignId" v:"required|min:1" dc:"活动 id"`
	OptionResults []string `json:"optionResults" dc:"选项原文加评价后缀，可空"`
	Comment       string   `json:"comment" dc:"更多意见，可空"`
}

// DeviceAppSurveySubmitRes 提交成功（含重复提交）。
type DeviceAppSurveySubmitRes struct{}
