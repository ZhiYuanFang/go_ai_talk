package voicectrl

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// VoicePredictImminentInternalCtrl 预测临近待办内部读（供 Python；须 InternalSecretMiddleware）。
type VoicePredictImminentInternalCtrl struct{}

// NewVoicePredictImminentInternalCtrl 构造内部读控制器。
func NewVoicePredictImminentInternalCtrl() *VoicePredictImminentInternalCtrl {
	return &VoicePredictImminentInternalCtrl{}
}

// Get GET /voice/internal/api/predict/imminent/pending?deviceNo= — 返回 Redis 当前列表；缺键空 events。
// 只读，不改推送/MQ 调度。
func (c *VoicePredictImminentInternalCtrl) Get(ctx context.Context, req *v1.VoiceInternalPredictImminentPendingGetReq) (res *v1.VoiceInternalPredictImminentPendingGetRes, err error) {
	_ = c
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	list, err := voice.LoadPredictImminentPending(ctx, deviceNo)
	if err != nil {
		return nil, err
	}
	out := make([]v1.PredictImminentPendingEventItem, 0, len(list))
	for _, it := range list {
		out = append(out, v1.PredictImminentPendingEventItem{
			EventId:           it.EventId,
			NextAt:            it.NextAt,
			Title:             it.Title,
			ExpectKind:        it.ExpectKind,
			ExpectDurationSec: it.ExpectDurationSec,
			ExpectAmount:      it.ExpectAmount,
			ExpectUnit:        it.ExpectUnit,
			ExpectLeafTitle:   it.ExpectLeafTitle,
		})
	}
	return &v1.VoiceInternalPredictImminentPendingGetRes{Events: out}, nil
}
