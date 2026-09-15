package ucgctrl

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"hello/internal/platform/httpmeta"
	ucgsvc "hello/internal/services/ucg"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ucgInternalPushByBizBody POST /ucg/internal/api/push/by-biz-type 请求体。
type ucgInternalPushByBizBody struct {
	WxId    int64             `json:"wxId"`
	BizType string            `json:"bizType"`
	Alert   string            `json:"alert"`
	Data    map[string]string `json:"data"`
}

// InternalPushByBizType POST /ucg/internal/api/push/by-biz-type — voice 等跨域可见推送。
// 鉴权：X-Device-Gateway-Internal-Secret。Token 复用 App 注册的 ucg_push_device。
func InternalPushByBizType(r *ghttp.Request) {
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
	var body ucgInternalPushByBizBody
	if len(raw) > 0 {
		if err = json.Unmarshal(raw, &body); err != nil {
			r.Response.WriteJson(g.Map{"code": 400, "message": "请求体无效"})
			return
		}
	}
	if body.WxId <= 0 || strings.TrimSpace(body.Alert) == "" || strings.TrimSpace(body.BizType) == "" {
		r.Response.WriteJson(g.Map{"code": 400, "message": "wxId/bizType/alert 必填"})
		return
	}
	ucgsvc.PushByBizType(ctx, body.WxId, body.BizType, body.Alert, body.Data)
	r.Response.WriteJson(g.Map{"code": 0, "message": "OK", "data": g.Map{}})
}
