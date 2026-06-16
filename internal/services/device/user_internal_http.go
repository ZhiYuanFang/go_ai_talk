package device

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"hello/internal/dao"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var (
	userInternalHTTPOnce sync.Once
	userInternalHTTPIns  *userInternalHTTPClient
)

type userInternalHTTPClient struct {
	base   string
	secret string
	client *http.Client
}

type userInternalGFEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func userInternalHTTP() *userInternalHTTPClient {
	userInternalHTTPOnce.Do(func() {
		base := strings.TrimRight(strings.TrimSpace(os.Getenv("DEVICE_SERVICE_URL")), "/")
		if base == "" {
			if v, err := g.Cfg().Get(context.Background(), "deviceServiceUrl"); err == nil && v != nil && !v.IsEmpty() {
				base = strings.TrimRight(strings.TrimSpace(v.String()), "/")
			}
		}
		if base == "" {
			base = "http://127.0.0.1:9803"
		}
		userInternalHTTPIns = &userInternalHTTPClient{
			base:   base,
			secret: resolveGatewayInternalSecret(),
			client: &http.Client{Timeout: 8 * time.Second},
		}
	})
	return userInternalHTTPIns
}

func resolveGatewayInternalSecret() string {
	if v := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET")); v != "" {
		return v
	}
	ctx := context.Background()
	for _, key := range []string{"gatewayApp.deviceInternalSecret", "ucg.deviceInternalSecret"} {
		v, err := g.Cfg().Get(ctx, key)
		if err != nil || v == nil || v.IsEmpty() {
			continue
		}
		if s := strings.TrimSpace(v.String()); s != "" {
			return s
		}
	}
	return ""
}

func (c *userInternalHTTPClient) doJSON(ctx context.Context, path string, query map[string]string, out interface{}) error {
	if c == nil || c.base == "" {
		return gerror.NewCode(gcode.CodeInternalError, "DEVICE_SERVICE_URL 未配置")
	}
	if strings.TrimSpace(c.secret) == "" {
		return gerror.NewCode(gcode.CodeInternalError, "DEVICE_GATEWAY_INTERNAL_SECRET 未配置")
	}
	u, err := url.Parse(c.base + path)
	if err != nil {
		return err
	}
	if len(query) > 0 {
		q := u.Query()
		for k, v := range query {
			if strings.TrimSpace(v) != "" {
				q.Set(k, strings.TrimSpace(v))
			}
		}
		u.RawQuery = q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set(HeaderDeviceGatewayInternalSecret, c.secret)
	resp, err := c.client.Do(req)
	if err != nil {
		return gerror.NewCode(gcode.CodeOperationFailed, fmt.Sprintf("device user internal 不可达: %v", err))
	}
	defer resp.Body.Close()
	rawBody, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return err
	}
	var env userInternalGFEnvelope
	if len(rawBody) > 0 {
		if err = json.Unmarshal(rawBody, &env); err != nil {
			return gerror.NewCode(gcode.CodeInternalError, "device user internal 响应非 JSON")
		}
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			msg = fmt.Sprintf("device user internal HTTP %d", resp.StatusCode)
		}
		return gerror.NewCode(gcode.CodeOperationFailed, msg)
	}
	if out != nil && len(env.Data) > 0 && string(env.Data) != "null" {
		if err = json.Unmarshal(env.Data, out); err != nil {
			return gerror.NewCode(gcode.CodeInternalError, "解析 device user internal 响应失败")
		}
	}
	return nil
}

// RemoteWxIDByDeviceNo 经 device user 域 internal API 按 deviceNo 反查 wxId。
func RemoteWxIDByDeviceNo(ctx context.Context, deviceNo string) (int64, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return 0, nil
	}
	var data struct {
		WxId int64 `json:"wxId"`
	}
	err := userInternalHTTP().doJSON(ctx, "/device/app/api/user/internal/wx-id-by-device-no", map[string]string{
		"deviceNo": deviceNo,
	}, &data)
	return data.WxId, err
}

// WxIDByDeviceNo 由 device_no 反查 wx 主键；未绑定返回 0（device-service 进程内直查）。
func WxIDByDeviceNo(ctx context.Context, deviceNo string) (int64, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return 0, nil
	}
	one, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().DeviceNo, deviceNo).Limit(1).One()
	if err != nil {
		return 0, err
	}
	if one.IsEmpty() {
		return 0, nil
	}
	return one["id"].Int64(), nil
}
