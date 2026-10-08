// =================================================================================
// xiaozhi_mcp_binding 实体：用户绑定的小智 MCP 接入点 token（一人可多音箱）。
// =================================================================================

package entity

// XiaozhiMcpBinding is the golang structure for table xiaozhi_mcp_binding.
type XiaozhiMcpBinding struct {
	Id         int64  `json:"id"         ` // 主键
	WxId       int64  `json:"wxId"       ` // 归属账号 wx 主键
	DeviceNo   string `json:"deviceNo"   ` // 喂养落点宝宝设备号
	McpToken   string `json:"mcpToken"   ` // 小智 MCP 接入点 token（敏感）
	SpeakerMac string `json:"speakerMac" ` // 音箱 MAC（规范化）
	Alias      string `json:"alias"      ` // 用户备注名
	Status     int    `json:"status"     ` // 1=active 0=disabled
	CreatedAt  int64  `json:"createdAt"  ` // 创建 unix 秒
	UpdatedAt  int64  `json:"updatedAt"  ` // 更新 unix 秒
}
