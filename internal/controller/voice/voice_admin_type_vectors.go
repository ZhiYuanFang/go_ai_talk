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

// VoiceAdminTypeVectorsCtrl 类型向量管理（代理 Python Chroma feeding_type_cues）。
//
// 业务说明：对标意图向量缓存；Hub 整理类型关键字飞轮，不写 MySQL。
type VoiceAdminTypeVectorsCtrl struct{}

// NewVoiceAdminTypeVectorsCtrl 构造控制器。
func NewVoiceAdminTypeVectorsCtrl() *VoiceAdminTypeVectorsCtrl {
	return &VoiceAdminTypeVectorsCtrl{}
}

// requireAdmin 校验 gateway 注入的 voice Admin 口令。
func (c *VoiceAdminTypeVectorsCtrl) requireAdmin(ctx context.Context) error {
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

// List GET /voice/admin/api/type-vectors
func (c *VoiceAdminTypeVectorsCtrl) List(ctx context.Context, req *v1.VoiceAdminTypeVectorsListReq) (res *v1.VoiceAdminTypeVectorsListRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	offset, limit := req.Offset, req.Limit
	if limit <= 0 {
		limit = 100
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.ListTypeCache(ctx, offset, limit)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminTypeVectorsListRes{
		Total:  out.Total,
		Offset: out.Offset,
		Limit:  out.Limit,
		Items:  out.Items,
	}, nil
}

// Bulk POST /voice/admin/api/type-vectors/bulk
func (c *VoiceAdminTypeVectorsCtrl) Bulk(ctx context.Context, req *v1.VoiceAdminTypeVectorsBulkReq) (res *v1.VoiceAdminTypeVectorsBulkRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	items := make([]voice.TypeCacheBulkItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, voice.TypeCacheBulkItem{
			Document: it.Document,
			Payload:  it.Payload,
		})
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.BulkUpsertTypeCache(ctx, items)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	added := out.Added
	if added == 0 && out.Ok > 0 {
		// 兼容旧 Python 仅回 ok 的情况
		added = out.Ok
	}
	return &v1.VoiceAdminTypeVectorsBulkRes{
		Ok:      out.Ok,
		Added:   added,
		Skipped: out.Skipped,
		Failed:  out.Failed,
		Ids:     out.Ids,
	}, nil
}

// Dedupe POST /voice/admin/api/type-vectors/dedupe
func (c *VoiceAdminTypeVectorsCtrl) Dedupe(ctx context.Context, req *v1.VoiceAdminTypeVectorsDedupeReq) (res *v1.VoiceAdminTypeVectorsDedupeRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	_ = req
	client := voice.PythonAIClientFromCfg()
	out, err := client.DedupeTypeCache(ctx)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminTypeVectorsDedupeRes{
		Kept:    out.Kept,
		Deleted: out.Deleted,
		Groups:  out.Groups,
	}, nil
}

// Delete DELETE /voice/admin/api/type-vectors/{id}
func (c *VoiceAdminTypeVectorsCtrl) Delete(ctx context.Context, req *v1.VoiceAdminTypeVectorsDeleteReq) (res *v1.VoiceAdminTypeVectorsDeleteRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(req.Id)
	client := voice.PythonAIClientFromCfg()
	if err = client.DeleteTypeCache(ctx, id); err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminTypeVectorsDeleteRes{Ok: true, Id: id}, nil
}
