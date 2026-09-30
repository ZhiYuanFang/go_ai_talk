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

// VoiceAdminIntentVectorsCtrl 意图向量管理（代理 Python Chroma feeding_intents）。
type VoiceAdminIntentVectorsCtrl struct{}

// NewVoiceAdminIntentVectorsCtrl 构造控制器。
func NewVoiceAdminIntentVectorsCtrl() *VoiceAdminIntentVectorsCtrl {
	return &VoiceAdminIntentVectorsCtrl{}
}

func (c *VoiceAdminIntentVectorsCtrl) requireAdmin(ctx context.Context) error {
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

// List GET /voice/admin/api/intent-vectors
func (c *VoiceAdminIntentVectorsCtrl) List(ctx context.Context, req *v1.VoiceAdminIntentVectorsListReq) (res *v1.VoiceAdminIntentVectorsListRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	offset, limit := req.Offset, req.Limit
	if limit <= 0 {
		limit = 100
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.ListIntentCache(ctx, offset, limit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminIntentVectorsListRes{
		Total:  out.Total,
		Offset: out.Offset,
		Limit:  out.Limit,
		Items:  out.Items,
	}, nil
}

// Bulk POST /voice/admin/api/intent-vectors/bulk
func (c *VoiceAdminIntentVectorsCtrl) Bulk(ctx context.Context, req *v1.VoiceAdminIntentVectorsBulkReq) (res *v1.VoiceAdminIntentVectorsBulkRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	items := make([]voice.IntentCacheBulkItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, voice.IntentCacheBulkItem{
			Document: it.Document,
			Payload:  it.Payload,
		})
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.BulkUpsertIntentCache(ctx, items)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminIntentVectorsBulkRes{
		Ok:     out.Ok,
		Failed: out.Failed,
		Ids:    out.Ids,
	}, nil
}

// Delete DELETE /voice/admin/api/intent-vectors/{id}
func (c *VoiceAdminIntentVectorsCtrl) Delete(ctx context.Context, req *v1.VoiceAdminIntentVectorsDeleteReq) (res *v1.VoiceAdminIntentVectorsDeleteRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(req.Id)
	client := voice.PythonAIClientFromCfg()
	if err = client.DeleteIntentCache(ctx, id); err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminIntentVectorsDeleteRes{Ok: true, Id: id}, nil
}
