// =================================================================================
// xiaozhi_mcp_binding DO（DAO Data 操作）。
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// XiaozhiMcpBinding is the golang structure of table xiaozhi_mcp_binding for DAO operations.
type XiaozhiMcpBinding struct {
	g.Meta     `orm:"table:xiaozhi_mcp_binding, do:true"`
	Id         interface{} //
	WxId       interface{} //
	DeviceNo   interface{} //
	McpToken   interface{} //
	SpeakerMac interface{} //
	Alias      interface{} //
	Status     interface{} //
	CreatedAt  interface{} //
	UpdatedAt  interface{} //
}
