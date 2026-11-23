package gatewayappctrl

import (
	"strings"
	"time"

	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

const historySSEHeartbeatInterval = 25 * time.Second

// gatewayAppHistorySSE GET /device/app/sse/history — Bearer 由中间件校验；本 handler 校验 deviceNo 对齐后长连接推送。
func gatewayAppHistorySSE(r *ghttp.Request) {
	ctx := r.Context()

	deviceNoFromJWT := strings.TrimSpace(r.Header.Get(gatewayapp.HeaderInternalDeviceNo))
	if deviceNoFromJWT == "" {
		WriteAuthJSON(r, 401, "未绑定设备，无法订阅历史推送")
		return
	}
	want := strings.TrimSpace(r.Get("deviceNo").String())
	if want == "" {
		want = strings.TrimSpace(r.Get("device_no").String())
	}
	if want == "" || want != deviceNoFromJWT {
		WriteAuthJSON(r, 400, "device_no 与 token 不一致")
		return
	}

	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	r.Response.Header().Set("X-Accel-Buffering", "no")
	r.Response.WriteHeader(200)
	r.Response.Flush()

	conn := gatewayapp.NewHistorySSEConn(r.Response)
	gatewayapp.HistorySSE().Register(want, conn)
	defer gatewayapp.HistorySSE().Unregister(want, conn)

	glog.Infof(ctx, "[gateway-app-sse] 已订阅历史推送 deviceNo=%s", want)

	ticker := time.NewTicker(historySSEHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := conn.WriteComment("ping"); err != nil {
				glog.Warningf(ctx, "[gateway-app-sse] 心跳失败 deviceNo=%s err=%v", want, err)
				return
			}
		}
	}
}
