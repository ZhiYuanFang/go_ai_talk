package ucg

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v4"
)

const ucgJWTIssuer = "gateway-app-server"

type ucgAccessClaims struct {
	jwt.RegisteredClaims
	DeviceNo string `json:"device_no,omitempty"`
}

// ParseWSAccessToken 校验 App access JWT（secret 须与 gateway-app jwtSecret 一致）。
func ParseWSAccessToken(ctx context.Context, tokenString string) (wxID int64, err error) {
	secret := wsJWTSecret(ctx)
	if secret == "" {
		return 0, fmt.Errorf("ucg jwtSecret 未配置")
	}
	token, err := jwt.ParseWithClaims(tokenString, &ucgAccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*ucgAccessClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}
	if claims.Issuer != "" && claims.Issuer != ucgJWTIssuer {
		return 0, fmt.Errorf("invalid issuer")
	}
	wxID, err = strconv.ParseInt(strings.TrimSpace(claims.Subject), 10, 64)
	if err != nil || wxID <= 0 {
		return 0, fmt.Errorf("invalid sub")
	}
	return wxID, nil
}

func wsJWTSecret(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("UCG_JWT_SECRET")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("GATEWAY_APP_JWT_SECRET")); v != "" {
		return v
	}
	return strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.jwtSecret").String())
}
