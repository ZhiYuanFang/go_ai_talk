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

// VoiceAdminLocateVectorsCtrl U/D 定位策略飞轮管理（代理 Python Chroma feeding_locate_cues）。
//
// 业务说明：Hub 整理整段 → strategy；禁止 historyId；无 dedupe；不写 MySQL。
type VoiceAdminLocateVectorsCtrl struct{}

// NewVoiceAdminLocateVectorsCtrl 构造控制器。
func NewVoiceAdminLocateVectorsCtrl() *VoiceAdminLocateVectorsCtrl {
	return &VoiceAdminLocateVectorsCtrl{}
}

// requireAdmin 校验 gateway 注入的 voice Admin 口令。
func (c *VoiceAdminLocateVectorsCtrl) requireAdmin(ctx context.Context) error {
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

// List GET /voice/admin/api/locate-vectors
func (c *VoiceAdminLocateVectorsCtrl) List(ctx context.Context, req *v1.VoiceAdminLocateVectorsListReq) (res *v1.VoiceAdminLocateVectorsListRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	offset, limit := req.Offset, req.Limit
	if limit <= 0 {
		limit = 100
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.ListLocateCache(ctx, offset, limit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminLocateVectorsListRes{
		Total:  out.Total,
		Offset: out.Offset,
		Limit:  out.Limit,
		Items:  out.Items,
	}, nil
}

// Bulk POST /voice/admin/api/locate-vectors/bulk
func (c *VoiceAdminLocateVectorsCtrl) Bulk(ctx context.Context, req *v1.VoiceAdminLocateVectorsBulkReq) (res *v1.VoiceAdminLocateVectorsBulkRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	items := make([]voice.LocateCacheBulkItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, voice.LocateCacheBulkItem{
			Document: it.Document,
			Payload:  it.Payload,
		})
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.BulkUpsertLocateCache(ctx, items)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	added := out.Added
	if added == 0 && out.Ok > 0 {
		added = out.Ok
	}
	return &v1.VoiceAdminLocateVectorsBulkRes{
		Ok:      out.Ok,
		Added:   added,
		Skipped: out.Skipped,
		Failed:  out.Failed,
		Ids:     out.Ids,
	}, nil
}

// Delete DELETE /voice/admin/api/locate-vectors/{id}
func (c *VoiceAdminLocateVectorsCtrl) Delete(ctx context.Context, req *v1.VoiceAdminLocateVectorsDeleteReq) (res *v1.VoiceAdminLocateVectorsDeleteRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(req.Id)
	client := voice.PythonAIClientFromCfg()
	if err = client.DeleteLocateCache(ctx, id); err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminLocateVectorsDeleteRes{Ok: true, Id: id}, nil
}
