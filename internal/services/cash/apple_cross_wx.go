// apple_cross_wx.go：Apple IAP 渠道交易号跨 wx 拒绝授权。
//
// 业务说明：
// Apple 非消耗/已购商品换 App 账号后常返回同一 transactionId；履约对已 paid 的
// channel_txn 若一律幂等成功，异号接口会空成功、误导客户端。本文件提供统一判断：
// 仅 channel=apple_iap 且已履约单 wx 与目标订单 wx 不同时拒绝，并返回固定中文文案。
//
// 设计思路：挂在 FulfillPaid / FulfillFeaturePaid / FulfillXiaozhiMcpPaid 的 txn 早退分支，
// 覆盖 verify、ASN、DispatchFulfillPaid；支付宝等非 apple_iap 保持原幂等语义。
package cash

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// ErrMsgAppleIAPOnePangbaoAccount Apple 异号复用同一交易时的固定业务文案（App 直接展示）。
const ErrMsgAppleIAPOnePangbaoAccount = "一个 Apple 账号仅支持开通一个胖宝账号"

// checkAppleIAPChannelTxnWx 校验已履约的 Apple 交易号是否可对目标 wx 幂等放行。
//
// 业务逻辑：非 apple_iap → 放行；同 wx → 放行；异 wx → 拒绝。
//
// Args:
//   - channel：支付渠道（与订单 channel 一致）
//   - paidWxID：已 paid 订单上的 wx_id
//   - targetWxID：本次履约目标订单上的 wx_id
//
// Returns: nil 表示调用方应幂等成功（return nil）；非 nil 为业务错误（勿授权益）。
// Side Effects: 无。
func checkAppleIAPChannelTxnWx(channel string, paidWxID, targetWxID int64) error {
	if strings.TrimSpace(channel) != ChannelAppleIAP {
		return nil
	}
	if paidWxID == targetWxID {
		return nil
	}
	return gerror.NewCode(gcode.CodeInvalidOperation, ErrMsgAppleIAPOnePangbaoAccount)
}
