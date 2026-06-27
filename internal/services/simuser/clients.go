package simuser

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
)

type loginSession struct {
	WxID        int64
	AccessToken string
}

func gatewayBase(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("GATEWAY_APP_URL")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return strings.TrimRight(g.Cfg().MustGet(ctx, "simUser.gatewayAppUrl").String(), "/")
}

func deviceBase(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("DEVICE_SERVICE_URL")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return strings.TrimRight(g.Cfg().MustGet(ctx, "simUser.deviceServiceUrl").String(), "/")
}

func ucgBase(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("UCG_SERVICE_URL")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return strings.TrimRight(g.Cfg().MustGet(ctx, "simUser.ucgServiceUrl").String(), "/")
}

func internalSecret(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET")); v != "" {
		return v
	}
	return strings.TrimSpace(g.Cfg().MustGet(ctx, "simUser.deviceInternalSecret").String())
}

func deviceInternalPost(ctx context.Context, path string, body interface{}, out interface{}) error {
	base := deviceBase(ctx)
	if base == "" {
		return fmt.Errorf("DEVICE_SERVICE_URL 未配置")
	}
	resp, err := gclient.New().
		SetHeader("X-Device-Gateway-Internal-Secret", internalSecret(ctx)).
		ContentJson().Post(ctx, base+path, body)
	if err != nil {
		return err
	}
	return parseEnvelope(resp.ReadAllString(), out)
}

// 这里是获取UCG的帖子列表
func ucgInternalPost(ctx context.Context, path string, body interface{}, out interface{}) error {
	// 等待UCG的请求速率
	if err := waitOutboundRate(ctx, "ucg-internal"); err != nil {
		return err
	}
	// 获取UCG的基址
	base := ucgBase(ctx)
	if base == "" {
		return fmt.Errorf("UCG_SERVICE_URL 未配置")
	}
	// 创建一个HTTP客户端
	// 设置内部密钥
	resp, err := gclient.New().
		// 设置内容类型为JSON
		SetHeader("X-Device-Gateway-Internal-Secret", internalSecret(ctx)).
		ContentJson().Post(ctx, base+path, body)
	if err != nil {
		return err
	}
	// 读取响应内容
	raw := resp.ReadAllString()
	if out != nil {
		return parseEnvelope(raw, out)
	}
	j := gjson.New(raw)
	if j.Get("code").Int() != 0 {
		return fmt.Errorf("%s", j.Get("message").String())
	}
	return nil
}

func parseEnvelope(raw string, out interface{}) error {
	j := gjson.New(raw)
	if j.Get("code").Int() != 0 {
		return fmt.Errorf("%s", j.Get("message").String())
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal([]byte(j.Get("data").String()), out)
}

func simRegister(ctx context.Context, account, password string) (int64, error) {
	var data struct {
		WxId int64 `json:"wxId"`
	}
	if err := deviceInternalPost(ctx, "/device/internal/api/sim/username/register", g.Map{
		"account": account, "password": password,
	}, &data); err != nil {
		return 0, err
	}
	return data.WxId, nil
}

func appPut(ctx context.Context, token, path string, body interface{}, out interface{}) error {
	if err := waitOutboundRate(ctx, "app-put"); err != nil {
		return err
	}
	base := gatewayBase(ctx)
	resp, err := gclient.New().
		SetHeader("Authorization", "Bearer "+token).
		ContentJson().Put(ctx, base+path, body)
	if err != nil {
		return err
	}
	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 {
		return fmt.Errorf("%s", j.Get("message").String())
	}
	if out != nil {
		return json.Unmarshal([]byte(j.Get("data").String()), out)
	}
	return nil
}

func countSimUsers(ctx context.Context) (int, error) {
	base := deviceBase(ctx)
	resp, err := gclient.New().
		SetHeader("X-Device-Gateway-Internal-Secret", internalSecret(ctx)).
		Get(ctx, base+"/device/internal/api/sim/wx/list?page=1&pageSize=1")
	if err != nil {
		return 0, err
	}
	var data struct {
		Total int `json:"total"`
	}
	if err = parseEnvelope(resp.ReadAllString(), &data); err != nil {
		return 0, err
	}
	return data.Total, nil
}

// simWxPick device random 接口返回的模拟用户摘要。
type simWxPick struct {
	WxId    int64  `json:"wxId"`
	Account string `json:"account"`
}

const simFollowRandomMaxTry = 8

func deviceInternalGet(ctx context.Context, path string, out interface{}) error {
	base := deviceBase(ctx)
	if base == "" {
		return fmt.Errorf("DEVICE_SERVICE_URL 未配置")
	}
	resp, err := gclient.New().
		SetHeader("X-Device-Gateway-Internal-Secret", internalSecret(ctx)).
		Get(ctx, base+path)
	if err != nil {
		return err
	}
	return parseEnvelope(resp.ReadAllString(), out)
}

// pickRandomSimWx 经 device internal random 接口全库随机取 1 个模拟用户。
func pickRandomSimWx(ctx context.Context) (simWxPick, error) {
	var data simWxPick
	if err := deviceInternalGet(ctx, "/device/internal/api/sim/wx/random", &data); err != nil {
		return simWxPick{}, err
	}
	if data.WxId <= 0 || strings.TrimSpace(data.Account) == "" {
		return simWxPick{}, fmt.Errorf("无模拟用户")
	}
	return data, nil
}

// pickTwoDistinctSimWx 随机取两个不同 wxId；仅 1 个 sim 用户时失败。
func pickTwoDistinctSimWx(ctx context.Context) (a, b simWxPick, err error) {
	a, err = pickRandomSimWx(ctx)
	if err != nil {
		return simWxPick{}, simWxPick{}, err
	}
	for i := 0; i < simFollowRandomMaxTry; i++ {
		candidate, pErr := pickRandomSimWx(ctx)
		if pErr != nil {
			return simWxPick{}, simWxPick{}, pErr
		}
		if candidate.WxId != a.WxId {
			return a, candidate, nil
		}
	}
	return simWxPick{}, simWxPick{}, fmt.Errorf("sim 用户不足")
}

func usernameLogin(ctx context.Context, account, password string) (loginSession, error) {
	base := gatewayBase(ctx)
	if base == "" {
		return loginSession{}, fmt.Errorf("GATEWAY_APP_URL 未配置")
	}
	resp, err := gclient.New().ContentJson().Post(ctx, base+"/device/app/api/username_login", g.Map{
		"account": account, "password": password,
	})
	if err != nil {
		return loginSession{}, err
	}
	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 {
		return loginSession{}, fmt.Errorf("%s", j.Get("message").String())
	}
	return loginSession{
		WxID:        j.Get("data.wxId").Int64(),
		AccessToken: j.Get("data.accessToken").String(),
	}, nil
}

func appGet(ctx context.Context, token, path string, out interface{}) error {
	if err := waitOutboundRate(ctx, "app-get"); err != nil {
		return err
	}
	base := gatewayBase(ctx)
	resp, err := gclient.New().
		SetHeader("Authorization", "Bearer "+token).
		Get(ctx, base+path)
	if err != nil {
		return err
	}
	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 {
		return fmt.Errorf("%s", j.Get("message").String())
	}
	if out != nil {
		return json.Unmarshal([]byte(j.Get("data").String()), out)
	}
	return nil
}

func appPost(ctx context.Context, token, path string, body interface{}, out interface{}) error {
	if err := waitOutboundRate(ctx, "app-post"); err != nil {
		return err
	}
	base := gatewayBase(ctx)
	resp, err := gclient.New().
		SetHeader("Authorization", "Bearer "+token).
		ContentJson().Post(ctx, base+path, body)
	if err != nil {
		return err
	}
	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 {
		return fmt.Errorf("%s", j.Get("message").String())
	}
	if out != nil {
		return json.Unmarshal([]byte(j.Get("data").String()), out)
	}
	return nil
}

func randomSimSession(ctx context.Context, password string) (loginSession, string, error) {
	item, err := pickRandomSimWx(ctx)
	if err != nil {
		return loginSession{}, "", err
	}
	sess, err := usernameLogin(ctx, item.Account, password)
	return sess, item.Account, err
}

func newClientMsgID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return "sim-" + hex.EncodeToString(b[:])
}

// simMediaTransformVersion 服务端直传无客户端变换管线，与 UCG blob 索引键一致。
const simMediaTransformVersion = "sim-raw"

func uploadImageFromURL(ctx context.Context, token, imageURL string) (objectKey string, err error) {
	data, contentType, err := downloadURLBytes(ctx, imageURL, 10<<20)
	if err != nil {
		return "", err
	}
	ext := inferMediaExtension(contentType, imageURL, "png")
	return uploadMediaBytes(ctx, token, 1, ext, data)
}

func uploadVideoFromURL(ctx context.Context, token, videoURL string) (objectKey string, err error) {
	data, contentType, err := downloadURLBytes(ctx, videoURL, 100<<20)
	if err != nil {
		return "", err
	}
	ext := inferMediaExtension(contentType, videoURL, "mp4")
	return uploadMediaBytes(ctx, token, 2, ext, data)
}

// downloadURLBytes 下载远程媒体并限制最大字节，返回 body 与 Content-Type。
func downloadURLBytes(ctx context.Context, rawURL string, maxBytes int64) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
	if err != nil {
		return nil, "", err
	}
	if len(data) == 0 {
		return nil, "", fmt.Errorf("下载内容为空")
	}
	return data, resp.Header.Get("Content-Type"), nil
}

// inferMediaExtension 从 Content-Type 或 URL 路径推断扩展名（不含点）；失败时回退 defaultExt。
func inferMediaExtension(contentType, rawURL, defaultExt string) string {
	if ext := extensionFromContentType(contentType); ext != "" {
		return ext
	}
	if ext := extensionFromURL(rawURL); ext != "" {
		return ext
	}
	return defaultExt
}

func extensionFromContentType(contentType string) string {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch ct {
	case "image/png":
		return "png"
	case "image/jpeg", "image/jpg":
		return "jpg"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	case "video/mp4":
		return "mp4"
	case "video/quicktime":
		return "mov"
	default:
		return ""
	}
}

func extensionFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(u.Path), "."))
	switch ext {
	case "jpeg":
		return "jpg"
	case "png", "jpg", "webp", "gif", "mp4", "mov":
		return ext
	default:
		return ""
	}
}

func sha256HexLower(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// uploadMediaBytes 对齐 UCG media presign/register：extension + contentHash + transformVersion。
func uploadMediaBytes(ctx context.Context, token string, mediaKind int, ext string, data []byte) (string, error) {
	var presign struct {
		UploadUrl string            `json:"uploadUrl"`
		ObjectKey string            `json:"objectKey"`
		Headers   map[string]string `json:"headers"`
	}
	if err := appPost(ctx, token, "/ucg/app/api/media/presign", g.Map{
		"mediaKind": mediaKind,
		"extension": ext,
	}, &presign); err != nil {
		return "", err
	}
	putReq, err := http.NewRequestWithContext(ctx, http.MethodPut, presign.UploadUrl, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	contentType := ""
	if presign.Headers != nil {
		contentType = strings.TrimSpace(presign.Headers["Content-Type"])
	}
	if contentType == "" {
		contentType = contentTypeForSimMedia(mediaKind, ext)
	}
	putReq.Header.Set("Content-Type", contentType)
	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		return "", err
	}
	putResp.Body.Close()
	if putResp.StatusCode >= 300 {
		return "", fmt.Errorf("OSS PUT %d", putResp.StatusCode)
	}
	if err = appPost(ctx, token, "/ucg/app/api/media/register", g.Map{
		"objectKey":        presign.ObjectKey,
		"contentHash":      sha256HexLower(data),
		"transformVersion": simMediaTransformVersion,
		"mediaKind":        mediaKind,
		"dedupHit":         false,
	}, nil); err != nil {
		return "", err
	}
	return presign.ObjectKey, nil
}

func contentTypeForSimMedia(mediaKind int, ext string) string {
	if mediaKind == 2 {
		switch ext {
		case "mov":
			return "video/quicktime"
		default:
			return "video/mp4"
		}
	}
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	case "gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func sendInternalChat(ctx context.Context, senderWxID int64, convID uint64, content string) error {
	return ucgInternalPost(ctx, "/ucg/internal/api/chat/send", g.Map{
		"senderWxId": senderWxID, "conversationId": convID,
		"clientMsgId": newClientMsgID(), "content": content,
	}, nil)
}
