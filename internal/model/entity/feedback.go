package entity

import "github.com/gogf/gf/v2/os/gtime"

// Feedback is the golang structure for table feedback.
type Feedback struct {
	Id            int64       `json:"id"            `
	WxId          int64       `json:"wxId"          `
	Question      string      `json:"question"      `
	OfficialReply string      `json:"officialReply" `
	Status        int         `json:"status"        `
	CreatedAt     *gtime.Time `json:"createdAt"     `
	UpdatedAt     *gtime.Time `json:"updatedAt"     `
	RepliedAt     *gtime.Time `json:"repliedAt"     `
}
