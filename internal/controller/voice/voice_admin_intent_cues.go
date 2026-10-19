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

// VoiceAdminIntentCuesCtrl 按原文跨九个飞轮查询。
//
// 业务说明：Hub 输入一整句 document，voice 转发 Python 聚合接口。
// 删除不在本控制器：各维度 id 可能撞车，门户按 dimension 调已有 *-vectors DELETE。
// 本路径属于 /voice/admin/api/，与其它飞轮管理接口同一类，不单独改 usage 排除名单。
type VoiceAdminIntentCuesCtrl struct{}

// NewVoiceAdminIntentCuesCtrl 构造控制器。
func NewVoiceAdminIntentCuesCtrl() *VoiceAdminIntentCuesCtrl {
	return &VoiceAdminIntentCuesCtrl{}
}

// requireAdmin 校验 gateway 注入的 voice Admin 口令。
func (c *VoiceAdminIntentCuesCtrl) requireAdmin(ctx context.Context) error {
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

// ByDocument GET /voice/admin/api/intent-cues/by-document
//
// 业务逻辑：空白原文直接拒绝，不调用 Python。非空则原样转发聚合查询结果。
func (c *VoiceAdminIntentCuesCtrl) ByDocument(ctx context.Context, req *v1.VoiceAdminIntentCuesByDocumentReq) (res *v1.VoiceAdminIntentCuesByDocumentRes, err error) {
	if err = c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	doc := strings.TrimSpace(req.Document)
	// 空串不扫九个集合。
	if doc == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "document 不能为空")
	}
	client := voice.PythonAIClientFromCfg()
	out, err := client.ListCuesByDocument(ctx, doc)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, err.Error())
	}
	items := make([]v1.VoiceAdminIntentCueHit, 0, len(out.Items))
	for _, it := range out.Items {
		items = append(items, v1.VoiceAdminIntentCueHit{
			Dimension:  it.Dimension,
			Collection: it.Collection,
			Id:         it.Id,
			Document:   it.Document,
			Payload:    it.Payload,
		})
	}
	return &v1.VoiceAdminIntentCuesByDocumentRes{Items: items}, nil
}
