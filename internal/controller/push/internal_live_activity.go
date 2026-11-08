package pushctrl

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"hello/internal/platform/httpmeta"
	pushsvc "hello/internal/services/push"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// pushInternalLiveActivityBody voice/history 经 clients/push 调用的 Live Activity 编排请求。
type pushInternalLiveActivityBody struct {
	WxId          int64                  `json:"wxId"`
	Action        string                 `json:"action"` // start | update | end | upsert
	Attributes    map[string]interface{} `json:"attributes"`
	ContentState  map[string]interface{} `json:"contentState"`
	AlertTitle    string                 `json:"alertTitle"`
	AlertBody     string                 `json:"alertBody"`
	DismissalDate int64                  `json:"dismissalDate"`
}

// InternalLiveActivity POST /push/internal/api/live-activity — 各域经 clients/push 扇出 LA。
func InternalLiveActivity(r *ghttp.Request) {
	if r.Method != http.MethodPost {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	secret := strings.TrimSpace(r.GetHeader(httpmeta.HeaderDeviceGatewayInternalSecret))
	if secret == "" {
		secret = strings.TrimSpace(r.GetHeader("X-Gateway-Internal-Secret"))
	}
	if !httpmeta.ValidateInternalSecret(secret) {
		g.Log().Warningf(ctx, "[push-la] skip reason=auth_denied path=/push/internal/api/live-activity")
		r.Response.Status = 403
		r.Response.WriteJson(g.Map{"code": 403, "message": "内部接口未授权"})
		r.ExitAll()
		return
	}
	raw, err := io.ReadAll(r.Request.Body)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": 400, "message": "读取请求体失败"})
		return
	}
	var body pushInternalLiveActivityBody
	if len(raw) > 0 {
		if err = json.Unmarshal(raw, &body); err != nil {
			r.Response.WriteJson(g.Map{"code": 400, "message": "请求体无效"})
			return
		}
	}
	if body.WxId <= 0 || strings.TrimSpace(body.Action) == "" {
		r.Response.WriteJson(g.Map{"code": 400, "message": "wxId/action 必填"})
		return
	}
	g.Log().Warningf(ctx, "[push-la] accepted wxId=%d action=%s",
		body.WxId, strings.TrimSpace(body.Action))
	pushsvc.FanoutLiveActivity(ctx, pushsvc.LiveActivityOp{
		WxID:          body.WxId,
		Action:        body.Action,
		Attributes:    body.Attributes,
		ContentState:  body.ContentState,
		AlertTitle:    body.AlertTitle,
		AlertBody:     body.AlertBody,
		DismissalDate: body.DismissalDate,
	})
	r.Response.WriteJson(g.Map{"code": 0, "message": "OK", "data": g.Map{}})
}
