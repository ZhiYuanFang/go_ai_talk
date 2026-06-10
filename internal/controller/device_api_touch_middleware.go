package controller

import (
	"strings"

	"hello/internal/services/device"
	"hello/internal/services/gatewayapp"
	"hello/internal/shared/deviceaccess"

	"github.com/gogf/gf/v2/net/ghttp"
)

// installDeviceAPIAccessTouchMiddleware 在网关边缘记录带 deviceNo 的 HTTP API 访问（不含 WebSocket）。
func installDeviceAPIAccessTouchMiddleware(s *ghttp.Server) {
	s.BindMiddleware("/*", func(r *ghttp.Request) {
		if deviceaccess.ShouldSkipTouch(r) {
			r.Middleware.Next()
			return
		}
		deviceNo, _ := deviceaccess.ExtractDeviceNo(r)
		if deviceNo == "" {
			r.Middleware.Next()
			return
		}
		apiPath := deviceaccess.FormatAPIPath(r)
		base := deviceServiceBaseForTouch(r)
		device.TouchAPIAccessAsync(base, deviceNo, apiPath)
		r.Middleware.Next()
	})
}

// deviceServiceBaseForTouch 解析 device-service 基址：App 网关优先环境变量，主网关回退反代目标。
func deviceServiceBaseForTouch(r *ghttp.Request) string {
	if r != nil {
		if base := gatewayapp.DeviceServiceBaseURL(r.Context()); base != "" {
			return base
		}
	}
	return gatewayDeviceServiceTarget()
}

// gatewayDeviceServiceTarget 主网关反代 device 时使用的下游基址（与 DEVICE_API_PROXY_URL 一致）。
func gatewayDeviceServiceTarget() string {
	cfg, _ := deviceProxyFromEnv()
	if t := strings.TrimRight(strings.TrimSpace(cfg.targetURL), "/"); t != "" {
		return t
	}
	return "http://127.0.0.1:9803"
}
