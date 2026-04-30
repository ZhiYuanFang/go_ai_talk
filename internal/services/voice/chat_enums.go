package voice

import "strings"

// ChatMode 表示语音对话模式。
type ChatMode string

const (
	ChatModeMaternity ChatMode = "maternity"
	ChatModeCasual    ChatMode = "casual"
)

func (m ChatMode) String() string {
	return string(m)
}

func ParseChatMode(raw string) ChatMode {
	switch ChatMode(strings.TrimSpace(strings.ToLower(raw))) {
	case ChatModeMaternity:
		return ChatModeMaternity
	default:
		return ChatModeCasual
	}
}
