package voicectrl

// voice_admin_miss_inbox.go：Hub 覆盖缺口（答不出）收件箱。
//
// 业务说明：代理 Python /v1/admin/miss-inbox 列表与物理删除；
// 独立于 rule_gaps 与时间飞轮；不落 Go 库。鉴权同 voice Admin 口令。

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	voice "hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// VoiceAdminMissInboxCtrl 覆盖缺口管理（代理 Python agent_miss）。
type VoiceAdminMissInboxCtrl struct{}

// NewVoiceAdminMissInboxCtrl 构造控制器。
func NewVoiceAdminMissInboxCtrl() *VoiceAdminMissInboxCtrl {
	return &VoiceAdminMissInboxCtrl{}
}

// requireAdmin 校验 gateway 注入的 voice Admin 口令。
func (c *VoiceAdminMissInboxCtrl) requireAdmin(ctx context.Context) error {
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

// List GET /voice/admin/api/miss-inbox
func (c *VoiceAdminMissInboxCtrl) List(ctx context.Context, req *v1.VoiceAdminMissInboxListReq) (res *v1.VoiceAdminMissInboxListRes, err error) {
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
	out, err := client.ListMissInbox(ctx, offset, limit, status, req.FailKind)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	return &v1.VoiceAdminMissInboxListRes{
		Total:  out.Total,
		Offset: out.Offset,
		Limit:  out.Limit,
		Items:  out.Items,
	}, nil
}

// Delete DELETE /voice/admin/api/miss-inbox/{id}
func (c *VoiceAdminMissInboxCtrl) Delete(ctx context.Context, req *v1.VoiceAdminMissInboxDeleteReq) (res *v1.VoiceAdminMissInboxDeleteRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(req.Id)
	client := voice.PythonAIClientFromCfg()
	if err = client.DeleteMissInbox(ctx, id); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "不存在") {
			return nil, gerror.NewCode(gcode.CodeNotFound, msg)
		}
		return nil, gerror.NewCode(gcode.CodeInternalError, msg)
	}
	return &v1.VoiceAdminMissInboxDeleteRes{Ok: true, Id: id}, nil
}
