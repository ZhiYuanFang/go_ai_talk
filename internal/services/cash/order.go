package cash

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// CreateOrderResult 建单响应（含渠道调起字段）；功能单与小智 MCP 共用。
type CreateOrderResult struct {
	OrderNo         string `json:"orderNo"`
	ProductCode     string `json:"productCode"`
	Channel         string `json:"channel"`
	AmountFen       int    `json:"amountFen"`
	AppleProductId  string `json:"appleProductId,omitempty"`
	AppAccountToken string `json:"appAccountToken,omitempty"` // Apple 购买必带；ASN 反查订单
	// AlipayOrderStr 支付宝 App 支付 orderStr；未配置密钥时为空并带 tip。
	AlipayOrderStr string `json:"alipayOrderStr,omitempty"`
	PayTip         string `json:"payTip,omitempty"`
}

// newOrderNo 生成订单号（功能/小智 MCP 共用；历史前缀 VIP 保留以免运维脚本依赖断裂）。
func newOrderNo() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("VIP%s%s", time.Now().Format("20060102150405"), hex.EncodeToString(b[:])), nil
}

// newAppAccountToken 生成 RFC4122 UUID v4（Apple StoreKit appAccountToken 要求）。
func newAppAccountToken() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // RFC4122 variant
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
