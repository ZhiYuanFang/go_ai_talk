package controller

import (
	"encoding/json"
	"strings"

	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

type gatewayAppAuthFrame struct {
	Type        string `json:"type"`
	AccessToken string `json:"access_token"`
	DeviceNo    string `json:"deviceNo"`
}

func gatewayAppHistoryWS(r *ghttp.Request) {
	ctx := r.Context()
	ws, err := r.WebSocket()
	if err != nil {
		r.Response.WriteStatusExit(400, []byte(`{"type":"error","message":"WebSocket 握手失败"}`))
		return
	}
	deviceNo := ""
	defer func() {
		if deviceNo != "" {
			gatewayapp.HistoryHub().Unregister(deviceNo, ws)
		}
	}()
	_, msg, err := ws.ReadMessage()
	if err != nil {
		_ = ws.WriteJSON(g.Map{"type": "error", "message": "读取首帧失败"})
		return
	}
	var af gatewayAppAuthFrame
	if err := json.Unmarshal(msg, &af); err != nil {
		_ = ws.WriteJSON(g.Map{"type": "error", "message": "首帧非合法 JSON"})
		return
	}
	if strings.TrimSpace(strings.ToLower(af.Type)) != "auth" {
		_ = ws.WriteJSON(g.Map{"type": "error", "message": "首帧必须为 auth"})
		return
	}
	wxID, deviceNoFromJWT, err := gatewayapp.ParseAccessClaims(ctx, strings.TrimSpace(af.AccessToken))
	if err != nil || wxID < 0 {
		_ = ws.WriteJSON(g.Map{"type": "error", "message": "access_token 无效"})
		return
	}
	if deviceNoFromJWT == "" {
		_ = ws.WriteJSON(g.Map{"type": "error", "message": "未绑定设备，无法订阅历史推送"})
		return
	}
	want := strings.TrimSpace(af.DeviceNo)
	if want == "" || want != deviceNoFromJWT {
		_ = ws.WriteJSON(g.Map{"type": "error", "message": "device_no 与 token 不一致"})
		return
	}
	deviceNo = want
	gatewayapp.HistoryHub().Register(deviceNo, ws)
	_ = ws.WriteJSON(g.Map{"type": "auth_ok", "code": 0})
	glog.Infof(ctx, "[gateway-app-ws] 已订阅历史推送 deviceNo=%s wxId=%d", deviceNo, wxID)
	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			return
		}
	}
}
