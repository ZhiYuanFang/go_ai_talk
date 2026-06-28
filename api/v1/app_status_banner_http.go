package v1

import "github.com/gogf/gf/v2/frame/g"

// AppStatusBannerGetReq App 无鉴权读取维护通知。
type AppStatusBannerGetReq struct {
	g.Meta `path:"/app/api/status/banner" method:"get" tags:"app-status" summary:"读取 App 维护/公告通知（无鉴权）"`
}

// AppStatusBannerGetRes 维护通知 payload；active=false 时仅 active 字段有意义。
type AppStatusBannerGetRes struct {
	Active        bool   `json:"active"`
	Title         string `json:"title,omitempty"`
	Message       string `json:"message,omitempty"`
	ExpectedEndAt *int64 `json:"expectedEndAt,omitempty"`
	Dismissible   bool   `json:"dismissible,omitempty"`
	UpdatedAt     int64  `json:"updatedAt,omitempty"`
	ContentKey    string `json:"contentKey,omitempty"`
}
