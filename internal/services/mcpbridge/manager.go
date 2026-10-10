package mcpbridge

// manager.go：多 Bridge 连接池（单副本进程内）。
//
// 业务：每条 active 小智绑定对应一条出站 MCP Bridge；Upsert/Remove/Reconcile 热更新。
// 约束：部署 replicas=1，避免同 token 多进程拨号。
// 连接态：进程内存 token→connected（真连通=读循环中）；供内部 HTTP 查询。

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
	WxId     int64 // 开通主体；tools/call 校验用
}

// bridgeSession 单条绑定的运行态。
type bridgeSession struct {
	cancel   context.CancelFunc
	id       int64
	token    string
	deviceNo string
	wxId     int64
}

// Manager 维护 token → Bridge 会话与连接态。
type Manager struct {
	mu           sync.Mutex
	sessions     map[string]*bridgeSession // key = mcp token
	connected    map[string]bool           // key = mcp token；true 仅当读循环中
	rootCtx      context.Context           // Bridge 生命周期父 ctx（进程级；禁止用 HTTP 请求 ctx）
	baseURL      string
	reconnectMin time.Duration
	reconnectMax time.Duration
}

// NewManager 构造 Manager。
// rootCtx：进程级 context（如 signal.NotifyContext）；Bridge 挂在此下，随进程退出而停。
// 切勿传入 HTTP 请求 ctx，否则写路径 Upsert 响应结束后桥会被立刻 cancel。
func NewManager(rootCtx context.Context, baseURL string, reconnectMin, reconnectMax time.Duration) *Manager {
	if rootCtx == nil {
		rootCtx = context.Background()
	}
	if reconnectMin <= 0 {
		reconnectMin = defaultReconnectMin
	}
	if reconnectMax <= 0 {
		reconnectMax = defaultReconnectMax
	}
	return &Manager{
		sessions:     make(map[string]*bridgeSession),
		connected:    make(map[string]bool),
		rootCtx:      rootCtx,
		baseURL:      baseURL,
		reconnectMin: reconnectMin,
		reconnectMax: reconnectMax,
	}
}

// Upsert 启动或更新一条 Bridge。
// 同 token 且 deviceNo/wxId 不变：仅刷新 binding id；变化则重建会话。
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
		if cur.deviceNo == deviceNo && cur.wxId == spec.WxId {
			cur.id = spec.Id
			glog.Infof(parent, "[mcp-manager] upsert noop token=%s deviceNo=%s wxId=%d id=%d", maskTokenValue(token), deviceNo, spec.WxId, spec.Id)
			return
		}
		glog.Infof(parent, "[mcp-manager] upsert rebuild token=%s oldDevice=%s newDevice=%s oldWx=%d newWx=%d",
			maskTokenValue(token), cur.deviceNo, deviceNo, cur.wxId, spec.WxId)
		cur.cancel()
		delete(m.sessions, token)
		delete(m.connected, token)
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
			delete(m.connected, token)
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
			delete(m.connected, k)
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
			delete(m.connected, token)
		}
	}
	// 启/更需要的
	for token, spec := range want {
		cur, ok := m.sessions[token]
		if !ok {
			m.startLocked(parent, spec)
			continue
		}
		if cur.deviceNo != spec.DeviceNo || cur.wxId != spec.WxId {
			glog.Infof(parent, "[mcp-manager] reconcile rebuild token=%s", maskTokenValue(token))
			cur.cancel()
			delete(m.sessions, token)
			delete(m.connected, token)
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

// ConnectionStatus 批量查询 token 是否已连通小智（读循环中）。
// 未知 / 无会话 / 重连中 → false。
func (m *Manager) ConnectionStatus(tokens []string) map[string]bool {
	out := make(map[string]bool, len(tokens))
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, raw := range tokens {
		t := strings.TrimSpace(raw)
		if t == "" {
			continue
		}
		out[t] = m.connected[t]
	}
	return out
}

// markConnected 由 Bridge 回调：仅当会话仍存在时写入；否则清理。
func (m *Manager) markConnected(token string, on bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sessions[token]; !ok {
		delete(m.connected, token)
		return
	}
	if on {
		m.connected[token] = true
	} else {
		delete(m.connected, token)
	}
}

// startLocked 在已持锁前提下启动会话。
// logCtx 仅用于日志；Bridge 的 runCtx 必须派生自 m.rootCtx，避免 HTTP Upsert 请求结束连带 cancel。
func (m *Manager) startLocked(logCtx context.Context, spec BindingSpec) {
	token := strings.TrimSpace(spec.McpToken)
	deviceNo := strings.TrimSpace(spec.DeviceNo)
	runCtx, cancel := context.WithCancel(m.rootCtx)
	sess := &bridgeSession{
		cancel:   cancel,
		id:       spec.Id,
		token:    token,
		deviceNo: deviceNo,
		wxId:     spec.WxId,
	}
	m.sessions[token] = sess
	delete(m.connected, token) // 新会话初始未连通
	bridge := NewBridge(m.baseURL, token, deviceNo, spec.WxId, m.reconnectMin, m.reconnectMax)
	tokenCopy := token
	bridge.SetOnConnectionChange(func(on bool) {
		m.markConnected(tokenCopy, on)
	})
	// 明确未开通时懒停：tools 拒答后回调 Remove。
	bridge.SetOnEntitlementDenied(func() {
		m.Remove(context.Background(), spec.Id, tokenCopy)
	})
	glog.Infof(logCtx, "[mcp-manager] start bridge id=%d token=%s deviceNo=%s wxId=%d",
		spec.Id, maskTokenValue(token), deviceNo, spec.WxId)
	go func() {
		if err := bridge.Run(runCtx); err != nil {
			if runCtx.Err() != nil {
				glog.Infof(context.Background(), "[mcp-manager] bridge stopped id=%d err=%v", spec.Id, err)
				return
			}
			glog.Errorf(context.Background(), "[mcp-manager] bridge exited id=%d err=%v", spec.Id, err)
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
