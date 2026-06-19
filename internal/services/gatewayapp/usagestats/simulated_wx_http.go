package usagestats

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
)

// fetchWxIsSimulatedHTTP 经 device internal batch 查询 isSimulated（gateway 无 device 库直连）。
func fetchWxIsSimulatedHTTP(ctx context.Context, wxID int64) (bool, error) {
	base := strings.TrimSpace(os.Getenv("DEVICE_SERVICE_URL"))
	if base == "" {
		base = strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.deviceServiceUrl").String())
	}
	base = strings.TrimRight(base, "/")
	if base == "" {
		return false, fmt.Errorf("DEVICE_SERVICE_URL 未配置")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		secret = strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.deviceInternalSecret").String())
	}
	url := base + "/device/internal/api/ucg/wx/batch"
	resp, err := gclient.New().
		SetHeader("X-Device-Gateway-Internal-Secret", secret).
		ContentJson().
		Post(ctx, url, g.Map{"wxIds": []int64{wxID}})
	if err != nil {
		return false, err
	}
	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 {
		return false, fmt.Errorf("device batch: %s", j.Get("message").String())
	}
	list := j.Get("data.list").Array()
	for _, item := range list {
		ji := gjson.New(item)
		if ji.Get("wxId").Int64() == wxID {
			return ji.Get("isSimulated").Bool(), nil
		}
	}
	return false, nil
}
