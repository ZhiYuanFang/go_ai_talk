package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	voice "hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// VoiceSuggestInternalCtrl 语音域内部建议 API（供 history 网关路径聚合调用）。
type VoiceSuggestInternalCtrl struct{}

// NewVoiceSuggestInternalCtrl 构造 VoiceSuggestInternalCtrl。
func NewVoiceSuggestInternalCtrl() *VoiceSuggestInternalCtrl {
	return &VoiceSuggestInternalCtrl{}
}

// List 成长建议列表。
func (c *VoiceSuggestInternalCtrl) List(ctx context.Context, req *v1.VoiceSuggestListReq) (res *v1.VoiceSuggestListRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	items, err := voice.ListSuggestItems(ctx, deviceNo)
	if err != nil {
		return nil, err
	}
	return &v1.VoiceSuggestListRes{List: items}, nil
}

// Delete 删除一条建议。
func (c *VoiceSuggestInternalCtrl) Delete(ctx context.Context, req *v1.VoiceSuggestDeleteReq) (res *v1.VoiceSuggestDeleteRes, err error) {
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" || req.Id <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "参数无效")
	}
	if err := voice.DeleteSuggestItem(ctx, req.Id, deviceNo); err != nil {
		return nil, err
	}
	return &v1.VoiceSuggestDeleteRes{}, nil
}
