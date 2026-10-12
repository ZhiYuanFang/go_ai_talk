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

// VoiceAdminReadSourceVectorsCtrl R·数据源飞轮管理（代理 Python feeding_read_source_cues）。
//
// 业务说明：Hub 整理整段 → record|predict；无 dedupe；不写 MySQL。
type VoiceAdminReadSourceVectorsCtrl struct{}

// NewVoiceAdminReadSourceVectorsCtrl 构造控制器。
func NewVoiceAdminReadSourceVectorsCtrl() *VoiceAdminReadSourceVectorsCtrl {
	return &VoiceAdminReadSourceVectorsCtrl{}
}

// requireAdmin 校验 gateway 注入的 voice Admin 口令。
func (c *VoiceAdminReadSourceVectorsCtrl) requireAdmin(ctx context.Context) error {
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

// List GET /voice/admin/api/read-source-vectors
func (c *VoiceAdminReadSourceVectorsCtrl) List(ctx context.Context, req *v1.VoiceAdminReadSourceVectorsListReq) (res *v1.VoiceAdminReadSourceVectorsListRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	offset, limit := req.Offset, req.Limit
	if limit <= 0 {
		limit = 100
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.ListReadSourceCache(ctx, offset, limit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminReadSourceVectorsListRes{
		Total:  out.Total,
		Offset: out.Offset,
		Limit:  out.Limit,
		Items:  out.Items,
	}, nil
}

// Bulk POST /voice/admin/api/read-source-vectors/bulk
func (c *VoiceAdminReadSourceVectorsCtrl) Bulk(ctx context.Context, req *v1.VoiceAdminReadSourceVectorsBulkReq) (res *v1.VoiceAdminReadSourceVectorsBulkRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	items := make([]voice.ReadSourceCacheBulkItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, voice.ReadSourceCacheBulkItem{
			Document: it.Document,
			Payload:  it.Payload,
		})
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.BulkUpsertReadSourceCache(ctx, items)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	added := out.Added
	if added == 0 && out.Ok > 0 {
		added = out.Ok
	}
	return &v1.VoiceAdminReadSourceVectorsBulkRes{
		Ok:      out.Ok,
		Added:   added,
		Skipped: out.Skipped,
		Failed:  out.Failed,
		Ids:     out.Ids,
	}, nil
}

// Delete DELETE /voice/admin/api/read-source-vectors/{id}
func (c *VoiceAdminReadSourceVectorsCtrl) Delete(ctx context.Context, req *v1.VoiceAdminReadSourceVectorsDeleteReq) (res *v1.VoiceAdminReadSourceVectorsDeleteRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(req.Id)
	client := voice.PythonAIClientFromCfg()
	if err = client.DeleteReadSourceCache(ctx, id); err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminReadSourceVectorsDeleteRes{Ok: true, Id: id}, nil
}
