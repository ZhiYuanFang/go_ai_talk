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

type ucgInternalChatSimUnreadSampleBody struct {
	SimWxIds     []int64 `json:"simWxIds"`
	MessageLimit int     `json:"messageLimit"`
}

// ucgInternalChatSimUnreadSample POST /ucg/internal/api/chat/sim-unread-sample — sim 未读会话轻量抽样。
func ucgInternalChatSimUnreadSample(r *ghttp.Request) {
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
	var body ucgInternalChatSimUnreadSampleBody
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
	if len(body.SimWxIds) == 0 {
		r.Response.WriteJson(g.Map{"code": 400, "message": "simWxIds 不能为空"})
		return
	}

	result, err := ucgsvc.SampleRandomSimUnreadChat(ctx, body.SimWxIds, body.MessageLimit)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": 500, "message": err.Error()})
		return
	}
	r.Response.WriteJson(g.Map{"code": 0, "message": "OK", "data": result})
}
