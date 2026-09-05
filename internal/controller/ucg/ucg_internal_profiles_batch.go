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

type ucgInternalProfilesBatchBody struct {
	WxIds []int64 `json:"wxIds"`
}

// ucgInternalProfilesBatch POST /ucg/internal/api/profiles/batch — 批量公开 profile 展示字段。
func InternalProfilesBatch(r *ghttp.Request) {
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
	var body ucgInternalProfilesBatchBody
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
	if body.WxIds == nil {
		body.WxIds = []int64{}
	}
	list, err := ucgsvc.BatchPublicProfilesForInternal(ctx, body.WxIds)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": 500, "message": err.Error()})
		return
	}
	items := make([]g.Map, 0, len(list))
	for _, row := range list {
		items = append(items, g.Map{
			"wxId":               int64(row.WxId),
			"nickname":           row.Nickname,
			"avatarUrl":          row.AvatarUrl,
			"avatarThumbnailUrl": row.AvatarThumbnailUrl,
		})
	}
	r.Response.WriteJson(g.Map{"code": 0, "message": "OK", "data": g.Map{"list": items}})
}
