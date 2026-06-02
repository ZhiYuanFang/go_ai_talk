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
	Type           string `json:"type"`
	AccessToken    string `json:"accessToken"`
	AccessTokenAlt string `json:"access_token"`
	DeviceNo       string `json:"deviceNo"`
	DeviceNoAlt    string `json:"device_no"`
}

func (af gatewayAppAuthFrame) token() string {
	if t := strings.TrimSpace(af.AccessToken); t != "" {
		return t
	}
	return strings.TrimSpace(af.AccessTokenAlt)
}

func (af gatewayAppAuthFrame) deviceNo() string {
	if d := strings.TrimSpace(af.DeviceNo); d != "" {
		return d
	}
	return strings.TrimSpace(af.DeviceNoAlt)
}

type gatewayAppWsFrame struct {
	Type string `json:"type"`
}

func gatewayAppHistoryWS(r *ghttp.Request) {
	ctx := r.Context()
	rawWS, err := r.WebSocket()
	if err != nil {
		r.Response.WriteStatusExit(400, []byte(`{"type":"error","message":"WebSocket 握手失败"}`))
		return
	}
	conn := gatewayapp.NewHistoryWSConn(rawWS)
	deviceNo := ""
	defer func() {
		if deviceNo != "" {
			gatewayapp.HistoryHub().Unregister(deviceNo, conn)
		}
	}()
	_, msg, err := conn.ReadMessage()
	if err != nil {
		_ = conn.WriteJSON(g.Map{"type": "error", "message": "读取首帧失败"})
		return
	}
	var af gatewayAppAuthFrame
	if err := json.Unmarshal(msg, &af); err != nil {
		_ = conn.WriteJSON(g.Map{"type": "error", "message": "首帧非合法 JSON"})
		return
	}
	if strings.TrimSpace(strings.ToLower(af.Type)) != "auth" {
		_ = conn.WriteJSON(g.Map{"type": "error", "message": "首帧必须为 auth"})
		return
	}
	accessToken := af.token()
	wxID, deviceNoFromJWT, err := gatewayapp.ParseAccessClaims(ctx, accessToken)
	if err != nil || wxID < 0 {
		_ = conn.WriteJSON(g.Map{"type": "error", "message": "accessToken 无效"})
		return
	}
	if deviceNoFromJWT == "" {
		_ = conn.WriteJSON(g.Map{"type": "error", "message": "未绑定设备，无法订阅历史推送"})
		return
	}
	want := af.deviceNo()
	if want == "" || want != deviceNoFromJWT {
		_ = conn.WriteJSON(g.Map{"type": "error", "message": "device_no 与 token 不一致"})
		return
	}
	deviceNo = want
	gatewayapp.HistoryHub().Register(deviceNo, conn)
	if err := conn.WriteJSON(g.Map{"type": "auth_ok", "code": 0}); err != nil {
		glog.Warningf(ctx, "[gateway-app-ws] auth_ok 下发失败 deviceNo=%s err=%v", deviceNo, err)
		return
	}
	glog.Infof(ctx, "[gateway-app-ws] 已订阅历史推送 deviceNo=%s wxId=%d", deviceNo, wxID)
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if strings.EqualFold(strings.TrimSpace(string(msg)), "ping") {
			_ = conn.WriteJSON(g.Map{"type": "pong"})
			continue
		}
		var frame gatewayAppWsFrame
		if err := json.Unmarshal(msg, &frame); err != nil {
			continue
		}
		if strings.TrimSpace(strings.ToLower(frame.Type)) == "ping" {
			_ = conn.WriteJSON(g.Map{"type": "pong"})
		}
	}
}
