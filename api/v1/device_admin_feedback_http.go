package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeviceAdminFeedbackListReq Admin 反馈分页列表。
type DeviceAdminFeedbackListReq struct {
	g.Meta        `path:"/device/admin/api/feedback/list" method:"get" tags:"admin" summary:"反馈分页列表"`
	Page          int  `json:"page" p:"page" dc:"页码，从 1 开始"`
	PageSize      int  `json:"pageSize" p:"pageSize" dc:"每页条数，默认 20，最大 100"`
	UnrepliedOnly bool `json:"unrepliedOnly" p:"unrepliedOnly" dc:"仅未回复"`
}

// DeviceAdminFeedbackItem Admin 反馈列表项。
type DeviceAdminFeedbackItem struct {
	Id            int64  `json:"id"`
	WxId          int64  `json:"wxId"`
	Question      string `json:"question"`
	OfficialReply string `json:"officialReply,omitempty"`
	Status        int    `json:"status"`
	CreatedAt     int64  `json:"createdAt"`
	RepliedAt     *int64 `json:"repliedAt,omitempty"`
}

// DeviceAdminFeedbackListRes Admin 反馈分页响应。
type DeviceAdminFeedbackListRes struct {
	List     []DeviceAdminFeedbackItem `json:"list"`
	Total    int                       `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"pageSize"`
}

// DeviceAdminFeedbackReplyReq Admin 官方回复（每条仅一次）。
type DeviceAdminFeedbackReplyReq struct {
	g.Meta        `path:"/device/admin/api/feedback/reply" method:"post" tags:"admin" summary:"官方回复反馈"`
	Id            int64  `json:"id" dc:"反馈 id"`
	OfficialReply string `json:"officialReply" dc:"官方回复，最多 2000 字"`
}

// DeviceAdminFeedbackReplyRes 回复成功。
type DeviceAdminFeedbackReplyRes struct{}
