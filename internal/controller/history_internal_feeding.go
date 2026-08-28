package controller

import (
	"context"
	"strings"

	"hello/internal/services/device"
	"hello/internal/services/history"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// HistoryInternalFeedingDayStatsReq GET 内部按日 history 条数（供 cash UCG 资格）。
type HistoryInternalFeedingDayStatsReq struct {
	g.Meta   `path:"/history/internal/api/feeding-day-stats" method:"get" tags:"history-internal" summary:"内部：设备近 N 上海日每日 history 条数"`
	DeviceNo string `json:"deviceNo" in:"query" v:"required"`
	Days     int    `json:"days" in:"query" d:"14" dc:"统计天数，默认 14，最大 31"`
}

// HistoryInternalFeedingDayStatsRes 内部按日统计 data。
type HistoryInternalFeedingDayStatsRes struct {
	DeviceNo string                   `json:"deviceNo"`
	Days     []history.FeedingDayCount `json:"days"`
}

// HistoryInternalCtrl history 内部契约（须共享密钥）。
type HistoryInternalCtrl struct{}

// FeedingDayStats GET /history/internal/api/feeding-day-stats
func (c *HistoryInternalCtrl) FeedingDayStats(ctx context.Context, req *HistoryInternalFeedingDayStatsReq) (*HistoryInternalFeedingDayStatsRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil || !device.ValidateGatewayInternalSecret(device.GatewayInternalSecretHeaderFromRequest(r)) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "内部接口未授权")
	}
	deviceNo := strings.TrimSpace(req.DeviceNo)
	stats, err := history.GetFeedingDayStats(ctx, deviceNo, req.Days)
	if err != nil {
		return nil, err
	}
	return &HistoryInternalFeedingDayStatsRes{DeviceNo: stats.DeviceNo, Days: stats.Days}, nil
}
