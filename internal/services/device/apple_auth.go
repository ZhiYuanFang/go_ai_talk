package device

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/golang-jwt/jwt/v4"
)

const (
	appleJWKSURL          = "https://appleid.apple.com/auth/keys"
	appleJWTIssuer        = "https://appleid.apple.com"
	appleJWKSCacheTTL     = 24 * time.Hour // 与 Apple 密钥轮换频率对齐；kid 未命中时强制刷新一次
)

var (
	ErrAppleIdentityTokenInvalid = errors.New("Apple 登录凭证无效或已过期")
	ErrAppleIdentityTokenEmpty   = errors.New("identityToken 不能为空")

	appleJWKS struct {
		mu        sync.RWMutex
		keys      map[string]*rsa.PublicKey
		fetchedAt time.Time
	}
)

type appleJWKSResponse struct {
	Keys []appleJWK `json:"keys"`
}

type appleJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type appleIdentityClaims struct {
	jwt.RegisteredClaims
}

// verifyAppleIdentityToken 校验 Apple identityToken（JWKS RS256），仅返回 JWT sub；不记录 token 原文。
// 服务端仅依赖 JWKS 验签，无需 Apple .p8 私钥（authorizationCode 换票不在本路径使用）。
func verifyAppleIdentityToken(ctx context.Context, identityToken string) (sub string, err error) {
	identityToken = strings.TrimSpace(identityToken)
	if identityToken == "" {
		return "", ErrAppleIdentityTokenEmpty
	}
	bundleID := strings.TrimSpace(g.Cfg().MustGet(ctx, "apple.ios.bundleId").String())
	if bundleID == "" {
		return "", errors.New("未配置 apple.ios.bundleId")
	}

	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
	var claims appleIdentityClaims
	token, err := parser.ParseWithClaims(identityToken, &claims, func(t *jwt.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		kid = strings.TrimSpace(kid)
		if kid == "" {
			return nil, errors.New("token 缺少 kid")
		}
		pub, keyErr := applePublicKeyForKid(ctx, kid)
		if keyErr != nil {
			return nil, keyErr
		}
		return pub, nil
	})
	if err != nil || token == nil || !token.Valid {
		glog.Warningf(ctx, "[apple-auth] identityToken 校验失败 err=%v", err)
		return "", ErrAppleIdentityTokenInvalid
	}
	if !claims.VerifyIssuer(appleJWTIssuer, true) {
		return "", ErrAppleIdentityTokenInvalid
	}
	if !claims.VerifyExpiresAt(time.Now(), true) {
		return "", ErrAppleIdentityTokenInvalid
	}
	if !appleAudienceMatches(claims.Audience, bundleID) {
		return "", ErrAppleIdentityTokenInvalid
	}
	sub = strings.TrimSpace(claims.Subject)
	if sub == "" {
		return "", ErrAppleIdentityTokenInvalid
	}
	return sub, nil
}

func appleAudienceMatches(aud jwt.ClaimStrings, bundleID string) bool {
	for _, a := range aud {
		if strings.TrimSpace(a) == bundleID {
			return true
		}
	}
	return false
}

func applePublicKeyForKid(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	appleJWKS.mu.RLock()
	if pub := appleJWKS.keys[kid]; pub != nil && time.Since(appleJWKS.fetchedAt) < appleJWKSCacheTTL {
		appleJWKS.mu.RUnlock()
		return pub, nil
	}
	appleJWKS.mu.RUnlock()

	if err := refreshAppleJWKS(ctx); err != nil {
		return nil, err
	}

	appleJWKS.mu.RLock()
	defer appleJWKS.mu.RUnlock()
	if pub := appleJWKS.keys[kid]; pub != nil {
		return pub, nil
	}
	return nil, fmt.Errorf("Apple JWKS 未找到 kid=%s", kid)
}

func refreshAppleJWKS(ctx context.Context) error {
	client := gclient.New().Timeout(8 * time.Second)
	resp, err := client.Get(ctx, appleJWKSURL)
	if err != nil {
		return fmt.Errorf("拉取 Apple JWKS 失败: %w", err)
	}
	defer resp.Close()
	body := resp.ReadAll()
	if resp.StatusCode != http.StatusOK {
		glog.Warningf(ctx, "[apple-auth] JWKS HTTP 状态异常 status=%d", resp.StatusCode)
		return fmt.Errorf("Apple JWKS HTTP 状态异常: %d", resp.StatusCode)
	}
	var parsed appleJWKSResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return fmt.Errorf("解析 Apple JWKS 失败: %w", err)
	}
	keys := make(map[string]*rsa.PublicKey, len(parsed.Keys))
	for _, k := range parsed.Keys {
		if strings.TrimSpace(k.Kid) == "" || strings.ToUpper(k.Kty) != "RSA" {
			continue
		}
		pub, convErr := rsaPublicKeyFromJWK(k.N, k.E)
		if convErr != nil {
			glog.Warningf(ctx, "[apple-auth] 跳过无效 JWK kid=%s err=%v", k.Kid, convErr)
			continue
		}
		keys[k.Kid] = pub
	}
	if len(keys) == 0 {
		return errors.New("Apple JWKS 无可用 RSA 公钥")
	}
	appleJWKS.mu.Lock()
	appleJWKS.keys = keys
	appleJWKS.fetchedAt = time.Now()
	appleJWKS.mu.Unlock()
	return nil
}

func rsaPublicKeyFromJWK(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes).Int64()
	if e <= 0 || e > int64(^uint(0)>>1) {
		return nil, errors.New("invalid RSA exponent")
	}
	return &rsa.PublicKey{N: n, E: int(e)}, nil
}
