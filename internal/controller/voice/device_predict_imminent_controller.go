package voicectrl

import (
	"context"

	v1 "hello/api/v1"
	"hello/internal/platform/httpmeta"
	"hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// DevicePredictImminentController 预测临近待办同步（宿主 voice-service，经 gateway /device/api/predict/*）。
type DevicePredictImminentController struct{}

// PendingPut PUT /device/api/predict/imminent/pending — 全量替换 Redis 待办并投递延时 MQ。
// 客户端须在预测更新后调用；离线收到推送需先注册 /ucg/app/api/push/register。
func (c *DevicePredictImminentController) PendingPut(ctx context.Context, req *v1.DevicePredictImminentPendingPutReq) (res *v1.DevicePredictImminentPendingPutRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少请求上下文")
	}
	wxID := httpmeta.ParseHeaderWxID(r.GetHeader(httpmeta.HeaderInternalWxId))
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	items := make([]voice.PredictImminentPendingItem, 0, len(req.Events))
	for _, e := range req.Events {
		items = append(items, voice.PredictImminentPendingItem{
			EventId: e.EventId,
			NextAt:  e.NextAt,
			Title:   e.Title,
		})
	}
	n, err := voice.SyncPredictImminentPending(ctx, wxID, req.DeviceNo, items)
	if err != nil {
		return nil, err
	}
	return &v1.DevicePredictImminentPendingPutRes{Count: n}, nil
}
