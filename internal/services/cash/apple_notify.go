package cash

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// appleNotificationOuter ASN V2 signedPayload 解码后外层结构（字段按需）。
type appleNotificationOuter struct {
	NotificationType string `json:"notificationType"`
	Subtype          string `json:"subtype"`
	NotificationUUID string `json:"notificationUUID"`
	Data             struct {
		BundleID              string `json:"bundleId"`
		Environment           string `json:"environment"`
		SignedTransactionInfo string `json:"signedTransactionInfo"`
	} `json:"data"`
}

// HandleAppleNotification 处理 Apple Server Notifications V2（同步履约，对齐支付宝 notify，无 MQ）。
//
// 业务：验签 signedPayload → 解析 notificationType / transaction → 按 appAccountToken 查单
// → 成功类履约；REFUND/REVOKE 退款撤销；其它类型记日志并成功返回（避免苹果无限重试）。
//
// Args: signedPayload — POST body 中的 JWS
// Returns: error 非空时 HTTP 应返回非 2xx 触发苹果重试
// Side Effects: 写订单/权益表；无 RabbitMQ、无 ticker
func HandleAppleNotification(ctx context.Context, signedPayload string) error {
	signedPayload = strings.TrimSpace(signedPayload)
	if signedPayload == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "signedPayload 为空")
	}

	// ASN 权威路径：始终验签，禁止 DEV_BYPASS 跳过。
	raw, err := VerifyAndDecodeAppleJWS(signedPayload)
	if err != nil {
		return err
	}
	var outer appleNotificationOuter
	if err := json.Unmarshal(raw, &outer); err != nil {
		return gerror.WrapCode(gcode.CodeInvalidParameter, err, "ASN payload 解析失败")
	}
	nType := strings.TrimSpace(outer.NotificationType)
	glog.Infof(ctx, "[cash] apple ASN type=%s subtype=%s uuid=%s env=%s",
		nType, outer.Subtype, outer.NotificationUUID, outer.Data.Environment)

	if bid := strings.TrimSpace(os.Getenv("CASH_APPLE_BUNDLE_ID")); bid != "" {
		if b := strings.TrimSpace(outer.Data.BundleID); b != "" && b != bid {
			return gerror.NewCode(gcode.CodeNotAuthorized, "ASN bundleId 不匹配")
		}
	}

	txnJWS := strings.TrimSpace(outer.Data.SignedTransactionInfo)
	if txnJWS == "" {
		glog.Infof(ctx, "[cash] apple ASN skip no signedTransactionInfo type=%s", nType)
		return nil
	}
	claims, err := DecodeAppleJWSPayloadMap(txnJWS)
	if err != nil {
		return err
	}
	txnID := strings.TrimSpace(claims["transactionId"])
	token := strings.TrimSpace(claims["appAccountToken"])
	productID := strings.TrimSpace(claims["productId"])
	if bid := strings.TrimSpace(os.Getenv("CASH_APPLE_BUNDLE_ID")); bid != "" {
		if b := strings.TrimSpace(claims["bundleId"]); b != "" && b != bid {
			return gerror.NewCode(gcode.CodeNotAuthorized, "transaction bundleId 不匹配")
		}
	}

	switch nType {
	case "REFUND", "REVOKE":
		return handleAppleRefund(ctx, token, txnID, productID, nType)
	case "ONE_TIME_CHARGE", "SUBSCRIBED", "DID_RENEW", "OFFER_REDEEMED":
		return handleAppleFulfill(ctx, token, txnID, productID, nType)
	default:
		glog.Infof(ctx, "[cash] apple ASN ignored type=%s txn=%s token=%s product=%s",
			nType, txnID, token, productID)
		return nil
	}
}

// handleAppleFulfill 成功扣款类通知：按 appAccountToken 履约（channel_txn_id=transactionId 幂等）。
func handleAppleFulfill(ctx context.Context, token, txnID, productID, nType string) error {
	if token == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "ASN 缺少 appAccountToken（客户端购买须传入建单 UUID）")
	}
	if txnID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "ASN 缺少 transactionId")
	}
	orderNo, err := resolveOrderNoByAppAccountToken(ctx, token)
	if err != nil {
		return err
	}
	// 已退款订单忽略再次履约，避免非 2xx 导致苹果无限重试。
	if fo, e := loadFeatureOrderByNo(ctx, orderNo); e == nil && fo != nil && fo.Status == OrderRefunded {
		glog.Infof(ctx, "[cash] apple ASN fulfill skip refunded feature orderNo=%s", orderNo)
		return nil
	}
	if vo, e := loadOrderByNo(ctx, orderNo); e == nil && vo != nil && vo.Status == OrderRefunded {
		glog.Infof(ctx, "[cash] apple ASN fulfill skip refunded vip orderNo=%s", orderNo)
		return nil
	}
	glog.Infof(ctx, "[cash] apple ASN fulfill type=%s orderNo=%s txn=%s product=%s", nType, orderNo, txnID, productID)
	// amountFen=0：跳过金额比对（Apple 通知不以人民币「分」回传）
	return DispatchFulfillPaid(ctx, orderNo, ChannelAppleIAP, txnID, 0)
}

// handleAppleRefund 退款/撤销：标记订单 refunded 并回退该笔权益（幂等）。
func handleAppleRefund(ctx context.Context, token, txnID, productID, nType string) error {
	if token == "" && txnID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "REFUND 缺少 appAccountToken 与 transactionId")
	}
	var vip *Order
	var feat *FeatureOrder
	var err error
	if token != "" {
		vip, err = loadVipOrderByAppAccountToken(ctx, token)
		if err != nil {
			return err
		}
		if vip == nil {
			feat, err = loadFeatureOrderByAppAccountToken(ctx, token)
			if err != nil {
				return err
			}
		}
	}
	if vip == nil && feat == nil && txnID != "" {
		vip, err = loadOrderByChannelTxn(ctx, ChannelAppleIAP, txnID)
		if err != nil {
			return err
		}
		if vip == nil {
			feat, err = loadFeatureOrderByChannelTxn(ctx, ChannelAppleIAP, txnID)
			if err != nil {
				return err
			}
		}
	}
	if vip == nil && feat == nil {
		glog.Warningf(ctx, "[cash] apple ASN %s 未找到订单 token=%s txn=%s product=%s", nType, token, txnID, productID)
		return nil
	}
	if vip != nil {
		return refundVipOrder(ctx, vip)
	}
	return refundFeatureOrder(ctx, feat)
}

func resolveOrderNoByAppAccountToken(ctx context.Context, token string) (string, error) {
	if fo, e := loadFeatureOrderByAppAccountToken(ctx, token); e != nil {
		return "", e
	} else if fo != nil {
		return fo.OrderNo, nil
	}
	vo, e := loadVipOrderByAppAccountToken(ctx, token)
	if e != nil {
		return "", e
	}
	if vo == nil {
		return "", gerror.NewCode(gcode.CodeNotFound, "appAccountToken 无对应订单")
	}
	return vo.OrderNo, nil
}

// refundVipOrder 将 VIP 订单标 refunded；若曾 paid 则按商品 duration 回退 expire_at。
func refundVipOrder(ctx context.Context, order *Order) error {
	if order.Status == OrderRefunded {
		return nil
	}
	wasPaid := order.Status == OrderPaid
	res, err := g.DB().Model("vip_order").Ctx(ctx).
		Where("id", order.Id).WhereNot("status", OrderRefunded).
		Data(g.Map{"status": OrderRefunded}).Update()
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil
	}
	if !wasPaid {
		glog.Infof(ctx, "[cash] apple refund vip (unpaid) orderNo=%s", order.OrderNo)
		return nil
	}
	days := ProductDurationD
	if prod, pErr := GetActiveProduct(ctx, order.ProductCode); pErr == nil && prod != nil && prod.DurationDays > 0 {
		days = prod.DurationDays
	}
	if err := ShrinkEntitlement(ctx, order.WxId, days); err != nil {
		return err
	}
	glog.Infof(ctx, "[cash] apple refund vip orderNo=%s wxId=%d shrinkDays=%d", order.OrderNo, order.WxId, days)
	return nil
}

// refundFeatureOrder 将功能订单标 refunded，并撤销该笔授予（权益到期 / 数量回退）。
func refundFeatureOrder(ctx context.Context, order *FeatureOrder) error {
	if order.Status == OrderRefunded {
		return nil
	}
	wasPaid := order.Status == OrderPaid
	res, err := g.DB().Model("feature_order").Ctx(ctx).
		Where("id", order.Id).WhereNot("status", OrderRefunded).
		Data(g.Map{"status": OrderRefunded}).Update()
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil
	}
	if !wasPaid {
		return nil
	}
	if err := RevokeFeatureGrantForOrder(ctx, order); err != nil {
		return err
	}
	glog.Infof(ctx, "[cash] apple refund feature orderNo=%s", order.OrderNo)
	return nil
}

// RevokeFeatureGrantForOrder 按付款订单撤销功能权益（幂等：已过期/已扣减可重复调用）。
func RevokeFeatureGrantForOrder(ctx context.Context, order *FeatureOrder) error {
	prod, err := loadFeatureProductAnyStatus(ctx, order.ProductCode)
	if err != nil {
		return err
	}
	grantKind := prod.GrantKind
	if grantKind == "" {
		grantKind = GrantKindEntitlement
	}
	if prod.FeatureId == FeatureIDPredictionUnlock || grantKind == GrantKindAllowedCountDelta {
		glog.Warningf(ctx, "[cash] prediction_unlock 退款跳过条数回退 orderNo=%s", order.OrderNo)
		return nil
	}
	now := time.Now().Unix()

	featureID := strings.TrimSpace(prod.FeatureId)
	if featureID == "" {
		return nil
	}
	subjType, subjKey, sErr := ResolveActivateSubject(ctx, featureID, order.DeviceNo, order.WxId)
	if sErr != nil {
		return sErr
	}
	if subjType == ActivationSubjectUser {
		wxID := order.WxId
		if wxID <= 0 {
			wxID = parseSubjectWxID(subjKey)
		}
		_, err = g.DB().Model("feature_user_entitlement").Ctx(ctx).
			Where("wx_id", wxID).Where("feature_id", featureID).
			Where("source_ref", order.OrderNo).
			Data(g.Map{"expires_at": now, "updated_at": now}).Update()
		return err
	}
	deviceNo := order.DeviceNo
	if subjKey != "" {
		deviceNo = subjKey
	}
	_, err = g.DB().Model("feature_entitlement").Ctx(ctx).
		Where("device_no", deviceNo).Where("feature_id", featureID).
		Where("source_ref", order.OrderNo).
		Data(g.Map{"expires_at": now, "updated_at": now}).Update()
	if err != nil {
		return err
	}
	invalidateDeviceFeatureCaches(ctx, deviceNo)
	return nil
}

// loadFeatureProductAnyStatus 退款用：含已停用 SKU，避免停售后无法撤销。
func loadFeatureProductAnyStatus(ctx context.Context, productCode string) (*FeatureProduct, error) {
	productCode = strings.TrimSpace(productCode)
	if productCode == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "productCode 不能为空")
	}
	// 商品不存在时空集正常，返回业务 NotFound，禁止裸 ErrNoRows。
	one, err := g.DB().Model("feature_product").Ctx(ctx).Where("product_code", productCode).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotFound, "功能商品不存在")
	}
	var r featureProductDB
	if err = one.Struct(&r); err != nil {
		return nil, err
	}
	return mapFeatureProduct(r), nil
}

func parseSubjectWxID(key string) int64 {
	key = strings.TrimSpace(key)
	var n int64
	for _, c := range key {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int64(c-'0')
	}
	return n
}
