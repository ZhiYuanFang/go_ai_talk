package cash

// 小智 MCP 永久开通域常量（与 feature_def / ActivateFeature 隔离）。

const (
	// XiaozhiMcpProductCodePerm 小智 MCP 一次性永久买断商品编码（种子固定，Admin 不可改码）。
	XiaozhiMcpProductCodePerm = "xiaozhi_mcp_perm"

	// XiaozhiMcpEntitlementActive 权益有效。
	XiaozhiMcpEntitlementActive = 1
	// XiaozhiMcpEntitlementRevoked 权益已撤销（退款或手工撤）。
	XiaozhiMcpEntitlementRevoked = 0

	// XiaozhiMcpProductOn 商品上架。
	XiaozhiMcpProductOn = 1
	// XiaozhiMcpProductOff 商品下架。
	XiaozhiMcpProductOff = 0

	// XiaozhiMcpUnlockPayment 支付开通。
	XiaozhiMcpUnlockPayment = "payment"
	// XiaozhiMcpUnlockAdmin Hub 手工授。
	XiaozhiMcpUnlockAdmin = "admin"
	// XiaozhiMcpUnlockTrial 试用开通（限时）。
	XiaozhiMcpUnlockTrial = "trial"

	// XiaozhiMcpTrialStatusUnused 试用未领取。
	XiaozhiMcpTrialStatusUnused = "unused"
	// XiaozhiMcpTrialStatusUsed 试用已领取。
	XiaozhiMcpTrialStatusUsed = "used"
)

// IsXiaozhiMcpProductCode 是否为小智 MCP 专用商品编码（履约/退款类型判断）。
func IsXiaozhiMcpProductCode(productCode string) bool {
	return productCode == XiaozhiMcpProductCodePerm
}
