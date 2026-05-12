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

// accessClaims access JWT 载荷：`sub` 为 wx 表主键；`device_no` 为已绑定设备（未绑定时省略）。
type accessClaims struct {
	jwt.RegisteredClaims
	DeviceNo string `json:"device_no,omitempty"`
}

// SignAccess 签发 access JWT（HS256）。`sub` MUST 为 wx 表主键 id；`device_no` 写入私有声明供网关注入下游头（未绑定时传空串则省略该 claim）。
func SignAccess(ctx context.Context, wxID int64, deviceNo string) (string, error) {
	secret := strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.jwtSecret").String())
	if secret == "" {
		return "", fmt.Errorf("gatewayApp.jwtSecret 未配置")
	}
	ttl := g.Cfg().MustGet(ctx, "gatewayApp.accessTtlSeconds").Int64()
	if ttl <= 0 {
		ttl = 1800
	}
	now := time.Now()
	ac := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(wxID, 10),
			Issuer:    jwtIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttl) * time.Second)),
		},
	}
	dn := strings.TrimSpace(deviceNo)
	if dn != "" {
		ac.DeviceNo = dn
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &ac)
	return t.SignedString([]byte(secret))
}

// ParseAccessClaims 校验 access JWT，返回 wx 主键 id 与声明中的 device_no（旧 token 无 device_no 时为空串）。
func ParseAccessClaims(ctx context.Context, tokenString string) (wxID int64, deviceNo string, err error) {
	secret := strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.jwtSecret").String())
	if secret == "" {
		return 0, "", fmt.Errorf("gatewayApp.jwtSecret 未配置")
	}
	token, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, "", err
	}
	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return 0, "", fmt.Errorf("invalid token")
	}
	wxID, err = strconv.ParseInt(strings.TrimSpace(claims.Subject), 10, 64)
	if err != nil {
		return 0, "", err
	}
	return wxID, strings.TrimSpace(claims.DeviceNo), nil
}

// ParseAccessWxID 校验 access JWT，仅返回 wx 主键 id（兼容旧调用方）。
func ParseAccessWxID(ctx context.Context, tokenString string) (int64, error) {
	wxID, _, err := ParseAccessClaims(ctx, tokenString)
	return wxID, err
}
