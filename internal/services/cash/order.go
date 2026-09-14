package cash

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Order VIP 订单行。
type Order struct {
	Id              int64  `json:"id"`
	OrderNo         string `json:"orderNo"`
	WxId            int64  `json:"wxId"`
	ProductCode     string `json:"productCode"`
	Channel         string `json:"channel"`
	AmountFen       int    `json:"amountFen"`
	Currency        string `json:"currency"`
	Status          string `json:"status"`
	ChannelTxnId    string `json:"channelTxnId"`
	AppAccountToken string `json:"appAccountToken,omitempty"` // Apple StoreKit UUID；仅 apple_iap
	CreatedAt       int64  `json:"createdAt"`
	PaidAt          int64  `json:"paidAt"`
}

// CreateOrderResult 建单响应（含渠道调起字段）。
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

// CreateOrder 创建 created 订单并返回调起参数。
func CreateOrder(ctx context.Context, wxID int64, productCode, channel string) (*CreateOrderResult, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	channel = strings.TrimSpace(channel)
	if channel != ChannelAlipay && channel != ChannelAppleIAP {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "channel 必须为 alipay 或 apple_iap")
	}
	prod, err := GetActiveProduct(ctx, productCode)
	if err != nil {
		return nil, err
	}
	orderNo, err := newOrderNo()
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	row := g.Map{
		"order_no":     orderNo,
		"wx_id":        wxID,
		"product_code": prod.ProductCode,
		"channel":      channel,
		"amount_fen":   prod.PriceFen,
		"currency":     "CNY",
		"status":       OrderCreated,
		"created_at":   now,
	}
	// Apple 渠道：生成 UUID 写入 app_account_token，供 StoreKit / ASN 关联（不可用 orderNo）。
	var appToken string
	if channel == ChannelAppleIAP {
		appToken, err = newAppAccountToken()
		if err != nil {
			return nil, err
		}
		row["app_account_token"] = appToken
	}
	_, err = g.DB().Model("vip_order").Ctx(ctx).Data(row).Insert()
	if err != nil {
		return nil, err
	}
	out := &CreateOrderResult{
		OrderNo:     orderNo,
		ProductCode: prod.ProductCode,
		Channel:     channel,
		AmountFen:   prod.PriceFen,
	}
	switch channel {
	case ChannelAppleIAP:
		out.AppleProductId = prod.AppleProductId
		out.AppAccountToken = appToken
		if out.AppleProductId == "" {
			out.PayTip = "未在开通功能管理配置 VIP 的 Apple 商品 ID，请在 ASC 建商品后填写"
		}
	case ChannelAlipay:
		str, tip, aErr := BuildAlipayAppPayOrderStr(ctx, orderNo, prod)
		if aErr != nil {
			return nil, aErr
		}
		out.AlipayOrderStr = str
		out.PayTip = tip
	}
	return out, nil
}

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

func loadOrderByNo(ctx context.Context, orderNo string) (*Order, error) {
	orderNo = strings.TrimSpace(orderNo)
	if orderNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "orderNo 不能为空")
	}
	one, err := g.DB().Model("vip_order").Ctx(ctx).Where("order_no", orderNo).Limit(1).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotFound, "订单不存在")
	}
	return scanOrder(one), nil
}

func loadOrderByChannelTxn(ctx context.Context, channel, txnID string) (*Order, error) {
	txnID = strings.TrimSpace(txnID)
	if txnID == "" {
		return nil, nil
	}
	one, err := g.DB().Model("vip_order").Ctx(ctx).
		Where("channel", channel).Where("channel_txn_id", txnID).Limit(1).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, nil
	}
	return scanOrder(one), nil
}

// loadVipOrderByAppAccountToken 按 Apple appAccountToken 查 VIP 订单。
func loadVipOrderByAppAccountToken(ctx context.Context, token string) (*Order, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, nil
	}
	one, err := g.DB().Model("vip_order").Ctx(ctx).Where("app_account_token", token).Limit(1).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, nil
	}
	return scanOrder(one), nil
}

func scanOrder(one gdb.Record) *Order {
	return &Order{
		Id:              one["id"].Int64(),
		OrderNo:         one["order_no"].String(),
		WxId:            one["wx_id"].Int64(),
		ProductCode:     one["product_code"].String(),
		Channel:         one["channel"].String(),
		AmountFen:       one["amount_fen"].Int(),
		Currency:        one["currency"].String(),
		Status:          one["status"].String(),
		ChannelTxnId:    one["channel_txn_id"].String(),
		AppAccountToken: one["app_account_token"].String(),
		CreatedAt:       one["created_at"].Int64(),
		PaidAt:          one["paid_at"].Int64(),
	}
}
