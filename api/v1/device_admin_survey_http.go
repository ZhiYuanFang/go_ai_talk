package v1

import "github.com/gogf/gf/v2/frame/g"

// DeviceAdminSurveyEditorReq 读取问卷编辑态。campaignId=0 表示最新一轮。
type DeviceAdminSurveyEditorReq struct {
	g.Meta     `path:"/device/admin/api/survey/editor" method:"get" tags:"admin" summary:"满意度问卷编辑态"`
	CampaignId int64 `json:"campaignId" dc:"活动 id，0=最新"`
}

// DeviceAdminSurveyOption 选项行。
type DeviceAdminSurveyOption struct {
	Id      int64  `json:"id"`
	Text    string `json:"text"`
	Sort    int    `json:"sort"`
	Enabled bool   `json:"enabled"`
}

// DeviceAdminSurveyCampaignBrief 活动摘要。
type DeviceAdminSurveyCampaignBrief struct {
	Id      int64 `json:"id"`
	Enabled bool  `json:"enabled"`
	StartAt int64 `json:"startAt"`
	EndAt   int64 `json:"endAt"`
}

// DeviceAdminSurveyEditorRes 编辑态。
type DeviceAdminSurveyEditorRes struct {
	CampaignId int64                            `json:"campaignId"`
	Enabled    bool                             `json:"enabled"`
	StartAt    int64                            `json:"startAt"`
	EndAt      int64                            `json:"endAt"`
	Intro      string                           `json:"intro"`
	Options    []DeviceAdminSurveyOption        `json:"options"`
	Campaigns  []DeviceAdminSurveyCampaignBrief `json:"campaigns"`
	IsLatest   bool                             `json:"isLatest"`
}

// DeviceAdminSurveySaveReq 保存当前编辑的活动。campaignId=0 时新建。
type DeviceAdminSurveySaveReq struct {
	g.Meta     `path:"/device/admin/api/survey/save" method:"post" tags:"admin" summary:"保存满意度问卷"`
	CampaignId int64                     `json:"campaignId"`
	Enabled    bool                      `json:"enabled"`
	StartAt    int64                     `json:"startAt" dc:"开始 unix 秒，上海墙钟换算"`
	EndAt      int64                     `json:"endAt" dc:"结束 unix 秒"`
	Intro      string                    `json:"intro"`
	Options    []DeviceAdminSurveyOption `json:"options"`
}

// DeviceAdminSurveySaveRes 保存结果。
type DeviceAdminSurveySaveRes struct {
	CampaignId int64 `json:"campaignId"`
}

// DeviceAdminSurveyOptionAddReq 新增一行选项并立即落库。
type DeviceAdminSurveyOptionAddReq struct {
	g.Meta  `path:"/device/admin/api/survey/option/add" method:"post" tags:"admin" summary:"新增满意度问卷选项"`
	Text    string `json:"text"`
	Enabled bool   `json:"enabled"`
}

// DeviceAdminSurveyOptionAddRes 新增后的行。
type DeviceAdminSurveyOptionAddRes struct {
	Id      int64  `json:"id"`
	Text    string `json:"text"`
	Sort    int    `json:"sort"`
	Enabled bool   `json:"enabled"`
}

// DeviceAdminSurveyOptionUpdateReq 更新一行文案与启用。
type DeviceAdminSurveyOptionUpdateReq struct {
	g.Meta  `path:"/device/admin/api/survey/option/update" method:"post" tags:"admin" summary:"更新满意度问卷选项"`
	Id      int64  `json:"id" v:"required|min:1"`
	Text    string `json:"text"`
	Enabled bool   `json:"enabled"`
}

// DeviceAdminSurveyOptionUpdateRes 更新成功。
type DeviceAdminSurveyOptionUpdateRes struct{}

// DeviceAdminSurveyOptionDeleteReq 删除一行选项。
type DeviceAdminSurveyOptionDeleteReq struct {
	g.Meta `path:"/device/admin/api/survey/option/delete" method:"post" tags:"admin" summary:"删除满意度问卷选项"`
	Id     int64 `json:"id" v:"required|min:1"`
}

// DeviceAdminSurveyOptionDeleteRes 删除成功。
type DeviceAdminSurveyOptionDeleteRes struct{}

// DeviceAdminSurveyNewReq 已停用：问卷只有一份，调用会被拒绝。
type DeviceAdminSurveyNewReq struct {
	g.Meta `path:"/device/admin/api/survey/new" method:"post" tags:"admin" summary:"开启新一轮满意度问卷"`
}

// DeviceAdminSurveyNewRes 新活动 id。
type DeviceAdminSurveyNewRes struct {
	CampaignId int64 `json:"campaignId"`
}

// DeviceAdminSurveyResponsesReq 分页查看提交。
type DeviceAdminSurveyResponsesReq struct {
	g.Meta     `path:"/device/admin/api/survey/responses" method:"get" tags:"admin" summary:"满意度问卷提交列表"`
	CampaignId int64 `json:"campaignId" dc:"0=最新活动"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
}

// DeviceAdminSurveyResponseItem 一条提交。
type DeviceAdminSurveyResponseItem struct {
	Id          int64    `json:"id"`
	WxId        int64    `json:"wxId"`
	BabyName    string   `json:"babyName"`
	DeviceNo    string   `json:"deviceNo"`
	Options     []string `json:"options"`
	Comment     string   `json:"comment"`
	SubmittedAt int64    `json:"submittedAt"`
	Valid       bool     `json:"valid"`
}

// DeviceAdminSurveyResponseValidReq 修改一条提交是否有效。
type DeviceAdminSurveyResponseValidReq struct {
	g.Meta `path:"/device/admin/api/survey/response/valid" method:"post" tags:"admin" summary:"设置满意度问卷提交是否有效"`
	Id     int64 `json:"id" v:"required|min:1"`
	Valid  bool  `json:"valid"`
}

// DeviceAdminSurveyResponseValidRes 设置成功。
type DeviceAdminSurveyResponseValidRes struct{}

// DeviceAdminSurveyResponsesRes 提交分页。
type DeviceAdminSurveyResponsesRes struct {
	List       []DeviceAdminSurveyResponseItem `json:"list"`
	Total      int                             `json:"total"`
	Page       int                             `json:"page"`
	PageSize   int                             `json:"pageSize"`
	CampaignId int64                           `json:"campaignId"`
}
