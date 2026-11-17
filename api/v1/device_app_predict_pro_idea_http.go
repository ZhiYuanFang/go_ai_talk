package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeviceAppPredictProIdeaListReq 想法公开列表。
type DeviceAppPredictProIdeaListReq struct {
	g.Meta `path:"/device/app/api/predict-pro-ideas/list" method:"get" tags:"app" summary:"预测 Pro 想法列表"`
}

// DeviceAppPredictProIdeaItem App 想法列表项（不下发 wxId）。
type DeviceAppPredictProIdeaItem struct {
	Id        int64  `json:"id"`
	Text      string `json:"text"`
	BabyName  string `json:"babyName"`
	LikeCount int    `json:"likeCount"`
	LikedByMe bool   `json:"likedByMe"`
	IsMine    bool   `json:"isMine"`
	CreatedAt int64  `json:"createdAt"`
}

// DeviceAppPredictProIdeaListRes 想法列表。
type DeviceAppPredictProIdeaListRes struct {
	List []DeviceAppPredictProIdeaItem `json:"list"`
}

// DeviceAppPredictProIdeaSubmitReq 提交想法。
type DeviceAppPredictProIdeaSubmitReq struct {
	g.Meta `path:"/device/app/api/predict-pro-ideas/submit" method:"post" tags:"app" summary:"提交预测 Pro 想法"`
	Text   string `json:"text" dc:"想法正文，最多 200 字"`
}

// DeviceAppPredictProIdeaSubmitRes 提交成功返回新建项。
type DeviceAppPredictProIdeaSubmitRes struct {
	Id        int64  `json:"id"`
	Text      string `json:"text"`
	BabyName  string `json:"babyName"`
	LikeCount int    `json:"likeCount"`
	LikedByMe bool   `json:"likedByMe"`
	IsMine    bool   `json:"isMine"`
	CreatedAt int64  `json:"createdAt"`
}

// DeviceAppPredictProIdeaLikeReq 点赞。
type DeviceAppPredictProIdeaLikeReq struct {
	g.Meta `path:"/device/app/api/predict-pro-ideas/like" method:"post" tags:"app" summary:"点赞预测 Pro 想法"`
	Id     int64 `json:"id" v:"required" dc:"想法 id"`
}

// DeviceAppPredictProIdeaUnlikeReq 取消点赞。
type DeviceAppPredictProIdeaUnlikeReq struct {
	g.Meta `path:"/device/app/api/predict-pro-ideas/unlike" method:"post" tags:"app" summary:"取消点赞预测 Pro 想法"`
	Id     int64 `json:"id" v:"required" dc:"想法 id"`
}

// DeviceAppPredictProIdeaLikeRes 点赞/取消结果。
type DeviceAppPredictProIdeaLikeRes struct {
	LikeCount int  `json:"likeCount"`
	LikedByMe bool `json:"likedByMe"`
}

// DeviceAppPredictProIdeaDeleteReq 作者自删想法。
type DeviceAppPredictProIdeaDeleteReq struct {
	g.Meta `path:"/device/app/api/predict-pro-ideas/delete" method:"post" tags:"app" summary:"作者删除自己的预测 Pro 想法"`
	Id     int64 `json:"id" v:"required" dc:"想法 id"`
}

// DeviceAppPredictProIdeaDeleteRes 删除成功（空 data）。
type DeviceAppPredictProIdeaDeleteRes struct{}
