package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	ucgsvc "hello/internal/services/ucg"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// UcgAdminCtrl UCG Admin HTTP API（Header: X-Admin-Password）。
type UcgAdminCtrl struct{}

// NewUcgAdminCtrl 构造 UCG Admin 控制器。
func NewUcgAdminCtrl() *UcgAdminCtrl { return &UcgAdminCtrl{} }

func (c *UcgAdminCtrl) requireAdmin(ctx context.Context) error {
	_ = c
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return gerror.NewCode(gcode.CodeNotAuthorized, "口令错误")
	}
	if !ucgsvc.VerifyUcgAdminPassword(ctx, strings.TrimSpace(r.GetHeader("X-Admin-Password"))) {
		return gerror.NewCode(gcode.CodeNotAuthorized, "口令错误")
	}
	return nil
}

// AiConfigGet GET /ucg/admin/api/ai-config
func (c *UcgAdminCtrl) AiConfigGet(ctx context.Context, req *v1.UcgAdminAiConfigGetReq) (res *v1.UcgAdminAiConfigGetRes, err error) {
	_ = req
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	dto := ucgsvc.GetAIConfigForAdmin(ctx)
	return &v1.UcgAdminAiConfigGetRes{
		VisionModel:         dto.VisionModel,
		MaxImagesPerRequest: dto.MaxImagesPerRequest,
		UpdatedAt:           dto.UpdatedAt,
		UpdatedBy:           dto.UpdatedBy,
		AllowedModels:       ucgsvc.AllowedVisionModels,
	}, nil
}

// AiConfigPut PUT /ucg/admin/api/ai-config
func (c *UcgAdminCtrl) AiConfigPut(ctx context.Context, req *v1.UcgAdminAiConfigPutReq) (res *v1.UcgAdminAiConfigPutRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	updatedBy := strings.TrimSpace(req.UpdatedBy)
	if updatedBy == "" {
		updatedBy = "admin"
	}
	if err = ucgsvc.UpdateAIConfigForAdmin(ctx, req.VisionModel, req.MaxImagesPerRequest, updatedBy); err != nil {
		return nil, err
	}
	dto := ucgsvc.GetAIConfigForAdmin(ctx)
	return &v1.UcgAdminAiConfigPutRes{
		VisionModel:         dto.VisionModel,
		MaxImagesPerRequest: dto.MaxImagesPerRequest,
		UpdatedAt:           dto.UpdatedAt,
		UpdatedBy:           dto.UpdatedBy,
	}, nil
}
