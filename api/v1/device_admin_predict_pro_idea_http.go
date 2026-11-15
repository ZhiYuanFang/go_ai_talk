package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeviceAdminPredictProIdeaListReq Admin 想法分页。
type DeviceAdminPredictProIdeaListReq struct {
	g.Meta   `path:"/device/admin/api/predict-pro-ideas/list" method:"get" tags:"admin" summary:"预测 Pro 想法分页列表"`
	Page     int `json:"page" p:"page" dc:"页码，从 1 开始"`
	PageSize int `json:"pageSize" p:"pageSize" dc:"每页条数，默认 20，最大 100"`
}

// DeviceAdminPredictProIdeaItem Admin 列表项。
type DeviceAdminPredictProIdeaItem struct {
	Id          int64  `json:"id"`
	WxId        int64  `json:"wxId"`
	BabyName    string `json:"babyName"`
	Content     string `json:"content"`
	DisplayText string `json:"displayText"`
	LikeCount   int    `json:"likeCount"`
	CreatedAt   int64  `json:"createdAt"`
}

// DeviceAdminPredictProIdeaListRes Admin 分页响应。
type DeviceAdminPredictProIdeaListRes struct {
	List     []DeviceAdminPredictProIdeaItem `json:"list"`
	Total    int                             `json:"total"`
	Page     int                             `json:"page"`
	PageSize int                             `json:"pageSize"`
}

// DeviceAdminPredictProIdeaUpdateReq Admin 改展示文案。
type DeviceAdminPredictProIdeaUpdateReq struct {
	g.Meta      `path:"/device/admin/api/predict-pro-ideas/update" method:"post" tags:"admin" summary:"修改预测 Pro 想法展示文案"`
	Id          int64  `json:"id" v:"required"`
	DisplayText string `json:"displayText" dc:"展示文案，最多 200 字"`
}

// DeviceAdminPredictProIdeaUpdateRes 更新成功。
type DeviceAdminPredictProIdeaUpdateRes struct{}

// DeviceAdminPredictProIdeaDeleteReq Admin 软删。
type DeviceAdminPredictProIdeaDeleteReq struct {
	g.Meta `path:"/device/admin/api/predict-pro-ideas/delete" method:"post" tags:"admin" summary:"删除预测 Pro 想法"`
	Id     int64 `json:"id" v:"required"`
}

// DeviceAdminPredictProIdeaDeleteRes 删除成功。
type DeviceAdminPredictProIdeaDeleteRes struct{}
