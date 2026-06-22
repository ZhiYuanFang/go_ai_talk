package cachekit

import (
	"fmt"
	"strconv"
)

// UCGChatMsgListKey 会话消息 LIST（永久 PERSIST）；MVP 不设 TTL。
func UCGChatMsgListKey(convID uint64) string {
	return fmt.Sprintf("ucg:chat:conv:%d:msgs", convID)
}

// UCGChatUnreadKey 会话未读计数 String。
func UCGChatUnreadKey(convID uint64, wxID int64) string {
	return fmt.Sprintf("ucg:chat:conv:%d:unread:%d", convID, wxID)
}

// UCGChatMsgSeqKey 会话消息自增 seq。
func UCGChatMsgSeqKey(convID uint64) string {
	return fmt.Sprintf("ucg:chat:conv:%d:seq", convID)
}

// UCGChatUserConvKey 用户会话 ZSET 索引。
func UCGChatUserConvKey(wxID int64) string {
	return fmt.Sprintf("ucg:chat:user:%d:conversations", wxID)
}

// UCGChatRebuildLockKey MySQL warm 重建 LIST 分布式锁；TTL 30s。
func UCGChatRebuildLockKey(convID uint64) string {
	return fmt.Sprintf("ucg:chat:conv:%d:rebuild", convID)
}

// UCGProfilePendingSetKey 待 Green 审核资料 wxId 集合。
func UCGProfilePendingSetKey() string {
	return "ucg:green:profile:pending"
}

// UCGProfilePendingDataKey 待审资料 patch JSON；TTL 7 天。
func UCGProfilePendingDataKey(wxID int64) string {
	return "ucg:green:profile:data:" + strconv.FormatInt(wxID, 10)
}

// UCGProfileRejectReasonKey 资料审核拒绝原因；TTL 7 天。
func UCGProfileRejectReasonKey(wxID int64) string {
	return "ucg:green:profile:reject:" + strconv.FormatInt(wxID, 10)
}

// UCGRecommendThrottleKey 推荐分 MQ 节流；TTL 见 LikeThrottleMs。
func UCGRecommendThrottleKey(postID uint64) string {
	return fmt.Sprintf("ucg:recommend:throttle:%d", postID)
}

// UCGIPLocationThrottleKey IP 属地更新节流；TTL 1h。
func UCGIPLocationThrottleKey(wxID int64) string {
	return "ucg:ip_location:throttle:" + strconv.FormatInt(wxID, 10)
}
