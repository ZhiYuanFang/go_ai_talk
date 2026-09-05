package ucgctrl

import (
	"hello/internal/platform/httpmeta"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	deviceclient "hello/internal/clients/device"
	ucgsvc "hello/internal/services/ucg"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type ucgInternalChatSendBody struct {
	SenderWxId     int64  `json:"senderWxId"`
	ConversationId uint64 `json:"conversationId"`
	ClientMsgId    string `json:"clientMsgId"`
	Content        string `json:"content"`
	ImageKey       string `json:"imageKey"`
	VideoKey       string `json:"videoKey"`
}

// ucgInternalChatSend POST /ucg/internal/api/chat/send — 仅允许 is_simulated 发送方。
func InternalChatSend(r *ghttp.Request) {
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
	var body ucgInternalChatSendBody
	raw, err := io.ReadAll(r.Request.Body)
	if err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidParameter.Code(), "message": "读取请求体失败"})
		return
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidParameter.Code(), "message": "请求体无效"})
		return
	}
	if body.SenderWxId <= 0 || body.ConversationId == 0 {
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidParameter.Code(), "message": "senderWxId/conversationId 无效"})
		return
	}
	batch, err := deviceclient.UcgAPI().BatchWx(ctx, []int64{body.SenderWxId})
	if err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeOperationFailed.Code(), "message": err.Error()})
		return
	}
	item, ok := batch[body.SenderWxId]
	if !ok || !item.Exists || !item.IsSimulated {
		r.Response.Status = 403
		r.Response.WriteJson(g.Map{"code": 403, "message": "仅模拟用户可调用内部发消息"})
		return
	}
	msg, err := ucgsvc.ProcessOutboundChatMessage(ctx, body.SenderWxId, body.ConversationId,
		strings.TrimSpace(body.ClientMsgId), strings.TrimSpace(body.Content),
		strings.TrimSpace(body.ImageKey), strings.TrimSpace(body.VideoKey))
	if err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeOperationFailed.Code(), "message": gerror.Current(err).Error()})
		return
	}
	// T5 闭环：sim 回复后清零发送方未读，避免下轮 sample 重复命中同一会话。
	if err = ucgsvc.MarkConversationRead(ctx, body.SenderWxId, body.ConversationId, msg.ID); err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeOperationFailed.Code(), "message": gerror.Current(err).Error()})
		return
	}
	r.Response.WriteJson(g.Map{"code": 0, "message": "OK", "data": g.Map{}})
}
