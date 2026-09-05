package ucgctrl

import (
	"hello/internal/platform/httpmeta"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	ucgsvc "hello/internal/services/ucg"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type ucgInternalForceAcquireBody struct {
	WxId int64  `json:"wxId"`
	Ref  string `json:"ref"`
}

// ucgInternalForceAcquire POST /ucg/internal/api/force/acquire — cash 获客成功后原力 +100。
func InternalForceAcquire(r *ghttp.Request) {
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
	var body ucgInternalForceAcquireBody
	if len(raw) > 0 {
		if err = json.Unmarshal(raw, &body); err != nil {
			r.Response.WriteJson(g.Map{"code": 400, "message": "请求体无效"})
			return
		}
	}
	if err = ucgsvc.AddInviteAcquisitionForce(ctx, body.WxId, body.Ref); err != nil {
		r.Response.WriteJson(g.Map{"code": 500, "message": err.Error()})
		return
	}
	r.Response.WriteJson(g.Map{"code": 0, "message": "OK", "data": g.Map{}})
}
