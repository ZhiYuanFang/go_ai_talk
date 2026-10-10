package voicectrl

// voice_admin_rule_gaps.go：Hub 规则缺口收件箱（代理 Python /v1/admin/rule-gaps）。
//
// 业务：多维本地规则未覆盖时 Python 落 JSONL；本控制器仅代理列表与标已处理，
// 不落 Go 库、不改规则代码。鉴权与意图向量 Admin 一致。

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	voice "hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// VoiceAdminRuleGapsCtrl 规则缺口管理（代理 Python rule_gaps）。
type VoiceAdminRuleGapsCtrl struct{}

// NewVoiceAdminRuleGapsCtrl 构造控制器。
func NewVoiceAdminRuleGapsCtrl() *VoiceAdminRuleGapsCtrl {
	return &VoiceAdminRuleGapsCtrl{}
}

func (c *VoiceAdminRuleGapsCtrl) requireAdmin(ctx context.Context) error {
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

// List GET /voice/admin/api/rule-gaps
func (c *VoiceAdminRuleGapsCtrl) List(ctx context.Context, req *v1.VoiceAdminRuleGapsListReq) (res *v1.VoiceAdminRuleGapsListRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	offset, limit := req.Offset, req.Limit
	if limit <= 0 {
		limit = 100
	}
	// 未传 status → 默认 open；显式 status=（空）→ 不过滤（与 Python 一致）。
	status := req.Status
	if r := ghttp.RequestFromCtx(ctx); r != nil {
		if _, ok := r.URL.Query()["status"]; !ok {
			status = "open"
		}
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.ListRuleGaps(ctx, offset, limit, req.Dimension, status)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminRuleGapsListRes{
		Total:  out.Total,
		Offset: out.Offset,
		Limit:  out.Limit,
		Items:  out.Items,
	}, nil
}

// Patch PATCH /voice/admin/api/rule-gaps/{id}
func (c *VoiceAdminRuleGapsCtrl) Patch(ctx context.Context, req *v1.VoiceAdminRuleGapsPatchReq) (res *v1.VoiceAdminRuleGapsPatchRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(req.Id)
	st := strings.TrimSpace(req.Status)
	client := voice.PythonAIClientFromCfg()
	out, err := client.PatchRuleGapStatus(ctx, id, st)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "不存在") {
			return nil, gerror.NewCode(gcode.CodeNotFound, msg)
		}
		return nil, gerror.NewCode(gcode.CodeInternalError, msg)
	}
	return &v1.VoiceAdminRuleGapsPatchRes{
		Ok:     out.Ok,
		Id:     out.Id,
		Status: out.Status,
	}, nil
}
