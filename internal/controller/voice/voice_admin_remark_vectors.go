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

// VoiceAdminRemarkVectorsCtrl 备注整段飞轮管理（代理 Python Chroma feeding_remark_cues）。
//
// 业务说明：Hub 整理整段 → remark（可空）；无 dedupe；不写 MySQL。
type VoiceAdminRemarkVectorsCtrl struct{}

// NewVoiceAdminRemarkVectorsCtrl 构造控制器。
func NewVoiceAdminRemarkVectorsCtrl() *VoiceAdminRemarkVectorsCtrl {
	return &VoiceAdminRemarkVectorsCtrl{}
}

// requireAdmin 校验 gateway 注入的 voice Admin 口令。
func (c *VoiceAdminRemarkVectorsCtrl) requireAdmin(ctx context.Context) error {
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

// List GET /voice/admin/api/remark-vectors
func (c *VoiceAdminRemarkVectorsCtrl) List(ctx context.Context, req *v1.VoiceAdminRemarkVectorsListReq) (res *v1.VoiceAdminRemarkVectorsListRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	offset, limit := req.Offset, req.Limit
	if limit <= 0 {
		limit = 100
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.ListRemarkCache(ctx, offset, limit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminRemarkVectorsListRes{
		Total:  out.Total,
		Offset: out.Offset,
		Limit:  out.Limit,
		Items:  out.Items,
	}, nil
}

// Bulk POST /voice/admin/api/remark-vectors/bulk
func (c *VoiceAdminRemarkVectorsCtrl) Bulk(ctx context.Context, req *v1.VoiceAdminRemarkVectorsBulkReq) (res *v1.VoiceAdminRemarkVectorsBulkRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	items := make([]voice.RemarkCacheBulkItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, voice.RemarkCacheBulkItem{
			Document: it.Document,
			Payload:  it.Payload,
		})
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.BulkUpsertRemarkCache(ctx, items)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	added := out.Added
	if added == 0 && out.Ok > 0 {
		added = out.Ok
	}
	return &v1.VoiceAdminRemarkVectorsBulkRes{
		Ok:      out.Ok,
		Added:   added,
		Skipped: out.Skipped,
		Failed:  out.Failed,
		Ids:     out.Ids,
	}, nil
}

// Delete DELETE /voice/admin/api/remark-vectors/{id}
func (c *VoiceAdminRemarkVectorsCtrl) Delete(ctx context.Context, req *v1.VoiceAdminRemarkVectorsDeleteReq) (res *v1.VoiceAdminRemarkVectorsDeleteRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(req.Id)
	client := voice.PythonAIClientFromCfg()
	if err = client.DeleteRemarkCache(ctx, id); err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminRemarkVectorsDeleteRes{Ok: true, Id: id}, nil
}
