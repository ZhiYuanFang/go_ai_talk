package cashctrl

import (
	"context"
	"strings"

	"hello/internal/platform/httpmeta"
	"hello/internal/services/cash"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// cashWxIDFromHeader 从网关注入的 X-Internal-Wx-Id 解析账号主键。
//
// Args: ctx 须携带 *ghttp.Request。
// Returns: wxId；缺头或非法时返回业务错误。
func cashWxIDFromHeader(ctx context.Context) (int64, error) {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	wxID := httpmeta.ParseHeaderWxID(r.GetHeader(httpmeta.HeaderInternalWxId))
	if wxID <= 0 {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	return wxID, nil
}

// requireCashAdmin 校验网关注入的 X-Admin-Password（浏览器不得自带）。
//
// Args: ctx 须携带 *ghttp.Request。
// Returns: 口令错误时 CodeNotAuthorized。
func requireCashAdmin(ctx context.Context) error {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return gerror.NewCode(gcode.CodeNotAuthorized, "口令错误")
	}
	if !cash.VerifyCashAdminPassword(strings.TrimSpace(r.GetHeader(cash.HeaderAdminPassword))) {
		return gerror.NewCode(gcode.CodeNotAuthorized, "口令错误")
	}
	return nil
}
