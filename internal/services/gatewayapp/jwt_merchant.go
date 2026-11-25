package gatewayapp

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hello/internal/platform/httpmeta"
	"hello/internal/platform/jwtcfg"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v4"
)

const merchantJWTIssuer = "gateway-app-merchant"
const merchantJWTTyp = "merchant"

type merchantAccessClaims struct {
	jwt.RegisteredClaims
	Typ string `json:"typ"`
}

// SignMerchantAccess 签发商户 access JWT（与 wx access 同 secret，issuer/typ 区分）。
func SignMerchantAccess(ctx context.Context, merchantID int64) (string, error) {
	if merchantID <= 0 {
		return "", fmt.Errorf("merchantId 无效")
	}
	secret := jwtcfg.GatewayAppSecret()
	if secret == "" {
		return "", fmt.Errorf("GATEWAY_APP_JWT_SECRET 未配置")
	}
	ttl := g.Cfg().MustGet(ctx, "gatewayApp.accessTtlSeconds").Int64()
	if ttl <= 0 {
		ttl = 1800
	}
	now := time.Now()
	ac := merchantAccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(merchantID, 10),
			Issuer:    merchantJWTIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttl) * time.Second)),
		},
		Typ: merchantJWTTyp,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &ac)
	return t.SignedString([]byte(secret))
}

// ParseMerchantAccessClaims 校验商户 JWT，返回 merchantId。
func ParseMerchantAccessClaims(ctx context.Context, tokenString string) (int64, error) {
	secret := jwtcfg.GatewayAppSecret()
	if secret == "" {
		return 0, fmt.Errorf("GATEWAY_APP_JWT_SECRET 未配置")
	}
	token, err := jwt.ParseWithClaims(tokenString, &merchantAccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*merchantAccessClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}
	if claims.Issuer != merchantJWTIssuer || claims.Typ != merchantJWTTyp {
		return 0, fmt.Errorf("not merchant token")
	}
	id, err := strconv.ParseInt(strings.TrimSpace(claims.Subject), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid merchant sub")
	}
	return id, nil
}

// IsMerchantPath 商户 API / cash 商户面路径。
func IsMerchantPath(path string) bool {
	p := strings.TrimSpace(path)
	return strings.HasPrefix(p, "/device/merchant/") ||
		strings.HasPrefix(p, "/cash/merchant/") ||
		strings.HasPrefix(p, "/merchant/")
}

// InjectMerchantHeadersFromBearer 解析商户 Bearer 并注入 X-Internal-Merchant-Id。
func InjectMerchantHeadersFromBearer(r *ghttp.Request) error {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	const pfx = "Bearer "
	if len(auth) < len(pfx) || !strings.EqualFold(auth[:len(pfx)], pfx) {
		return ErrGatewayBearerMissing
	}
	raw := strings.TrimSpace(auth[len(pfx):])
	mid, err := ParseMerchantAccessClaims(r.Context(), raw)
	if err != nil {
		return ErrGatewayBearerInvalid
	}
	SetInternalHeader(r, httpmeta.HeaderInternalMerchantId, strconv.FormatInt(mid, 10))
	return nil
}

// PeekIsMerchantJWT 粗判 Bearer 是否商户令牌（不写头）。
func PeekIsMerchantJWT(ctx context.Context, raw string) bool {
	_, err := ParseMerchantAccessClaims(ctx, strings.TrimSpace(raw))
	return err == nil
}
