package voicectrl

import (
	"context"
	"fmt"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/platform/httpmeta"
	"hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// DevicePredictImminentController 预测临近待办同步（宿主 voice-service，经 gateway /device/api/predict/*）。
type DevicePredictImminentController struct{}

// PendingPut PUT /device/api/predict/imminent/pending — 全量替换 Redis 待办并投递延时 MQ。
// 客户端须在预测更新后调用；离线收到推送需先注册 /app/api/push/register。
func (c *DevicePredictImminentController) PendingPut(ctx context.Context, req *v1.DevicePredictImminentPendingPutReq) (res *v1.DevicePredictImminentPendingPutRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少请求上下文")
	}
	wxID := httpmeta.ParseHeaderWxID(r.GetHeader(httpmeta.HeaderInternalWxId))
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}

	// 入口可观测：成功同步本身不打业务日志；用 sync_recv 确认客户端是否上报预测时间。
	// deviceNo 仅记长度，与 fire 侧 [predict-imminent] 风格一致。
	deviceNo := strings.TrimSpace(req.DeviceNo)
	glog.Infof(ctx, "[predict-imminent] sync_recv wxId=%d deviceNoLen=%d rawCount=%d events=%s",
		wxID, len(deviceNo), len(req.Events), formatPredictImminentSyncRecvEvents(req.Events))

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

// formatPredictImminentSyncRecvEvents 把原始 events 压成一行摘要，便于 grep 核对 nextAt。
// Args: events 客户端上报列表（可空）。
// Returns: 形如 [{id:1 nextAt:1710000000 titleLen:3},...]；空列表返回 []。
func formatPredictImminentSyncRecvEvents(events []v1.PredictImminentPendingEventItem) string {
	if len(events) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(events))
	for _, e := range events {
		parts = append(parts, fmt.Sprintf("{id:%d nextAt:%d titleLen:%d}",
			e.EventId, e.NextAt, len(strings.TrimSpace(e.Title))))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
