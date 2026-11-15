// Admin 手工授功能：补付款未开通与推广赠送。
//
// 业务：复用 feature_order（channel=admin、amount=0、status=paid）落审查；
// grant_reason 必填；权益续期语义与支付一致。VIP 手工授已下线。
package cash

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	// grantReasonMaxRunes 授权理由最大字符数（与 VARCHAR(256) 对齐，按 rune 计）。
	grantReasonMaxRunes = 200
)

// AdminGrantFeatureResult 授功能成功结果。
type AdminGrantFeatureResult struct {
	OrderNo   string `json:"orderNo"`
	FeatureId string `json:"featureId"`
}

// normalizeGrantReason 校验并规范化授权理由（必填）。
func normalizeGrantReason(reason string) (string, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "授权理由必填")
	}
	if utf8.RuneCountInString(reason) > grantReasonMaxRunes {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "授权理由过长")
	}
	return reason, nil
}

func newAdminOrderNo(prefix string) (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s%s", prefix, time.Now().Format("20060102150405"), hex.EncodeToString(b[:])), nil
}

// AdminGrantFeatureInput 手工授功能入参。
type AdminGrantFeatureInput struct {
	FeatureID     string
	WxID          int64
	DeviceNo      string
	DurationDays  int
	GrantQuantity int
	Reason        string
}

// resolveAdminFeatureProductCode 优先该功能在售 SKU，否则 admin_<featureId>。
func resolveAdminFeatureProductCode(ctx context.Context, featureID string) (string, *FeatureProduct, error) {
	featureID = strings.TrimSpace(featureID)
	one, err := g.DB().Model("feature_product").Ctx(ctx).
		Where("feature_id", featureID).Where("status", 1).
		OrderAsc("product_code").Limit(1).One()
	if err != nil {
		return "", nil, err
	}
	if !one.IsEmpty() {
		var r featureProductDB
		if err = one.Struct(&r); err != nil {
			return "", nil, err
		}
		return r.ProductCode, mapFeatureProduct(r), nil
	}
	return "admin_" + featureID, nil, nil
}

// AdminGrantFeature 手工授功能：写 admin feature_order + ActivateFeature。
func AdminGrantFeature(ctx context.Context, in AdminGrantFeatureInput) (*AdminGrantFeatureResult, error) {
	featureID := strings.TrimSpace(in.FeatureID)
	if featureID == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "featureId 必填")
	}
	if featureID == FeatureIDPredictionUnlock {
		glog.Warningf(ctx, "[cash] prediction_unlock 已下线，拒绝手工授 featureId=%s", featureID)
		return nil, gerror.NewCode(gcode.CodeInvalidOperation, "预测事项开通数量已下线")
	}
	reason, err := normalizeGrantReason(in.Reason)
	if err != nil {
		return nil, err
	}
	subj, err := GetFeatureActivationSubject(ctx, featureID)
	if err != nil {
		return nil, err
	}
	deviceNo := strings.TrimSpace(in.DeviceNo)
	wxID := in.WxID
	var subjectKey string
	switch subj {
	case ActivationSubjectUser:
		if wxID <= 0 {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, "该功能须填写 wxId")
		}
		subjectKey = fmt.Sprintf("%d", wxID)
		deviceNo = "" // 订单 device_no 可空串
	default:
		if deviceNo == "" {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, "该功能须填写 deviceNo")
		}
		subjectKey = deviceNo
	}

	prodCode, prod, err := resolveAdminFeatureProductCode(ctx, featureID)
	if err != nil {
		return nil, err
	}
	grantKind := GrantKindEntitlement
	grantQty := 1
	durationDays := in.DurationDays
	if prod != nil {
		grantKind = strings.TrimSpace(prod.GrantKind)
		if grantKind == "" {
			grantKind = GrantKindEntitlement
		}
		if durationDays <= 0 {
			durationDays = prod.DurationDays
		}
	}
	if grantKind == GrantKindAllowedCountDelta || grantKind != GrantKindEntitlement {
		glog.Warningf(ctx, "[cash] 非权益类手工授已拒绝 featureId=%s kind=%s", featureID, grantKind)
		return nil, gerror.NewCode(gcode.CodeInvalidOperation, "预测条数开通已下线")
	}
	if durationDays < 1 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "durationDays 须≥1")
	}

	orderNo, err := newAdminOrderNo("FAD")
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	txnID := "admin-" + orderNo
	_, err = g.DB().Model("feature_order").Ctx(ctx).Data(g.Map{
		"order_no":       orderNo,
		"device_no":      deviceNo,
		"wx_id":          wxID,
		"product_code":   prodCode,
		"channel":        ChannelAdmin,
		"amount_fen":     0,
		"currency":       "CNY",
		"status":         OrderPaid,
		"channel_txn_id": txnID,
		"grant_reason":   reason,
		"created_at":     now,
		"paid_at":        now,
	}).Insert()
	if err != nil {
		return nil, err
	}

	actErr := ActivateFeature(ctx, ActivateFeatureRequest{
		FeatureID:    featureID,
		SubjectType:  subj,
		SubjectKey:   subjectKey,
		Channel:      UnlockMethodAdmin,
		ChannelRef:   orderNo,
		ActorWxID:    wxID,
		GrantKind:    grantKind,
		GrantQty:     grantQty,
		DurationDays: durationDays,
	})
	if actErr != nil {
		_, _ = g.DB().Model("feature_order").Ctx(ctx).Where("order_no", orderNo).Data(g.Map{
			"status": OrderFailed,
		}).Update()
		glog.Warningf(ctx, "[cash-admin-grant] 功能履约失败 orderNo=%s featureId=%s err=%v", orderNo, featureID, actErr)
		return nil, actErr
	}
	return &AdminGrantFeatureResult{OrderNo: orderNo, FeatureId: featureID}, nil
}
