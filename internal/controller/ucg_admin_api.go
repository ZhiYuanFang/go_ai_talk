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

// PostsList GET /ucg/admin/api/posts/list
func (c *UcgAdminCtrl) PostsList(ctx context.Context, req *v1.UcgAdminPostsListReq) (res *v1.UcgAdminPostsListRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := ucgsvc.ListPostsForAdmin(ctx, req.Page, req.PageSize, req.Status)
	if err != nil {
		return nil, err
	}
	items := make([]v1.UcgAdminPostItem, 0)
	if dtos, ok := result.List.([]*ucgsvc.PostDTO); ok {
		items = make([]v1.UcgAdminPostItem, 0, len(dtos))
		for _, dto := range dtos {
			items = append(items, mapAdminPostItem(dto))
		}
	}
	return &v1.UcgAdminPostsListRes{
		List:     items,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}

// PostsReject POST /ucg/admin/api/posts/reject
func (c *UcgAdminCtrl) PostsReject(ctx context.Context, req *v1.UcgAdminPostsRejectReq) (res *v1.UcgAdminPostsRejectRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := ucgsvc.AdminBatchRejectPosts(ctx, req.PostIds, req.Reason)
	if err != nil {
		return nil, err
	}
	return &v1.UcgAdminPostsRejectRes{
		Rejected: result.Rejected,
		Skipped:  result.Skipped,
		Failed:   result.Failed,
	}, nil
}

// AIQuotaDefaultGet GET /ucg/admin/api/ai-quota/default
func (c *UcgAdminCtrl) AIQuotaDefaultGet(ctx context.Context, req *v1.UcgAdminAIQuotaDefaultGetReq) (res *v1.UcgAdminAIQuotaDefaultGetRes, err error) {
	_ = req
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	dto, err := ucgsvc.GetAIQuotaDefaultAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.UcgAdminAIQuotaDefaultGetRes{
		PolishMonthlyLimit:  dto.PolishMonthlyLimit,
		VoiceAiMonthlyLimit: dto.VoiceAiMonthlyLimit,
		UpdatedAt:           dto.UpdatedAt,
	}, nil
}

// AIQuotaDefaultPut PUT /ucg/admin/api/ai-quota/default
func (c *UcgAdminCtrl) AIQuotaDefaultPut(ctx context.Context, req *v1.UcgAdminAIQuotaDefaultPutReq) (res *v1.UcgAdminAIQuotaDefaultPutRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	dto, err := ucgsvc.UpdateAIQuotaDefaultAdmin(ctx, req.PolishMonthlyLimit, req.VoiceAiMonthlyLimit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
	}
	return &v1.UcgAdminAIQuotaDefaultPutRes{
		PolishMonthlyLimit:  dto.PolishMonthlyLimit,
		VoiceAiMonthlyLimit: dto.VoiceAiMonthlyLimit,
		UpdatedAt:           dto.UpdatedAt,
	}, nil
}

// AIQuotaUserGet GET /ucg/admin/api/ai-quota/user
func (c *UcgAdminCtrl) AIQuotaUserGet(ctx context.Context, req *v1.UcgAdminAIQuotaUserGetReq) (res *v1.UcgAdminAIQuotaUserGetRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	dto, err := ucgsvc.GetAIQuotaUserOverrideAdmin(ctx, req.WxId)
	if err != nil {
		return nil, err
	}
	return &v1.UcgAdminAIQuotaUserGetRes{
		WxId:                dto.WxId,
		PolishMonthlyLimit:  dto.PolishMonthlyLimit,
		VoiceAiMonthlyLimit: dto.VoiceAiMonthlyLimit,
		UpdatedAt:           dto.UpdatedAt,
	}, nil
}

// AIQuotaUserPut PUT /ucg/admin/api/ai-quota/user
func (c *UcgAdminCtrl) AIQuotaUserPut(ctx context.Context, req *v1.UcgAdminAIQuotaUserPutReq) (res *v1.UcgAdminAIQuotaUserPutRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	dto, err := ucgsvc.UpdateAIQuotaUserOverrideAdmin(ctx, req.WxId, req.PolishMonthlyLimit, req.VoiceAiMonthlyLimit, req.ClearPolish, req.ClearVoiceAi)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, err.Error())
	}
	return &v1.UcgAdminAIQuotaUserPutRes{
		WxId:                dto.WxId,
		PolishMonthlyLimit:  dto.PolishMonthlyLimit,
		VoiceAiMonthlyLimit: dto.VoiceAiMonthlyLimit,
		UpdatedAt:           dto.UpdatedAt,
	}, nil
}

func mapAdminPostItem(dto *ucgsvc.PostDTO) v1.UcgAdminPostItem {
	if dto == nil {
		return v1.UcgAdminPostItem{}
	}
	item := v1.UcgAdminPostItem{
		Id:           dto.Id,
		AuthorWxId:   dto.AuthorWxId,
		Content:      dto.Content,
		Status:       dto.Status,
		RejectReason: dto.RejectReason,
		CreatedAt:    dto.CreatedAt,
		UpdatedAt:    dto.UpdatedAt,
		PublishedAt:  dto.PublishedAt,
	}
	if len(dto.Media) > 0 {
		item.Media = make([]v1.UcgAdminPostMediaItem, 0, len(dto.Media))
		for _, m := range dto.Media {
			item.Media = append(item.Media, v1.UcgAdminPostMediaItem{
				ObjectKey:    m.ObjectKey,
				CdnUrl:       m.CdnUrl,
				ThumbnailUrl: m.ThumbnailUrl,
				MediaKind:    m.MediaKind,
			})
		}
	}
	if dto.Author != nil {
		item.Author = &v1.UcgAdminPostAuthor{Nickname: dto.Author.Nickname}
	}
	return item
}

// ProfileAuditJobsList GET /ucg/admin/api/profile-audit-jobs/list
func (c *UcgAdminCtrl) ProfileAuditJobsList(ctx context.Context, req *v1.UcgAdminProfileAuditJobsListReq) (res *v1.UcgAdminProfileAuditJobsListRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := ucgsvc.ListProfileAuditJobsForAdmin(ctx, req.Page, req.PageSize, req.Status)
	if err != nil {
		return nil, err
	}
	items := make([]v1.UcgAdminProfileAuditJobItem, 0)
	if rows, ok := result.List.([]ucgsvc.ProfileAuditJobAdminItem); ok {
		items = make([]v1.UcgAdminProfileAuditJobItem, 0, len(rows))
		for _, row := range rows {
			items = append(items, v1.UcgAdminProfileAuditJobItem{
				JobId:        row.JobId,
				WxId:         row.WxId,
				AuditVersion: row.AuditVersion,
				Status:       row.Status,
				Nickname:     row.Nickname,
				AvatarKey:    row.AvatarKey,
				Bio:          row.Bio,
				RejectReason: row.RejectReason,
				CreatedAt:    row.CreatedAt,
				UpdatedAt:    row.UpdatedAt,
			})
		}
	}
	return &v1.UcgAdminProfileAuditJobsListRes{
		List:     items,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}

// ProfileAuditJobResolve POST /ucg/admin/api/profile-audit-jobs/resolve
func (c *UcgAdminCtrl) ProfileAuditJobResolve(ctx context.Context, req *v1.UcgAdminProfileAuditJobResolveReq) (res *v1.UcgAdminProfileAuditJobResolveRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err = ucgsvc.ResolveProfileAuditJobAdmin(ctx, req.JobId, req.Action, req.Reason); err != nil {
		return nil, err
	}
	return &v1.UcgAdminProfileAuditJobResolveRes{Ok: true}, nil
}
