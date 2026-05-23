package device

import (
	"fmt"
	"strings"
)

// 事件主档 event_type 枚举（与 ai_voice_device.event.event_type 一致）。
const (
	EventTypeNumber = "number" // 计数
	EventTypeTime   = "time"   // 计时
	EventTypeOne    = "one"    // 一次性
)

// ValidateEventType 校验事件类型；创建/更新时 MUST 为合法枚举。
func ValidateEventType(eventType string) error {
	t := strings.TrimSpace(strings.ToLower(eventType))
	switch t {
	case EventTypeNumber, EventTypeTime, EventTypeOne:
		return nil
	default:
		return fmt.Errorf("eventType 须为 number、time 或 one")
	}
}

// NormalizeEventType 规范化事件类型；非法或空时默认 time（对齐原「计时」主路径）。
func NormalizeEventType(eventType string) string {
	t := strings.TrimSpace(strings.ToLower(eventType))
	switch t {
	case EventTypeNumber, EventTypeTime, EventTypeOne:
		return t
	default:
		return EventTypeTime
	}
}
