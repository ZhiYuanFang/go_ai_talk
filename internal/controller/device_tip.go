package controller

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v1 "hello/api/v1"
	contracts "hello/internal/services/contracts"
	voice "hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// TipCtrl 小贴士 API（流式生成）。
type TipCtrl struct{}

func NewTipCtrl() *TipCtrl {
	return &TipCtrl{}
}

// Generate 流式生成小贴士（SSE）。
// 逐帧推送 thinking/answer/done 事件，最终以 data: [DONE] 结束。
func (c *TipCtrl) Generate(ctx context.Context, req *v1.DeviceTipGenerateReq) (res *v1.DeviceTipGenerateRes, err error) {
	_ = c
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, "HTTP 请求上下文缺失")
	}
	// 1. 设置 SSE 响应头
	var rw http.ResponseWriter = r.Response.Writer
	rw.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	rw.Header().Set("Cache-Control", "no-cache, no-transform")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("X-Accel-Buffering", "no")
	rw.WriteHeader(http.StatusOK)
	if flusher, ok := rw.(http.Flusher); ok {
		flusher.Flush()
	}
	// 2. 调用 voice 服务流式小贴士生成（月龄/时间由 Python 派生，此处不传）
	_, streamErr := voice.Voice().TipStream(ctx, deviceNo, req.EventId, req.EventName, &contracts.TipStreamCallback{
		OnThinking: func(delta string) error {
			return writeSSEEvent(rw, "thinking", delta)
		},
		OnAnswer: func(delta string) error {
			return writeSSEEvent(rw, "answer", delta)
		},
		OnDone: func(answerID string) error {
			return writeSSEEvent(rw, "done", fmt.Sprintf(`{"answerId":"%s"}`, answerID))
		},
	})
	// 3. 错误处理：AI 不可用时返回固定提示
	if streamErr != nil {
		_ = writeSSEEvent(rw, "error", "AI服务暂时不可用，请稍后再试")
	}
	// 4. 结束标记
	_, _ = rw.Write([]byte("data: [DONE]\n\n"))
	if flusher, ok := rw.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil, nil
}

// writeSSEEvent 写入一个 SSE 事件帧并立即 flush。
func writeSSEEvent(rw http.ResponseWriter, event, data string) error {
	if _, err := rw.Write([]byte("event: " + event + "\n")); err != nil {
		return err
	}
	// 按行拆分 data，每行前加 "data: "
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
