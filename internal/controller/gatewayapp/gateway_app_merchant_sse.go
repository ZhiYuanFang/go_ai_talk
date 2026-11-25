package gatewayappctrl

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hello/internal/platform/httpmeta"
	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

var errMerchantSSEDeviceURL = errors.New("DEVICE_SERVICE_URL 未配置")

// gatewayAppMerchantSSE GET /device/merchant/sse/babies — 商户 JWT；一连接多 device。
func gatewayAppMerchantSSE(r *ghttp.Request) {
	ctx := r.Context()
	midStr := strings.TrimSpace(r.Header.Get(httpmeta.HeaderInternalMerchantId))
	merchantID, err := strconv.ParseInt(midStr, 10, 64)
	if err != nil || merchantID <= 0 {
		WriteAuthJSON(r, 401, "缺少商户身份")
		return
	}

	devices, err := loadMerchantDevicesForSSE(ctx, r.Header.Get("Authorization"))
	if err != nil {
		glog.Warningf(ctx, "[gateway-merchant-sse] 拉取宝宝列表失败 merchantId=%d err=%v", merchantID, err)
		WriteAuthJSON(r, 502, "无法加载商户宝宝列表")
		return
	}

	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	r.Response.Header().Set("X-Accel-Buffering", "no")
	r.Response.WriteHeader(200)
	r.Response.Flush()

	base := gatewayapp.NewHistorySSEConn(r.Response)
	conn := gatewayapp.NewMerchantSSEConn(base, merchantID, devices)
	gatewayapp.MerchantSSE().Register(conn)
	defer gatewayapp.MerchantSSE().Unregister(conn)

	glog.Infof(ctx, "[gateway-merchant-sse] 已订阅 merchantId=%d devices=%d", merchantID, len(devices))

	ticker := time.NewTicker(historySSEHeartbeatInterval)
	defer ticker.Stop()
	reload := time.NewTicker(60 * time.Second)
	defer reload.Stop()

	auth := r.Header.Get("Authorization")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := conn.WriteComment("ping"); err != nil {
				glog.Warningf(ctx, "[gateway-merchant-sse] 心跳失败 merchantId=%d err=%v", merchantID, err)
				return
			}
		case <-reload.C:
			if next, e2 := loadMerchantDevicesForSSE(ctx, auth); e2 == nil {
				conn.SetDevices(next)
			}
		}
	}
}

func loadMerchantDevicesForSSE(ctx context.Context, authorization string) ([]string, error) {
	base := deviceServiceBase(ctx)
	if base == "" {
		return nil, errMerchantSSEDeviceURL
	}
	url := strings.TrimRight(base, "/") + "/device/merchant/api/baby/list"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(authorization)
	const pfx = "Bearer "
	if len(raw) >= len(pfx) && strings.EqualFold(raw[:len(pfx)], pfx) {
		raw = strings.TrimSpace(raw[len(pfx):])
	}
	mid, err := gatewayapp.ParseMerchantAccessClaims(ctx, raw)
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpmeta.HeaderInternalMerchantId, strconv.FormatInt(mid, 10))
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return nil, err
	}
	var env struct {
		Code int `json:"code"`
		Data struct {
			DeviceNos []string `json:"deviceNos"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		return nil, errors.New("baby list failed")
	}
	return env.Data.DeviceNos, nil
}
