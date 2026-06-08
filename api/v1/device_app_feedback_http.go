package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeviceAppFeedbackListReq 当前用户反馈历史列表。
type DeviceAppFeedbackListReq struct {
	g.Meta `path:"/device/app/api/feedback/list" method:"get" tags:"app" summary:"反馈列表"`
}

// DeviceAppFeedbackItem App 反馈列表项。
type DeviceAppFeedbackItem struct {
	Id            int64  `json:"id"`
	Question      string `json:"question"`
	OfficialReply string `json:"officialReply,omitempty"`
	Status        int    `json:"status"`
	CreatedAt     int64  `json:"createdAt"`
	RepliedAt     *int64 `json:"repliedAt,omitempty"`
}

// DeviceAppFeedbackListRes 反馈列表响应。
type DeviceAppFeedbackListRes struct {
	List []DeviceAppFeedbackItem `json:"list"`
}

// DeviceAppFeedbackSubmitReq 提交反馈。
type DeviceAppFeedbackSubmitReq struct {
	g.Meta   `path:"/device/app/api/feedback/submit" method:"post" tags:"app" summary:"提交反馈"`
	Question string `json:"question" dc:"问题正文，最多 2000 字"`
}

// DeviceAppFeedbackSubmitRes 提交成功返回新建记录。
type DeviceAppFeedbackSubmitRes struct {
	Id            int64  `json:"id"`
	Question      string `json:"question"`
	OfficialReply string `json:"officialReply,omitempty"`
	Status        int    `json:"status"`
	CreatedAt     int64  `json:"createdAt"`
}
