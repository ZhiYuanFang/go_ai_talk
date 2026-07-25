package voice

import (
	"strings"
	"sync"
	"time"
)

// pendingConfirmEntry 待确认意图条目
// 当 Python 意图分析返回 need_confirm=true 时保存该条目，用于后续 confirm/reject 流程
type pendingConfirmEntry struct {
	ConversationID string    // 会话 ID（用于恢复 Python 侧中断的图执行）
	EventName      string    // 事件名称（记录用，便于日志追踪）
	Action         string    // 动作类型（记录用，便于日志追踪）
	CreatedAt      time.Time // 创建时间（用于 60 秒超时懒清理）
}

// pendingConfirmStateStruct 待确认意图状态管理
// 与 pendingChild/pendingQuantity 一致采用内存 map + 读写锁
// 服务重启后状态丢失符合业务预期
type pendingConfirmStateStruct struct {
	mu      sync.RWMutex                   // 读写锁保护并发访问
	entries map[string]*pendingConfirmEntry // key: deviceNo
}

// pendingConfirmState 全局待确认意图状态管理器（包级单例）
var pendingConfirmState = &pendingConfirmStateStruct{
	entries: make(map[string]*pendingConfirmEntry),
}

// set 设置待确认意图（意图分析返回 need_confirm=true 时调用）
func (s *pendingConfirmStateStruct) set(deviceNo string, entry *pendingConfirmEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[deviceNo] = entry
}

// get 获取待确认意图（chatWithResult 开头调用）
// 返回 nil 表示无条目或已超时（60 秒）
func (s *pendingConfirmStateStruct) get(deviceNo string) *pendingConfirmEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[deviceNo]
	if !ok {
		return nil
	}
	// 60 秒超时懒检查
	if time.Since(entry.CreatedAt) > 60*time.Second {
		return nil
	}
	return entry
}

// clear 清除待确认意图（confirm/reject 处理后或回退常规流程时调用）
func (s *pendingConfirmStateStruct) clear(deviceNo string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, deviceNo)
}

// parseConfirmFeedback 解析用户反馈为 confirm 或 reject
// 肯定词命中返回 "confirm"，否定词命中返回 "reject"，无法识别返回 ""
func parseConfirmFeedback(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))

	// 肯定词列表
	confirmWords := []string{"确认", "是的", "对", "好的", "没错", "嗯", "ok", "yes", "是", "对的", "可以"}
	for _, w := range confirmWords {
		if strings.Contains(text, w) {
			return "confirm"
		}
	}

	// 否定词列表
	rejectWords := []string{"取消", "不是", "错", "不对", "没有", "no", "nope", "不", "错的", "不对的"}
	for _, w := range rejectWords {
		if strings.Contains(text, w) {
			return "reject"
		}
	}

	// 无法识别
	return ""
}
