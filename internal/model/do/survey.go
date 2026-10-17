package do

import "github.com/gogf/gf/v2/frame/g"

// SurveyCampaign 活动表 DO。
type SurveyCampaign struct {
	g.Meta    `orm:"table:survey_campaign, do:true"`
	Id        interface{}
	Enabled   interface{}
	StartAt   interface{}
	EndAt     interface{}
	Intro     interface{}
	CreatedAt interface{}
}

// SurveyOption 选项表 DO。
type SurveyOption struct {
	g.Meta     `orm:"table:survey_option, do:true"`
	Id         interface{}
	CampaignId interface{}
	Text       interface{}
	Sort       interface{}
	Enabled    interface{}
}

// SurveyResponse 提交表 DO。
type SurveyResponse struct {
	g.Meta         `orm:"table:survey_response, do:true"`
	Id             interface{}
	CampaignId     interface{}
	WxId           interface{}
	DeviceNo       interface{}
	BabyName       interface{}
	OptionSnapshot interface{}
	Comment        interface{}
	SubmittedAt    interface{}
}
