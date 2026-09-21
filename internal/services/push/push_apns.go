// APNs 通道：经 Apple HTTP/2 API 下发通知/静默推送。
//
// 业务说明：使用 .p8 私钥签发 Provider Authentication JWT（非设备 device token）。
// Apple 限制同一 Key 下 Provider JWT 更新不得过频（约 20 分钟），故进程内缓存复用 JWT，
// 对齐 HMS 的 hmsAccessToken 模式。运维轮换 .p8 / keyId 后须重启 push-service 以清空缓存。
package push

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

	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v4"
)

// ApnsSender 经 Apple Push Notification service (HTTP/2) 向单个 device token 发送。
type ApnsSender struct{}

func NewApnsSender() *ApnsSender { return &ApnsSender{} }

func (s *ApnsSender) Channel() string { return PushChannelAPNs }

// apnsJWTReuseTTL 为 Provider JWT 进程内复用窗口。
// Apple JWT 最长约 1h，且更换间隔不宜短于约 20min；取 45min 兼顾两者。
const apnsJWTReuseTTL = 45 * time.Minute

var (
	apnsClientOnce sync.Once
	apnsHTTPClient *http.Client

	// Provider JWT 进程内缓存（与设备 device token 无关）。
	apnsJWTMu     sync.Mutex
	apnsJWTCache  string
	apnsJWTExp    time.Time
	apnsJWTKeyID  string
	apnsJWTTeamID string
)

// Send 向单个 APNs device token 发送推送。
//
// Returns: invalidToken 表示应删除本地设备 token；err 非 nil 时 dispatcher 记 send_failed。
// Side Effects: 可能读 .p8 并更新进程内 Provider JWT 缓存；不写库。
func (s *ApnsSender) Send(ctx context.Context, token string, payload PushPayload) (invalidToken bool, err error) {
	cfg := loadPushConfig(ctx)
	if !apnsConfigured(cfg) {
		g.Log().Warningf(ctx, "[push] APNs skipped: credentials not configured")
		return false, fmt.Errorf("credentials not configured")
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
	// 自定义键放在 aps 之外，点击后出现在 userInfo 根上（含 bizType）。
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

// isApnsInvalidToken 判断是否应删除本地设备 device token。
//
// 业务：仅设备侧失效（410 / BadDeviceToken 等）才返回 true。
// Provider JWT 限流（429 TooManyProviderTokenUpdates）或其它 Provider 鉴权问题
// MUST NOT 当作设备 token 无效，否则会误删 push_device 行。
func isApnsInvalidToken(status int, reason string) bool {
	// Provider token 更新过频：发送失败，但不删设备。
	if status == http.StatusTooManyRequests {
		return false
	}
	if status == http.StatusGone {
		return true
	}
	switch strings.ToUpper(reason) {
	case "BADDEVICE_TOKEN", "BADDEVICETOKEN", "UNREGISTERED", "DEVICE_TOKEN_NOT_FOR_TOPIC", "DEVICETOKENNOTFORTOPIC":
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

// apnsBearerToken 返回可复用的 APNs Provider JWT。
//
// 业务逻辑：缓存命中（未过期且 keyId/teamId 未变）直接返回，避免每次 Send 换 iat
// 触发 Apple TooManyProviderTokenUpdates。未命中时读 .p8 签发并写入进程缓存。
// 运维轮换密钥文件后须重启本进程，否则可能继续使用旧 JWT 直至 TTL 到期。
//
// Args: cfg 含 ApnsKeyPath / KeyID / TeamID。
// Returns: bearer 字符串（不含 "bearer " 前缀）；签发失败时 error。
// Side Effects: 更新包级 apnsJWT* 缓存变量。
func apnsBearerToken(cfg pushConfig) (string, error) {
	apnsJWTMu.Lock()
	defer apnsJWTMu.Unlock()

	now := time.Now()
	if apnsJWTCache != "" &&
		now.Before(apnsJWTExp) &&
		apnsJWTKeyID == cfg.ApnsKeyID &&
		apnsJWTTeamID == cfg.ApnsTeamID {
		return apnsJWTCache, nil
	}

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
		"iat": now.Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	t.Header["kid"] = cfg.ApnsKeyID
	signed, err := t.SignedString(ecKey)
	if err != nil {
		return "", err
	}

	apnsJWTCache = signed
	apnsJWTExp = now.Add(apnsJWTReuseTTL)
	apnsJWTKeyID = cfg.ApnsKeyID
	apnsJWTTeamID = cfg.ApnsTeamID
	return apnsJWTCache, nil
}
