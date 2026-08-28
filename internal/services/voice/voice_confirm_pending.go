package voice

import (
	"sync"
	"time"
)

// pendingConfirmEntry 澄清续聊会话条目
// 业务说明：Python 返回 need_confirm=true 时，Go 仅按设备保存 conversation_id，
// 下一轮统一 AnalyzeIntent / AnalyzeIntentStream 带回该 cid；自由文本解析在 Python 侧完成。
type pendingConfirmEntry struct {
	ConversationID string    // 会话 ID（下一轮 intent 请求带回）
	EventName      string    // 事件名称（可观测，便于日志）
	Action         string    // 动作类型（可观测，便于日志）
	CreatedAt      time.Time // 创建时间（用于 60 秒超时懒清理）
}

// pendingConfirmStateStruct 按设备隔离的 conversation_id 便签
// 与 pendingQuantity 一致采用内存 map + 读写锁；进程重启后丢失符合预期。
type pendingConfirmStateStruct struct {
	mu      sync.RWMutex                    // 读写锁保护并发访问
	entries map[string]*pendingConfirmEntry // key: deviceNo
}

// pendingConfirmState 全局澄清 cid 状态管理器（包级单例）
var pendingConfirmState = &pendingConfirmStateStruct{
	entries: make(map[string]*pendingConfirmEntry),
}

// set 保存澄清 conversation_id（need_confirm=true 时调用）
func (s *pendingConfirmStateStruct) set(deviceNo string, entry *pendingConfirmEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[deviceNo] = entry
}

// get 获取澄清条目；无条目或已超时（60 秒）返回 nil。
// 调用方据此决定是否在 intent 请求中附带 conversation_id；失败路径不应 clear，以便重试续聊。
func (s *pendingConfirmStateStruct) get(deviceNo string) *pendingConfirmEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[deviceNo]
	if !ok {
		return nil
	}
	// 60 秒超时懒检查：过期视为无 pending，不在此删除（下次 set/clear 覆盖）
	if time.Since(entry.CreatedAt) > 60*time.Second {
		return nil
	}
	return entry
}

// clear 清除澄清 conversation_id（need_confirm=false 的成功落地后调用）
func (s *pendingConfirmStateStruct) clear(deviceNo string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, deviceNo)
}

// pendingConversationID 读取设备当前有效的澄清 conversation_id；无则返回空串。
// 仅供统一意图调用路径使用；成长建议/历史问答等独立 AnalyzeIntent MUST NOT 调用本函数。
func pendingConversationID(deviceNo string) string {
	if entry := pendingConfirmState.get(deviceNo); entry != nil {
		return entry.ConversationID
	}
	return ""
}
