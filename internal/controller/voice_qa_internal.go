package controller

import (
	"context"

	v1 "hello/api/v1"
	voice "hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// VoiceQaInternalCtrl 语音域内部问答库 API（供 device-service 委派，避免跨库直连 qa 表）。
type VoiceQaInternalCtrl struct{}

// NewVoiceQaInternalCtrl 构造 VoiceQaInternalCtrl。
func NewVoiceQaInternalCtrl() *VoiceQaInternalCtrl {
	return &VoiceQaInternalCtrl{}
}

// List 问答库分页列表（id 倒序）。
func (c *VoiceQaInternalCtrl) List(ctx context.Context, req *v1.VoiceInternalQaListReq) (res *v1.VoiceInternalQaListRes, err error) {
	result, err := voice.ListQaPage(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return &v1.VoiceInternalQaListRes{
		List:     result.List,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}

// Delete 删除问答库行。
func (c *VoiceQaInternalCtrl) Delete(ctx context.Context, req *v1.VoiceInternalQaDeleteReq) (res *v1.VoiceInternalQaDeleteRes, err error) {
	if req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id 无效")
	}
	if err := voice.DeleteQa(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.VoiceInternalQaDeleteRes{}, nil
}
