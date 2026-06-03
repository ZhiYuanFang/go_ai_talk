package deviceaccess

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	headerInternalDeviceNo = "X-Internal-Device-No"
	headerDeviceNo         = "X-Device-No"
	maxAPIPathLen          = 256
)

// FormatAPIPath 生成落库的对外路径：METHOD /path（不含 query）。
func FormatAPIPath(r *ghttp.Request) string {
	if r == nil {
		return ""
	}
	p := strings.TrimSpace(r.Method) + " " + strings.TrimSpace(r.URL.Path)
	if len(p) > maxAPIPathLen {
		return p[:maxAPIPathLen]
	}
	return p
}

// ShouldSkipTouch 判断是否跳过 API 访问 touch（WebSocket、internal、静态等）。
func ShouldSkipTouch(r *ghttp.Request) bool {
	if r == nil {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket") {
		return true
	}
	path := r.URL.Path
	if strings.HasPrefix(path, "/device/internal/") {
		return true
	}
	return false
}

// ExtractDeviceNo 从 query、可信头或 JSON body 解析 deviceNo；若读取了 body 则返回还原后的 Body 供下游使用。
func ExtractDeviceNo(r *ghttp.Request) (deviceNo string, restoredBody io.ReadCloser) {
	if r == nil {
		return "", nil
	}
	if dn := strings.TrimSpace(r.GetQuery("deviceNo").String()); dn != "" {
		return dn, nil
	}
	if dn := strings.TrimSpace(r.GetHeader(headerInternalDeviceNo)); dn != "" {
		return dn, nil
	}
	if dn := strings.TrimSpace(r.GetHeader(headerDeviceNo)); dn != "" {
		return dn, nil
	}
	if !methodMayHaveJSONBody(r.Method) {
		return "", nil
	}
	ct := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if !strings.Contains(ct, "json") {
		return "", nil
	}
	raw, err := io.ReadAll(r.Request.Body)
	if err != nil {
		return "", nil
	}
	_ = r.Request.Body.Close()
	restoredBody = io.NopCloser(bytes.NewReader(raw))
	r.Request.Body = restoredBody
	if len(raw) == 0 {
		return "", restoredBody
	}
	dn := strings.TrimSpace(gjson.New(raw).Get("deviceNo").String())
	return dn, restoredBody
}

func methodMayHaveJSONBody(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return false
	}
}
