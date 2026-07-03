package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"hello/internal/services/device"
	ucgsvc "hello/internal/services/ucg"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type ucgInternalPostsSampleBody struct {
	Mode               string  `json:"mode"`
	Limit              int     `json:"limit"`
	ExcludeMediaTypes  []int   `json:"excludeMediaTypes"`
	ExcludeAuthorWxIds []int64 `json:"excludeAuthorWxIds"`
	ExcludeDebate      bool    `json:"excludeDebate"`
	OnlyDebate         bool    `json:"onlyDebate"`
}

// ucgInternalPostsSample POST /ucg/internal/api/posts/sample — 轻量已发布帖抽样。
func ucgInternalPostsSample(r *ghttp.Request) {
	if r.Method != http.MethodPost {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	secret := strings.TrimSpace(r.GetHeader(device.HeaderDeviceGatewayInternalSecret))
	if secret == "" {
		secret = strings.TrimSpace(r.GetHeader("X-Gateway-Internal-Secret"))
	}
	if !device.ValidateGatewayInternalSecretHeader(secret) {
		r.Response.Status = 403
		r.Response.WriteJson(g.Map{"code": 403, "message": "内部接口未授权"})
		r.ExitAll()
		return
	}
	var body ucgInternalPostsSampleBody
	raw, err := io.ReadAll(r.Request.Body)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": 400, "message": "读取请求体失败"})
		return
	}
	if len(raw) > 0 {
		if err = json.Unmarshal(raw, &body); err != nil {
			r.Response.WriteJson(g.Map{"code": 400, "message": "请求体无效"})
			return
		}
	}

	if body.ExcludeDebate && body.OnlyDebate {
		r.Response.WriteJson(g.Map{"code": 400, "message": "excludeDebate 与 onlyDebate 不可同时为 true"})
		return
	}
	filter := ucgsvc.PostSampleFilter{
		ExcludeMediaTypes:  body.ExcludeMediaTypes,
		ExcludeAuthorWxIds: body.ExcludeAuthorWxIds,
		ExcludeDebate:      body.ExcludeDebate,
		OnlyDebate:         body.OnlyDebate,
	}

	mode := strings.ToLower(strings.TrimSpace(body.Mode))
	var list []ucgsvc.PostSampleItem
	switch mode {
	case "", "latest":
		list, err = ucgsvc.SamplePublishedPosts(ctx, body.Limit, filter)
	case "random":
		list, err = ucgsvc.SampleRandomPublishedPost(ctx, filter)
	default:
		r.Response.WriteJson(g.Map{"code": 400, "message": "mode 无效，支持 latest 或 random"})
		return
	}
	if err != nil {
		r.Response.WriteJson(g.Map{"code": 500, "message": err.Error()})
		return
	}
	if list == nil {
		list = []ucgsvc.PostSampleItem{}
	}
	r.Response.WriteJson(g.Map{"code": 0, "message": "OK", "data": g.Map{"list": list}})
}

