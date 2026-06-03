// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgConversationMember is the golang structure for table ucg_conversation_member.
type UcgConversationMember struct {
	Id             uint64 `json:"id"             ` //
	ConversationId uint64 `json:"conversationId" ` //
	WxId           uint64 `json:"wxId"           ` //
	Pinned         int    `json:"pinned"         ` //
	DeletedAt      int64  `json:"deletedAt"      ` // user soft delete
	LastReadMsgId  uint64 `json:"lastReadMsgId"  ` //
	UnreadCount    uint   `json:"unreadCount"    ` //
	UpdatedAt      int64  `json:"updatedAt"      ` // last activity; drives idx_wx_list sort
}
