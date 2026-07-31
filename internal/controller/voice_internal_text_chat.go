package controller

import (
	"context"
	"net/http"
	"strings"

	v1 "hello/api/v1"
	contracts "hello/internal/services/contracts"
	"hello/internal/services/gatewayapp"
	voice "hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// VoiceInternalTextChatCtrl voice 域内部文本对话（供 history-service HTTP 委派）。
type VoiceInternalTextChatCtrl struct{}

func NewVoiceInternalTextChatCtrl() *VoiceInternalTextChatCtrl {
	return &VoiceInternalTextChatCtrl{}
}

// Chat 执行文本智能对话；wxId 由调用方经 X-Internal-Wx-Id 注入，用于额度校验。
func (c *VoiceInternalTextChatCtrl) Chat(ctx context.Context, req *v1.VoiceInternalTextChatReq) (res *v1.VoiceInternalTextChatRes, err error) {
	_ = c
	deviceNo := strings.TrimSpace(req.DeviceNo)
	transcript := strings.TrimSpace(req.Transcript)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if transcript == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "transcript 不能为空")
	}
	chatCtx := ctx
	if r := ghttp.RequestFromCtx(ctx); r != nil {
		wxID := voice.ParseHeaderWxID(r.GetHeader(gatewayapp.HeaderInternalWxId))
		chatCtx = voice.WithVoiceWxID(ctx, wxID)
	}
	reply, err := voice.Voice().TextChat(chatCtx, deviceNo, transcript)
	if err != nil {
		return nil, mapAIQuotaErr(err)
	}
	return &v1.VoiceInternalTextChatRes{Reply: reply}, nil
}

// ChatStream 执行流式文本智能对话（SSE）：thinking 实时推送，落地后推送业务 Reply 为 answer。
func (c *VoiceInternalTextChatCtrl) ChatStream(ctx context.Context, req *v1.VoiceInternalTextChatStreamReq) (res *v1.VoiceInternalTextChatStreamRes, err error) {
	_ = c
	deviceNo := strings.TrimSpace(req.DeviceNo)
	transcript := strings.TrimSpace(req.Transcript)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if transcript == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "transcript 不能为空")
	}
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, "HTTP 请求上下文缺失")
	}
	chatCtx := ctx
	wxID := voice.ParseHeaderWxID(r.GetHeader(gatewayapp.HeaderInternalWxId))
	chatCtx = voice.WithVoiceWxID(ctx, wxID)
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
	// 2. 调用流式意图分析：过程中仅推 thinking；业务话术在落地后写入 answer
	chatRes, streamErr := voice.Voice().HandleTranscriptForIntentStream(chatCtx, deviceNo, transcript, &contracts.IntentStreamCallback{
		OnThinking: func(delta string) error {
			return writeSSEEvent(rw, "thinking", delta)
		},
	})
	// 3. 错误处理与业务话术
	if streamErr != nil {
		_ = writeSSEEvent(rw, "error", "AI服务暂时不可用，请稍后再试")
	}
	if chatRes != nil && strings.TrimSpace(chatRes.Reply) != "" {
		_ = writeSSEEvent(rw, "answer", chatRes.Reply)
	}
	// 4. 结束标记
	_, _ = rw.Write([]byte("data: [DONE]\n\n"))
	if flusher, ok := rw.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil, nil
}
