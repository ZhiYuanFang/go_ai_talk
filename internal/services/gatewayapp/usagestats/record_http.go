package usagestats

import (
	"net/http"
	"strings"
	"time"

	"hello/internal/services/gatewayapp"
	"hello/internal/services/gatewayapp/apiregistry"

	"github.com/gogf/gf/v2/net/ghttp"
)

// ShouldSkipHTTPRecord 判断是否跳过 HTTP 使用统计（供标准库 *http.Request 使用）。
func ShouldSkipHTTPRecord(req *http.Request) bool {
	if req == nil || req.URL == nil {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(req.Header.Get("Upgrade")), "websocket") {
		return true
	}
	path := req.URL.Path
	if strings.HasPrefix(path, "/device/internal/") {
		return true
	}
	if strings.HasPrefix(path, "/device/admin/api/") {
		return true
	}
	if isStaticOrShellPath(path) {
		return true
	}
	if isMaintenanceAPI(req.Method, path) {
		return true
	}
	return false
}

// RecordHTTPRequest 在响应状态码确定后写入统计（适用于反代 ModifyResponse 等场景）。
func RecordHTTPRequest(req *http.Request, status int) {
	if req == nil || req.URL == nil || !IsSuccessStatus(status) || ShouldSkipHTTPRecord(req) {
		return
	}
	apiregistry.Init()
	apiKey, _ := apiregistry.Normalize(req.Method, req.URL.Path)
	wxID := wxIDFromHTTPRequest(req)
	RecordAsync(wxID, apiKey, time.Now())
}

// RecordGHTTPRequest 在 ghttp 请求完成且状态码已知后写入统计（适用于网关本机 Handler）。
func RecordGHTTPRequest(r *ghttp.Request) {
	if r == nil || r.Request == nil {
		return
	}
	if ShouldSkipRecord(r) {
		return
	}
	status := r.Response.Status
	if status == 0 {
		status = http.StatusOK
	}
	if !IsSuccessStatus(status) {
		return
	}
	RecordHTTPRequest(r.Request, status)
}

func wxIDFromHTTPRequest(req *http.Request) int64 {
	if req == nil {
		return 0
	}
	raw := strings.TrimSpace(req.Header.Get(gatewayapp.HeaderInternalWxId))
	if raw == "" {
		return 0
	}
	return parseWxID(raw)
}
