package controller

import (
	"encoding/json"
	"strings"
	"sync"

	"hello/internal/services/contracts"
	"hello/internal/services/gatewayapp"
	voice "hello/internal/services/voice"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

type clinicAuthFrame struct {
	Type           string `json:"type"`
	AccessToken    string `json:"accessToken"`
	AccessTokenAlt string `json:"access_token"`
	DeviceNo       string `json:"deviceNo"`
	DeviceNoAlt    string `json:"device_no"`
}

func (f clinicAuthFrame) token() string {
	if t := strings.TrimSpace(f.AccessToken); t != "" {
		return t
	}
	return strings.TrimSpace(f.AccessTokenAlt)
}

func (f clinicAuthFrame) deviceNo() string {
	if d := strings.TrimSpace(f.DeviceNo); d != "" {
		return d
	}
	return strings.TrimSpace(f.DeviceNoAlt)
}

type clinicQuestionFrame struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func registerVoiceClinicWS(s *ghttp.Server) {
	// 胖宝 WS 不注册 VoiceWSManager，与 ASR WS 同策略。
	s.BindHandler("/voice/clinic/ws", voiceClinicWS)
}

func voiceClinicWS(r *ghttp.Request) {
	ctx := r.Context()
	ws, err := r.WebSocket()
	if err != nil {
		r.Response.Status = 400
		r.Response.WriteJson(map[string]interface{}{
			"type":    "error",
			"code":    1,
			"stage":   "handshake",
			"message": "WebSocket 握手失败",
		})
		return
	}

	authenticated := false
	var wxID int64
	var deviceNo string
	writeMu := sync.Mutex{}
	writeJSON := func(v interface{}) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		return ws.WriteJSON(v)
	}

	// 首帧 MUST 为 auth：解析 JWT 得 wxId>0，deviceNo 须与 JWT device_no 一致；禁止 ResolveVoiceWxID 反查替代登录。
	_, firstMsg, err := ws.ReadMessage()
	if err != nil {
		_ = writeJSON(map[string]interface{}{"type": "error", "message": "读取首帧失败"})
		return
	}
	var authFrame clinicAuthFrame
	if err := json.Unmarshal(firstMsg, &authFrame); err != nil {
		_ = writeJSON(map[string]interface{}{"type": "error", "message": "首帧非合法 JSON"})
		return
	}
	if strings.ToLower(strings.TrimSpace(authFrame.Type)) != "auth" {
		_ = writeJSON(map[string]interface{}{"type": "error", "message": "首帧必须为 auth"})
		return
	}
	token := authFrame.token()
	clientDeviceNo := authFrame.deviceNo()
	parsedWxID, deviceNoFromJWT, pErr := gatewayapp.ParseAccessClaims(ctx, token)
	if pErr != nil || parsedWxID <= 0 {
		_ = writeJSON(map[string]interface{}{"type": "error", "code": contracts.CodeAINotLoggedIn, "message": "请先登录账号"})
		return
	}
	if clientDeviceNo == "" {
		_ = writeJSON(map[string]interface{}{"type": "error", "code": 400, "message": "deviceNo 不能为空"})
		return
	}
	if deviceNoFromJWT != "" && !strings.EqualFold(clientDeviceNo, deviceNoFromJWT) {
		_ = writeJSON(map[string]interface{}{"type": "error", "code": 400, "message": "deviceNo 与登录凭证不一致"})
		return
	}
	wxID = parsedWxID
	deviceNo = clientDeviceNo
	authenticated = true
	if err := writeJSON(map[string]interface{}{"type": "auth_ok", "code": 0}); err != nil {
		return
	}

	svc := voice.Clinic()
	// auth_ok 后立即下发 session_sync（每次重连重复）；读 Redis 失败时下发空 turns，不阻断后续 question。
	syncPayload, syncErr := svc.BuildSessionSync(ctx, wxID)
	if syncErr != nil {
		glog.Warningf(ctx, "[胖宝WS] session_sync 读取失败 wxId=%d err=%v", wxID, syncErr)
		syncPayload = voice.SessionSyncPayload{Type: "session_sync", Turns: []voice.SessionSyncTurn{}, ExpiresAt: 0}
	}
	if err := writeJSON(syncPayload); err != nil {
		return
	}

	for {
		_, msg, rErr := ws.ReadMessage()
		if rErr != nil {
			return
		}
		if !authenticated {
			_ = writeJSON(map[string]interface{}{"type": "error", "message": "未鉴权"})
			continue
		}
		var base struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(msg, &base); err != nil {
			_ = writeJSON(map[string]interface{}{"type": "error", "message": "非法 JSON"})
			continue
		}
		switch strings.ToLower(strings.TrimSpace(base.Type)) {
		case "question":
			var q clinicQuestionFrame
			if err := json.Unmarshal(msg, &q); err != nil {
				_ = writeJSON(map[string]interface{}{"type": "error", "message": "question 帧格式错误"})
				continue
			}
			if err := svc.HandleQuestion(ctx, wxID, deviceNo, q.Text, writeJSON); err != nil {
				glog.Warningf(ctx, "[胖宝WS] 处理问题失败 wxId=%d err=%v", wxID, err)
			}
		default:
			_ = writeJSON(map[string]interface{}{"type": "error", "message": "未知帧类型"})
		}
	}
}
