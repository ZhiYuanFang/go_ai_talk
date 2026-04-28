package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func newMockBaiduTTSServers(t *testing.T, audioChunks [][]byte) (*httptest.Server, *httptest.Server) {
	t.Helper()

	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "token",
			"expires_in":   3600,
		})
	}))

	upgrader := websocket.Upgrader{}
	wsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("access_token"); got != "token" {
			t.Fatalf("unexpected access_token: %s", got)
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		defer conn.Close()

		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if mt != websocket.TextMessage {
				continue
			}
			var obj map[string]interface{}
			if err := json.Unmarshal(msg, &obj); err != nil {
				t.Fatalf("bad client json: %v", err)
			}
			switch obj["type"] {
			case "system.start":
				_ = conn.WriteJSON(map[string]interface{}{"type": "system.started", "code": 0})
			case "text":
				for _, chunk := range audioChunks {
					if err := conn.WriteMessage(websocket.BinaryMessage, chunk); err != nil {
						t.Fatalf("write binary failed: %v", err)
					}
				}
			case "system.finish":
				_ = conn.WriteJSON(map[string]interface{}{"type": "system.finished", "code": 0})
				return
			}
		}
	}))

	return tokenSrv, wsSrv
}

func TestVoiceServiceHandleHappyPath(t *testing.T) {
	sttSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "hello"})
	}))
	defer sttSrv.Close()

	chatSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{"content": "hi"}},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer chatSrv.Close()

	ttsAudio := []byte{1, 2, 3, 4}
	tokenSrv, ttsSrv := newMockBaiduTTSServers(t, [][]byte{ttsAudio})
	defer tokenSrv.Close()
	defer ttsSrv.Close()

	cfg := VoiceChatConfig{}
	cfg.STT.Endpoint = sttSrv.URL
	cfg.DeepSeek.Endpoint = chatSrv.URL
	cfg.TTS.Provider = "baidu"
	cfg.TTS.StreamEnabled = true
	cfg.TTS.StreamEndpoint = "ws" + strings.TrimPrefix(ttsSrv.URL, "http")
	cfg.TTS.TokenEndpoint = tokenSrv.URL
	cfg.TTS.APIKey = "tts-key"
	cfg.TTS.APISecret = "tts-secret"
	cfg.TTS.CUID = "device"
	svc := NewVoiceService(cfg)

	pcm := []byte("test pcm data")
	meta := AudioMeta{SampleRate: 16000, Bits: 16, Channels: 1, Length: len(pcm)}
	input := base64.StdEncoding.EncodeToString(pcm)
	audio, outMeta, _, err := svc.Handle(context.Background(), "device-a", meta, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(audio) != len(ttsAudio) {
		t.Fatalf("unexpected audio length: got %d want %d", len(audio), len(ttsAudio))
	}
	if outMeta != meta {
		t.Fatalf("unexpected meta: %+v", outMeta)
	}
}

func TestVoiceServiceValidateBits(t *testing.T) {
	cfg := VoiceChatConfig{}
	svc := NewVoiceService(cfg)
	meta := AudioMeta{SampleRate: 16000, Bits: 8, Channels: 1, Length: 3}
	input := base64.StdEncoding.EncodeToString([]byte{1, 2, 3})
	_, _, _, err := svc.Handle(context.Background(), "device-a", meta, input)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if se, ok := err.(StageError); !ok || se.Stage != "validate" {
		t.Fatalf("unexpected error type: %v", err)
	}
}

func TestVoiceServiceDurationLimit(t *testing.T) {
	cfg := VoiceChatConfig{}
	cfg.Audio.MaxDurationSec = 1
	svc := NewVoiceService(cfg)
	meta := AudioMeta{SampleRate: 16000, Bits: 16, Channels: 1, Length: 40000}
	// 16000 * 1 * 2 = 32000 bytes for 1s; exceed it.
	bigRaw := bytes.Repeat([]byte{1}, 40000)
	big := base64.StdEncoding.EncodeToString(bigRaw)
	_, _, _, err := svc.Handle(context.Background(), "device-a", meta, big)
	if err == nil {
		t.Fatal("expected duration limit error")
	}
	if se, ok := err.(StageError); !ok || se.Stage != "validate" {
		t.Fatalf("unexpected error type: %v", err)
	}
}

func TestVoiceServiceBaiduProviders(t *testing.T) {
	expectedSampleRate := 16000
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "token",
			"expires_in":   3600,
		})
	}))
	defer tokenSrv.Close()

	sttSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		_ = json.Unmarshal(body, &payload)
		if payload["speech"] == "" {
			t.Fatalf("speech payload missing: %s", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"err_no": 0,
			"result": []string{"recognized"},
		})
	}))
	defer sttSrv.Close()

	chatSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]interface{}{
			"choices": []map[string]interface{}{{"message": map[string]interface{}{"content": "reply"}}},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer chatSrv.Close()

	ttsAudio := []byte{7, 8, 9}
	upgrader := websocket.Upgrader{}
	ttsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("access_token") != "token" {
			t.Fatalf("unexpected access_token: %s", r.URL.Query().Get("access_token"))
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		defer conn.Close()
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if mt != websocket.TextMessage {
				continue
			}
			var obj map[string]interface{}
			if err := json.Unmarshal(msg, &obj); err != nil {
				t.Fatalf("bad client json: %v", err)
			}
			switch obj["type"] {
			case "system.start":
				_ = conn.WriteJSON(map[string]interface{}{"type": "system.started", "code": 0})
			case "text":
				_ = conn.WriteMessage(websocket.BinaryMessage, ttsAudio)
			case "system.finish":
				_ = conn.WriteJSON(map[string]interface{}{"type": "system.finished", "code": 0})
				return
			}
		}
	}))
	defer ttsSrv.Close()

	cfg := VoiceChatConfig{}
	cfg.STT.Provider = "baidu"
	cfg.STT.Endpoint = sttSrv.URL
	cfg.STT.TokenEndpoint = tokenSrv.URL
	cfg.STT.APIKey = "stt-key"
	cfg.STT.APISecret = "stt-secret"
	cfg.STT.CUID = "device"
	cfg.TTS.Provider = "baidu"
	cfg.TTS.StreamEnabled = true
	cfg.TTS.StreamEndpoint = "ws" + strings.TrimPrefix(ttsSrv.URL, "http")
	cfg.TTS.TokenEndpoint = tokenSrv.URL
	cfg.TTS.APIKey = "tts-key"
	cfg.TTS.APISecret = "tts-secret"
	cfg.TTS.CUID = "device"
	cfg.TTS.Voice = "0"
	cfg.DeepSeek.Endpoint = chatSrv.URL

	svc := NewVoiceService(cfg)
	raw := []byte("test pcm data")
	meta := AudioMeta{SampleRate: expectedSampleRate, Bits: 16, Channels: 1, Length: len(raw)}
	input := base64.StdEncoding.EncodeToString(raw)
	audio, _, _, err := svc.Handle(context.Background(), "device-a", meta, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(audio, ttsAudio) {
		t.Fatalf("unexpected audio bytes: %v", audio)
	}
}

func TestVoiceServiceConcurrencyLimits(t *testing.T) {
	var sttMu sync.Mutex
	var sttCurrent, sttMax int
	sttSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sttMu.Lock()
		sttCurrent++
		if sttCurrent > sttMax {
			sttMax = sttCurrent
		}
		sttMu.Unlock()

		time.Sleep(50 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "ok"})

		sttMu.Lock()
		sttCurrent--
		sttMu.Unlock()
	}))
	defer sttSrv.Close()

	var chatMu sync.Mutex
	var chatCurrent, chatMax int
	chatSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chatMu.Lock()
		chatCurrent++
		if chatCurrent > chatMax {
			chatMax = chatCurrent
		}
		chatMu.Unlock()

		time.Sleep(50 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		payload := map[string]interface{}{
			"choices": []map[string]interface{}{{"message": map[string]interface{}{"content": "reply"}}},
		}
		_ = json.NewEncoder(w).Encode(payload)

		chatMu.Lock()
		chatCurrent--
		chatMu.Unlock()
	}))
	defer chatSrv.Close()

	cfg := VoiceChatConfig{}
	cfg.STT.Endpoint = sttSrv.URL
	cfg.STT.MaxConcurrency = 2
	cfg.DeepSeek.Endpoint = chatSrv.URL
	cfg.DeepSeek.Model = "deepseek-chat"
	cfg.DeepSeek.MaxConcurrency = 2

	svc := NewVoiceService(cfg)
	pcm := []byte("test pcm data")
	meta := AudioMeta{SampleRate: 16000, Bits: 16, Channels: 1, Length: len(pcm)}
	encoded := base64.StdEncoding.EncodeToString(pcm)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := svc.transcribe(context.Background(), meta, encoded); err != nil {
				t.Errorf("transcribe error: %v", err)
			}
		}()
	}
	wg.Wait()
	if sttMax > 2 {
		t.Fatalf("STT concurrency exceeded: %d", sttMax)
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if _, err := svc.chat(context.Background(), fmt.Sprintf("dev-%d", idx), "hello"); err != nil {
				t.Errorf("chat error: %v", err)
			}
		}(i)
	}
	wg.Wait()
	if chatMax > 2 {
		t.Fatalf("Chat concurrency exceeded: %d", chatMax)
	}
}

func TestVoiceServiceChatDeviceHistory(t *testing.T) {
	var historyLens []int
	var mu sync.Mutex
	chatSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Messages []map[string]string `json:"messages"`
		}
		_ = json.Unmarshal(body, &payload)

		msgLen := 0
		if len(payload.Messages) > 0 {
			msgLen = len(payload.Messages)
		}
		mu.Lock()
		historyLens = append(historyLens, msgLen)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{{"message": map[string]interface{}{"content": "ok"}}},
		})
	}))
	defer chatSrv.Close()

	cfg := VoiceChatConfig{}
	cfg.DeepSeek.Endpoint = chatSrv.URL
	cfg.DeepSeek.Model = "deepseek-chat"
	cfg.Session.MaxRounds = 10
	svc := NewVoiceService(cfg)

	if _, err := svc.chat(context.Background(), "device-1", "第一句"); err != nil {
		t.Fatalf("first chat error: %v", err)
	}
	if _, err := svc.chat(context.Background(), "device-1", "第二句"); err != nil {
		t.Fatalf("second chat error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(historyLens) != 4 {
		t.Fatalf("unexpected call count: %d", len(historyLens))
	}
	if historyLens[0] != 1 {
		t.Fatalf("first classify call should only have current user message, got %d", historyLens[0])
	}
	if historyLens[3] != 3 {
		t.Fatalf("second chat call should include previous round + current user, got %d", historyLens[3])
	}
}

func TestVoiceServiceChatDeviceIsolationAndTTL(t *testing.T) {
	var allLens []int
	var mu sync.Mutex
	chatSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Messages []map[string]string `json:"messages"`
		}
		_ = json.Unmarshal(body, &payload)
		mu.Lock()
		allLens = append(allLens, len(payload.Messages))
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{{"message": map[string]interface{}{"content": "ok"}}},
		})
	}))
	defer chatSrv.Close()

	cfg := VoiceChatConfig{}
	cfg.DeepSeek.Endpoint = chatSrv.URL
	cfg.DeepSeek.Model = "deepseek-chat"
	cfg.Session.MaxRounds = 10
	cfg.Session.TTLSeconds = 1
	svc := NewVoiceService(cfg)

	if _, err := svc.chat(context.Background(), "device-a", "A1"); err != nil {
		t.Fatalf("chat device-a error: %v", err)
	}
	if _, err := svc.chat(context.Background(), "device-b", "B1"); err != nil {
		t.Fatalf("chat device-b error: %v", err)
	}

	svc.sessionMu.Lock()
	if sess, ok := svc.sessions["device-a"]; ok {
		sess.LastActive = time.Now().Add(-2 * time.Second)
	}
	svc.sessionMu.Unlock()

	if _, err := svc.chat(context.Background(), "device-a", "A2"); err != nil {
		t.Fatalf("chat device-a second error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(allLens) != 6 {
		t.Fatalf("unexpected call count: %d", len(allLens))
	}
	if allLens[3] != 1 {
		t.Fatalf("different device should not share history, got %d", allLens[3])
	}
	if allLens[5] != 1 {
		t.Fatalf("expired session should reset history, got %d", allLens[5])
	}
}

func TestVoiceServiceChatTrimMaxRounds(t *testing.T) {
	var allLens []int
	var mu sync.Mutex
	chatSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Messages []map[string]string `json:"messages"`
		}
		_ = json.Unmarshal(body, &payload)
		mu.Lock()
		allLens = append(allLens, len(payload.Messages))
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{{"message": map[string]interface{}{"content": "ok"}}},
		})
	}))
	defer chatSrv.Close()

	cfg := VoiceChatConfig{}
	cfg.DeepSeek.Endpoint = chatSrv.URL
	cfg.DeepSeek.Model = "deepseek-chat"
	cfg.Session.MaxRounds = 1
	svc := NewVoiceService(cfg)

	if _, err := svc.chat(context.Background(), "device-z", "Q1"); err != nil {
		t.Fatalf("chat1 error: %v", err)
	}
	if _, err := svc.chat(context.Background(), "device-z", "Q2"); err != nil {
		t.Fatalf("chat2 error: %v", err)
	}
	if _, err := svc.chat(context.Background(), "device-z", "Q3"); err != nil {
		t.Fatalf("chat3 error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(allLens) != 6 {
		t.Fatalf("unexpected call count: %d", len(allLens))
	}
	if allLens[5] != 3 {
		t.Fatalf("with maxRounds=1, third call should include one previous round + current user, got %d", allLens[5])
	}
}

func TestVoiceServiceHandleRejectsShortTranscriptWithoutDeepSeekCall(t *testing.T) {
	sttSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "你"})
	}))
	defer sttSrv.Close()

	var (
		mu       sync.Mutex
		chatCall int
	)
	chatSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		chatCall++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{{"message": map[string]interface{}{"content": "ok"}}},
		})
	}))
	defer chatSrv.Close()

	cfg := VoiceChatConfig{}
	cfg.STT.Endpoint = sttSrv.URL
	cfg.DeepSeek.Endpoint = chatSrv.URL
	cfg.DeepSeek.Model = "deepseek-chat"
	cfg.DeepSeek.MinTextLength = 2
	svc := NewVoiceService(cfg)

	raw := []byte("test pcm data")
	meta := AudioMeta{SampleRate: 16000, Bits: 16, Channels: 1, Length: len(raw)}
	input := base64.StdEncoding.EncodeToString(raw)

	_, _, _, err := svc.Handle(context.Background(), "device-short", meta, input)
	if err == nil {
		t.Fatal("expected short transcript error")
	}
	se, ok := err.(StageError)
	if !ok {
		t.Fatalf("expected StageError, got %T", err)
	}
	if se.Stage != "chat" {
		t.Fatalf("unexpected stage: %s", se.Stage)
	}
	if !strings.Contains(se.Detail, "文本长度不能小于2") {
		t.Fatalf("unexpected detail: %s", se.Detail)
	}

	mu.Lock()
	defer mu.Unlock()
	if chatCall != 0 {
		t.Fatalf("deepseek should not be called, got %d", chatCall)
	}
}

func TestVoiceServiceSynthesizeBaiduChunks(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "token",
			"expires_in":   3600,
		})
	}))
	defer tokenSrv.Close()

	upgrader := websocket.Upgrader{}
	wsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		defer conn.Close()
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if mt != websocket.TextMessage {
				continue
			}
			var obj map[string]interface{}
			if err := json.Unmarshal(msg, &obj); err != nil {
				t.Fatalf("bad client json: %v", err)
			}
			switch obj["type"] {
			case "system.start":
				_ = conn.WriteJSON(map[string]interface{}{"type": "system.started", "code": 0})
			case "text":
				payload, _ := obj["payload"].(map[string]interface{})
				text := strings.TrimSpace(anyString(payload["text"]))
				if text == "" {
					t.Fatal("missing text payload")
				}
				_ = conn.WriteMessage(websocket.BinaryMessage, []byte{byte(len(text) % 255)})
			case "system.finish":
				_ = conn.WriteJSON(map[string]interface{}{"type": "system.finished", "code": 0})
				return
			}
		}
	}))
	defer wsSrv.Close()

	cfg := VoiceChatConfig{}
	cfg.TTS.Provider = "baidu"
	cfg.TTS.StreamEnabled = true
	cfg.TTS.StreamEndpoint = "ws" + strings.TrimPrefix(wsSrv.URL, "http")
	cfg.TTS.TokenEndpoint = tokenSrv.URL
	cfg.TTS.APIKey = "key"
	cfg.TTS.APISecret = "secret"
	cfg.TTS.CUID = "device"
	cfg.TTS.Voice = "0"
	cfg.TTS.TimeoutSeconds = 5
	cfg.TTS.StreamFinishTimeoutSeconds = 3

	svc := NewVoiceService(cfg)
	meta := AudioMeta{SampleRate: 16000, Bits: 16, Channels: 1}
	reply := strings.Repeat("你好世界。", 200)
	audio, err := svc.synthesizeBaidu(context.Background(), meta, reply)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(audio) != 1 {
		t.Fatalf("unexpected audio length: %d", len(audio))
	}
}

func TestVoiceServiceStreamReplyWithBaiduStreamingTTS(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "token",
			"expires_in":   3600,
		})
	}))
	defer tokenSrv.Close()

	upgrader := websocket.Upgrader{}
	wsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("access_token"); got != "token" {
			t.Fatalf("unexpected access_token: %s", got)
		}
		if got := r.URL.Query().Get("per"); got != "0" {
			t.Fatalf("unexpected per: %s", got)
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		defer conn.Close()

		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if mt != websocket.TextMessage {
				continue
			}
			var obj map[string]interface{}
			if err := json.Unmarshal(msg, &obj); err != nil {
				t.Fatalf("bad client json: %v", err)
			}
			switch obj["type"] {
			case "system.start":
				_ = conn.WriteJSON(map[string]interface{}{"type": "system.started", "code": 0})
			case "text":
				payload, _ := obj["payload"].(map[string]interface{})
				if strings.TrimSpace(anyString(payload["text"])) == "" {
					t.Fatal("missing text payload")
				}
				if err := conn.WriteMessage(websocket.BinaryMessage, []byte{1, 2, 3, 4}); err != nil {
					t.Fatalf("write binary failed: %v", err)
				}
			case "system.finish":
				_ = conn.WriteJSON(map[string]interface{}{"type": "system.finished", "code": 0})
				return
			}
		}
	}))
	defer wsSrv.Close()

	cfg := VoiceChatConfig{}
	cfg.TTS.Provider = "baidu"
	cfg.TTS.StreamEnabled = true
	cfg.TTS.StreamEndpoint = "ws" + strings.TrimPrefix(wsSrv.URL, "http")
	cfg.TTS.TokenEndpoint = tokenSrv.URL
	cfg.TTS.APIKey = "key"
	cfg.TTS.APISecret = "secret"
	cfg.TTS.CUID = "device"
	cfg.TTS.Voice = "0"
	cfg.TTS.TimeoutSeconds = 5
	cfg.TTS.StreamFinishTimeoutSeconds = 3

	svc := NewVoiceService(cfg)
	meta := AudioMeta{SampleRate: 16000, Bits: 16, Channels: 1}

	var got [][]byte
	chunks, err := svc.StreamReplyWithBaiduTTS(context.Background(), meta, "你好", func(audio []byte, meta AudioMeta, seq int) error {
		got = append(got, append([]byte(nil), audio...))
		if seq != 1 {
			t.Fatalf("unexpected seq: %d", seq)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chunks != 1 {
		t.Fatalf("unexpected chunks: %d", chunks)
	}
	if len(got) != 1 || !bytes.Equal(got[0], []byte{1, 2, 3, 4}) {
		t.Fatalf("unexpected audio chunks: %+v", got)
	}
}

func TestDecodeBase64AudioVariants(t *testing.T) {
	pcm := []byte{0, 1, 2, 3, 4, 5, 6}
	std := base64.StdEncoding.EncodeToString(pcm)
	rawStd := base64.RawStdEncoding.EncodeToString(pcm)
	urlEnc := base64.RawURLEncoding.EncodeToString(pcm)
	noisy := "  data:audio/wav;base64," + std[:4] + "\n" + std[4:] + "  "
	zwsp := string('\u200b') + std + string('\u200b')
	zwspInside := std[:3] + string('\u200c') + std[3:]
	bom := string('\ufeff') + std
	noPad := strings.TrimRight(std, "=")

	inputs := map[string]string{
		"std":        std,
		"raw":        rawStd,
		"url":        urlEnc,
		"noisy":      noisy,
		"zwsp":       zwsp,
		"zwspInside": zwspInside,
		"bom":        bom,
		"noPad":      noPad,
	}

	for name, input := range inputs {
		got, err := decodeBase64Audio(input)
		if err != nil {
			t.Fatalf("%s decode failed: %v", name, err)
		}
		if !bytes.Equal(got, pcm) {
			t.Fatalf("%s decoded mismatch", name)
		}
	}
}

func TestDecodeBase64AudioInvalid(t *testing.T) {
	if _, err := decodeBase64Audio("invalid***payload"); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestVoiceServiceHandlePersistsTalkRecord(t *testing.T) {
	sttSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "今天天气怎么样"})
	}))
	defer sttSrv.Close()

	chatSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{{"message": map[string]interface{}{"content": "今天天气不错"}}},
		})
	}))
	defer chatSrv.Close()

	tokenSrv, ttsSrv := newMockBaiduTTSServers(t, [][]byte{{1, 2, 3}})
	defer tokenSrv.Close()
	defer ttsSrv.Close()

	cfg := VoiceChatConfig{}
	cfg.STT.Endpoint = sttSrv.URL
	cfg.DeepSeek.Endpoint = chatSrv.URL
	cfg.TTS.Provider = "baidu"
	cfg.TTS.StreamEnabled = true
	cfg.TTS.StreamEndpoint = "ws" + strings.TrimPrefix(ttsSrv.URL, "http")
	cfg.TTS.TokenEndpoint = tokenSrv.URL
	cfg.TTS.APIKey = "tts-key"
	cfg.TTS.APISecret = "tts-secret"
	cfg.TTS.CUID = "device"
	svc := NewVoiceService(cfg)
	svc.ensureDeviceRegistered = func(ctx context.Context, deviceNo string) error { return nil }

	called := 0
	var gotAsk, gotAnswer string
	svc.persistTalkRecord = func(ctx context.Context, deviceNo, ask, answer string) error {
		called++
		gotAsk = ask
		gotAnswer = answer
		return nil
	}

	raw := []byte("test pcm data")
	meta := AudioMeta{SampleRate: 16000, Bits: 16, Channels: 1, Length: len(raw)}
	input := base64.StdEncoding.EncodeToString(raw)
	if _, _, _, err := svc.Handle(context.Background(), "device-a", meta, input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called != 1 {
		t.Fatalf("persistTalkRecord should be called once, got %d", called)
	}
	if gotAsk == "" || gotAnswer == "" {
		t.Fatalf("ask/answer should be captured, got ask=%q answer=%q", gotAsk, gotAnswer)
	}
}

func TestVoiceServiceGetLastUserMessageFromSessionMemory(t *testing.T) {
	cfg := VoiceChatConfig{}
	cfg.Session.MaxRounds = 10
	svc := NewVoiceService(cfg)

	svc.appendChatHistory("device-1", "你好", "你好呀")
	got := svc.getLastUserMessageFromSession(context.Background(), "device-1", time.Now())
	if got == "" {
		t.Fatal("expected non-empty last user message")
	}
}

func TestVoiceServiceUseRedisSessionFlag(t *testing.T) {
	old := os.Getenv("VOICE_SESSION_BACKEND")
	defer func() { _ = os.Setenv("VOICE_SESSION_BACKEND", old) }()

	cfg := VoiceChatConfig{}
	svc := NewVoiceService(cfg)

	_ = os.Setenv("VOICE_SESSION_BACKEND", "memory")
	if svc.useRedisSession() {
		t.Fatal("memory backend should not enable redis session")
	}

	_ = os.Setenv("VOICE_SESSION_BACKEND", "redis")
	if !svc.useRedisSession() {
		t.Fatal("redis backend should enable redis session")
	}
}

func TestVoiceServiceUseRedisGuardsFlag(t *testing.T) {
	old := os.Getenv("VOICE_GUARD_REDIS_ENABLED")
	defer func() { _ = os.Setenv("VOICE_GUARD_REDIS_ENABLED", old) }()

	cfg := VoiceChatConfig{}
	svc := NewVoiceService(cfg)

	_ = os.Setenv("VOICE_GUARD_REDIS_ENABLED", "false")
	if svc.useRedisGuards() {
		t.Fatal("false should disable redis guards")
	}

	_ = os.Setenv("VOICE_GUARD_REDIS_ENABLED", "true")
	if !svc.useRedisGuards() {
		t.Fatal("true should enable redis guards")
	}
}

func TestExtractChatReplyDetectsExitMarker(t *testing.T) {
	body := []byte(`{"choices":[{"message":{"content":"{\"exit\":true}"}}]}`)
	reply, exit, err := extractChatReply(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exit {
		t.Fatal("expected exit=true")
	}
	if reply != "" {
		t.Fatalf("expected empty reply, got %q", reply)
	}
}
