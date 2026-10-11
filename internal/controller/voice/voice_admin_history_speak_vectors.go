package voicectrl

// voice_admin_history_speak_vectors.go：Hub 查史播报向量管理（代理 Python feeding_history_speak_cues）。
//
// 业务：整理点查播报策略关键字（point_pick / open_speak），与类型/事件飞轮隔离。
// 鉴权同 type-vectors；不写 MySQL、不直连 Chroma。

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	voice "hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// VoiceAdminHistorySpeakVectorsCtrl 查史播报向量管理。
type VoiceAdminHistorySpeakVectorsCtrl struct{}

// NewVoiceAdminHistorySpeakVectorsCtrl 构造控制器。
func NewVoiceAdminHistorySpeakVectorsCtrl() *VoiceAdminHistorySpeakVectorsCtrl {
	return &VoiceAdminHistorySpeakVectorsCtrl{}
}

// requireAdmin 校验 gateway 注入的 voice Admin 口令。
func (c *VoiceAdminHistorySpeakVectorsCtrl) requireAdmin(ctx context.Context) error {
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

// List GET /voice/admin/api/history-speak-vectors
func (c *VoiceAdminHistorySpeakVectorsCtrl) List(ctx context.Context, req *v1.VoiceAdminHistorySpeakVectorsListReq) (res *v1.VoiceAdminHistorySpeakVectorsListRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	offset, limit := req.Offset, req.Limit
	if limit <= 0 {
		limit = 100
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.ListHistorySpeakCache(ctx, offset, limit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminHistorySpeakVectorsListRes{
		Total:  out.Total,
		Offset: out.Offset,
		Limit:  out.Limit,
		Items:  out.Items,
	}, nil
}

// Bulk POST /voice/admin/api/history-speak-vectors/bulk
func (c *VoiceAdminHistorySpeakVectorsCtrl) Bulk(ctx context.Context, req *v1.VoiceAdminHistorySpeakVectorsBulkReq) (res *v1.VoiceAdminHistorySpeakVectorsBulkRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	items := make([]voice.HistorySpeakCacheBulkItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, voice.HistorySpeakCacheBulkItem{
			Document: it.Document,
			Payload:  it.Payload,
		})
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.BulkUpsertHistorySpeakCache(ctx, items)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	added := out.Added
	if added == 0 && out.Ok > 0 {
		added = out.Ok
	}
	return &v1.VoiceAdminHistorySpeakVectorsBulkRes{
		Ok:      out.Ok,
		Added:   added,
		Skipped: out.Skipped,
		Failed:  out.Failed,
		Ids:     out.Ids,
	}, nil
}

// Dedupe POST /voice/admin/api/history-speak-vectors/dedupe
func (c *VoiceAdminHistorySpeakVectorsCtrl) Dedupe(ctx context.Context, req *v1.VoiceAdminHistorySpeakVectorsDedupeReq) (res *v1.VoiceAdminHistorySpeakVectorsDedupeRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	_ = req
	client := voice.PythonAIClientFromCfg()
	out, err := client.DedupeHistorySpeakCache(ctx)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminHistorySpeakVectorsDedupeRes{
		Kept:    out.Kept,
		Deleted: out.Deleted,
		Groups:  out.Groups,
	}, nil
}

// Delete DELETE /voice/admin/api/history-speak-vectors/{id}
func (c *VoiceAdminHistorySpeakVectorsCtrl) Delete(ctx context.Context, req *v1.VoiceAdminHistorySpeakVectorsDeleteReq) (res *v1.VoiceAdminHistorySpeakVectorsDeleteRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(req.Id)
	client := voice.PythonAIClientFromCfg()
	if err = client.DeleteHistorySpeakCache(ctx, id); err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminHistorySpeakVectorsDeleteRes{Ok: true, Id: id}, nil
}
