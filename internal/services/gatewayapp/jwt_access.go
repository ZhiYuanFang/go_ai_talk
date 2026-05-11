package gatewayapp

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v4"
)

const jwtIssuer = "gateway-app-server"

// SignAccess 签发 access JWT（HS256），`sub` 为 wx 表主键 id。
func SignAccess(ctx context.Context, wxID int64) (string, error) {
	secret := strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.jwtSecret").String())
	if secret == "" {
		return "", fmt.Errorf("gatewayApp.jwtSecret 未配置")
	}
	ttl := g.Cfg().MustGet(ctx, "gatewayApp.accessTtlSeconds").Int64()
	if ttl <= 0 {
		ttl = 1800
	}
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(wxID, 10),
		Issuer:    jwtIssuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttl) * time.Second)),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// ParseAccessWxID 校验 access JWT，返回 wx 主键 id（来自 `sub`）。
func ParseAccessWxID(ctx context.Context, tokenString string) (int64, error) {
	secret := strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.jwtSecret").String())
	if secret == "" {
		return 0, fmt.Errorf("gatewayApp.jwtSecret 未配置")
	}
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}
	return strconv.ParseInt(strings.TrimSpace(claims.Subject), 10, 64)
}
