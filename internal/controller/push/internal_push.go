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

type pushInternalByBizBody struct {
	WxId    int64             `json:"wxId"`
	BizType string            `json:"bizType"`
	Alert   string            `json:"alert"`
	Badge   int               `json:"badge"`
	Silent  bool              `json:"silent"`
	Data    map[string]string `json:"data"`
}

// InternalByBizType POST /push/internal/api/by-biz-type — 各域经 clients/push 调用。
// 业务：校验内部密钥后异步下发；Warning 级日志便于 GF_LOGGER_LEVEL=prod 下对齐调用方。
func InternalByBizType(r *ghttp.Request) {
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
		// prod 可见：区分「请求未进业务」与「进了但无 token」。
		g.Log().Warningf(ctx, "[push] skip reason=auth_denied path=/push/internal/api/by-biz-type")
		r.Response.Status = 403
		r.Response.WriteJson(g.Map{"code": 403, "message": "内部接口未授权"})
		r.ExitAll()
		return
	}
	raw, err := io.ReadAll(r.Request.Body)
	if err != nil {
		g.Log().Warningf(ctx, "[push] skip reason=bad_request detail=read_body_failed err=%v", err)
		r.Response.WriteJson(g.Map{"code": 400, "message": "读取请求体失败"})
		return
	}
	var body pushInternalByBizBody
	if len(raw) > 0 {
		if err = json.Unmarshal(raw, &body); err != nil {
			g.Log().Warningf(ctx, "[push] skip reason=bad_request detail=invalid_json err=%v", err)
			r.Response.WriteJson(g.Map{"code": 400, "message": "请求体无效"})
			return
		}
	}
	if body.WxId <= 0 || strings.TrimSpace(body.BizType) == "" {
		g.Log().Warningf(ctx, "[push] skip reason=bad_request detail=wxId_or_bizType wxId=%d bizTypeLen=%d",
			body.WxId, len(strings.TrimSpace(body.BizType)))
		r.Response.WriteJson(g.Map{"code": 400, "message": "wxId/bizType 必填"})
		return
	}
	if !body.Silent && strings.TrimSpace(body.Alert) == "" && strings.TrimSpace(strings.ToLower(body.BizType)) != pushsvc.PushBizUcgSilentBadge {
		g.Log().Warningf(ctx, "[push] skip reason=bad_request detail=alert_required wxId=%d bizType=%s",
			body.WxId, strings.TrimSpace(body.BizType))
		r.Response.WriteJson(g.Map{"code": 400, "message": "非静默推送 alert 必填"})
		return
	}
	alertLen := len([]rune(strings.TrimSpace(body.Alert)))
	// 受理成功：不打印 alert 全文，仅 alertLen。
	g.Log().Warningf(ctx, "[push] accepted wxId=%d bizType=%s badge=%d silent=%v alertLen=%d",
		body.WxId, strings.TrimSpace(body.BizType), body.Badge, body.Silent, alertLen)
	pushsvc.PushByBizType(ctx, body.WxId, body.BizType, body.Alert, body.Badge, body.Silent, body.Data)
	r.Response.WriteJson(g.Map{"code": 0, "message": "OK", "data": g.Map{}})
}
