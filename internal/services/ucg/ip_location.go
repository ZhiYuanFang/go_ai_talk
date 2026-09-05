package ucg

import (
	"context"
	"strings"
	"time"

	deviceclient "hello/internal/clients/device"
	"hello/internal/platform/cachekit"
	"hello/internal/platform/httpmeta"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

const ipLocationThrottleTTL = time.Hour

// ClientIPFromRequest 读取网关注入的真实客户端 IP。
func ClientIPFromRequest(r *ghttp.Request) string {
	if r == nil {
		return ""
	}
	return strings.TrimSpace(r.GetHeader(httpmeta.HeaderInternalClientIP))
}

// MaybeUpdateWxIpLocation 解析 clientIP → 属地，节流后写入 device wx 行；返回当前展示属地。
func MaybeUpdateWxIpLocation(ctx context.Context, wxID int64, clientIP string) (string, error) {
	if wxID <= 0 {
		return "", nil
	}
	location := ResolveIPLocation(ctx, clientIP)
	if location == "" {
		display, err := loadWxIpLocation(ctx, wxID)
		return display, err
	}
	throttleKey := cachekit.UCGIPLocationThrottleKey(wxID)
	throttled, err := ucgCache.Exists(ctx, throttleKey)
	if err == nil && throttled {
		return loadWxIpLocation(ctx, wxID)
	}
	if err := deviceclient.UcgAPI().UpdateIpLocation(ctx, wxID, location); err != nil {
		glog.Warningf(ctx, "[ucg-ip-location] 更新 wx IP 属地失败 wxId=%d err=%v", wxID, err)
		return loadWxIpLocation(ctx, wxID)
	}
	_ = ucgCache.SetEX(ctx, throttleKey, "1", ipLocationThrottleTTL)
	return location, nil
}

func loadWxIpLocation(ctx context.Context, wxID int64) (string, error) {
	batch, err := deviceclient.UcgAPI().BatchWx(ctx, []int64{wxID})
	if err != nil {
		return "", err
	}
	if item, ok := batch[wxID]; ok {
		return strings.TrimSpace(item.IpLocation), nil
	}
	return "", nil
}

// IpLocationForWxIDs 批量读取 wx IP 属地（device batch）。
func IpLocationForWxIDs(ctx context.Context, wxIDs []int64) (map[int64]string, error) {
	out := make(map[int64]string, len(wxIDs))
	if len(wxIDs) == 0 {
		return out, nil
	}
	batch, err := deviceclient.UcgAPI().BatchWx(ctx, wxIDs)
	if err != nil {
		return nil, err
	}
	for id, item := range batch {
		if loc := strings.TrimSpace(item.IpLocation); loc != "" {
			out[id] = loc
		}
	}
	return out, nil
}

// SnapshotPostIpLocation 发帖时快照当前 IP 属地（服务端解析，不接受客户端 body）。
func SnapshotPostIpLocation(ctx context.Context, clientIP string) string {
	return ResolveIPLocation(ctx, clientIP)
}
