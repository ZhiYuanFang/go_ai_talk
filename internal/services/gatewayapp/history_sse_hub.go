package gatewayapp

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// HistorySSEConn 封装单条历史 SSE 连接；写操作串行化，避免 Hub 广播与心跳并发写。
type HistorySSEConn struct {
	resp *ghttp.Response
	mu   sync.Mutex
}

// NewHistorySSEConn 包装 ghttp Response（须已写入 SSE 响应头）。
func NewHistorySSEConn(resp *ghttp.Response) *HistorySSEConn {
	return &HistorySSEConn{resp: resp}
}

// WriteEvent 写入一条 SSE event/data 帧并 Flush。
func (c *HistorySSEConn) WriteEvent(event, data string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.resp.Writer.Write([]byte("event: " + event + "\ndata: " + data + "\n\n")); err != nil {
		return err
	}
	c.resp.Flush()
	return nil
}

// WriteComment 写入 SSE comment（如心跳 `: ping`）并 Flush。
func (c *HistorySSEConn) WriteComment(text string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.resp.Writer.Write([]byte(": " + text + "\n\n")); err != nil {
		return err
	}
	c.resp.Flush()
	return nil
}

// HistorySSEHub 按 device_no 管理已认证的历史 SSE 推送连接（与 HistoryWSHub 并列，互不共用写路径）。
type HistorySSEHub struct {
	mu    sync.RWMutex
	byDev map[string]map[*HistorySSEConn]struct{}
}

var defaultHistorySSEHub = &HistorySSEHub{byDev: make(map[string]map[*HistorySSEConn]struct{})}

// HistorySSE 返回进程内 SSE Hub 单例。
func HistorySSE() *HistorySSEHub {
	return defaultHistorySSEHub
}

// Register 将已通过鉴权的 SSE 连接加入设备分组。
func (h *HistorySSEHub) Register(deviceNo string, conn *HistorySSEConn) {
	if deviceNo == "" || conn == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	m, ok := h.byDev[deviceNo]
	if !ok {
		m = make(map[*HistorySSEConn]struct{})
		h.byDev[deviceNo] = m
	}
	m[conn] = struct{}{}
}

// Unregister 连接关闭时移除。
func (h *HistorySSEHub) Unregister(deviceNo string, conn *HistorySSEConn) {
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

// Broadcast 向某设备下所有 SSE 连接写入 event:history + Redis 原文 data（忽略写失败）。
func (h *HistorySSEHub) Broadcast(ctx context.Context, deviceNo, payload string) {
	h.mu.RLock()
	m := h.byDev[deviceNo]
	conns := make([]*HistorySSEConn, 0, len(m))
	for conn := range m {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()
	for _, conn := range conns {
		if err := conn.WriteEvent("history", payload); err != nil {
			glog.Warningf(ctx, "[gateway-app-sse] 下发失败 deviceNo=%s err=%v", deviceNo, err)
		}
	}
}
