// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DomainOutbox is the golang structure of table domain_outbox for DAO operations like Where/Data.
type DomainOutbox struct {
	g.Meta      `orm:"table:domain_outbox, do:true"`
	Id          interface{} //
	EventId     interface{} //
	EventType   interface{} //
	RoutingKey  interface{} //
	Payload     interface{} //
	Status      interface{} //
	Attempts    interface{} //
	LastError   interface{} //
	PublishedAt interface{} //
	CreatedAt   interface{} //
	UpdatedAt   interface{} //
}
