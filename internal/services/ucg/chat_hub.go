package ucg

import (
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
)

// ChatHub 按 wxId 维护 WS 连接，用于 message_delivered / audit_failed 推送。
type ChatHub struct {
	mu    sync.RWMutex
	conns map[int64]map[*ghttp.WebSocket]struct{}
}

var defaultChatHub = &ChatHub{conns: make(map[int64]map[*ghttp.WebSocket]struct{})}

// ChatWSHub 返回进程内聊天 WS Hub。
func ChatWSHub() *ChatHub { return defaultChatHub }

func (h *ChatHub) Register(wxID int64, ws *ghttp.WebSocket) {
	if wxID <= 0 || ws == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[wxID] == nil {
		h.conns[wxID] = make(map[*ghttp.WebSocket]struct{})
	}
	h.conns[wxID][ws] = struct{}{}
}

func (h *ChatHub) Unregister(wxID int64, ws *ghttp.WebSocket) {
	if wxID <= 0 || ws == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	m := h.conns[wxID]
	if m == nil {
		return
	}
	delete(m, ws)
	if len(m) == 0 {
		delete(h.conns, wxID)
	}
}

func (h *ChatHub) SendJSON(wxID int64, payload interface{}) {
	h.mu.RLock()
	m := h.conns[wxID]
	conns := make([]*ghttp.WebSocket, 0, len(m))
	for ws := range m {
		conns = append(conns, ws)
	}
	h.mu.RUnlock()
	for _, ws := range conns {
		_ = ws.WriteJSON(payload)
	}
}
