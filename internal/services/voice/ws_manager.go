package voice

import (
	"strings"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
)

// WSManager 管理设备与 WS 连接映射。
type WSManager struct {
	mu      sync.Mutex
	clients map[string]*ghttp.WebSocket
}

var audioWSManager = &WSManager{
	clients: make(map[string]*ghttp.WebSocket),
}

// VoiceWSManager 返回全局 WS 管理器。
func VoiceWSManager() *WSManager {
	return audioWSManager
}

// Register 注册设备连接；若已有连接则返回被替换连接。
func (m *WSManager) Register(deviceNo string, ws *ghttp.WebSocket) (replaced *ghttp.WebSocket) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" || ws == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	replaced = m.clients[deviceNo]
	m.clients[deviceNo] = ws
	return replaced
}

// Unregister 注销设备连接。
func (m *WSManager) Unregister(deviceNo string, ws *ghttp.WebSocket) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	cur, ok := m.clients[deviceNo]
	if !ok {
		return
	}
	if ws == nil || cur == ws {
		delete(m.clients, deviceNo)
	}
}

// Get 按设备号获取当前连接。
func (m *WSManager) Get(deviceNo string) (*ghttp.WebSocket, bool) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	ws, ok := m.clients[deviceNo]
	return ws, ok
}
