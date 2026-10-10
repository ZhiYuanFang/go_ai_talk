package v1

import "github.com/gogf/gf/v2/frame/g"

// DeviceAppXiaozhiMcpBindingListReq 列出当前用户的小智 MCP 绑定（脱敏）。
type DeviceAppXiaozhiMcpBindingListReq struct {
	g.Meta `path:"/device/app/api/xiaozhi-mcp/bindings" method:"get" tags:"device" summary:"小智音箱绑定列表"`
}

// DeviceAppXiaozhiMcpBindingItem 列表项。
type DeviceAppXiaozhiMcpBindingItem struct {
	Id         int64  `json:"id" dc:"绑定主键"`
	Alias      string `json:"alias" dc:"备注名"`
	TokenMask  string `json:"tokenMask" dc:"脱敏 token"`
	SpeakerMac string `json:"speakerMac" dc:"音箱 MAC（规范化）"`
	DeviceNo   string `json:"deviceNo" dc:"喂养落点宝宝设备号"`
	Status     int    `json:"status" dc:"1=active"`
	Connected  bool   `json:"connected" dc:"是否已连接小智 MCP WebSocket"`
	CreatedAt  int64  `json:"createdAt" dc:"创建 unix 秒"`
	UpdatedAt  int64  `json:"updatedAt" dc:"更新 unix 秒"`
}

// DeviceAppXiaozhiMcpBindingListRes 列表响应。
type DeviceAppXiaozhiMcpBindingListRes struct {
	List []DeviceAppXiaozhiMcpBindingItem `json:"list"`
}

// DeviceAppXiaozhiMcpBindingAddReq 添加小智绑定；deviceNo 由服务端取当前绑机；同 MAC 再添加更新 token。
type DeviceAppXiaozhiMcpBindingAddReq struct {
	g.Meta     `path:"/device/app/api/xiaozhi-mcp/bindings" method:"post" tags:"device" summary:"添加小智音箱绑定"`
	McpToken   string `json:"mcpToken" v:"required" dc:"小智 MCP token；可粘贴完整 wss://…/mcp/?token=xxx，服务端只取 token"`
	SpeakerMac string `json:"speakerMac" v:"required" dc:"音箱 MAC；如 3c:dc:75:fc:7f:c4"`
	Alias      string `json:"alias" v:"required" dc:"备注名（如客厅音箱）"`
}

// DeviceAppXiaozhiMcpBindingAddRes 添加成功。
type DeviceAppXiaozhiMcpBindingAddRes struct {
	Id         int64  `json:"id" dc:"绑定主键"`
	Alias      string `json:"alias"`
	TokenMask  string `json:"tokenMask"`
	SpeakerMac string `json:"speakerMac"`
	DeviceNo   string `json:"deviceNo"`
}

// DeviceAppXiaozhiMcpBindingAliasPutReq 更新备注。
type DeviceAppXiaozhiMcpBindingAliasPutReq struct {
	g.Meta `path:"/device/app/api/xiaozhi-mcp/bindings/{id}/alias" method:"put" tags:"device" summary:"更新小智音箱绑定备注"`
	Id     int64  `json:"id" in:"path" v:"required|min:1" dc:"绑定主键"`
	Alias  string `json:"alias" v:"required" dc:"备注名"`
}

// DeviceAppXiaozhiMcpBindingAliasPutRes 更新成功。
type DeviceAppXiaozhiMcpBindingAliasPutRes struct{}

// DeviceAppXiaozhiMcpBindingDeleteReq 删除绑定。
type DeviceAppXiaozhiMcpBindingDeleteReq struct {
	g.Meta `path:"/device/app/api/xiaozhi-mcp/bindings/{id}" method:"delete" tags:"device" summary:"删除小智音箱绑定"`
	Id     int64 `json:"id" in:"path" v:"required|min:1" dc:"绑定主键"`
}

// DeviceAppXiaozhiMcpBindingDeleteRes 删除成功。
type DeviceAppXiaozhiMcpBindingDeleteRes struct{}

// DeviceAppXiaozhiMcpBindingReconnectReq 强制重连小智 MCP Bridge（不改 DB；不等待 WS 已连通）。
type DeviceAppXiaozhiMcpBindingReconnectReq struct {
	g.Meta `path:"/device/app/api/xiaozhi-mcp/bindings/{id}/reconnect" method:"post" tags:"device" summary:"强制重连小智音箱绑定"`
	Id     int64 `json:"id" in:"path" v:"required|min:1" dc:"绑定主键"`
}

// DeviceAppXiaozhiMcpBindingReconnectRes 已触发重连（客户端应再刷列表看 connected）。
type DeviceAppXiaozhiMcpBindingReconnectRes struct{}
