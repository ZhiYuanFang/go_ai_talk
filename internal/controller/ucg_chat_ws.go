package controller

import (
	"encoding/json"
	"strings"
	"time"

	ucgsvc "hello/internal/services/ucg"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

type ucgWSInFrame struct {
	Type           string `json:"type"`
	Token          string `json:"token"`
	ClientMsgID    string `json:"clientMsgId"`
	ConversationID uint64 `json:"conversationId"`
	Content        string `json:"content"`
	ImageKey       string `json:"imageKey"`
	VideoKey       string `json:"videoKey"`
}

func registerUcgChatWS(s *ghttp.Server) {
	s.BindHandler("/ws/chat", ucgChatWS)
}

func ucgChatWS(r *ghttp.Request) {
	ctx := r.Context()
	ws, err := r.WebSocket()
	if err != nil {
		r.Response.Status = 400
		r.Response.WriteJson(map[string]interface{}{
			"type": "error", "stage": "handshake", "message": "WebSocket 握手失败",
		})
		return
	}

	var wxID int64
	authenticated := false
	defer func() {
		if wxID > 0 {
			ucgsvc.ChatWSHub().Unregister(wxID, ws)
		}
	}()

	readDeadline := 90 * time.Second
	for {
		_ = ws.SetReadDeadline(time.Now().Add(readDeadline))
		_, payload, rErr := ws.ReadMessage()
		if rErr != nil {
			return
		}
		var frame ucgWSInFrame
		if uErr := json.Unmarshal(payload, &frame); uErr != nil {
			_ = ws.WriteJSON(map[string]interface{}{"type": "error", "message": "invalid json"})
			continue
		}
		switch strings.ToLower(strings.TrimSpace(frame.Type)) {
		case "ping":
			_ = ws.WriteJSON(map[string]interface{}{"type": "pong"})
		case "auth":
			if authenticated {
				continue
			}
			id, aErr := ucgsvc.ParseWSAccessToken(ctx, strings.TrimSpace(frame.Token))
			if aErr != nil || id <= 0 {
				_ = ws.WriteJSON(map[string]interface{}{"type": "error", "stage": "auth", "message": "invalid token"})
				return
			}
			wxID = id
			authenticated = true
			ucgsvc.ChatWSHub().Register(wxID, ws)
			_ = ws.WriteJSON(map[string]interface{}{"type": "auth_ok", "wxId": wxID})
		case "message":
			if !authenticated {
				_ = ws.WriteJSON(map[string]interface{}{"type": "error", "stage": "auth", "message": "auth required"})
				continue
			}
			content := strings.TrimSpace(frame.Content)
			imageKey := strings.TrimSpace(frame.ImageKey)
			videoKey := strings.TrimSpace(frame.VideoKey)
			if frame.ConversationID == 0 || (content == "" && imageKey == "" && videoKey == "") {
				_ = ws.WriteJSON(map[string]interface{}{"type": "error", "message": "conversationId/content or media required"})
				continue
			}
			if pErr := ucgsvc.ProcessOutboundChatMessage(ctx, wxID, frame.ConversationID, strings.TrimSpace(frame.ClientMsgID), content, imageKey, videoKey); pErr != nil {
				glog.Warningf(ctx, "ucg chat message failed wxId=%d conv=%d: %v", wxID, frame.ConversationID, pErr)
				_ = ws.WriteJSON(map[string]interface{}{"type": "error", "message": pErr.Error()})
			}
		default:
			_ = ws.WriteJSON(map[string]interface{}{"type": "error", "message": "unknown type"})
		}
	}
}
