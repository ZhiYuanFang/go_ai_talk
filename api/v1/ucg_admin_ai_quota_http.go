package v1

import "github.com/gogf/gf/v2/frame/g"

// UcgAdminAIQuotaDefaultGetReq UCG 管理读取润笔额度全局默认。
type UcgAdminAIQuotaDefaultGetReq struct {
	g.Meta `path:"/ucg/admin/api/ai-quota/default" method:"get" tags:"ucg-admin" summary:"读取润笔额度全局默认"`
}

type UcgAdminAIQuotaDefaultGetRes struct {
	PolishMonthlyLimit int   `json:"polishMonthlyLimit"`
	UpdatedAt          int64 `json:"updatedAt"`
}

type UcgAdminAIQuotaDefaultPutReq struct {
	g.Meta             `path:"/ucg/admin/api/ai-quota/default" method:"put" tags:"ucg-admin" summary:"更新润笔额度全局默认"`
	PolishMonthlyLimit int `json:"polishMonthlyLimit" v:"required|min:1"`
}

type UcgAdminAIQuotaDefaultPutRes = UcgAdminAIQuotaDefaultGetRes

type UcgAdminAIQuotaUserGetReq struct {
	g.Meta `path:"/ucg/admin/api/ai-quota/user" method:"get" tags:"ucg-admin" summary:"读取 wxId 润笔额度 override"`
	WxId   int64 `json:"wxId" p:"wxId" v:"required|min:1"`
}

type UcgAdminAIQuotaUserGetRes struct {
	WxId               int64 `json:"wxId"`
	PolishMonthlyLimit *int  `json:"polishMonthlyLimit,omitempty"`
	UpdatedAt          int64 `json:"updatedAt"`
}

type UcgAdminAIQuotaUserPutReq struct {
	g.Meta             `path:"/ucg/admin/api/ai-quota/user" method:"put" tags:"ucg-admin" summary:"更新 wxId 润笔额度 override"`
	WxId               int64 `json:"wxId" v:"required|min:1"`
	PolishMonthlyLimit *int  `json:"polishMonthlyLimit"`
	ClearPolish        bool  `json:"clearPolish"`
}

type UcgAdminAIQuotaUserPutRes = UcgAdminAIQuotaUserGetRes
