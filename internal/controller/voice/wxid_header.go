package voicectrl

import (
	"hello/internal/platform/httpmeta"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// wxIDFromAppUserHeader 从网关注入头解析 wx 主键（voice App API 共用）。
func wxIDFromAppUserHeader(r *ghttp.Request) (int64, error) {
	id, msg := httpmeta.RequireHeaderWxID(r.GetHeader(httpmeta.HeaderInternalWxId))
	if msg != "" {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, msg)
	}
	return id, nil
}
