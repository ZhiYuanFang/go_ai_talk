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

// VoiceAdminOpVectorsCtrl OP 轨飞轮管理（代理 Python Chroma feeding_op_cues）。
//
// 业务说明：Hub 整理整段 → op_rail；无 dedupe 路由；不写 MySQL。
type VoiceAdminOpVectorsCtrl struct{}

// NewVoiceAdminOpVectorsCtrl 构造控制器。
func NewVoiceAdminOpVectorsCtrl() *VoiceAdminOpVectorsCtrl {
	return &VoiceAdminOpVectorsCtrl{}
}

// requireAdmin 校验 gateway 注入的 voice Admin 口令。
func (c *VoiceAdminOpVectorsCtrl) requireAdmin(ctx context.Context) error {
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

// List GET /voice/admin/api/op-vectors
func (c *VoiceAdminOpVectorsCtrl) List(ctx context.Context, req *v1.VoiceAdminOpVectorsListReq) (res *v1.VoiceAdminOpVectorsListRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	offset, limit := req.Offset, req.Limit
	if limit <= 0 {
		limit = 100
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.ListOpCache(ctx, offset, limit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminOpVectorsListRes{
		Total:  out.Total,
		Offset: out.Offset,
		Limit:  out.Limit,
		Items:  out.Items,
	}, nil
}

// Bulk POST /voice/admin/api/op-vectors/bulk
func (c *VoiceAdminOpVectorsCtrl) Bulk(ctx context.Context, req *v1.VoiceAdminOpVectorsBulkReq) (res *v1.VoiceAdminOpVectorsBulkRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	items := make([]voice.OpCacheBulkItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, voice.OpCacheBulkItem{
			Document: it.Document,
			Payload:  it.Payload,
		})
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.BulkUpsertOpCache(ctx, items)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	added := out.Added
	if added == 0 && out.Ok > 0 {
		added = out.Ok
	}
	return &v1.VoiceAdminOpVectorsBulkRes{
		Ok:      out.Ok,
		Added:   added,
		Skipped: out.Skipped,
		Failed:  out.Failed,
		Ids:     out.Ids,
	}, nil
}

// Delete DELETE /voice/admin/api/op-vectors/{id}
func (c *VoiceAdminOpVectorsCtrl) Delete(ctx context.Context, req *v1.VoiceAdminOpVectorsDeleteReq) (res *v1.VoiceAdminOpVectorsDeleteRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(req.Id)
	client := voice.PythonAIClientFromCfg()
	if err = client.DeleteOpCache(ctx, id); err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminOpVectorsDeleteRes{Ok: true, Id: id}, nil
}
