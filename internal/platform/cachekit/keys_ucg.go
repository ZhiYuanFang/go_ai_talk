package cachekit

import (
	"fmt"
	"strconv"
	"strings"
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

// UCGFeedGeoKey 已发布且有坐标的帖子 GEO 索引；member=postId，coord=(lng,lat)。
// 无 TTL，下架/删帖时 ZREM；与 MySQL published 状态由写路径同步。
func UCGFeedGeoKey() string {
	return "ucg:feed:geo"
}

// UCGRecommendScoreKey 推荐 baseScore ZSET；member=postId，score=baseScore。
// 无 TTL，下架 ZREM；MQ/reconciler/publish 写路径维护。
func UCGRecommendScoreKey() string {
	return "ucg:recommend:score"
}

// UCGPostSnapshotKey 帖子 Feed 展示 JSON 快照（含 server-only lat/lng）；TTL 见 ucg.feed.snapshotTtlDays。
func UCGPostSnapshotKey(postID uint64) string {
	return fmt.Sprintf("ucg:post:snapshot:%d", postID)
}

// UCGProfileSnapshotKey 作者公开 profile JSON 快照；TTL 见 ucg.feed.snapshotTtlDays。
func UCGProfileSnapshotKey(wxID uint64) string {
	return fmt.Sprintf("ucg:profile:snapshot:%d", wxID)
}

// UCGUserLikedPostsKey 用户点赞 postId SET；无 TTL，like/unlike 写路径维护。
func UCGUserLikedPostsKey(wxID int64) string {
	return fmt.Sprintf("ucg:user:%d:liked-posts", wxID)
}

// UCGPostVoteCountsKey 辩论帖投票计数 Hash；field=left|right，value=票数。无 TTL，VotePost 写路径 HINCRBY 维护。
func UCGPostVoteCountsKey(postID uint64) string {
	return fmt.Sprintf("ucg:post:%d:vote-counts", postID)
}

// UCGUserPostVotesKey 用户辩论帖投票立场 Hash；field=postId，value=left|right。无 TTL，VotePost 写路径 HSET 维护。
func UCGUserPostVotesKey(wxID int64) string {
	return fmt.Sprintf("ucg:user:%d:post-votes", wxID)
}

// UCGFeedSessionKey Feed 分页 session 已下发 postId SET；TTL 见 ucg.feed.sessionTtlMinutes（默认 30min）。
func UCGFeedSessionKey(sessionID string) string {
	return fmt.Sprintf("ucg:feed:session:%s", strings.TrimSpace(sessionID))
}

// UCGFeedIndexWarmLockKey 推荐 Feed 索引冷启动 warm 分布式锁；TTL 见 ucg.feed.indexWarmLockSeconds（默认 60s）。
// 非 session 键；并发 Feed 请求仅一方执行 MySQL→Redis 灌库，其余短退避后读 ZCARD。
func UCGFeedIndexWarmLockKey() string {
	return "ucg:feed:index:warm:lock"
}

// UCGPostCommentsKey 帖子已发布评论 ZSET 索引；score=created_at，member=commentId。
// 无 TTL；删帖/下架时 DEL；publish/delete 写路径 ZADD/ZREM 维护。
func UCGPostCommentsKey(postID uint64) string {
	return fmt.Sprintf("ucg:post:%d:comments", postID)
}

// UCGCommentSnapshotKey 已发布评论 JSON 快照；TTL 见 ucg.feed.snapshotTtlDays。
func UCGCommentSnapshotKey(commentID uint64) string {
	return fmt.Sprintf("ucg:comment:snapshot:%d", commentID)
}
