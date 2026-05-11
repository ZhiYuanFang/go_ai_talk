// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// DomainOutbox is the golang structure for table domain_outbox.
type DomainOutbox struct {
	Id          int64  `json:"id"          ` //
	EventId     string `json:"eventId"     ` //
	EventType   string `json:"eventType"   ` //
	RoutingKey  string `json:"routingKey"  ` //
	Payload     string `json:"payload"     ` //
	Status      string `json:"status"      ` //
	Attempts    int    `json:"attempts"    ` //
	LastError   string `json:"lastError"   ` //
	PublishedAt int64  `json:"publishedAt" ` //
	CreatedAt   int64  `json:"createdAt"   ` //
	UpdatedAt   int64  `json:"updatedAt"   ` //
}
