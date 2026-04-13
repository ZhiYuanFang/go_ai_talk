package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/service"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// VoiceTextCtrl 文本对话 API（Header: X-Admin-Password）。
type VoiceTextCtrl struct {
	Voice service.VoiceContract
	Admin service.DeviceAdminContract
}

// NewVoiceTextCtrl 构造 VoiceTextCtrl。
func NewVoiceTextCtrl(voice service.VoiceContract, admin service.DeviceAdminContract) *VoiceTextCtrl {
	return &VoiceTextCtrl{Voice: voice, Admin: admin}
}

// Chat 触发内部 chat 对话推理。
func (c *VoiceTextCtrl) Chat(ctx context.Context, req *v1.VoiceTextChatReq) (res *v1.VoiceTextChatRes, err error) {
	if !c.Admin.VerifyPassword(adminPassword(ctx)) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "口令错误")
	}
	deviceNo := strings.TrimSpace(req.DeviceNo)
	transcript := strings.TrimSpace(req.Transcript)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if transcript == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "transcript 不能为空")
	}
	reply, err := c.Voice.TextChat(ctx, deviceNo, transcript)
	if err != nil {
		return nil, err
	}
	return &v1.VoiceTextChatRes{Reply: reply}, nil
}
