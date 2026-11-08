package historyctrl

import (
	"context"
	"strings"

	"hello/internal/platform/httpmeta"
	"hello/internal/services/history"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// HistoryInternalLiveActivityRefreshReq POST 内部触发统一 Live Activity Refresh（voice pending PUT / MQ 到点）。
type HistoryInternalLiveActivityRefreshReq struct {
	g.Meta   `path:"/history/internal/api/live-activity/refresh" method:"post" tags:"history-internal" summary:"内部：按 deviceNo 刷新统一 Live Activity"`
	DeviceNo string `json:"deviceNo" v:"required"`
}

// HistoryInternalLiveActivityRefreshRes 空 data；扇出失败不抬 5xx（Refresh 内 WARN）。
type HistoryInternalLiveActivityRefreshRes struct{}

// LiveActivityRefresh POST /history/internal/api/live-activity/refresh
// 业务：经 trailing 300ms 合并后重读 pending + 未闭合 history 再 upsert/end；须共享内部密钥。
func (c *HistoryInternalCtrl) LiveActivityRefresh(ctx context.Context, req *HistoryInternalLiveActivityRefreshReq) (*HistoryInternalLiveActivityRefreshRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil || !httpmeta.ValidateInternalSecret(httpmeta.InternalSecretFromRequest(r)) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "内部接口未授权")
	}
	deviceNo := strings.TrimSpace(req.DeviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	// 预测 PUT / 到点连打时合并为一次投影；立即 200，实际扇出在 coalesce 窗后。
	history.ScheduleRefreshBabyLiveActivity(deviceNo)
	return &HistoryInternalLiveActivityRefreshRes{}, nil
}
