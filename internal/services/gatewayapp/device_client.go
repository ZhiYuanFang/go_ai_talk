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

// FetchWxCodeByID 调用 device 内部接口解析 wxId → wxCode。
func FetchWxCodeByID(ctx context.Context, wxID int64) (string, error) {
	base := deviceBaseURL(ctx)
	if base == "" {
		return "", fmt.Errorf("DEVICE_SERVICE_URL 未配置")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		secret = strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.deviceInternalSecret").String())
	}
	url := fmt.Sprintf("%s/device/app/api/user/internal/by-id?id=%d", base, wxID)
	resp, err := gclient.New().SetHeader("X-Gateway-Internal-Secret", secret).Get(ctx, url)
	if err != nil {
		return "", err
	}
	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 {
		return "", fmt.Errorf("device 内部接口失败: %s", j.Get("message").String())
	}
	return strings.TrimSpace(j.Get("data.wxCode").String()), nil
}

// FetchWxDeviceNo 使用内部 wxCode 头查询 device_no（用于 WS 订阅校验）。
func FetchWxDeviceNo(ctx context.Context, wxCode string) (string, error) {
	base := deviceBaseURL(ctx)
	if base == "" {
		return "", fmt.Errorf("DEVICE_SERVICE_URL 未配置")
	}
	url := fmt.Sprintf("%s/device/app/api/user/detail", base)
	resp, err := gclient.New().SetHeader("X-Internal-Wx-Code", wxCode).Get(ctx, url)
	if err != nil {
		return "", err
	}
	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 {
		return "", fmt.Errorf("wx detail 失败: %s", j.Get("message").String())
	}
	return strings.TrimSpace(j.Get("data.deviceNo").String()), nil
}
