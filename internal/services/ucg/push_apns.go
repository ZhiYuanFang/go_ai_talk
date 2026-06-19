package ucg

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gogf/gf/v2/frame/g"
)

// ApnsSender sends via Apple Push Notification service (HTTP/2).
type ApnsSender struct{}

func NewApnsSender() *ApnsSender { return &ApnsSender{} }

func (s *ApnsSender) Channel() string { return PushChannelAPNs }

var (
	apnsClientOnce sync.Once
	apnsHTTPClient *http.Client
)

func (s *ApnsSender) Send(ctx context.Context, token string, payload PushPayload) (invalidToken bool, err error) {
	cfg := loadPushConfig(ctx)
	if !apnsConfigured(cfg) {
		g.Log().Debug(ctx, "[ucg-push] APNs skipped: credentials not configured")
		return false, nil
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return false, nil
	}
	body, err := buildApnsBody(payload)
	if err != nil {
		return false, err
	}
	host := "https://api.sandbox.push.apple.com"
	if cfg.ApnsProduction {
		host = "https://api.push.apple.com"
	}
	url := host + "/3/device/" + token
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return false, err
	}
	bearer, err := apnsBearerToken(cfg)
	if err != nil {
		return false, err
	}
	req.Header.Set("authorization", "bearer "+bearer)
	req.Header.Set("apns-topic", cfg.ApnsBundleID)
	req.Header.Set("apns-push-type", apnsPushType(payload))
	req.Header.Set("content-type", "application/json")
	client := apnsHTTPClientFor(cfg)
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode == http.StatusOK {
		return false, nil
	}
	reason := parseApnsReason(respBody)
	if isApnsInvalidToken(resp.StatusCode, reason) {
		return true, fmt.Errorf("apns invalid token: %s", reason)
	}
	return false, fmt.Errorf("apns status=%d reason=%s body=%s", resp.StatusCode, reason, string(respBody))
}

func apnsPushType(payload PushPayload) string {
	if payload.Silent {
		return "background"
	}
	return "alert"
}

func buildApnsBody(payload PushPayload) (string, error) {
	aps := map[string]interface{}{"badge": payload.Badge}
	if payload.Silent {
		aps["content-available"] = 1
	} else if strings.TrimSpace(payload.Alert) != "" {
		aps["alert"] = payload.Alert
		aps["sound"] = "default"
	}
	root := map[string]interface{}{"aps": aps}
	if len(payload.Data) > 0 {
		for k, v := range payload.Data {
			root[k] = v
		}
	}
	b, err := json.Marshal(root)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func parseApnsReason(body []byte) string {
	var m map[string]interface{}
	if json.Unmarshal(body, &m) != nil {
		return ""
	}
	if r, ok := m["reason"].(string); ok {
		return r
	}
	return ""
}

func isApnsInvalidToken(status int, reason string) bool {
	if status == http.StatusGone {
		return true
	}
	switch strings.ToUpper(reason) {
	case "BADDEVICE_TOKEN", "UNREGISTERED", "DEVICE_TOKEN_NOT_FOR_TOPIC", "INVALIDPROVIDER_TOKEN":
		return true
	default:
		return false
	}
}

func apnsHTTPClientFor(cfg pushConfig) *http.Client {
	apnsClientOnce.Do(func() {
		apnsHTTPClient = &http.Client{Timeout: 15 * time.Second}
	})
	return apnsHTTPClient
}

func apnsBearerToken(cfg pushConfig) (string, error) {
	keyData, err := os.ReadFile(cfg.ApnsKeyPath)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(keyData)
	if block == nil {
		return "", fmt.Errorf("apns key pem decode failed")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}
	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("apns key is not ecdsa")
	}
	claims := jwt.MapClaims{
		"iss": cfg.ApnsTeamID,
		"iat": time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	t.Header["kid"] = cfg.ApnsKeyID
	return t.SignedString(ecKey)
}
