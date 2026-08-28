package cash

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

// AppleVerifyInput App 提交的 IAP 验单载荷。
type AppleVerifyInput struct {
	OrderNo             string `json:"orderNo"`
	TransactionId       string `json:"transactionId"`
	ProductId           string `json:"productId"`
	SignedTransaction   string `json:"signedTransaction"` // JWS，可选；有则解析声明
}

// VerifyAppleIAP 校验 Apple IAP 并履约开通（VIP 或功能订单共用入口）。
// 生产应配置真实验签；CASH_PAYMENT_DEV_BYPASS=1 时仅校验 productId 映射与订单归属。
func VerifyAppleIAP(ctx context.Context, wxID int64, in AppleVerifyInput) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	txn := strings.TrimSpace(in.TransactionId)
	productID := strings.TrimSpace(in.ProductId)
	orderNo := strings.TrimSpace(in.OrderNo)
	if txn == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "transactionId 不能为空")
	}

	if jws := strings.TrimSpace(in.SignedTransaction); jws != "" {
		claims, cErr := decodeJWSPayload(jws)
		if cErr != nil {
			return gerror.WrapCode(gcode.CodeInvalidParameter, cErr, "signedTransaction 解析失败")
		}
		if v := strings.TrimSpace(claims["transactionId"]); v != "" {
			txn = v
		}
		if v := strings.TrimSpace(claims["productId"]); v != "" {
			productID = v
		}
		if bid := strings.TrimSpace(os.Getenv("CASH_APPLE_BUNDLE_ID")); bid != "" {
			if b := strings.TrimSpace(claims["bundleId"]); b != "" && b != bid {
				return gerror.NewCode(gcode.CodeNotAuthorized, "bundleId 不匹配")
			}
		}
		if !paymentDevBypass() {
			glog.Infof(ctx, "[cash] apple verify using JWS payload fields txn=%s product=%s", txn, productID)
		}
	} else if !paymentDevBypass() {
		return gerror.NewCode(gcode.CodeInvalidParameter, "生产环境须提交 signedTransaction（JWS）；开发可设 CASH_PAYMENT_DEV_BYPASS=1")
	}

	// 功能订单优先：已有 orderNo 属于 feature_order，或 Apple ID 命中功能 SKU。
	if orderNo != "" {
		if ok, _ := EnsureFeatureOrderExists(ctx, orderNo); ok {
			fo, err := loadFeatureOrderByNo(ctx, orderNo)
			if err != nil {
				return err
			}
			if fo.WxId != wxID && fo.WxId != 0 {
				return gerror.NewCode(gcode.CodeNotAuthorized, "订单不属于当前账号")
			}
			prod, err := GetActiveFeatureProduct(ctx, fo.ProductCode)
			if err != nil {
				return err
			}
			if expect := strings.TrimSpace(prod.AppleProductId); expect != "" && productID != expect {
				return gerror.NewCode(gcode.CodeInvalidParameter, "productId 与功能商品不匹配")
			}
			return FulfillFeaturePaid(ctx, orderNo, ChannelAppleIAP, txn, prod.PriceFen)
		}
	}
	if fp, err := GetFeatureProductByAppleID(ctx, productID); err == nil && fp != nil {
		if orderNo == "" {
			return gerror.NewCode(gcode.CodeInvalidParameter, "功能 Apple 验单须提供 orderNo")
		}
		return FulfillFeaturePaid(ctx, orderNo, ChannelAppleIAP, txn, fp.PriceFen)
	}

	prod, err := GetActiveProduct(ctx, ProductMonthly19)
	if err != nil {
		return err
	}
	expectPID := strings.TrimSpace(prod.AppleProductId)
	if expectPID == "" {
		expectPID = strings.TrimSpace(os.Getenv("CASH_APPLE_PRODUCT_ID"))
	}
	if expectPID == "" {
		return gerror.NewCode(gcode.CodeInternalError, "未配置 Apple productId（CASH_APPLE_PRODUCT_ID）")
	}
	if productID != expectPID {
		return gerror.NewCode(gcode.CodeInvalidParameter, "productId 与一期 VIP 商品不匹配")
	}

	if orderNo == "" {
		created, cErr := CreateOrder(ctx, wxID, ProductMonthly19, ChannelAppleIAP)
		if cErr != nil {
			return cErr
		}
		orderNo = created.OrderNo
	} else {
		order, oErr := loadOrderByNo(ctx, orderNo)
		if oErr != nil {
			return oErr
		}
		if order.WxId != wxID {
			return gerror.NewCode(gcode.CodeNotAuthorized, "订单不属于当前账号")
		}
		if order.Channel != ChannelAppleIAP {
			return gerror.NewCode(gcode.CodeInvalidParameter, "订单渠道不是 apple_iap")
		}
	}
	return FulfillPaid(ctx, orderNo, ChannelAppleIAP, txn, prod.PriceFen)
}

func decodeJWSPayload(jws string) (map[string]string, error) {
	parts := strings.Split(jws, ".")
	if len(parts) < 2 {
		return nil, gerror.New("JWS 格式非法")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		// 兼容 padding
		raw, err = base64.URLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, err
		}
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		switch t := v.(type) {
		case string:
			out[k] = t
		case float64:
			out[k] = strings.TrimSpace(strings.Split(jsonNumber(t), ".")[0])
		default:
			b, _ := json.Marshal(t)
			out[k] = string(b)
		}
	}
	return out, nil
}

func jsonNumber(f float64) string {
	b, _ := json.Marshal(f)
	return string(b)
}
