package gatewayapp

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/os/glog"
)

// MerchantSSEConn 商户 SSE：一条连接订阅多个 device_no。
type MerchantSSEConn struct {
	*HistorySSEConn
	merchantID int64
	mu         sync.RWMutex
	devices    map[string]struct{}
}

// NewMerchantSSEConn 包装已写好 SSE 头的响应。
func NewMerchantSSEConn(base *HistorySSEConn, merchantID int64, deviceNos []string) *MerchantSSEConn {
	m := make(map[string]struct{}, len(deviceNos))
	for _, d := range deviceNos {
		if d == "" {
			continue
		}
		m[d] = struct{}{}
	}
	return &MerchantSSEConn{
		HistorySSEConn: base,
		merchantID:     merchantID,
		devices:        m,
	}
}

// SetDevices 热更新订阅集合（分发/新建宝宝后可由客户端重连或后续扩展推送触发）。
func (c *MerchantSSEConn) SetDevices(deviceNos []string) {
	m := make(map[string]struct{}, len(deviceNos))
	for _, d := range deviceNos {
		if d != "" {
			m[d] = struct{}{}
		}
	}
	c.mu.Lock()
	c.devices = m
	c.mu.Unlock()
}

func (c *MerchantSSEConn) hasDevice(deviceNo string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.devices[deviceNo]
	return ok
}

// MerchantSSEHub 按 merchant 连接广播；按 device 过滤。
type MerchantSSEHub struct {
	mu    sync.RWMutex
	conns map[*MerchantSSEConn]struct{}
}

var defaultMerchantSSEHub = &MerchantSSEHub{conns: make(map[*MerchantSSEConn]struct{})}

// MerchantSSE 进程内商户 SSE Hub。
func MerchantSSE() *MerchantSSEHub {
	return defaultMerchantSSEHub
}

// Register 登记商户 SSE 连接。
func (h *MerchantSSEHub) Register(conn *MerchantSSEConn) {
	if conn == nil {
		return
	}
	h.mu.Lock()
	h.conns[conn] = struct{}{}
	h.mu.Unlock()
}

// Unregister 移除连接。
func (h *MerchantSSEHub) Unregister(conn *MerchantSSEConn) {
	if conn == nil {
		return
	}
	h.mu.Lock()
	delete(h.conns, conn)
	h.mu.Unlock()
}

// BroadcastDevice 向订阅了该 deviceNo 的商户连接推送（data 为 Redis 原文，前端可解析 device_no）。
func (h *MerchantSSEHub) BroadcastDevice(ctx context.Context, deviceNo, payload string) {
	if deviceNo == "" {
		return
	}
	h.mu.RLock()
	list := make([]*MerchantSSEConn, 0, len(h.conns))
	for c := range h.conns {
		list = append(list, c)
	}
	h.mu.RUnlock()
	for _, c := range list {
		if !c.hasDevice(deviceNo) {
			continue
		}
		if err := c.WriteEvent("history", payload); err != nil {
			glog.Warningf(ctx, "[gateway-merchant-sse] 下发失败 merchantId=%d deviceNo=%s err=%v",
				c.merchantID, deviceNo, err)
		}
	}
}
