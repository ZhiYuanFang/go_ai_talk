package gatewayapp

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// HistoryWSConn 封装单条历史 WS 连接；写操作串行化，避免 Hub 广播与 handler 回包并发写帧。
type HistoryWSConn struct {
	ws *ghttp.WebSocket
	mu sync.Mutex
}

// NewHistoryWSConn 包装 ghttp WebSocket。
func NewHistoryWSConn(ws *ghttp.WebSocket) *HistoryWSConn {
	return &HistoryWSConn{ws: ws}
}

// ReadMessage 读下一帧（读路径无需加锁）。
func (c *HistoryWSConn) ReadMessage() (messageType int, p []byte, err error) {
	return c.ws.ReadMessage()
}

// WriteJSON 线程安全写 JSON 文本帧。
func (c *HistoryWSConn) WriteJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ws.WriteJSON(v)
}

// WriteText 线程安全写 UTF-8 文本帧。
func (c *HistoryWSConn) WriteText(text string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ws.WriteMessage(1, []byte(text))
}

// HistoryWSHub 按 device_no 管理已认证的历史推送连接。
type HistoryWSHub struct {
	mu    sync.RWMutex
	byDev map[string]map[*HistoryWSConn]struct{}
}

var defaultHistoryHub = &HistoryWSHub{byDev: make(map[string]map[*HistoryWSConn]struct{})}

// HistoryHub 返回进程内单例 Hub。
func HistoryHub() *HistoryWSHub {
	return defaultHistoryHub
}

// Register 将已通过鉴权的连接加入设备分组。
func (h *HistoryWSHub) Register(deviceNo string, conn *HistoryWSConn) {
	if deviceNo == "" || conn == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	m, ok := h.byDev[deviceNo]
	if !ok {
		m = make(map[*HistoryWSConn]struct{})
		h.byDev[deviceNo] = m
	}
	m[conn] = struct{}{}
}

// Unregister 连接关闭时移除。
func (h *HistoryWSHub) Unregister(deviceNo string, conn *HistoryWSConn) {
	if deviceNo == "" || conn == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if m, ok := h.byDev[deviceNo]; ok {
		delete(m, conn)
		if len(m) == 0 {
			delete(h.byDev, deviceNo)
		}
	}
}

// BroadcastText 向某设备下所有连接广播文本帧（忽略写失败）。
func (h *HistoryWSHub) BroadcastText(ctx context.Context, deviceNo, text string) {
	h.mu.RLock()
	m := h.byDev[deviceNo]
	conns := make([]*HistoryWSConn, 0, len(m))
	for conn := range m {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()
	for _, conn := range conns {
		if err := conn.WriteText(text); err != nil {
			glog.Warningf(ctx, "[gateway-app-ws] 下发失败 deviceNo=%s err=%v", deviceNo, err)
		}
	}
}
