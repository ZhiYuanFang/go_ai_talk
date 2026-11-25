package v1

import "github.com/gogf/gf/v2/frame/g"

// —— 商户认证（device-service；JWT 由 gateway 签发）——

type MerchantRegisterReq struct {
	g.Meta    `path:"/device/merchant/api/register" method:"post" tags:"merchant" summary:"商户注册"`
	StoreName string `json:"storeName"`
	Password  string `json:"password"`
}

type MerchantRegisterRes struct {
	MerchantId int64  `json:"merchantId"`
	StoreName  string `json:"storeName"`
}

type MerchantLoginReq struct {
	g.Meta    `path:"/device/merchant/api/login" method:"post" tags:"merchant" summary:"商户登录业务（无 JWT）"`
	StoreName string `json:"storeName"`
	Password  string `json:"password"`
}

type MerchantLoginRes struct {
	MerchantId int64  `json:"merchantId"`
	StoreName  string `json:"storeName"`
}

type MerchantCreateStaffReq struct {
	g.Meta   `path:"/device/merchant/api/staff/create" method:"post" tags:"merchant" summary:"创建 staff 子账户"`
	Account  string `json:"account"`
	Password string `json:"password"`
}

type MerchantCreateStaffRes struct {
	WxId int64 `json:"wxId"`
}

type MerchantCreateBabyReq struct {
	g.Meta   `path:"/device/merchant/api/baby/create" method:"post" tags:"merchant" summary:"商户创建虚拟宝宝"`
	BabyName string `json:"babyName"`
	Birthday int64  `json:"birthday"`
	Sex      int    `json:"sex"`
}

type MerchantCreateBabyRes struct {
	DeviceNo string `json:"deviceNo"`
}

type MerchantListBabiesReq struct {
	g.Meta `path:"/device/merchant/api/baby/list" method:"get" tags:"merchant" summary:"商户宝宝列表"`
}

type MerchantListBabiesRes struct {
	DeviceNos []string `json:"deviceNos"`
}

type MerchantAssignBabyReq struct {
	g.Meta    `path:"/device/merchant/api/baby/assign" method:"post" tags:"merchant" summary:"分发宝宝给 staff"`
	WxId      int64  `json:"wxId"`
	DeviceNo  string `json:"deviceNo"`
	SetFocus  bool   `json:"setFocus"`
}

type MerchantAssignBabyRes struct{}

type MerchantUnassignBabyReq struct {
	g.Meta   `path:"/device/merchant/api/baby/unassign" method:"post" tags:"merchant" summary:"解绑 staff 宝宝"`
	WxId     int64  `json:"wxId"`
	DeviceNo string `json:"deviceNo"`
}

type MerchantUnassignBabyRes struct{}

type MerchantUpdateBabyProfileReq struct {
	g.Meta   `path:"/device/merchant/api/baby/profile" method:"post" tags:"merchant" summary:"商户更新宝宝画像"`
	DeviceNo string `json:"deviceNo"`
	BabyName string `json:"babyName"`
	Birthday int64  `json:"birthday"`
	Sex      int    `json:"sex"`
}

type MerchantUpdateBabyProfileRes struct{}

// —— App 多宝宝 ——

type DeviceBabyListReq struct {
	g.Meta `path:"/device/app/api/user/babies" method:"get" tags:"device" summary:"列出账号成员宝宝"`
}

type DeviceBabyListItem struct {
	DeviceNo string `json:"deviceNo"`
	BabyName string `json:"babyName"`
	Source   string `json:"source"`
	IsFocus  bool   `json:"isFocus"`
}

type DeviceBabyListRes struct {
	Items []DeviceBabyListItem `json:"items"`
}

type DeviceBabyFocusReq struct {
	g.Meta   `path:"/device/app/api/user/babies/focus" method:"post" tags:"device" summary:"切换焦点宝宝"`
	DeviceNo string `json:"deviceNo"`
}

type DeviceBabyFocusRes struct {
	DeviceNo string `json:"deviceNo"`
}

type DeviceBabyCreateReq struct {
	g.Meta   `path:"/device/app/api/user/babies/create" method:"post" tags:"device" summary:"parent 创建虚拟宝宝"`
	BabyName string `json:"babyName"`
	Birthday int64  `json:"birthday"`
	Sex      int    `json:"sex"`
}

type DeviceBabyCreateRes struct {
	DeviceNo string `json:"deviceNo"`
}

type DeviceBabyUnbindReq struct {
	g.Meta   `path:"/device/app/api/user/babies/unbind" method:"post" tags:"device" summary:"客户端解绑（恒拒绝）"`
	DeviceNo string `json:"deviceNo"`
}

type DeviceBabyUnbindRes struct{}

// —— 内部写闸 ——

type DeviceInternalMerchantWriteGateReq struct {
	g.Meta   `path:"/device/internal/api/merchant/history-write-gate" method:"get" tags:"device-internal" summary:"商户欠费禁 create 探活"`
	DeviceNo string `json:"deviceNo" in:"query"`
}

type DeviceInternalMerchantWriteGateRes struct {
	BlockCreate bool  `json:"blockCreate"`
	MerchantId  int64 `json:"merchantId"`
}

// —— cash 商户 ——

type CashMerchantProductReq struct {
	g.Meta `path:"/cash/merchant/api/product" method:"get" tags:"cash-merchant" summary:"商户订阅商品读价"`
}

type CashMerchantProductRes struct {
	ProductCode string `json:"productCode"`
	Title       string `json:"title"`
	PriceFen    int    `json:"priceFen"`
	PeriodDays  int    `json:"periodDays"`
}

type CashMerchantEntitlementReq struct {
	g.Meta `path:"/cash/merchant/api/entitlement" method:"get" tags:"cash-merchant" summary:"商户自身权益状态"`
}

type CashMerchantEntitlementRes struct {
	ExpireAt int64  `json:"expireAt"`
	Source   string `json:"source"`
	Entitled bool   `json:"entitled"`
	Arrears  bool   `json:"arrears"`
}

type CashMerchantSubscribeReq struct {
	g.Meta `path:"/cash/merchant/api/subscribe" method:"post" tags:"cash-merchant" summary:"商户连续订阅建单"`
}

type CashMerchantSubscribeRes struct {
	OrderNo    string `json:"orderNo"`
	AmountFen  int    `json:"amountFen"`
	PayPayload string `json:"payPayload"`
	Tip        string `json:"tip,omitempty"`
}

// CashMerchantSubscribeSimulateReq DEV_BYPASS 下模拟支付宝成功通知履约。
type CashMerchantSubscribeSimulateReq struct {
	g.Meta  `path:"/cash/merchant/api/subscribe/simulate-paid" method:"post" tags:"cash-merchant" summary:"模拟商户订阅支付成功（仅 DEV_BYPASS）"`
	OrderNo string `json:"orderNo"`
}

type CashMerchantSubscribeSimulateRes struct {
	Ok bool `json:"ok"`
}

type CashAdminMerchantListReq struct {
	g.Meta `path:"/cash/admin/api/merchant/entitlements" method:"get" tags:"cash-admin" summary:"Hub 商户权益列表"`
}

type CashAdminMerchantListRes struct {
	Items []CashAdminMerchantItem `json:"items"`
}

type CashAdminMerchantItem struct {
	MerchantId int64  `json:"merchantId"`
	ExpireAt   int64  `json:"expireAt"`
	Source     string `json:"source"`
	Entitled   bool   `json:"entitled"`
	Arrears    bool   `json:"arrears"`
}

type CashAdminMerchantGrantReq struct {
	g.Meta     `path:"/cash/admin/api/merchant/grant" method:"post" tags:"cash-admin" summary:"Hub 绝对授期"`
	MerchantId int64 `json:"merchantId"`
	ExpireAt   int64 `json:"expireAt"`
}

type CashAdminMerchantGrantRes struct {
	ExpireAt int64 `json:"expireAt"`
}

type CashAdminMerchantPriceReq struct {
	g.Meta   `path:"/cash/admin/api/merchant/price" method:"post" tags:"cash-admin" summary:"Hub 改月费"`
	PriceFen int `json:"priceFen"`
}

type CashAdminMerchantPriceRes struct {
	PriceFen int `json:"priceFen"`
}

type CashInternalMerchantEntitledReq struct {
	g.Meta     `path:"/cash/internal/api/merchant/entitled" method:"get" tags:"cash-internal" summary:"商户是否有效权益"`
	MerchantId int64 `json:"merchantId" in:"query"`
}

type CashInternalMerchantEntitledRes struct {
	Entitled bool  `json:"entitled"`
	ExpireAt int64 `json:"expireAt"`
}
