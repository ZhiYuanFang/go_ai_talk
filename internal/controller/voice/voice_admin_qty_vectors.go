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

// VoiceAdminQtyVectorsCtrl 数量飞轮管理（代理 Python Chroma feeding_qty_cues）。
//
// 业务说明：Hub 整理整段口语量 → quantity；无 dedupe；不写 MySQL。
type VoiceAdminQtyVectorsCtrl struct{}

// NewVoiceAdminQtyVectorsCtrl 构造控制器。
func NewVoiceAdminQtyVectorsCtrl() *VoiceAdminQtyVectorsCtrl {
	return &VoiceAdminQtyVectorsCtrl{}
}

// requireAdmin 校验 gateway 注入的 voice Admin 口令。
func (c *VoiceAdminQtyVectorsCtrl) requireAdmin(ctx context.Context) error {
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

// List GET /voice/admin/api/qty-vectors
func (c *VoiceAdminQtyVectorsCtrl) List(ctx context.Context, req *v1.VoiceAdminQtyVectorsListReq) (res *v1.VoiceAdminQtyVectorsListRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	offset, limit := req.Offset, req.Limit
	if limit <= 0 {
		limit = 100
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.ListQtyCache(ctx, offset, limit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminQtyVectorsListRes{
		Total:  out.Total,
		Offset: out.Offset,
		Limit:  out.Limit,
		Items:  out.Items,
	}, nil
}

// Bulk POST /voice/admin/api/qty-vectors/bulk
func (c *VoiceAdminQtyVectorsCtrl) Bulk(ctx context.Context, req *v1.VoiceAdminQtyVectorsBulkReq) (res *v1.VoiceAdminQtyVectorsBulkRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	items := make([]voice.QtyCacheBulkItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, voice.QtyCacheBulkItem{
			Document: it.Document,
			Payload:  it.Payload,
		})
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.BulkUpsertQtyCache(ctx, items)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	added := out.Added
	if added == 0 && out.Ok > 0 {
		added = out.Ok
	}
	return &v1.VoiceAdminQtyVectorsBulkRes{
		Ok:      out.Ok,
		Added:   added,
		Skipped: out.Skipped,
		Failed:  out.Failed,
		Ids:     out.Ids,
	}, nil
}

// Delete DELETE /voice/admin/api/qty-vectors/{id}
func (c *VoiceAdminQtyVectorsCtrl) Delete(ctx context.Context, req *v1.VoiceAdminQtyVectorsDeleteReq) (res *v1.VoiceAdminQtyVectorsDeleteRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(req.Id)
	client := voice.PythonAIClientFromCfg()
	if err = client.DeleteQtyCache(ctx, id); err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminQtyVectorsDeleteRes{Ok: true, Id: id}, nil
}
