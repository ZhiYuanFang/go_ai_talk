package gatewayapp

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// HistoryWSHub 按 device_no 管理已认证的历史推送连接。
type HistoryWSHub struct {
	mu      sync.RWMutex
	byDev   map[string]map[*ghttp.WebSocket]struct{}
}

var defaultHistoryHub = &HistoryWSHub{byDev: make(map[string]map[*ghttp.WebSocket]struct{})}

// HistoryHub 返回进程内单例 Hub。
func HistoryHub() *HistoryWSHub {
	return defaultHistoryHub
}

// Register 将已通过鉴权的连接加入设备分组。
func (h *HistoryWSHub) Register(deviceNo string, ws *ghttp.WebSocket) {
	if deviceNo == "" || ws == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	m, ok := h.byDev[deviceNo]
	if !ok {
		m = make(map[*ghttp.WebSocket]struct{})
		h.byDev[deviceNo] = m
	}
	m[ws] = struct{}{}
}

// Unregister 连接关闭时移除。
func (h *HistoryWSHub) Unregister(deviceNo string, ws *ghttp.WebSocket) {
	if deviceNo == "" || ws == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if m, ok := h.byDev[deviceNo]; ok {
		delete(m, ws)
		if len(m) == 0 {
			delete(h.byDev, deviceNo)
		}
	}
}

// BroadcastText 向某设备下所有连接广播文本帧（忽略写失败）。
func (h *HistoryWSHub) BroadcastText(ctx context.Context, deviceNo, text string) {
	h.mu.RLock()
	m := h.byDev[deviceNo]
	conns := make([]*ghttp.WebSocket, 0, len(m))
	for ws := range m {
		conns = append(conns, ws)
	}
	h.mu.RUnlock()
	for _, ws := range conns {
		if err := ws.WriteMessage(1, []byte(text)); err != nil {
			glog.Warningf(ctx, "[gateway-app-ws] 下发失败 deviceNo=%s err=%v", deviceNo, err)
		}
	}
}
