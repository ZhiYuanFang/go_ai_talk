package service

import (
	"strings"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
)

type voiceWSManager struct {
	mu      sync.Mutex
	clients map[string]*ghttp.WebSocket
}

var audioWSManager = &voiceWSManager{
	clients: make(map[string]*ghttp.WebSocket),
}

func VoiceWSManager() *voiceWSManager {
	return audioWSManager
}

func (m *voiceWSManager) Register(deviceNo string, ws *ghttp.WebSocket) (replaced *ghttp.WebSocket) {
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

func (m *voiceWSManager) Unregister(deviceNo string, ws *ghttp.WebSocket) {
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

func (m *voiceWSManager) Get(deviceNo string) (*ghttp.WebSocket, bool) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	ws, ok := m.clients[deviceNo]
	return ws, ok
}
