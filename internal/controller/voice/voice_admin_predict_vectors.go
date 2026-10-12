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

// VoiceAdminPredictVectorsCtrl R·predict 两维飞轮管理（代理 Python Chroma feeding_predict_cues）。
//
// 业务说明：Hub 整理整段 → predict_channel + time_answer；无 dedupe；不写 MySQL。
type VoiceAdminPredictVectorsCtrl struct{}

// NewVoiceAdminPredictVectorsCtrl 构造控制器。
func NewVoiceAdminPredictVectorsCtrl() *VoiceAdminPredictVectorsCtrl {
	return &VoiceAdminPredictVectorsCtrl{}
}

// requireAdmin 校验 gateway 注入的 voice Admin 口令。
func (c *VoiceAdminPredictVectorsCtrl) requireAdmin(ctx context.Context) error {
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

// List GET /voice/admin/api/predict-vectors
func (c *VoiceAdminPredictVectorsCtrl) List(ctx context.Context, req *v1.VoiceAdminPredictVectorsListReq) (res *v1.VoiceAdminPredictVectorsListRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	offset, limit := req.Offset, req.Limit
	if limit <= 0 {
		limit = 100
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.ListPredictCache(ctx, offset, limit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminPredictVectorsListRes{
		Total:  out.Total,
		Offset: out.Offset,
		Limit:  out.Limit,
		Items:  out.Items,
	}, nil
}

// Bulk POST /voice/admin/api/predict-vectors/bulk
func (c *VoiceAdminPredictVectorsCtrl) Bulk(ctx context.Context, req *v1.VoiceAdminPredictVectorsBulkReq) (res *v1.VoiceAdminPredictVectorsBulkRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	items := make([]voice.PredictCacheBulkItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, voice.PredictCacheBulkItem{
			Document: it.Document,
			Payload:  it.Payload,
		})
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.BulkUpsertPredictCache(ctx, items)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	added := out.Added
	if added == 0 && out.Ok > 0 {
		added = out.Ok
	}
	return &v1.VoiceAdminPredictVectorsBulkRes{
		Ok:      out.Ok,
		Added:   added,
		Skipped: out.Skipped,
		Failed:  out.Failed,
		Ids:     out.Ids,
	}, nil
}

// Delete DELETE /voice/admin/api/predict-vectors/{id}
func (c *VoiceAdminPredictVectorsCtrl) Delete(ctx context.Context, req *v1.VoiceAdminPredictVectorsDeleteReq) (res *v1.VoiceAdminPredictVectorsDeleteRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(req.Id)
	client := voice.PythonAIClientFromCfg()
	if err = client.DeletePredictCache(ctx, id); err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminPredictVectorsDeleteRes{Ok: true, Id: id}, nil
}
