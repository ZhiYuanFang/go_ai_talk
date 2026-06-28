package controller

import (
	"context"

	v1 "hello/api/v1"
	"hello/internal/services/appstatus"

	"github.com/gogf/gf/v2/net/ghttp"
)

// AppStatusBannerCtrl App 无鉴权 banner 读 API。
type AppStatusBannerCtrl struct{}

func NewAppStatusBannerCtrl() *AppStatusBannerCtrl { return &AppStatusBannerCtrl{} }

func bannerToPublicRes(s appstatus.BannerState) *v1.AppStatusBannerGetRes {
	if !s.Active {
		return &v1.AppStatusBannerGetRes{Active: false}
	}
	return &v1.AppStatusBannerGetRes{
		Active:        true,
		Title:         s.Title,
		Message:       s.Message,
		ExpectedEndAt: s.ExpectedEndAt,
		Dismissible:   s.Dismissible,
		UpdatedAt:     s.UpdatedAt,
		ContentKey:    appstatus.ContentKey(s.Title, s.Message),
	}
}

func bannerToAdminRes(s appstatus.BannerState) *v1.AppStatusAdminBannerGetRes {
	return &v1.AppStatusAdminBannerGetRes{
		Active:        s.Active,
		Title:         s.Title,
		Message:       s.Message,
		ExpectedEndAt: s.ExpectedEndAt,
		Dismissible:   s.Dismissible,
		UpdatedAt:     s.UpdatedAt,
		ContentKey:    appstatus.ContentKey(s.Title, s.Message),
	}
}

// BannerGet GET /app/api/status/banner
func (c *AppStatusBannerCtrl) BannerGet(ctx context.Context, req *v1.AppStatusBannerGetReq) (res *v1.AppStatusBannerGetRes, err error) {
	_ = c
	_ = req
	if r := ghttp.RequestFromCtx(ctx); r != nil {
		r.Response.Header().Set("Cache-Control", "public, max-age=30")
	}
	return bannerToPublicRes(appstatus.Snapshot()), nil
}
