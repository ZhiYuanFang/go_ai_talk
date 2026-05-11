package gatewayapp

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var wxBearerWxCodeCache = cachekit.WithObserver(cachekit.NewRedisCache(), cachekit.LoggingObserver{})

const cacheKeyWxBearerWxID = "gw:app:wxid2code:"

// 与 gateway-app 中间件历史文案保持一致，便于客户端统一处理。
var (
	ErrGatewayBearerMissing = errors.New("缺少或无效的 Authorization")
	ErrGatewayBearerInvalid = errors.New("access_token 无效或已过期")
	ErrGatewayWxCodeResolve = errors.New("无法解析 wx 身份")
)

// InjectWxCodeFromBearer 解析 Authorization Bearer JWT，将 wxCode 写入 r 的请求头。
// gateway-app-server 在 HookBeforeServe 中统一调用，先于各反代 BindMiddleware，使 /device/app/api/user/* 等到 device 的流量
// 在无需各 *_route_proxy 重复鉴权的情况下也能带上 X-Internal-Wx-Code（白名单路径由 controller 侧豁免）。
func InjectWxCodeFromBearer(r *ghttp.Request) error {
	ctx := r.Context()
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	const pfx = "Bearer "
	if len(auth) < len(pfx) || !strings.EqualFold(auth[:len(pfx)], pfx) {
		return ErrGatewayBearerMissing
	}
	raw := strings.TrimSpace(auth[len(pfx):])
	wxID, err := ParseAccessWxID(ctx, raw)
	if err != nil || wxID <= 0 {
		return ErrGatewayBearerInvalid
	}
	cacheKey := cacheKeyWxBearerWxID + strconv.FormatInt(wxID, 10)
	wxCode := ""
	if v, ok, e2 := wxBearerWxCodeCache.Get(ctx, cacheKey); e2 == nil && ok && strings.TrimSpace(v) != "" {
		wxCode = strings.TrimSpace(v)
	} else {
		wxCode, err = FetchWxCodeByID(ctx, wxID)
		if err != nil || wxCode == "" {
			return ErrGatewayWxCodeResolve
		}
		ttlSec := g.Cfg().MustGet(ctx, "gatewayApp.wxIdCodeCacheSeconds").Int64()
		if ttlSec <= 0 {
			ttlSec = 120
		}
		_ = wxBearerWxCodeCache.SetEX(ctx, cacheKey, wxCode, time.Duration(ttlSec)*time.Second)
	}
	r.Header.Set("X-Internal-Wx-Code", wxCode)
	r.Request.Header.Set("X-Internal-Wx-Code", wxCode)
	return nil
}
