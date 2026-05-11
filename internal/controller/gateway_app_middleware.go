package controller

import (
	"strconv"
	"strings"
	"time"

	"hello/internal/platform/cachekit"
	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

var gatewayWxCodeCache = cachekit.WithObserver(cachekit.NewRedisCache(), cachekit.LoggingObserver{})

const cacheKeyGatewayWxID = "gw:app:wxid2code:"

// installGatewayAppBearerMiddleware 为下游代理请求注入 X-Internal-Wx-Code（基于 JWT access）。
func installGatewayAppBearerMiddleware(s *ghttp.Server) {
	s.BindMiddleware("/*", func(r *ghttp.Request) {
		if r.Method == "OPTIONS" {
			r.Middleware.Next()
			return
		}
		path := r.URL.Path
		if !gatewayAppPathNeedsBearer(path) {
			r.Middleware.Next()
			return
		}
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		const pfx = "Bearer "
		if len(auth) < len(pfx) || !strings.EqualFold(auth[:len(pfx)], pfx) {
			r.Response.Status = 401
			r.Response.WriteJson(g.Map{"code": 401, "message": "缺少或无效的 Authorization"})
			r.ExitAll()
			return
		}
		raw := strings.TrimSpace(auth[len(pfx):])
		ctx := r.Context()
		wxID, err := gatewayapp.ParseAccessWxID(ctx, raw)
		if err != nil || wxID <= 0 {
			r.Response.Status = 401
			r.Response.WriteJson(g.Map{"code": 401, "message": "access_token 无效或已过期"})
			r.ExitAll()
			return
		}
		cacheKey := cacheKeyGatewayWxID + strconv.FormatInt(wxID, 10)
		wxCode := ""
		if v, ok, e2 := gatewayWxCodeCache.Get(ctx, cacheKey); e2 == nil && ok && strings.TrimSpace(v) != "" {
			wxCode = strings.TrimSpace(v)
		} else {
			wxCode, err = gatewayapp.FetchWxCodeByID(ctx, wxID)
			if err != nil || wxCode == "" {
				r.Response.Status = 401
				r.Response.WriteJson(g.Map{"code": 401, "message": "无法解析 wx 身份"})
				r.ExitAll()
				return
			}
			ttlSec := g.Cfg().MustGet(ctx, "gatewayApp.wxIdCodeCacheSeconds").Int64()
			if ttlSec <= 0 {
				ttlSec = 120
			}
			_ = gatewayWxCodeCache.SetEX(ctx, cacheKey, wxCode, time.Duration(ttlSec)*time.Second)
		}
		r.Header.Set("X-Internal-Wx-Code", wxCode)
		r.Request.Header.Set("X-Internal-Wx-Code", wxCode)
		r.Middleware.Next()
	})
}

func gatewayAppPathNeedsBearer(path string) bool {
	if path == "/device/wx/api/detail" {
		return true
	}
	if strings.HasPrefix(path, "/device/history/api/") {
		return true
	}
	if strings.HasPrefix(path, "/device/profile/api/") {
		return true
	}
	if strings.HasPrefix(path, "/device/admin/api/") {
		return true
	}
	if strings.HasPrefix(path, "/voice/text/") {
		return true
	}
	return false
}
