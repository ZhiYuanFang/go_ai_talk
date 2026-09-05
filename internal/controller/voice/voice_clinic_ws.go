package voicectrl

import (
	"context"
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
	Type   string `json:"type"`
	Text   string `json:"text"`
	TurnID string `json:"turnId"`
}

type clinicCancelFrame struct {
	Type   string `json:"type"`
	TurnID string `json:"turnId"`
}

// clinicConnState 连接级 active turn 状态；读循环非阻塞，LLM 在 goroutine 中处理。
type clinicConnState struct {
	activeTurnID string
	cancelTurn   context.CancelFunc
	turnMu       sync.Mutex
}

func RegisterVoiceClinicWS(s *ghttp.Server) {
	// 胖宝 WS 不注册 VoiceWSManager，与 ASR WS 同策略。
	s.BindHandler("/voice/clinic/ws", voiceClinicWS)
}

func voiceClinicWS(r *ghttp.Request) {
	connCtx := r.Context()
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

	connState := &clinicConnState{}
	// WS 关闭时取消 active LLM ctx；连接已断 MUST NOT 下发 turn_cancelled（无 reason:disconnected）。
	defer func() {
		connState.turnMu.Lock()
		if connState.cancelTurn != nil {
			connState.cancelTurn()
			connState.cancelTurn = nil
		}
		connState.turnMu.Unlock()
	}()

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
	parsedWxID, deviceNoFromJWT, pErr := gatewayapp.ParseAccessClaims(connCtx, token)
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
	// 业务说明：不再下发 session_sync。UI 对话历史由 Flutter 本地存储；
	// agent 多轮上下文由 Python companion_session 负责；Go 仅鉴权/额度/限流/流转发。
	// auth_ok 后即可进入读循环接收 question/cancel。

	// 非阻塞读循环：question 在 goroutine 中处理，cancel/supersede 可即时打断 LLM。
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
			turnID := strings.TrimSpace(q.TurnID)
			if turnID == "" {
				_ = writeJSON(map[string]interface{}{"type": "error", "code": 400, "message": "缺少 turnId"})
				continue
			}
			text := strings.TrimSpace(q.Text)
			if text == "" {
				_ = writeJSON(map[string]interface{}{"type": "error", "code": 400, "message": "问题不能为空"})
				continue
			}
			// 新 question supersede 旧 active turn：先 cancel 旧 ctx 并下发 turn_cancelled。
			connState.turnMu.Lock()
			if connState.activeTurnID != "" && connState.cancelTurn != nil {
				oldTurnID := connState.activeTurnID
				connState.cancelTurn()
				connState.cancelTurn = nil
				connState.activeTurnID = ""
				connState.turnMu.Unlock()
				_ = voice.EmitTurnCancelled(writeJSON, oldTurnID, "superseded")
			} else {
				connState.turnMu.Unlock()
			}
			turnCtx, cancelTurn := context.WithCancel(connCtx)
			connState.turnMu.Lock()
			connState.activeTurnID = turnID
			connState.cancelTurn = cancelTurn
			connState.turnMu.Unlock()
			go func(tid, qText string, tCtx context.Context) {
				defer func() {
					connState.turnMu.Lock()
					if connState.activeTurnID == tid {
						connState.activeTurnID = ""
						connState.cancelTurn = nil
					}
					connState.turnMu.Unlock()
				}()
				if err := svc.HandleQuestion(tCtx, wxID, deviceNo, qText, tid, writeJSON); err != nil {
					glog.Warningf(connCtx, "[胖宝WS] 处理问题失败 wxId=%d turnId=%s err=%v", wxID, tid, err)
				}
			}(turnID, text, turnCtx)
		case "cancel":
			var c clinicCancelFrame
			if err := json.Unmarshal(msg, &c); err != nil {
				_ = writeJSON(map[string]interface{}{"type": "error", "message": "cancel 帧格式错误"})
				continue
			}
			cancelTurnID := strings.TrimSpace(c.TurnID)
			if cancelTurnID == "" {
				_ = writeJSON(map[string]interface{}{"type": "error", "code": 400, "message": "缺少 turnId"})
				continue
			}
			connState.turnMu.Lock()
			if connState.activeTurnID == cancelTurnID && connState.cancelTurn != nil {
				activeID := connState.activeTurnID
				connState.cancelTurn()
				connState.cancelTurn = nil
				connState.activeTurnID = ""
				connState.turnMu.Unlock()
				_ = voice.EmitTurnCancelled(writeJSON, activeID, "cancelled")
			} else {
				// 已结束或不匹配的 turn：静默忽略
				connState.turnMu.Unlock()
			}
		default:
			_ = writeJSON(map[string]interface{}{"type": "error", "message": "未知帧类型"})
		}
	}
}
