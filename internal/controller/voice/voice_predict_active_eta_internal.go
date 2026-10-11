package voicectrl

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// VoicePredictActiveEtaInternalCtrl 进行中 ETA 内部读（供 Python；须 InternalSecretMiddleware）。
type VoicePredictActiveEtaInternalCtrl struct{}

// NewVoicePredictActiveEtaInternalCtrl 构造内部读控制器。
func NewVoicePredictActiveEtaInternalCtrl() *VoicePredictActiveEtaInternalCtrl {
	return &VoicePredictActiveEtaInternalCtrl{}
}

// Get GET /voice/internal/api/predict/active-eta?deviceNo= — 返回 Redis 当前列表；缺键空 items。
func (c *VoicePredictActiveEtaInternalCtrl) Get(ctx context.Context, req *v1.VoiceInternalPredictActiveEtaGetReq) (res *v1.VoiceInternalPredictActiveEtaGetRes, err error) {
	_ = c
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	list, err := voice.LoadPredictActiveEta(ctx, deviceNo)
	if err != nil {
		return nil, err
	}
	out := make([]v1.PredictActiveEtaItem, 0, len(list))
	for _, it := range list {
		out = append(out, v1.PredictActiveEtaItem{
			RootEventId: it.RootEventId,
			LeafEventId: it.LeafEventId,
			HistoryId:   it.HistoryId,
			Title:       it.Title,
			StartAt:     it.StartAt,
			ExpectEndAt: it.ExpectEndAt,
		})
	}
	return &v1.VoiceInternalPredictActiveEtaGetRes{Items: out}, nil
}
