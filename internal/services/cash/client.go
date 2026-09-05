package cash

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var (
	remoteOnce sync.Once
	remoteIns  *remoteClient
)

type remoteClient struct {
	base   string
	secret string
	client *http.Client
}

type gfEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func remoteHTTP() *remoteClient {
	remoteOnce.Do(func() {
		base := strings.TrimRight(strings.TrimSpace(os.Getenv("CASH_SERVICE_URL")), "/")
		if base == "" {
			if v, err := g.Cfg().Get(context.Background(), "cashServiceUrl"); err == nil && v != nil && !v.IsEmpty() {
				base = strings.TrimRight(strings.TrimSpace(v.String()), "/")
			}
		}
		if base == "" {
			base = "http://127.0.0.1:9807"
		}
		secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
		remoteIns = &remoteClient{
			base:   base,
			secret: secret,
			client: &http.Client{Timeout: 5 * time.Second},
		}
	})
	return remoteIns
}

// RemoteIsVipByWxID 供 voice 等跨进程调用 cash internal VIP 接口。
// err != nil 时调用方应降级为非 VIP。
func RemoteIsVipByWxID(ctx context.Context, wxID int64) (bool, error) {
	if wxID <= 0 {
		return false, nil
	}
	c := remoteHTTP()
	if c == nil || c.base == "" {
		return false, gerror.NewCode(gcode.CodeInternalError, "CASH_SERVICE_URL 未配置")
	}
	if strings.TrimSpace(c.secret) == "" {
		return false, gerror.NewCode(gcode.CodeInternalError, "DEVICE_GATEWAY_INTERNAL_SECRET 未配置")
	}
	u, err := url.Parse(c.base + "/cash/internal/api/vip/by-wx-id")
	if err != nil {
		return false, err
	}
	q := u.Query()
	q.Set("wxId", strconv.FormatInt(wxID, 10))
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set(HeaderInternalSecret, c.secret)
	resp, err := c.client.Do(req)
	if err != nil {
		return false, gerror.NewCode(gcode.CodeOperationFailed, fmt.Sprintf("cash-service 不可达: %v", err))
	}
	defer resp.Body.Close()
	rawBody, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return false, err
	}
	var env gfEnvelope
	if len(rawBody) > 0 {
		if err = json.Unmarshal(rawBody, &env); err != nil {
			return false, gerror.NewCode(gcode.CodeInternalError, "cash VIP 响应非 JSON")
		}
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			msg = fmt.Sprintf("cash VIP HTTP %d", resp.StatusCode)
		}
		return false, gerror.NewCode(gcode.CodeOperationFailed, msg)
	}
	var data struct {
		IsVip bool `json:"isVip"`
	}
	if len(env.Data) > 0 && string(env.Data) != "null" {
		if err = json.Unmarshal(env.Data, &data); err != nil {
			return false, gerror.NewCode(gcode.CodeInternalError, "解析 cash VIP 响应失败")
		}
	}
	return data.IsVip, nil
}

// CareAlertAccessRemote 值得留意可看合成（voice 门禁用）。
type CareAlertAccessRemote struct {
	Allowed              bool  `json:"allowed"`
	FeedingQualified     bool  `json:"feedingQualified"`
	FeatureActive        bool  `json:"featureActive"`
	EntitlementExpiresAt int64 `json:"entitlementExpiresAt,omitempty"`
}

// RemoteCareAlertAccess 供 voice 调用 cash internal 值得留意 access。
// err != nil 时调用方 MUST fail-closed（不得当作已开通放行）。
func RemoteCareAlertAccess(ctx context.Context, deviceNo string, wxID int64) (*CareAlertAccessRemote, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	c := remoteHTTP()
	if c == nil || c.base == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, "CASH_SERVICE_URL 未配置")
	}
	if strings.TrimSpace(c.secret) == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, "DEVICE_GATEWAY_INTERNAL_SECRET 未配置")
	}
	u, err := url.Parse(c.base + "/cash/internal/api/care-alert/access")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("deviceNo", deviceNo)
	q.Set("wxId", strconv.FormatInt(wxID, 10))
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(HeaderInternalSecret, c.secret)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeOperationFailed, fmt.Sprintf("cash-service 不可达: %v", err))
	}
	defer resp.Body.Close()
	rawBody, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, err
	}
	var env gfEnvelope
	if len(rawBody) > 0 {
		if err = json.Unmarshal(rawBody, &env); err != nil {
			return nil, gerror.NewCode(gcode.CodeInternalError, "cash care-alert access 响应非 JSON")
		}
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			msg = fmt.Sprintf("cash care-alert access HTTP %d", resp.StatusCode)
		}
		return nil, gerror.NewCode(gcode.CodeOperationFailed, msg)
	}
	var data CareAlertAccessRemote
	if len(env.Data) > 0 && string(env.Data) != "null" {
		if err = json.Unmarshal(env.Data, &data); err != nil {
			return nil, gerror.NewCode(gcode.CodeInternalError, "解析 cash care-alert access 失败")
		}
	}
	return &data, nil
}
