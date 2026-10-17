package entity

// SurveyCampaign 满意度问卷活动。
type SurveyCampaign struct {
	Id        int64  `json:"id"`
	Enabled   int    `json:"enabled"`
	StartAt   int64  `json:"startAt"`
	EndAt     int64  `json:"endAt"`
	Intro     string `json:"intro"`
	CreatedAt int64  `json:"createdAt"`
}

// SurveyOption 满意度问卷选项行。
type SurveyOption struct {
	Id         int64  `json:"id"`
	CampaignId int64  `json:"campaignId"`
	Text       string `json:"text"`
	Sort       int    `json:"sort"`
	Enabled    int    `json:"enabled"`
}

// SurveyResponse 满意度问卷提交行。
type SurveyResponse struct {
	Id             int64  `json:"id"`
	CampaignId     int64  `json:"campaignId"`
	WxId           int64  `json:"wxId"`
	DeviceNo       string `json:"deviceNo"`
	BabyName       string `json:"babyName"`
	OptionSnapshot string `json:"optionSnapshot"`
	Comment        string `json:"comment"`
	SubmittedAt    int64  `json:"submittedAt"`
	Valid          int    `json:"valid"`
}
