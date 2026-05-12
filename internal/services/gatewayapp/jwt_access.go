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

// accessClaims access JWT 载荷：`sub` 为 wx 表主键，设备号直连登录无 wx 时为 "0"；`device_no` 为会话设备（wx 登录未绑机时可省略，纯设备会话必填）。
type accessClaims struct {
	jwt.RegisteredClaims
	DeviceNo string `json:"device_no,omitempty"`
}

// SignAccess 签发 access JWT（HS256）。`sub` 一般为 wx 表主键；无 wx 的设备号会话传 wxID=0 且须传非空 deviceNo（写入 device_no claim）。`device_no` 供网关注入下游头（传空串则省略该 claim）。
func SignAccess(ctx context.Context, wxID int64, deviceNo string) (string, error) {
	if wxID == 0 && strings.TrimSpace(deviceNo) == "" {
		return "", fmt.Errorf("纯设备会话签发 access 时 deviceNo 不能为空")
	}
	if wxID < 0 {
		return "", fmt.Errorf("wxId 无效")
	}
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
