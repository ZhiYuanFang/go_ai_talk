package cash

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// FeatureProduct 功能 SKU。
type FeatureProduct struct {
	ProductCode      string `json:"productCode"`
	FeatureId        string `json:"featureId"`
	GrantKind        string `json:"grantKind"`
	GrantQuantity    int    `json:"grantQuantity"`
	PriceFen         int    `json:"priceFen"`
	OriginalPriceFen int    `json:"originalPriceFen"`
	DurationDays     int    `json:"durationDays"`
	AppleProductId   string `json:"appleProductId"`
	Status           int    `json:"status"`
}

type featureProductDB struct {
	ProductCode      string `json:"product_code"`
	FeatureId        string `json:"feature_id"`
	GrantKind        string `json:"grant_kind"`
	GrantQuantity    int    `json:"grant_quantity"`
	PriceFen         int    `json:"price_fen"`
	OriginalPriceFen int    `json:"original_price_fen"`
	DurationDays     int    `json:"duration_days"`
	AppleProductId   string `json:"apple_product_id"`
	Status           int    `json:"status"`
}

func mapFeatureProduct(r featureProductDB) *FeatureProduct {
	return &FeatureProduct{
		ProductCode: r.ProductCode, FeatureId: r.FeatureId, GrantKind: r.GrantKind,
		GrantQuantity: r.GrantQuantity, PriceFen: r.PriceFen, OriginalPriceFen: r.OriginalPriceFen,
		DurationDays: r.DurationDays, AppleProductId: r.AppleProductId, Status: r.Status,
	}
}

// GetActiveFeatureProduct 按 product_code 取启用功能 SKU。
func GetActiveFeatureProduct(ctx context.Context, productCode string) (*FeatureProduct, error) {
	productCode = strings.TrimSpace(productCode)
	if productCode == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "productCode 不能为空")
	}
	var r featureProductDB
	err := g.DB().Model("feature_product").Ctx(ctx).
		Where("product_code", productCode).Where("status", 1).Scan(&r)
	if err != nil {
		return nil, err
	}
	if r.ProductCode == "" {
		return nil, gerror.NewCode(gcode.CodeNotFound, "功能商品不存在或已停用")
	}
	return mapFeatureProduct(r), nil
}

// GetFeatureProductByAppleID 按 Apple productId 查找启用功能 SKU。
func GetFeatureProductByAppleID(ctx context.Context, applePID string) (*FeatureProduct, error) {
	applePID = strings.TrimSpace(applePID)
	if applePID == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "apple productId 为空")
	}
	var r featureProductDB
	err := g.DB().Model("feature_product").Ctx(ctx).
		Where("apple_product_id", applePID).Where("status", 1).Scan(&r)
	if err != nil {
		return nil, err
	}
	if r.ProductCode == "" {
		return nil, gerror.NewCode(gcode.CodeNotFound, "未找到功能 Apple 商品映射")
	}
	return mapFeatureProduct(r), nil
}

// FeatureOrder 功能订单行。
type FeatureOrder struct {
	Id              int64  `json:"id"`
	OrderNo         string `json:"orderNo"`
	DeviceNo        string `json:"deviceNo"`
	WxId            int64  `json:"wxId"`
	ProductCode     string `json:"productCode"`
	Channel         string `json:"channel"`
	AmountFen       int    `json:"amountFen"`
	Status          string `json:"status"`
	ChannelTxnId    string `json:"channelTxnId"`
	AppAccountToken string `json:"appAccountToken,omitempty"`
}

type featureOrderDB struct {
	Id              int64  `json:"id"`
	OrderNo         string `json:"order_no"`
	DeviceNo        string `json:"device_no"`
	WxId            int64  `json:"wx_id"`
	ProductCode     string `json:"product_code"`
	Channel         string `json:"channel"`
	AmountFen       int    `json:"amount_fen"`
	Status          string `json:"status"`
	ChannelTxnId    string `json:"channel_txn_id"`
	AppAccountToken string `json:"app_account_token"`
}

func loadFeatureOrderByNo(ctx context.Context, orderNo string) (*FeatureOrder, error) {
	var r featureOrderDB
	err := g.DB().Model("feature_order").Ctx(ctx).Where("order_no", orderNo).Scan(&r)
	if err != nil {
		return nil, err
	}
	if r.OrderNo == "" {
		return nil, gerror.NewCode(gcode.CodeNotFound, "功能订单不存在")
	}
	return &FeatureOrder{
		Id: r.Id, OrderNo: r.OrderNo, DeviceNo: r.DeviceNo, WxId: r.WxId,
		ProductCode: r.ProductCode, Channel: r.Channel, AmountFen: r.AmountFen,
		Status: r.Status, ChannelTxnId: r.ChannelTxnId, AppAccountToken: r.AppAccountToken,
	}, nil
}

func loadFeatureOrderByChannelTxn(ctx context.Context, channel, txn string) (*FeatureOrder, error) {
	if txn == "" {
		return nil, nil
	}
	var r featureOrderDB
	err := g.DB().Model("feature_order").Ctx(ctx).
		Where("channel", channel).Where("channel_txn_id", txn).Scan(&r)
	if err != nil {
		return nil, err
	}
	if r.OrderNo == "" {
		return nil, nil
	}
	return &FeatureOrder{
		Id: r.Id, OrderNo: r.OrderNo, DeviceNo: r.DeviceNo, WxId: r.WxId,
		ProductCode: r.ProductCode, Channel: r.Channel, AmountFen: r.AmountFen,
		Status: r.Status, ChannelTxnId: r.ChannelTxnId, AppAccountToken: r.AppAccountToken,
	}, nil
}

// loadFeatureOrderByAppAccountToken 按 Apple appAccountToken 查功能订单。
func loadFeatureOrderByAppAccountToken(ctx context.Context, token string) (*FeatureOrder, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, nil
	}
	var r featureOrderDB
	err := g.DB().Model("feature_order").Ctx(ctx).Where("app_account_token", token).Scan(&r)
	if err != nil {
		return nil, err
	}
	if r.OrderNo == "" {
		return nil, nil
	}
	return &FeatureOrder{
		Id: r.Id, OrderNo: r.OrderNo, DeviceNo: r.DeviceNo, WxId: r.WxId,
		ProductCode: r.ProductCode, Channel: r.Channel, AmountFen: r.AmountFen,
		Status: r.Status, ChannelTxnId: r.ChannelTxnId, AppAccountToken: r.AppAccountToken,
	}, nil
}

// CreateFeatureOrder 创建功能订单并返回调起参数。
func CreateFeatureOrder(ctx context.Context, deviceNo string, wxID int64, productCode, channel string) (*CreateOrderResult, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少设备号")
	}
	channel = strings.TrimSpace(channel)
	if channel != ChannelAlipay && channel != ChannelAppleIAP {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "channel 必须为 alipay 或 apple_iap")
	}
	prod, err := GetActiveFeatureProduct(ctx, productCode)
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
		"device_no":    deviceNo,
		"wx_id":        wxID,
		"product_code": prod.ProductCode,
		"channel":      channel,
		"amount_fen":   prod.PriceFen,
		"currency":     "CNY",
		"status":       OrderCreated,
		"created_at":   now,
	}
	var appToken string
	if channel == ChannelAppleIAP {
		appToken, err = newAppAccountToken()
		if err != nil {
			return nil, err
		}
		row["app_account_token"] = appToken
	}
	_, err = g.DB().Model("feature_order").Ctx(ctx).Data(row).Insert()
	if err != nil {
		return nil, err
	}
	out := &CreateOrderResult{
		OrderNo: orderNo, ProductCode: prod.ProductCode, Channel: channel, AmountFen: prod.PriceFen,
	}
	if channel == ChannelAppleIAP {
		out.AppleProductId = prod.AppleProductId
		out.AppAccountToken = appToken
		return out, nil
	}
	// 复用支付宝拼串：构造临时 VIP Product 形状。
	vipShape := &Product{
		ProductCode: prod.ProductCode, Title: prod.FeatureId, PriceFen: prod.PriceFen,
		DurationDays: prod.DurationDays, AppleProductId: prod.AppleProductId,
	}
	orderStr, tip, aErr := BuildAlipayAppPayOrderStr(ctx, orderNo, vipShape)
	if aErr != nil {
		return nil, aErr
	}
	out.AlipayOrderStr = orderStr
	out.PayTip = tip
	return out, nil
}

// FulfillFeaturePaid 功能订单履约（幂等）。
func FulfillFeaturePaid(ctx context.Context, orderNo, channel, channelTxnID string, amountFen int) error {
	orderNo = strings.TrimSpace(orderNo)
	if orderNo == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "orderNo 不能为空")
	}
	if channelTxnID != "" {
		if existed, err := loadFeatureOrderByChannelTxn(ctx, channel, channelTxnID); err != nil {
			return err
		} else if existed != nil && existed.Status == OrderPaid {
			return nil
		}
	}
	order, err := loadFeatureOrderByNo(ctx, orderNo)
	if err != nil {
		return err
	}
	if order.Channel != channel && channel != "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "支付渠道与订单不一致")
	}
	if amountFen > 0 && order.AmountFen != amountFen {
		return gerror.NewCode(gcode.CodeInvalidParameter, "支付金额与订单不一致")
	}
	if order.Status == OrderPaid {
		return nil
	}
	if order.Status != OrderCreated {
		return gerror.NewCode(gcode.CodeInvalidOperation, "订单状态不可支付")
	}
	prod, err := GetActiveFeatureProduct(ctx, order.ProductCode)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	res, err := g.DB().Model("feature_order").Ctx(ctx).
		Where("id", order.Id).Where("status", OrderCreated).
		Data(g.Map{
			"status":         OrderPaid,
			"channel_txn_id": channelTxnID,
			"paid_at":        now,
		}).Update()
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil
	}
	grantKind := prod.GrantKind
	if grantKind == "" {
		grantKind = GrantKindEntitlement
	}
	if prod.FeatureId == FeatureIDPredictionUnlock || grantKind == GrantKindAllowedCountDelta {
		glog.Warningf(ctx, "[cash] prediction_unlock 已下线，拒绝履约 orderNo=%s featureId=%s", order.OrderNo, prod.FeatureId)
		return gerror.NewCode(gcode.CodeInvalidOperation, "预测事项开通数量已下线")
	}
	// 按功能定义选择设备或账号主体（成长轨迹等 user 功能写付款人 wx）。
	subjType, subjKey, sErr := ResolveActivateSubject(ctx, prod.FeatureId, order.DeviceNo, order.WxId)
	if sErr != nil {
		return sErr
	}
	return ActivateFeature(ctx, ActivateFeatureRequest{
		FeatureID:    prod.FeatureId,
		SubjectType:  subjType,
		SubjectKey:   subjKey,
		Channel:      UnlockMethodPayment,
		ChannelRef:   order.OrderNo,
		ActorWxID:    order.WxId,
		GrantKind:    grantKind,
		GrantQty:     prod.GrantQuantity,
		DurationDays: prod.DurationDays,
	})
}

// DispatchFulfillPaid 按订单号分流 VIP / 功能履约（共用支付回调入口）。
func DispatchFulfillPaid(ctx context.Context, orderNo, channel, channelTxnID string, amountFen int) error {
	orderNo = strings.TrimSpace(orderNo)
	// 先查功能订单，再查 VIP。
	var fo featureOrderDB
	_ = g.DB().Model("feature_order").Ctx(ctx).Where("order_no", orderNo).Scan(&fo)
	if fo.OrderNo != "" {
		return FulfillFeaturePaid(ctx, orderNo, channel, channelTxnID, amountFen)
	}
	return FulfillPaid(ctx, orderNo, channel, channelTxnID, amountFen)
}

// EnsureFeatureOrderExists 判断 order_no 是否功能单（供 Apple 分流）。
func EnsureFeatureOrderExists(ctx context.Context, orderNo string) (bool, error) {
	n, err := g.DB().Model("feature_order").Ctx(ctx).Where("order_no", orderNo).Count()
	return n > 0, err
}
