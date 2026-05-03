package controller

import (
	"context"

	v1 "hello/api/v1"
	voice "hello/internal/services/voice"
)

// VoiceQaInternalCtrl 语音域内部问答库 API（供 device-service 委派，避免跨库直连 qa 表）。
type VoiceQaInternalCtrl struct{}

// NewVoiceQaInternalCtrl 构造 VoiceQaInternalCtrl。
func NewVoiceQaInternalCtrl() *VoiceQaInternalCtrl {
	return &VoiceQaInternalCtrl{}
}

// List 问答库全表列表。
func (c *VoiceQaInternalCtrl) List(ctx context.Context, req *v1.VoiceInternalQaListReq) (res *v1.VoiceInternalQaListRes, err error) {
	items, err := voice.ListQaAllForAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.VoiceInternalQaListRes{List: items}, nil
}
