// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgChatMessageOutbox is the golang structure for table ucg_chat_message_outbox.
type UcgChatMessageOutbox struct {
	Id             uint64 `json:"id"             ` //
	ConversationId uint64 `json:"conversationId" ` //
	Payload        string `json:"payload"        ` // ChatMessage JSON
	Status         string `json:"status"         ` // pending|done|failed
	Attempts       uint   `json:"attempts"       ` //
	LastError      string `json:"lastError"      ` //
	CreatedAt      int64  `json:"createdAt"      ` //
}
