package v1

import "github.com/gogf/gf/v2/frame/g"

// DeviceInternalXiaozhiMcpBindingListReq 内部：全量 active 小智绑定（含完整 token）。
// 须经 InternalSecretMiddleware；供 xiaozhi-mcp-service 启动与 reconcile。
type DeviceInternalXiaozhiMcpBindingListReq struct {
	g.Meta `path:"/device/internal/api/xiaozhi-mcp/bindings" method:"get" tags:"device" summary:"内部-小智绑定全量"`
}

// DeviceInternalXiaozhiMcpBindingItem 内部绑定项。
type DeviceInternalXiaozhiMcpBindingItem struct {
	Id       int64  `json:"id"`
	WxId     int64  `json:"wxId"`
	DeviceNo string `json:"deviceNo"`
	McpToken string `json:"mcpToken"`
	Alias    string `json:"alias"`
	Status   int    `json:"status"`
}

// DeviceInternalXiaozhiMcpBindingListRes 全量列表。
type DeviceInternalXiaozhiMcpBindingListRes struct {
	List []DeviceInternalXiaozhiMcpBindingItem `json:"list"`
}
