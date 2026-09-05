package voicectrl

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
		CareAlert: v1.VoiceAppAIQuotaFeatureStatus{
			Used:     status.CareAlert.Used,
			Limit:    status.CareAlert.Limit,
			Degraded: status.CareAlert.Degraded,
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
		VoiceAiMonthlyLimit:   dto.VoiceAiMonthlyLimit,
		ClinicAiMonthlyLimit:  dto.ClinicAiMonthlyLimit,
		CareAlertMonthlyLimit: dto.CareAlertMonthlyLimit,
		UpdatedAt:             dto.UpdatedAt,
	}, nil
}

func (c *VoiceAdminAIQuotaCtrl) DefaultPut(ctx context.Context, req *v1.VoiceAdminAIQuotaDefaultPutReq) (res *v1.VoiceAdminAIQuotaDefaultPutRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	dto, err := voice.UpdateVoiceAIQuotaDefaultForAdmin(ctx, req.VoiceAiMonthlyLimit, req.ClinicAiMonthlyLimit, req.CareAlertMonthlyLimit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
	}
	return &v1.VoiceAdminAIQuotaDefaultPutRes{
		VoiceAiMonthlyLimit:   dto.VoiceAiMonthlyLimit,
		ClinicAiMonthlyLimit:  dto.ClinicAiMonthlyLimit,
		CareAlertMonthlyLimit: dto.CareAlertMonthlyLimit,
		UpdatedAt:             dto.UpdatedAt,
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
		WxId:                  dto.WxId,
		VoiceAiMonthlyLimit:   dto.VoiceAiMonthlyLimit,
		ClinicAiMonthlyLimit:  dto.ClinicAiMonthlyLimit,
		CareAlertMonthlyLimit: dto.CareAlertMonthlyLimit,
		UpdatedAt:             dto.UpdatedAt,
	}, nil
}

func (c *VoiceAdminAIQuotaCtrl) UserPut(ctx context.Context, req *v1.VoiceAdminAIQuotaUserPutReq) (res *v1.VoiceAdminAIQuotaUserPutRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	voiceLimit := req.VoiceAiMonthlyLimit
	clinicLimit := req.ClinicAiMonthlyLimit
	careLimit := req.CareAlertMonthlyLimit
	if req.ClearVoiceAi {
		voiceLimit = nil
	}
	if req.ClearClinicAi {
		clinicLimit = nil
	}
	if req.ClearCareAlert {
		careLimit = nil
	}
	dto, err := voice.UpdateVoiceAIQuotaUserOverrideForAdmin(ctx, req.WxId, voiceLimit, clinicLimit, careLimit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
	}
	return &v1.VoiceAdminAIQuotaUserPutRes{
		WxId:                  dto.WxId,
		VoiceAiMonthlyLimit:   dto.VoiceAiMonthlyLimit,
		ClinicAiMonthlyLimit:  dto.ClinicAiMonthlyLimit,
		CareAlertMonthlyLimit: dto.CareAlertMonthlyLimit,
		UpdatedAt:             dto.UpdatedAt,
	}, nil
}

// UsersGet 分页列出全部真实 wx 的有效额度与身份字段。
func (c *VoiceAdminAIQuotaCtrl) UsersGet(ctx context.Context, req *v1.VoiceAdminAIQuotaUsersGetReq) (res *v1.VoiceAdminAIQuotaUsersGetRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := voice.ListVoiceAIQuotaUsersForAdmin(ctx, req.Page, req.PageSize, req.DeviceNo)
	if err != nil {
		return nil, err
	}
	list := make([]v1.VoiceAdminAIQuotaUsersItem, 0, len(result.List))
	for _, it := range result.List {
		list = append(list, v1.VoiceAdminAIQuotaUsersItem{
			DeviceNo:  it.DeviceNo,
			WxId:      it.WxId,
			Account:   it.Account,
			BabyName:  it.BabyName,
			VoiceAi:   v1.VoiceAdminAIQuotaUsersFeature{Used: it.VoiceAi.Used, Limit: it.VoiceAi.Limit},
			ClinicAi:  v1.VoiceAdminAIQuotaUsersFeature{Used: it.ClinicAi.Used, Limit: it.ClinicAi.Limit},
			CareAlert: v1.VoiceAdminAIQuotaUsersFeature{Used: it.CareAlert.Used, Limit: it.CareAlert.Limit},
		})
	}
	return &v1.VoiceAdminAIQuotaUsersGetRes{
		List:     list,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}
