package voicectrl

import (
	"context"
	"net/http"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// DeviceGrowthTrajectoryController 成长轨迹 latest + turn SSE（宿主 voice；经 gateway /device/api/growth-trajectory/* 反代）。
type DeviceGrowthTrajectoryController struct{}

// Latest GET /device/api/growth-trajectory/latest
func (c *DeviceGrowthTrajectoryController) Latest(ctx context.Context, req *v1.DeviceGrowthTrajectoryLatestReq) (res *v1.DeviceGrowthTrajectoryLatestRes, err error) {
	wxID, err := careAlertRequireWxID(ctx)
	if err != nil {
		return nil, err
	}
	return voice.GrowthTrajectoryLatest(ctx, req.DeviceNo, wxID)
}

// Turn POST /device/api/growth-trajectory/turn — 预检失败返回普通 gerror；通过后 raw SSE。
func (c *DeviceGrowthTrajectoryController) Turn(ctx context.Context, req *v1.DeviceGrowthTrajectoryTurnReq) (res *v1.DeviceGrowthTrajectoryTurnRes, err error) {
	wxID, err := careAlertRequireWxID(ctx)
	if err != nil {
		return nil, err
	}
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, "HTTP 请求上下文缺失")
	}

	// 预检（开通 / 日限 / 参数）在开 SSE 头之前完成，便于 Flutter Toast message。
	deviceNo := strings.TrimSpace(req.DeviceNo)
	action := strings.TrimSpace(req.Action)
	sessionID := strings.TrimSpace(req.SessionId)

	// 先做轻量参数校验；完整门禁在 GrowthTrajectoryTurn 内（含 access/日限）。
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}

	// 为了让预检错误走 JSON envelope：先 dry-run 门禁路径不可行（会占 flight），
	// 因此由 GrowthTrajectoryTurn 在写 SSE 前返回 err；控制器仅在 err==nil 时不应再包 envelope。
	// 实际做法：Turn 内部若在回调前失败，此处尚未写 SSE 头，直接 return err。

	var startedSSE bool
	var sseWriter http.ResponseWriter
	startSSE := func() http.ResponseWriter {
		if startedSSE {
			return sseWriter
		}
		startedSSE = true
		// 显式落到 net/http.ResponseWriter，以便 Flusher 断言（对齐已下线 tip）。
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

	streamErr := voice.GrowthTrajectoryTurn(ctx, deviceNo, action, sessionID, req.Answer, wxID, &voice.GrowthTrajectoryTurnCallback{
		OnThinking: func(dataJSON string) error {
			return writeGrowthTrajectorySSEEvent(startSSE(), "thinking", dataJSON)
		},
		OnQuestion: func(dataJSON string) error {
			return writeGrowthTrajectorySSEEvent(startSSE(), "question", dataJSON)
		},
		OnResult: func(dataJSON string) error {
			return writeGrowthTrajectorySSEEvent(startSSE(), "result", dataJSON)
		},
		OnError: func(dataJSON string) error {
			return writeGrowthTrajectorySSEEvent(startSSE(), "error", dataJSON)
		},
		OnDone: func(dataJSON string) error {
			if strings.TrimSpace(dataJSON) == "" {
				dataJSON = "{}"
			}
			return writeGrowthTrajectorySSEEvent(startSSE(), "done", dataJSON)
		},
	})

	// 预检失败：尚未写 SSE 头 → 普通 JSON envelope。
	if streamErr != nil && !startedSSE {
		return nil, streamErr
	}

	var rw http.ResponseWriter = startSSE()
	if streamErr != nil {
		// 已开流：以 SSE error 事件告知客户端。
		_ = writeGrowthTrajectorySSEEvent(rw, "error", `{"code":"GROWTH_TRAJECTORY_STREAM","message":"成长轨迹预测暂时不可用，请稍后再试"}`)
	}
	_, _ = rw.Write([]byte("data: [DONE]\n\n"))
	if flusher, ok := rw.(http.Flusher); ok {
		flusher.Flush()
	}
	r.ExitAll()
	return nil, nil
}

// writeGrowthTrajectorySSEEvent 写入一个 SSE 事件帧并立即 flush（对齐已下线 tip 的 writeSSEEvent）。
func writeGrowthTrajectorySSEEvent(rw http.ResponseWriter, event, data string) error {
	if _, err := rw.Write([]byte("event: " + event + "\n")); err != nil {
		return err
	}
	for _, line := range strings.Split(data, "\n") {
		if _, err := rw.Write([]byte("data: " + line + "\n")); err != nil {
			return err
		}
	}
	if _, err := rw.Write([]byte("\n")); err != nil {
		return err
	}
	if flusher, ok := rw.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}
