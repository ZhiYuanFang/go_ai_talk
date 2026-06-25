package gatewayapp

import (
	"context"
	"crypto/subtle"
	"fmt"
	"os"
	"strings"
	"time"

	"hello/internal/platform/jwtcfg"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v4"
)

const (
	// AdminJWTAudience Admin JWT aud 固定值，与用户 access JWT 隔离。
	AdminJWTAudience = "gateway-admin"
	// AdminJWTSubject Admin JWT sub 固定字面量，不承载 wxId。
	AdminJWTSubject = "gateway-admin"
	// HeaderInternalAdminVerified 网关在 Admin JWT 校验通过后注入，下游 handler 可读；客户端不可伪造。
	HeaderInternalAdminVerified = "X-Internal-Admin-Verified"
)

// adminClaims 运维 Hub 登录 JWT 载荷。
type adminClaims struct {
	jwt.RegisteredClaims
	Username string `json:"username,omitempty"`
}

// AdminAccessTTL 返回 Admin JWT 有效期，默认 8 小时。
func AdminAccessTTL(ctx context.Context) time.Duration {
	sec := g.Cfg().MustGet(ctx, "gatewayApp.admin.sessionTtlSeconds").Int64()
	if sec <= 0 {
		sec = g.Cfg().MustGet(ctx, "gatewayApp.versionAdmin.sessionTtlSeconds").Int64()
	}
	if sec <= 0 {
		sec = 8 * 3600
	}
	return time.Duration(sec) * time.Second
}

// SignAdminAccess 签发 Admin JWT（HS256，aud=gateway-admin）。
func SignAdminAccess(ctx context.Context, username string) (token string, expiresIn int64, err error) {
	secret := strings.TrimSpace(jwtcfg.GatewayAppSecret())
	if secret == "" {
		return "", 0, fmt.Errorf("GATEWAY_APP_JWT_SECRET 未配置")
	}
	ttl := AdminAccessTTL(ctx)
	expiresIn = int64(ttl.Seconds())
	now := time.Now()
	claims := adminClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   AdminJWTSubject,
			Audience:  jwt.ClaimStrings{AdminJWTAudience},
			Issuer:    jwtIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		Username: strings.TrimSpace(username),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	token, err = t.SignedString([]byte(secret))
	return token, expiresIn, err
}

// ParseAdminClaims 校验 Admin JWT；非 admin aud 或无效 token 返回 error。
func ParseAdminClaims(ctx context.Context, tokenString string) (*adminClaims, error) {
	secret := strings.TrimSpace(jwtcfg.GatewayAppSecret())
	if secret == "" {
		return nil, fmt.Errorf("GATEWAY_APP_JWT_SECRET 未配置")
	}
	token, err := jwt.ParseWithClaims(tokenString, &adminClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*adminClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid admin token")
	}
	if !claims.verifyAudience() {
		return nil, fmt.Errorf("invalid admin audience")
	}
	if strings.TrimSpace(claims.Subject) != AdminJWTSubject {
		return nil, fmt.Errorf("invalid admin subject")
	}
	return claims, nil
}

func (c *adminClaims) verifyAudience() bool {
	for _, aud := range c.Audience {
		if strings.TrimSpace(aud) == AdminJWTAudience {
			return true
		}
	}
	return false
}

// IsAdminJWTString 判断 raw token 是否为 Admin JWT。
func IsAdminJWTString(ctx context.Context, tokenString string) bool {
	_, err := ParseAdminClaims(ctx, tokenString)
	return err == nil
}

// AdminUsername 读取 Hub 登录用户名 env，默认 admin。
func AdminUsername() string {
	if v := strings.TrimSpace(os.Getenv("GATEWAY_APP_ADMIN_USERNAME")); v != "" {
		return v
	}
	return "admin"
}

// AdminPassword 读取 Hub 登录密码 env；未配置返回空。
func AdminPassword() string {
	return strings.TrimSpace(os.Getenv("GATEWAY_APP_ADMIN_PASSWORD"))
}

// DeviceAdminPassword 网关注入 device 反代用的 X-Admin-Password。
func DeviceAdminPassword() string {
	if v := strings.TrimSpace(os.Getenv("DEVICE_ADMIN_PASSWORD")); v != "" {
		return v
	}
	return AdminPassword()
}

// UcgAdminPassword 网关注入 ucg 反代用的 X-Admin-Password。
func UcgAdminPassword() string {
	if v := strings.TrimSpace(os.Getenv("UCG_ADMIN_PASSWORD")); v != "" {
		return v
	}
	return AdminPassword()
}

// VoiceAdminPassword 网关注入 voice 反代用的 X-Admin-Password。
func VoiceAdminPassword() string {
	if v := strings.TrimSpace(os.Getenv("VOICE_ADMIN_PASSWORD")); v != "" {
		return v
	}
	return AdminPassword()
}

// SimAdminPassword 网关注入 sim-user-service 反代用的 X-Admin-Password。
func SimAdminPassword() string {
	if v := strings.TrimSpace(os.Getenv("SIM_ADMIN_PASSWORD")); v != "" {
		return v
	}
	return AdminPassword()
}

// AdminLoginEnabled Hub 登录是否可用（须配置密码）。
func AdminLoginEnabled() bool {
	return AdminPassword() != ""
}

// ConstantTimeEqual 常量时间字符串比较，用于登录口令校验。
func ConstantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		dummy := strings.Repeat("\x00", 64)
		_ = subtle.ConstantTimeCompare([]byte(dummy), []byte(dummy))
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// VerifyAdminLogin 校验 Hub 账号密码。
func VerifyAdminLogin(username, password string) bool {
	if !AdminLoginEnabled() {
		return false
	}
	return ConstantTimeEqual(strings.TrimSpace(username), AdminUsername()) &&
		ConstantTimeEqual(strings.TrimSpace(password), AdminPassword())
}

// IsGatewayAdminAPIPath 是否须 Admin JWT 的管理 API（不含 login）。
func IsGatewayAdminAPIPath(path string) bool {
	path = strings.TrimSpace(path)
	if path == "/device/admin/api/login" {
		return false
	}
	if strings.HasPrefix(path, "/device/admin/api/") {
		return true
	}
	if strings.HasPrefix(path, "/ucg/admin/api/") {
		return true
	}
	if strings.HasPrefix(path, "/voice/admin/api/") {
		return true
	}
	if strings.HasPrefix(path, "/sim/admin/api/") {
		return true
	}
	if strings.HasPrefix(path, "/device/app/api/version/admin/") {
		return true
	}
	return false
}

// IsAdminJWTVerified 请求是否已通过 Admin JWT 校验（由网关 Hook 设置）。
func IsAdminJWTVerified(headerGetter func(string) string) bool {
	if headerGetter == nil {
		return false
	}
	return strings.TrimSpace(headerGetter(HeaderInternalAdminVerified)) == "1"
}
