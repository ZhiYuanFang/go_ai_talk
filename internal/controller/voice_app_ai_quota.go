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

// VoiceAppAIQuotaCtrl App 读 voice 域 AI 额度。
type VoiceAppAIQuotaCtrl struct{}

func NewVoiceAppAIQuotaCtrl() *VoiceAppAIQuotaCtrl { return &VoiceAppAIQuotaCtrl{} }

func (c *VoiceAppAIQuotaCtrl) Get(ctx context.Context, req *v1.VoiceAppAIQuotaGetReq) (res *v1.VoiceAppAIQuotaGetRes, err error) {
	_ = c
	_ = req
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	status, err := voice.GetVoiceAIQuotaAppStatus(ctx, wxID)
	if err != nil {
		return nil, mapAIQuotaErr(err)
	}
	return &v1.VoiceAppAIQuotaGetRes{
		VoiceAi: v1.VoiceAppAIQuotaFeatureStatus{
			Used:     status.VoiceAi.Used,
			Limit:    status.VoiceAi.Limit,
			Degraded: status.VoiceAi.Degraded,
		},
		ClinicAi: v1.VoiceAppAIQuotaFeatureStatus{
			Used:     status.ClinicAi.Used,
			Limit:    status.ClinicAi.Limit,
			Degraded: status.ClinicAi.Degraded,
		},
	}, nil
}

// VoiceAdminAIQuotaCtrl voice 域 Admin AI 额度配置。
type VoiceAdminAIQuotaCtrl struct{}

func NewVoiceAdminAIQuotaCtrl() *VoiceAdminAIQuotaCtrl { return &VoiceAdminAIQuotaCtrl{} }

func (c *VoiceAdminAIQuotaCtrl) requireAdmin(ctx context.Context) error {
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

func (c *VoiceAdminAIQuotaCtrl) DefaultGet(ctx context.Context, req *v1.VoiceAdminAIQuotaDefaultGetReq) (res *v1.VoiceAdminAIQuotaDefaultGetRes, err error) {
	_ = req
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	dto, err := voice.GetVoiceAIQuotaDefaultForAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.VoiceAdminAIQuotaDefaultGetRes{
		VoiceAiMonthlyLimit:  dto.VoiceAiMonthlyLimit,
		ClinicAiMonthlyLimit: dto.ClinicAiMonthlyLimit,
		UpdatedAt:            dto.UpdatedAt,
	}, nil
}

func (c *VoiceAdminAIQuotaCtrl) DefaultPut(ctx context.Context, req *v1.VoiceAdminAIQuotaDefaultPutReq) (res *v1.VoiceAdminAIQuotaDefaultPutRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	dto, err := voice.UpdateVoiceAIQuotaDefaultForAdmin(ctx, req.VoiceAiMonthlyLimit, req.ClinicAiMonthlyLimit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
	}
	return &v1.VoiceAdminAIQuotaDefaultPutRes{
		VoiceAiMonthlyLimit:  dto.VoiceAiMonthlyLimit,
		ClinicAiMonthlyLimit: dto.ClinicAiMonthlyLimit,
		UpdatedAt:            dto.UpdatedAt,
	}, nil
}

func (c *VoiceAdminAIQuotaCtrl) UserGet(ctx context.Context, req *v1.VoiceAdminAIQuotaUserGetReq) (res *v1.VoiceAdminAIQuotaUserGetRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	dto, err := voice.GetVoiceAIQuotaUserOverrideForAdmin(ctx, req.WxId)
	if err != nil {
		return nil, err
	}
	return &v1.VoiceAdminAIQuotaUserGetRes{
		WxId:                 dto.WxId,
		VoiceAiMonthlyLimit:  dto.VoiceAiMonthlyLimit,
		ClinicAiMonthlyLimit: dto.ClinicAiMonthlyLimit,
		UpdatedAt:            dto.UpdatedAt,
	}, nil
}

func (c *VoiceAdminAIQuotaCtrl) UserPut(ctx context.Context, req *v1.VoiceAdminAIQuotaUserPutReq) (res *v1.VoiceAdminAIQuotaUserPutRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	voiceLimit := req.VoiceAiMonthlyLimit
	clinicLimit := req.ClinicAiMonthlyLimit
	if req.ClearVoiceAi {
		voiceLimit = nil
	}
	if req.ClearClinicAi {
		clinicLimit = nil
	}
	dto, err := voice.UpdateVoiceAIQuotaUserOverrideForAdmin(ctx, req.WxId, voiceLimit, clinicLimit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
	}
	return &v1.VoiceAdminAIQuotaUserPutRes{
		WxId:                 dto.WxId,
		VoiceAiMonthlyLimit:  dto.VoiceAiMonthlyLimit,
		ClinicAiMonthlyLimit: dto.ClinicAiMonthlyLimit,
		UpdatedAt:            dto.UpdatedAt,
	}, nil
}
