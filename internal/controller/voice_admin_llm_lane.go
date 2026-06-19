package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	voice "hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// VoiceAdminLLMLanesCtrl voice 域 LLM lane Admin API。
type VoiceAdminLLMLanesCtrl struct{}

// NewVoiceAdminLLMLanesCtrl 构造控制器。
func NewVoiceAdminLLMLanesCtrl() *VoiceAdminLLMLanesCtrl { return &VoiceAdminLLMLanesCtrl{} }

func (c *VoiceAdminLLMLanesCtrl) requireAdmin(ctx context.Context) error {
	_ = c
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return gerror.NewCode(gcode.CodeNotAuthorized, "口令错误")
	}
	if !voice.VerifyVoiceAdminPassword(ctx, strings.TrimSpace(r.GetHeader("X-Admin-Password"))) {
		return gerror.NewCode(gcode.CodeNotAuthorized, "口令错误")
	}
	return nil
}

func (c *VoiceAdminLLMLanesCtrl) Get(ctx context.Context, req *v1.VoiceAdminLLMLanesGetReq) (res *v1.VoiceAdminLLMLanesGetRes, err error) {
	_ = req
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	dto, err := voice.GetLLMLanesForAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return mapLLMLanesRes(dto), nil
}

func (c *VoiceAdminLLMLanesCtrl) Put(ctx context.Context, req *v1.VoiceAdminLLMLanesPutReq) (res *v1.VoiceAdminLLMLanesPutRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	updatedBy := strings.TrimSpace(req.UpdatedBy)
	if updatedBy == "" {
		updatedBy = "admin"
	}
	if err = voice.UpdateLLMLanesForAdmin(ctx, req.VoiceUnderstanding.ToLaneDTO(), req.Clinic.ToLaneDTO(), updatedBy); err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
	}
	dto, err := voice.GetLLMLanesForAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return mapLLMLanesRes(dto), nil
}

func mapLLMLanesRes(dto voice.LLMLanesAdminDTO) *v1.VoiceAdminLLMLanesGetRes {
	return &v1.VoiceAdminLLMLanesGetRes{
		VoiceUnderstanding: v1.VoiceAdminLLMLaneItem{
			Provider:    dto.VoiceUnderstanding.Provider,
			Model:       dto.VoiceUnderstanding.Model,
			MaxInFlight: dto.VoiceUnderstanding.MaxInFlight,
			MaxWaiters:  dto.VoiceUnderstanding.MaxWaiters,
			UpdatedAt:   dto.VoiceUnderstanding.UpdatedAt,
			UpdatedBy:   dto.VoiceUnderstanding.UpdatedBy,
		},
		Clinic: v1.VoiceAdminLLMLaneItem{
			Provider:    dto.Clinic.Provider,
			Model:       dto.Clinic.Model,
			MaxInFlight: dto.Clinic.MaxInFlight,
			MaxWaiters:  dto.Clinic.MaxWaiters,
			UpdatedAt:   dto.Clinic.UpdatedAt,
			UpdatedBy:   dto.Clinic.UpdatedBy,
		},
		Allowlist: dto.Allowlist,
	}
}
