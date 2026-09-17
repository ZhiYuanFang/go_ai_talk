package voicectrl

import (
	"context"
	"net/http"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/platform/httpmeta"
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

// Daily GET /device/api/care-alert/daily — 读 latest；force 时生成并计次。
func (c *DeviceCareAlertController) Daily(ctx context.Context, req *v1.DeviceCareAlertDailyReq) (res *v1.DeviceCareAlertDailyRes, err error) {
	wxID, err := careAlertRequireWxID(ctx)
	if err != nil {
		return nil, err
	}
	force := careAlertForceTruthy(req.Force)
	day, items, used, limit, err := voice.CareAlertDaily(ctx, req.DeviceNo, wxID, force)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []v1.CareAlertItemDTO{}
	}
	return &v1.DeviceCareAlertDailyRes{Day: day, Items: items, UsedToday: used, DailyLimit: limit}, nil
}

// careAlertForceTruthy 解析 force 查询：1/true/yes（大小写不敏感）为真。
func careAlertForceTruthy(raw string) bool {
	s := strings.TrimSpace(strings.ToLower(raw))
	return s == "1" || s == "true" || s == "yes"
}

// DailyStream GET /device/api/care-alert/daily/stream — 强制生成 SSE；预检失败返回普通 envelope。
func (c *DeviceCareAlertController) DailyStream(ctx context.Context, req *v1.DeviceCareAlertDailyStreamReq) (res *v1.DeviceCareAlertDailyStreamRes, err error) {
	wxID, err := careAlertRequireWxID(ctx)
	if err != nil {
		return nil, err
	}
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, "HTTP 请求上下文缺失")
	}
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}

	var startedSSE bool
	var sseWriter http.ResponseWriter
	startSSE := func() http.ResponseWriter {
		if startedSSE {
			return sseWriter
		}
		startedSSE = true
		sseWriter = r.Response.Writer
		sseWriter.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		sseWriter.Header().Set("Cache-Control", "no-cache, no-transform")
		sseWriter.Header().Set("Connection", "keep-alive")
		sseWriter.Header().Set("X-Accel-Buffering", "no")
		sseWriter.WriteHeader(http.StatusOK)
		if flusher, ok := sseWriter.(http.Flusher); ok {
			flusher.Flush()
		}
		return sseWriter
	}

	streamErr := voice.CareAlertDailyStream(ctx, deviceNo, wxID, &voice.CareAlertDailyStreamCallback{
		OnThinking: func(dataJSON string) error {
			return writeGrowthTrajectorySSEEvent(startSSE(), "thinking", dataJSON)
		},
		OnResult: func(dataJSON string) error {
			return writeGrowthTrajectorySSEEvent(startSSE(), "result", dataJSON)
		},
		OnError: func(dataJSON string) error {
			return writeGrowthTrajectorySSEEvent(startSSE(), "error", dataJSON)
		},
	})

	// 预检失败：尚未写 SSE 头 → 普通 JSON envelope。
	if streamErr != nil && !startedSSE {
		return nil, streamErr
	}

	var rw http.ResponseWriter = startSSE()
	if streamErr != nil {
		_ = writeGrowthTrajectorySSEEvent(rw, "error", `{"code":"CARE_ALERT_STREAM","message":"护理留意分析暂时不可用，请稍后再试"}`)
	}
	_, _ = rw.Write([]byte("data: [DONE]\n\n"))
	if flusher, ok := rw.(http.Flusher); ok {
		flusher.Flush()
	}
	r.ExitAll()
	return nil, nil
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
