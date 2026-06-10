package controller

import (
	"strconv"
	"strings"
	"time"

	"hello/internal/services/gatewayapp"
	"hello/internal/services/gatewayapp/apiregistry"
	"hello/internal/services/gatewayapp/usagestats"

	"github.com/gogf/gf/v2/net/ghttp"
)

// installGatewayAppAPIUsageStatsMiddleware 在响应完成后记录 2xx App API 使用统计。
func installGatewayAppAPIUsageStatsMiddleware(s *ghttp.Server) {
	apiregistry.Init()
	s.BindMiddleware("/*", func(r *ghttp.Request) {
		r.Middleware.Next()
		if usagestats.ShouldSkipRecord(r) {
			return
		}
		if !usagestats.IsSuccessStatus(r.Response.Status) {
			return
		}
		apiKey, _ := apiregistry.Normalize(r.Method, r.URL.Path)
		wxId := extractWxIDFromRequest(r)
		usagestats.RecordAsync(wxId, apiKey, time.Now())
	})
}

// extractWxIDFromRequest 从网关注入头解析 wxId；无效或缺失返回 0。
func extractWxIDFromRequest(r *ghttp.Request) int64 {
	if r == nil {
		return 0
	}
	raw := strings.TrimSpace(r.GetHeader(gatewayapp.HeaderInternalWxId))
	if raw == "" {
		raw = strings.TrimSpace(r.Request.Header.Get(gatewayapp.HeaderInternalWxId))
	}
	if raw == "" {
		return 0
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 0 {
		return 0
	}
	return n
}
