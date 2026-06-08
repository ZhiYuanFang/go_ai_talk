package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Feedback is the golang structure of table feedback for DAO operations like Where/Data.
type Feedback struct {
	g.Meta        `orm:"table:feedback, do:true"`
	Id            interface{}
	WxId          interface{}
	Question      interface{}
	OfficialReply interface{}
	Status        interface{}
	CreatedAt     interface{}
	UpdatedAt     interface{}
	RepliedAt     interface{}
}
