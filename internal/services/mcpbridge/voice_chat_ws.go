package mcpbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// resolveVoiceChatWSURL 解析喂养 chat WS 地址。
// 优先 VOICE_CHAT_WS_URL；否则由 VOICE_SERVICE_URL（http/https）推导 ws/wss + /voice/chat/ws。
func resolveVoiceChatWSURL() (string, error) {
	if u := strings.TrimSpace(os.Getenv("VOICE_CHAT_WS_URL")); u != "" {
		return u, nil
	}
	base := strings.TrimSpace(os.Getenv("VOICE_SERVICE_URL"))
	if base == "" {
		return "", fmt.Errorf("未配置 VOICE_CHAT_WS_URL 或 VOICE_SERVICE_URL")
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("VOICE_SERVICE_URL 非法: %w", err)
	}
	switch strings.ToLower(u.Scheme) {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
		// 已是 WS
	default:
		return "", fmt.Errorf("VOICE_SERVICE_URL scheme 不支持: %s", u.Scheme)
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/voice/chat/ws"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

// ChatViaVoiceWS 经 /voice/chat/ws 文入文出完成一轮喂养对话，返回 answer 文本。
// start 仍携带音频占位字段（与协议约定一致，不放宽校验）。
func ChatViaVoiceWS(ctx context.Context, deviceNo, transcript string) (string, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	transcript = strings.TrimSpace(transcript)
	if deviceNo == "" {
		return "", fmt.Errorf("deviceNo 不能为空")
	}
	if transcript == "" {
		return "", fmt.Errorf("transcript 不能为空")
	}
	wsURL, err := resolveVoiceChatWSURL()
	if err != nil {
		return "", err
	}
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, _, err := dialer.DialContext(ctx, wsURL, http.Header{})
	if err != nil {
		return "", fmt.Errorf("连接 voice chat WS 失败: %w", err)
	}
	defer conn.Close()

	deadline := time.Now().Add(90 * time.Second)
	if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
		deadline = dl
	}
	_ = conn.SetReadDeadline(deadline)
	_ = conn.SetWriteDeadline(deadline)

	// 文模式：音频元数据用横屏同款占位，满足服务端必填校验。
	startPayload, _ := json.Marshal(map[string]interface{}{
		"type":           "start",
		"deviceNo":       deviceNo,
		"sampleRate":     16000,
		"bits":           16,
		"channels":       1,
		"length":         32000,
		"mode":           "stream",
		"inputModality":  "text",
		"outputModality": "text",
	})
	if err := conn.WriteMessage(websocket.TextMessage, startPayload); err != nil {
		return "", fmt.Errorf("发送 start 失败: %w", err)
	}
	if err := waitWSType(conn, "started"); err != nil {
		return "", err
	}
	textPayload, _ := json.Marshal(map[string]interface{}{
		"type": "text",
		"text": transcript,
	})
	if err := conn.WriteMessage(websocket.TextMessage, textPayload); err != nil {
		return "", fmt.Errorf("发送 text 失败: %w", err)
	}
	answer, err := waitWSAnswer(conn)
	if err != nil {
		return "", err
	}
	endPayload, _ := json.Marshal(map[string]interface{}{"type": "end"})
	_ = conn.WriteMessage(websocket.TextMessage, endPayload)
	return answer, nil
}

func waitWSType(conn *websocket.Conn, want string) error {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("等待 %s 失败: %w", want, err)
		}
		var frame map[string]interface{}
		if err := json.Unmarshal(msg, &frame); err != nil {
			continue
		}
		typ, _ := frame["type"].(string)
		typ = strings.ToLower(strings.TrimSpace(typ))
		if typ == "error" {
			return fmt.Errorf("WS error: %v", frame["message"])
		}
		if typ == want {
			return nil
		}
	}
}

func waitWSAnswer(conn *websocket.Conn) (string, error) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return "", fmt.Errorf("等待 answer 失败: %w", err)
		}
		var frame map[string]interface{}
		if err := json.Unmarshal(msg, &frame); err != nil {
			continue
		}
		typ, _ := frame["type"].(string)
		typ = strings.ToLower(strings.TrimSpace(typ))
		switch typ {
		case "thinking_delta":
			continue
		case "answer":
			text, _ := frame["text"].(string)
			text = strings.TrimSpace(text)
			if text == "" {
				return "", fmt.Errorf("answer 为空")
			}
			return text, nil
		case "error":
			return "", fmt.Errorf("WS error: %v", frame["message"])
		case "exit":
			return "", fmt.Errorf("会话退出且无 answer")
		}
	}
}
