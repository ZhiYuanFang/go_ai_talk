package gatewayapp

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
)

// DeviceServiceBaseURL 下游 device-service 基址（环境变量优先）。
func DeviceServiceBaseURL(ctx context.Context) string {
	u := strings.TrimSpace(os.Getenv("DEVICE_SERVICE_URL"))
	if u == "" {
		u = strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.deviceServiceUrl").String())
	}
	return strings.TrimRight(u, "/")
}

func deviceBaseURL(ctx context.Context) string {
	return DeviceServiceBaseURL(ctx)
}

// FetchDeviceNoByWxID 网关刷新 access 时调用：按 wx 主键取当前绑定 device_no（内部密钥）。
func FetchDeviceNoByWxID(ctx context.Context, wxID int64) (string, error) {
	base := deviceBaseURL(ctx)
	if base == "" {
		return "", fmt.Errorf("DEVICE_SERVICE_URL 未配置")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		secret = strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.deviceInternalSecret").String())
	}
	url := fmt.Sprintf("%s/device/app/api/user/internal/device-no-by-wx-id?wxId=%d", base, wxID)
	resp, err := gclient.New().SetHeader("X-Gateway-Internal-Secret", secret).Get(ctx, url)
	if err != nil {
		return "", err
	}
	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 {
		return "", fmt.Errorf("device 内部接口失败: %s", j.Get("message").String())
	}
	return strings.TrimSpace(j.Get("data.deviceNo").String()), nil
}
