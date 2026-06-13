// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgChatMessage is the golang structure for table ucg_chat_message.
type UcgChatMessage struct {
	Id                uint64 `json:"id"                ` // 会话内消息序号，与 Redis seq 一致
	ConversationId    uint64 `json:"conversationId"    ` //
	ClientMsgId       string `json:"clientMsgId"       ` // 客户端幂等 ID，空表示无
	SenderWxId        uint64 `json:"senderWxId"        ` //
	Content           string `json:"content"           ` //
	ImageKey          string `json:"imageKey"          ` //
	VideoKey          string `json:"videoKey"          ` //
	MediaCdnUrl       string `json:"mediaCdnUrl"       ` //
	CreatedAt         int64  `json:"createdAt"         ` //
	Status            string `json:"status"            ` //
	AuditStatus       string `json:"auditStatus"       ` // pending|approved|rejected|moderation_failed
	AuditVersion      int    `json:"auditVersion"      ` // 审核轮次
	ModerationVerdict int    `json:"moderationVerdict" ` // 0=未审 1=pass 2=reject
	ModerationReason  string `json:"moderationReason"  ` //
	ModerationAt      int64  `json:"moderationAt"      ` //
	ApplyAttempts     int    `json:"applyAttempts"     ` //
	ApplyFailedAt     int64  `json:"applyFailedAt"     ` //
	RejectReason      string `json:"rejectReason"      ` //
}
