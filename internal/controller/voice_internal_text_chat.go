package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
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
