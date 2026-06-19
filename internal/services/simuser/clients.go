package simuser

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

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

func ucgInternalPost(ctx context.Context, path string, body interface{}, out interface{}) error {
	if err := waitOutboundRate(ctx, "ucg-internal"); err != nil {
		return err
	}
	base := ucgBase(ctx)
	if base == "" {
		return fmt.Errorf("UCG_SERVICE_URL 未配置")
	}
	resp, err := gclient.New().
		SetHeader("X-Device-Gateway-Internal-Secret", internalSecret(ctx)).
		ContentJson().Post(ctx, base+path, body)
	if err != nil {
		return err
	}
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

func listSimWxIDs(ctx context.Context) ([]int64, error) {
	base := deviceBase(ctx)
	resp, err := gclient.New().
		SetHeader("X-Device-Gateway-Internal-Secret", internalSecret(ctx)).
		Get(ctx, base+"/device/internal/api/sim/wx/list?page=1&pageSize=200")
	if err != nil {
		return nil, err
	}
	var data struct {
		List []struct {
			WxId int64 `json:"wxId"`
		} `json:"list"`
	}
	if err = parseEnvelope(resp.ReadAllString(), &data); err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(data.List))
	for _, item := range data.List {
		out = append(out, item.WxId)
	}
	return out, nil
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
	ids, err := listSimWxIDs(ctx)
	if err != nil || len(ids) == 0 {
		return loginSession{}, "", fmt.Errorf("无模拟用户")
	}
	idx := time.Now().UnixNano() % int64(len(ids))
	// 需要 account 登录：从 list 拿 account
	base := deviceBase(ctx)
	resp, err := gclient.New().
		SetHeader("X-Device-Gateway-Internal-Secret", internalSecret(ctx)).
		Get(ctx, base+"/device/internal/api/sim/wx/list?page=1&pageSize=200")
	if err != nil {
		return loginSession{}, "", err
	}
	var data struct {
		List []struct {
			WxId    int64  `json:"wxId"`
			Account string `json:"account"`
		} `json:"list"`
	}
	if err = parseEnvelope(resp.ReadAllString(), &data); err != nil {
		return loginSession{}, "", err
	}
	if len(data.List) == 0 {
		return loginSession{}, "", fmt.Errorf("无模拟用户")
	}
	item := data.List[idx%int64(len(data.List))]
	sess, err := usernameLogin(ctx, item.Account, password)
	return sess, item.Account, err
}

func newClientMsgID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return "sim-" + hex.EncodeToString(b[:])
}

func uploadImageFromURL(ctx context.Context, token, imageURL string) (objectKey string, err error) {
	imgResp, err := http.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer imgResp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(imgResp.Body, 10<<20))
	if err != nil {
		return "", err
	}
	var presign struct {
		UploadUrl string `json:"uploadUrl"`
		ObjectKey string `json:"objectKey"`
	}
	if err = appPost(ctx, token, "/ucg/app/api/media/presign", g.Map{"mediaKind": 1, "contentType": "image/png"}, &presign); err != nil {
		return "", err
	}
	putReq, err := http.NewRequestWithContext(ctx, http.MethodPut, presign.UploadUrl, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	putReq.Header.Set("Content-Type", "image/png")
	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		return "", err
	}
	putResp.Body.Close()
	if putResp.StatusCode >= 300 {
		return "", fmt.Errorf("OSS PUT %d", putResp.StatusCode)
	}
	if err = appPost(ctx, token, "/ucg/app/api/media/register", g.Map{
		"objectKey": presign.ObjectKey, "mediaKind": 1, "sizeBytes": len(data),
	}, nil); err != nil {
		return "", err
	}
	return presign.ObjectKey, nil
}

func uploadVideoFromURL(ctx context.Context, token, videoURL string) (objectKey string, err error) {
	vResp, err := http.Get(videoURL)
	if err != nil {
		return "", err
	}
	defer vResp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(vResp.Body, 100<<20))
	if err != nil {
		return "", err
	}
	var presign struct {
		UploadUrl string `json:"uploadUrl"`
		ObjectKey string `json:"objectKey"`
	}
	if err = appPost(ctx, token, "/ucg/app/api/media/presign", g.Map{"mediaKind": 2, "contentType": "video/mp4"}, &presign); err != nil {
		return "", err
	}
	putReq, err := http.NewRequestWithContext(ctx, http.MethodPut, presign.UploadUrl, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	putReq.Header.Set("Content-Type", "video/mp4")
	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		return "", err
	}
	putResp.Body.Close()
	if putResp.StatusCode >= 300 {
		return "", fmt.Errorf("OSS PUT %d", putResp.StatusCode)
	}
	if err = appPost(ctx, token, "/ucg/app/api/media/register", g.Map{
		"objectKey": presign.ObjectKey, "mediaKind": 2, "sizeBytes": len(data),
	}, nil); err != nil {
		return "", err
	}
	return presign.ObjectKey, nil
}

func sendInternalChat(ctx context.Context, senderWxID int64, convID uint64, content string) error {
	return ucgInternalPost(ctx, "/ucg/internal/api/chat/send", g.Map{
		"senderWxId": senderWxID, "conversationId": convID,
		"clientMsgId": newClientMsgID(), "content": content,
	}, nil)
}
