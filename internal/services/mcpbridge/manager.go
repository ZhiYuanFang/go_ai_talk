package mcpbridge

// manager.go：多 Bridge 连接池（单副本进程内）。
//
// 业务：每条 active 小智绑定对应一条出站 MCP Bridge；Upsert/Remove/Reconcile 热更新。
// 约束：部署 replicas=1，避免同 token 多进程拨号。

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/glog"
)

// BindingSpec 一条需要拨号的绑定规格。
type BindingSpec struct {
	Id       int64
	McpToken string
	DeviceNo string
}

// bridgeSession 单条绑定的运行态。
type bridgeSession struct {
	cancel   context.CancelFunc
	id       int64
	token    string
	deviceNo string
}

// Manager 维护 token → Bridge 会话。
type Manager struct {
	mu           sync.Mutex
	sessions     map[string]*bridgeSession // key = mcp token
	baseURL      string
	reconnectMin time.Duration
	reconnectMax time.Duration
}

// NewManager 构造 Manager。
func NewManager(baseURL string, reconnectMin, reconnectMax time.Duration) *Manager {
	if reconnectMin <= 0 {
		reconnectMin = defaultReconnectMin
	}
	if reconnectMax <= 0 {
		reconnectMax = defaultReconnectMax
	}
	return &Manager{
		sessions:     make(map[string]*bridgeSession),
		baseURL:      baseURL,
		reconnectMin: reconnectMin,
		reconnectMax: reconnectMax,
	}
}

// Upsert 启动或更新一条 Bridge。
// 同 token 且 deviceNo 不变：仅刷新 binding id；deviceNo 变化：重建会话。
func (m *Manager) Upsert(parent context.Context, spec BindingSpec) {
	token := strings.TrimSpace(spec.McpToken)
	deviceNo := strings.TrimSpace(spec.DeviceNo)
	if token == "" || deviceNo == "" {
		glog.Warningf(parent, "[mcp-manager] upsert skip empty token/deviceNo id=%d", spec.Id)
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if cur, ok := m.sessions[token]; ok {
		if cur.deviceNo == deviceNo {
			cur.id = spec.Id
			glog.Infof(parent, "[mcp-manager] upsert noop token=%s deviceNo=%s id=%d", maskTokenValue(token), deviceNo, spec.Id)
			return
		}
		// deviceNo 变更：停旧桥再启新桥。
		glog.Infof(parent, "[mcp-manager] upsert rebuild token=%s oldDevice=%s newDevice=%s", maskTokenValue(token), cur.deviceNo, deviceNo)
		cur.cancel()
		delete(m.sessions, token)
	}
	m.startLocked(parent, spec)
}

// Remove 按 token（优先）或 id 停止 Bridge。
func (m *Manager) Remove(ctx context.Context, id int64, token string) {
	token = strings.TrimSpace(token)
	m.mu.Lock()
	defer m.mu.Unlock()
	if token != "" {
		if cur, ok := m.sessions[token]; ok {
			glog.Infof(ctx, "[mcp-manager] remove by token=%s id=%d", maskTokenValue(token), cur.id)
			cur.cancel()
			delete(m.sessions, token)
		}
		return
	}
	if id <= 0 {
		return
	}
	for k, cur := range m.sessions {
		if cur.id == id {
			glog.Infof(ctx, "[mcp-manager] remove by id=%d token=%s", id, maskTokenValue(k))
			cur.cancel()
			delete(m.sessions, k)
			return
		}
	}
}

// Reconcile 与 desired 全量对齐：该连的连、多余的停、deviceNo 变则重建。
func (m *Manager) Reconcile(parent context.Context, desired []BindingSpec) {
	want := make(map[string]BindingSpec, len(desired))
	for _, d := range desired {
		token := strings.TrimSpace(d.McpToken)
		deviceNo := strings.TrimSpace(d.DeviceNo)
		if token == "" || deviceNo == "" {
			continue
		}
		d.McpToken = token
		d.DeviceNo = deviceNo
		want[token] = d
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	// 停掉多余
	for token, cur := range m.sessions {
		if _, ok := want[token]; !ok {
			glog.Infof(parent, "[mcp-manager] reconcile remove token=%s", maskTokenValue(token))
			cur.cancel()
			delete(m.sessions, token)
		}
	}
	// 启/更需要的
	for token, spec := range want {
		cur, ok := m.sessions[token]
		if !ok {
			m.startLocked(parent, spec)
			continue
		}
		if cur.deviceNo != spec.DeviceNo {
			glog.Infof(parent, "[mcp-manager] reconcile rebuild token=%s", maskTokenValue(token))
			cur.cancel()
			delete(m.sessions, token)
			m.startLocked(parent, spec)
			continue
		}
		cur.id = spec.Id
	}
	glog.Infof(parent, "[mcp-manager] reconcile done active=%d desired=%d", len(m.sessions), len(want))
}

// ActiveCount 当前活跃 Bridge 数（观测用）。
func (m *Manager) ActiveCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sessions)
}

// startLocked 在已持锁前提下启动会话。
func (m *Manager) startLocked(parent context.Context, spec BindingSpec) {
	token := strings.TrimSpace(spec.McpToken)
	deviceNo := strings.TrimSpace(spec.DeviceNo)
	runCtx, cancel := context.WithCancel(parent)
	sess := &bridgeSession{
		cancel:   cancel,
		id:       spec.Id,
		token:    token,
		deviceNo: deviceNo,
	}
	m.sessions[token] = sess
	bridge := NewBridge(m.baseURL, token, deviceNo, m.reconnectMin, m.reconnectMax)
	glog.Infof(parent, "[mcp-manager] start bridge id=%d token=%s deviceNo=%s", spec.Id, maskTokenValue(token), deviceNo)
	go func() {
		if err := bridge.Run(runCtx); err != nil {
			if runCtx.Err() != nil {
				glog.Infof(runCtx, "[mcp-manager] bridge stopped id=%d err=%v", spec.Id, err)
				return
			}
			glog.Errorf(runCtx, "[mcp-manager] bridge exited id=%d err=%v", spec.Id, err)
		}
	}()
}

// maskTokenValue 脱敏裸 token（非 URL）。
func maskTokenValue(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 6 {
		return "***"
	}
	return token[:6] + "***"
}
