// =================================================================================
// appointment_next DO（DAO Data 操作）。
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AppointmentNext is the golang structure of table appointment_next for DAO operations.
type AppointmentNext struct {
	g.Meta   `orm:"table:appointment_next, do:true"`
	Id       interface{} //
	DeviceNo interface{} //
	EventId  interface{} //
	NextAt   interface{} //
}
