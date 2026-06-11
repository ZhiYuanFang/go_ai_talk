package usagestats

import (
	"strconv"
	"strings"

	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/net/ghttp"
)

// WxIDFromRequest 从网关注入头解析 wxId；无效或缺失返回 0。
func WxIDFromRequest(r *ghttp.Request) int64 {
	if r == nil {
		return 0
	}
	raw := strings.TrimSpace(r.GetHeader(gatewayapp.HeaderInternalWxId))
	if raw == "" && r.Request != nil {
		raw = strings.TrimSpace(r.Request.Header.Get(gatewayapp.HeaderInternalWxId))
	}
	return parseWxID(raw)
}

func parseWxID(raw string) int64 {
	if raw == "" {
		return 0
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 0 {
		return 0
	}
	return n
}
