package cashctrl

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"hello/internal/services/cash"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterAlipayNotify 支付宝异步通知：纯文本 success/failure，不用标准 JSON envelope。
//
// 业务说明：URL 历史路径仍为 /cash/app/api/vip/alipay/notify，供功能单与小智 MCP 共用；
// VIP 履约已下线，回调经 DispatchFulfillPaid 仅履约 feature_order。
//
// Side Effects: 注册 POST 路由；成功履约写订单/权益。
func RegisterAlipayNotify(s *ghttp.Server) {
	s.BindHandler("POST:/cash/app/api/vip/alipay/notify", func(r *ghttp.Request) {
		_ = r.Request.ParseForm()
		form := map[string]string{}
		for k, vals := range r.Request.Form {
			if len(vals) > 0 {
				form[k] = vals[0]
			}
		}
		if len(form) == 0 {
			var body map[string]interface{}
			if err := r.Parse(&body); err == nil {
				for k, v := range body {
					form[k] = stringifyNotifyVal(v)
				}
			}
		}
		ack, err := cash.HandleAlipayNotify(r.Context(), form)
		if err != nil || ack != "success" {
			r.Response.WriteStatus(200)
			r.Response.Write("failure")
			return
		}
		r.Response.WriteStatus(200)
		r.Response.Write("success")
	})
}

// RegisterAppleNotifications Apple ASN V2：读 JSON signedPayload，同步验签履约；失败非 2xx 触发苹果重试。
//
// 业务说明：URL 历史路径仍为 /cash/app/api/vip/apple/notifications，供功能单与小智 MCP 共用；
// 无 MQ、无标准 JSON envelope（对齐支付宝 notify 特殊写出风格）。
//
// Side Effects: 注册 POST 路由；成功履约或退款写订单/权益。
func RegisterAppleNotifications(s *ghttp.Server) {
	s.BindHandler("POST:/cash/app/api/vip/apple/notifications", func(r *ghttp.Request) {
		var body struct {
			SignedPayload string `json:"signedPayload"`
		}
		if err := r.Parse(&body); err != nil || strings.TrimSpace(body.SignedPayload) == "" {
			// 兼容原始 body 仅为 JWS 字符串的误配（少见）
			raw := strings.TrimSpace(r.GetBodyString())
			if strings.HasPrefix(raw, "eyJ") {
				body.SignedPayload = raw
			}
		}
		if err := cash.HandleAppleNotification(r.Context(), body.SignedPayload); err != nil {
			r.Response.WriteStatus(http.StatusInternalServerError)
			r.Response.Write([]byte(err.Error()))
			return
		}
		r.Response.WriteStatus(http.StatusOK)
		r.Response.Write([]byte("ok"))
	})
}

// stringifyNotifyVal 将支付宝 notify 字段值规范为字符串（表单/JSON 混用兼容）。
func stringifyNotifyVal(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}
