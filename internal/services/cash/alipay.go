package cash

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

// BuildAlipayAppPayOrderStr 组装支付宝 App 支付 orderStr；密钥未配置时返回空串与提示（订单已创建）。
func BuildAlipayAppPayOrderStr(ctx context.Context, orderNo string, prod *Product) (orderStr string, tip string, err error) {
	appID := strings.TrimSpace(os.Getenv("CASH_ALIPAY_APP_ID"))
	privPEM := strings.TrimSpace(os.Getenv("CASH_ALIPAY_PRIVATE_KEY"))
	notifyURL := strings.TrimSpace(os.Getenv("CASH_ALIPAY_NOTIFY_URL"))
	if appID == "" || privPEM == "" {
		return "", "未配置 CASH_ALIPAY_APP_ID / CASH_ALIPAY_PRIVATE_KEY，订单已创建但无法调起支付宝", nil
	}
	if notifyURL == "" {
		return "", "未配置 CASH_ALIPAY_NOTIFY_URL", nil
	}
	biz := fmt.Sprintf(`{"out_trade_no":"%s","total_amount":"%.2f","subject":"%s","product_code":"QUICK_MSECURITY_PAY"}`,
		orderNo, float64(prod.PriceFen)/100.0, escapeJSON(prod.Title))
	params := map[string]string{
		"app_id":      appID,
		"method":      "alipay.trade.app.pay",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  notifyURL,
		"biz_content": biz,
	}
	content := alipaySignContent(params)
	sign, err := rsa2Sign(privPEM, content)
	if err != nil {
		return "", "", gerror.WrapCode(gcode.CodeInternalError, err, "支付宝私钥签名失败")
	}
	params["sign"] = sign
	return alipayEncodeOrderStr(params), "", nil
}

// HandleAlipayNotify 处理异步通知；成功返回 "success"。
func HandleAlipayNotify(ctx context.Context, form map[string]string) (string, error) {
	if !paymentDevBypass() {
		pub := strings.TrimSpace(os.Getenv("CASH_ALIPAY_PUBLIC_KEY"))
		if pub == "" {
			return "failure", gerror.NewCode(gcode.CodeNotAuthorized, "未配置 CASH_ALIPAY_PUBLIC_KEY")
		}
		if err := verifyAlipayNotifySign(pub, form); err != nil {
			glog.Warningf(ctx, "[cash] alipay notify sign fail err=%v", err)
			return "failure", err
		}
	} else {
		glog.Warningf(ctx, "[cash] alipay notify DEV_BYPASS 跳过验签")
	}
	tradeStatus := strings.TrimSpace(form["trade_status"])
	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		return "success", nil
	}
	orderNo := strings.TrimSpace(form["out_trade_no"])
	txn := strings.TrimSpace(form["trade_no"])
	total := strings.TrimSpace(form["total_amount"])
	amountFen, err := yuanToFen(total)
	if err != nil {
		return "failure", err
	}
	if err := FulfillPaid(ctx, orderNo, ChannelAlipay, txn, amountFen); err != nil {
		glog.Warningf(ctx, "[cash] alipay fulfill fail orderNo=%s err=%v", orderNo, err)
		return "failure", err
	}
	return "success", nil
}

func paymentDevBypass() bool {
	return strings.TrimSpace(os.Getenv("CASH_PAYMENT_DEV_BYPASS")) == "1"
}

func yuanToFen(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "total_amount 为空")
	}
	var yuan float64
	if _, err := fmt.Sscanf(s, "%f", &yuan); err != nil {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "total_amount 非法")
	}
	return int(yuan*100 + 0.5), nil
}

func alipaySignContent(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "sign" || v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	return strings.Join(parts, "&")
}

func alipayEncodeOrderStr(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+url.QueryEscape(params[k]))
	}
	return strings.Join(parts, "&")
}

func rsa2Sign(privateKeyPEM, content string) (string, error) {
	key, err := parseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256([]byte(content))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func verifyAlipayNotifySign(publicKeyPEM string, form map[string]string) error {
	sign := strings.TrimSpace(form["sign"])
	if sign == "" {
		return gerror.NewCode(gcode.CodeNotAuthorized, "缺少 sign")
	}
	content := alipaySignContent(form)
	pub, err := parseRSAPublicKey(publicKeyPEM)
	if err != nil {
		return err
	}
	raw, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return err
	}
	h := sha256.Sum256([]byte(content))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, h[:], raw); err != nil {
		return gerror.NewCode(gcode.CodeNotAuthorized, "支付宝验签失败")
	}
	return nil
}

func parseRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	pemStr = normalizePEM(pemStr, "RSA PRIVATE KEY")
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, gerror.New("私钥 PEM 解析失败")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rk, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, gerror.New("非 RSA 私钥")
	}
	return rk, nil
}

func parseRSAPublicKey(pemStr string) (*rsa.PublicKey, error) {
	pemStr = normalizePEM(pemStr, "PUBLIC KEY")
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, gerror.New("公钥 PEM 解析失败")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// 兼容支付宝有时提供的 PKCS1
		if pk, err2 := x509.ParsePKCS1PublicKey(block.Bytes); err2 == nil {
			return pk, nil
		}
		return nil, err
	}
	rk, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, gerror.New("非 RSA 公钥")
	}
	return rk, nil
}

func normalizePEM(s, typ string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	if strings.Contains(s, "BEGIN") {
		return s
	}
	return "-----BEGIN " + typ + "-----\n" + s + "\n-----END " + typ + "-----\n"
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
