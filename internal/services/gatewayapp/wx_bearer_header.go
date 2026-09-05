package gatewayapp

import (
	"errors"
	"strconv"
	"strings"

	"hello/internal/platform/httpmeta"

	"github.com/gogf/gf/v2/net/ghttp"
)

// 与 gateway-app 中间件历史文案保持一致，便于客户端统一处理。
var (
	ErrGatewayBearerMissing = errors.New("缺少或无效的 Authorization")
	ErrGatewayBearerInvalid = errors.New("access_token 无效或已过期")
)

// HeaderInternalWxId 下游识别 wx 行（主键），由网关注入（别名 httpmeta）。
const HeaderInternalWxId = httpmeta.HeaderInternalWxId

// HeaderInternalDeviceNo 下游识别当前会话绑定设备号（别名 httpmeta）。
const HeaderInternalDeviceNo = httpmeta.HeaderInternalDeviceNo

// InjectAccessHeadersFromBearer 解析 Authorization Bearer JWT，写入 X-Internal-Wx-Id 与可选 X-Internal-Device-No。
// 不再调用 device internal/by-id 做 id→unionid；unionid 仅保留在 device 登录换票写库路径。
func InjectAccessHeadersFromBearer(r *ghttp.Request) error {
	ctx := r.Context()
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	const pfx = "Bearer "
	if len(auth) < len(pfx) || !strings.EqualFold(auth[:len(pfx)], pfx) {
		return ErrGatewayBearerMissing
	}
	raw := strings.TrimSpace(auth[len(pfx):])
	wxID, deviceNo, err := ParseAccessClaims(ctx, raw)
	if err != nil || wxID < 0 {
		return ErrGatewayBearerInvalid
	}
	// sub=0 表示仅设备会话：必须带 device_no claim，否则无法与伪造空 sub 区分且下游无法识别设备。
	if wxID == 0 && strings.TrimSpace(deviceNo) == "" {
		return ErrGatewayBearerInvalid
	}
	r.Header.Set(HeaderInternalWxId, strconv.FormatInt(wxID, 10))
	r.Request.Header.Set(HeaderInternalWxId, strconv.FormatInt(wxID, 10))
	if deviceNo != "" {
		r.Header.Set(HeaderInternalDeviceNo, deviceNo)
		r.Request.Header.Set(HeaderInternalDeviceNo, deviceNo)
	}
	return nil
}
