package cash

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// Apple Root CA - G3（ASN V2 / App Store Server API JWS 证书链根）。
// 来源：https://www.apple.com/certificateauthority/AppleRootCA-G3.cer
// 可通过 CASH_APPLE_ROOT_CA_PEM 覆盖（完整 PEM 文本）以便证书轮换。
const appleRootCAG3PEM = `-----BEGIN CERTIFICATE-----
MIICQzCCAcmgAwIBAgIILcX8iNLFS5UwCgYIKoZIzj0EAwMwZzEbMBkGA1UEAwwS
QXBwbGUgUm9vdCBDQSAtIEczMSYwJAYDVQQLDB1BcHBsZSBDZXJ0aWZpY2F0aW9u
IEF1dGhvcml0eTETMBEGA1UECgwKQXBwbGUgSW5jLjELMAkGA1UEBhMCVVMwHhcN
MTQwNDMwMTgxOTA2WhcNMzkwNDMwMTgxOTA2WjBnMRswGQYDVQQDDBJBcHBsZSBS
b290IENBIC0gRzMxJjAkBgNVBAsMHUFwcGxlIENlcnRpZmljYXRpb24gQXV0aG9y
aXR5MRMwEQYDVQQKDApBcHBsZSBJbmMuMQswCQYDVQQGEwJVUzB2MBAGByqGSM49
AgEGBSuBBAAiA2IABJjpLz1AcqTtkyJygRMc3RCV8cWjTnHcFBbZDuWmBSp3ZHtf
TjjTuxxEtX/1H7YyYl3J6YRbTzBPEVoA/VhYDKX1DyxNB0cTddqXl5dvMVztK517
IDvYuVTZXpmkOlEKMaNCMEAwHQYDVR0OBBYEFLuw3qFYM4iapIqZ3r6966/ayySr
MA8GA1UdEwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgEGMAoGCCqGSM49BAMDA2gA
MGUCMQCD6cHEFl4aXTQY2e3v9GwOAEZLuN+yRhHFD/3meoyhpmvOwgPUnPWTxnS4
at+qIxUCMG1mihDK1A3UT82NQz60imOlM27jbdoXt2QfyFMm+YhidDkLF1vLUagM
6BgD56KyKA==
-----END CERTIFICATE-----`

var (
	appleRootPoolOnce sync.Once
	appleRootPool     *x509.CertPool
	appleRootPoolErr  error
)

// appleJWSHeader ASN / transaction JWS 头部（含 x5c 证书链）。
type appleJWSHeader struct {
	Alg string   `json:"alg"`
	X5c []string `json:"x5c"`
}

// VerifyAndDecodeAppleJWS 校验 Apple 签发的 JWS（ES256 + x5c → Apple Root CA），返回 payload 原始 JSON。
//
// 业务：ASN signedPayload、signedTransactionInfo 与客户端 verify 的 signedTransaction 共用。
// 禁止用 CASH_PAYMENT_DEV_BYPASS 跳过本函数（权威开通/退款路径必须密码学可信）。
//
// Args: jws — 三段式 compact JWS
// Returns: payload 字节；验签失败返回 CodeNotAuthorized
func VerifyAndDecodeAppleJWS(jws string) ([]byte, error) {
	jws = strings.TrimSpace(jws)
	parts := strings.Split(jws, ".")
	if len(parts) != 3 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "JWS 须为三段")
	}
	hdrRaw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInvalidParameter, err, "JWS header 解码失败")
	}
	var hdr appleJWSHeader
	if err := json.Unmarshal(hdrRaw, &hdr); err != nil {
		return nil, gerror.WrapCode(gcode.CodeInvalidParameter, err, "JWS header 解析失败")
	}
	if !strings.EqualFold(hdr.Alg, "ES256") {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "不支持的 JWS alg（须 ES256）")
	}
	if len(hdr.X5c) == 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "JWS 缺少 x5c 证书链")
	}

	certs := make([]*x509.Certificate, 0, len(hdr.X5c))
	for i, b64 := range hdr.X5c {
		der, dErr := base64.StdEncoding.DecodeString(b64)
		if dErr != nil {
			return nil, gerror.WrapCode(gcode.CodeNotAuthorized, dErr, fmt.Sprintf("x5c[%d] 解码失败", i))
		}
		c, pErr := x509.ParseCertificate(der)
		if pErr != nil {
			return nil, gerror.WrapCode(gcode.CodeNotAuthorized, pErr, fmt.Sprintf("x5c[%d] 解析失败", i))
		}
		certs = append(certs, c)
	}
	leaf := certs[0]
	roots, rErr := getAppleRootPool()
	if rErr != nil {
		return nil, rErr
	}
	intermediates := x509.NewCertPool()
	for _, c := range certs[1:] {
		intermediates.AddCert(c)
	}
	if _, err := leaf.Verify(x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}); err != nil {
		return nil, gerror.WrapCode(gcode.CodeNotAuthorized, err, "Apple 证书链校验失败")
	}

	pub, ok := leaf.PublicKey.(*ecdsa.PublicKey)
	if !ok || pub.Curve != elliptic.P256() {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "叶子证书公钥不是 P-256 ECDSA")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInvalidParameter, err, "JWS signature 解码失败")
	}
	if len(sig) != 64 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "ES256 签名长度非法")
	}
	signingInput := []byte(parts[0] + "." + parts[1])
	sum := sha256.Sum256(signingInput)
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	if !ecdsa.Verify(pub, sum[:], r, s) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "JWS 签名校验失败")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInvalidParameter, err, "JWS payload 解码失败")
	}
	return payload, nil
}

// DecodeAppleJWSPayloadMap 验签后将 payload 转为 string map（数值压成整数字符串）。
func DecodeAppleJWSPayloadMap(jws string) (map[string]string, error) {
	raw, err := VerifyAndDecodeAppleJWS(jws)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, gerror.WrapCode(gcode.CodeInvalidParameter, err, "JWS payload JSON 非法")
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

func getAppleRootPool() (*x509.CertPool, error) {
	appleRootPoolOnce.Do(func() {
		pemData := strings.TrimSpace(os.Getenv("CASH_APPLE_ROOT_CA_PEM"))
		if pemData == "" {
			pemData = appleRootCAG3PEM
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(pemData)) {
			appleRootPoolErr = errors.New("无法加载 Apple Root CA PEM")
			return
		}
		appleRootPool = pool
	})
	if appleRootPoolErr != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, appleRootPoolErr, "Apple Root CA 初始化失败")
	}
	return appleRootPool, nil
}
