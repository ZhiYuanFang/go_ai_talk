package voicectrl

import (
	"hello/internal/platform/httpmeta"
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// DeviceCareAlertController 护理留意日缓存 API（宿主 voice-service，经 gateway /device/api/care-alert/* 反代）。
// 不扣 clinic 配额；编排 Python KG+LLM；VIP 选模见 voice.resolveCareAlertModelProfile（按触发者 wxId）。
// 三条接口均要求 X-Internal-Wx-Id>0，不支持纯设备会话。
type DeviceCareAlertController struct{}

// careAlertRequireWxID 从网关注入头解析 wx 主键；缺失或非正则拒绝（禁止 deviceNo 反查旁路）。
func careAlertRequireWxID(ctx context.Context) (int64, error) {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	wxID := httpmeta.ParseHeaderWxID(r.GetHeader(httpmeta.HeaderInternalWxId))
	if wxID <= 0 {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	return wxID, nil
}

// Daily GET /device/api/care-alert/daily — 宝宝日缓存列表；未命中 single-flight 阻塞生成。
// force=1/true 时先删当日日缓存再生成（仍要求 wxId>0，无鉴权旁路）。
func (c *DeviceCareAlertController) Daily(ctx context.Context, req *v1.DeviceCareAlertDailyReq) (res *v1.DeviceCareAlertDailyRes, err error) {
	wxID, err := careAlertRequireWxID(ctx)
	if err != nil {
		return nil, err
	}
	force := careAlertForceTruthy(req.Force)
	day, items, err := voice.CareAlertDaily(ctx, req.DeviceNo, wxID, force)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []v1.CareAlertItemDTO{}
	}
	return &v1.DeviceCareAlertDailyRes{Day: day, Items: items}, nil
}

// careAlertForceTruthy 解析 force 查询：1/true/yes（大小写不敏感）为真。
func careAlertForceTruthy(raw string) bool {
	s := strings.TrimSpace(strings.ToLower(raw))
	return s == "1" || s == "true" || s == "yes"
}

// DailyItemDelete DELETE /device/api/care-alert/daily/item — 仅删当日缓存中该 suggestionId。
func (c *DeviceCareAlertController) DailyItemDelete(ctx context.Context, req *v1.DeviceCareAlertDailyItemDeleteReq) (res *v1.DeviceCareAlertDailyItemDeleteRes, err error) {
	wxID, err := careAlertRequireWxID(ctx)
	if err != nil {
		return nil, err
	}
	day, items, err := voice.CareAlertDeleteItem(ctx, req.DeviceNo, req.SuggestionId, wxID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []v1.CareAlertItemDTO{}
	}
	return &v1.DeviceCareAlertDailyItemDeleteRes{Day: day, Items: items}, nil
}

// Feedback POST /device/api/care-alert/feedback — 固定意图 ignore|follow_up，无 NLP。
func (c *DeviceCareAlertController) Feedback(ctx context.Context, req *v1.DeviceCareAlertFeedbackReq) (res *v1.DeviceCareAlertFeedbackRes, err error) {
	wxID, err := careAlertRequireWxID(ctx)
	if err != nil {
		return nil, err
	}
	intent := strings.TrimSpace(req.Intent)
	if intent != "ignore" && intent != "follow_up" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "intent 必须为 ignore 或 follow_up")
	}
	if err := voice.CareAlertFeedback(ctx, req.DeviceNo, req.SuggestionId, intent, wxID); err != nil {
		return nil, err
	}
	return &v1.DeviceCareAlertFeedbackRes{}, nil
}
