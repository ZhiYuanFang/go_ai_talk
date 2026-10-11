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

// DevicePredictActiveEtaController 进行中计时 ETA 同步（宿主 voice-service，经 gateway /device/api/predict/*）。
type DevicePredictActiveEtaController struct{}

// PendingPut PUT /device/api/predict/active-eta/pending — 全量替换 Redis ETA；不发 MQ、不推送。
// 客户端须在进行中计时变化后调用；须登录且 deviceNo 与绑机一致。
func (c *DevicePredictActiveEtaController) PendingPut(ctx context.Context, req *v1.DevicePredictActiveEtaPendingPutReq) (res *v1.DevicePredictActiveEtaPendingPutRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少请求上下文")
	}
	wxID := httpmeta.ParseHeaderWxID(r.GetHeader(httpmeta.HeaderInternalWxId))
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}

	deviceNo := strings.TrimSpace(req.DeviceNo)
	glog.Infof(ctx, "[predict-active-eta] sync_recv wxId=%d deviceNoLen=%d rawCount=%d items=%s",
		wxID, len(deviceNo), len(req.Items), formatPredictActiveEtaSyncRecvItems(req.Items))

	items := make([]voice.PredictActiveEtaItem, 0, len(req.Items))
	for _, e := range req.Items {
		items = append(items, voice.PredictActiveEtaItem{
			RootEventId: e.RootEventId,
			LeafEventId: e.LeafEventId,
			HistoryId:   e.HistoryId,
			Title:       e.Title,
			StartAt:     e.StartAt,
			ExpectEndAt: e.ExpectEndAt,
		})
	}
	n, err := voice.SyncPredictActiveEta(ctx, wxID, req.DeviceNo, items)
	if err != nil {
		return nil, err
	}
	return &v1.DevicePredictActiveEtaPendingPutRes{Count: n}, nil
}

// formatPredictActiveEtaSyncRecvItems 把原始 items 压成一行摘要，便于 grep 联调。
// Args: items 客户端上报列表（可空）。
// Returns: 形如 [{root:1 leaf:2 start:… end:…},...]；空列表返回 []。
func formatPredictActiveEtaSyncRecvItems(items []v1.PredictActiveEtaItem) string {
	if len(items) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(items))
	for _, e := range items {
		parts = append(parts, fmt.Sprintf("{root:%d leaf:%d histLen:%d titleLen:%d start:%d end:%d}",
			e.RootEventId, e.LeafEventId, len(strings.TrimSpace(e.HistoryId)),
			len(strings.TrimSpace(e.Title)), e.StartAt, e.ExpectEndAt))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
