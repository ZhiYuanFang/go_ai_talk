package ucg

import "fmt"

// Redis 聊天键（MVP 永久保留，不设 TTL）。
func redisChatMsgListKey(convID uint64) string {
	return fmt.Sprintf("ucg:chat:conv:%d:msgs", convID)
}

func redisChatUnreadKey(convID uint64, wxID int64) string {
	return fmt.Sprintf("ucg:chat:conv:%d:unread:%d", convID, wxID)
}

func redisChatMsgSeqKey(convID uint64) string {
	return fmt.Sprintf("ucg:chat:conv:%d:seq", convID)
}

func redisChatUserConvKey(wxID int64) string {
	return fmt.Sprintf("ucg:chat:user:%d:conversations", wxID)
}
