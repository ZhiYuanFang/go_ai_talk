package voice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/glog"
	"github.com/gorilla/websocket"
)

type baiduStreamTTSSession struct {
	svc          *VoiceService
	meta         AudioMeta
	conn         *websocket.Conn
	onAudioChunk func(audio []byte, meta AudioMeta) error

	mu           sync.Mutex
	closed       bool
	started      bool
	finished     bool
	audioFrames  int
	audioBytes   int
	createdAt    time.Time
	firstAudioAt time.Time

	startAckCh   chan struct{}
	finishAckCh  chan struct{}
	errCh        chan error
}

func newBaiduStreamTTSSession(ctx context.Context, svc *VoiceService, meta AudioMeta, onAudioChunk func(audio []byte, meta AudioMeta) error) (StreamTTSSession, error) {
	if svc == nil {
		return nil, StageError{Stage: "tts", Detail: "voice service 为空"}
	}
	token, err := svc.getBaiduAccessToken(ctx, &svc.ttsToken, svc.cfg.TTS.APIKey, svc.cfg.TTS.APISecret, svc.cfg.TTS.TokenEndpoint, svc.cfg.TTS.TimeoutSeconds)
	if err != nil {
		return nil, StageError{Stage: "tts", Detail: err.Error()}
	}
	dialURL, err := buildBaiduStreamingTTSURL(svc.cfg.TTS.StreamEndpoint, token, svc.cfg.TTS.Voice)
	if err != nil {
		return nil, StageError{Stage: "tts", Detail: err.Error()}
	}
	glog.Infof(ctx, "[TTS请求] 建立百度流式TTS连接。url=%q voice=%q aue=%q", dialURL, strings.TrimSpace(svc.cfg.TTS.Voice), strings.TrimSpace(svc.cfg.TTS.AUE))

	dialer := websocket.Dialer{HandshakeTimeout: 8 * time.Second}
	conn, resp, err := dialer.DialContext(ctx, dialURL, nil)
	if err != nil {
		statusCode := 0
		statusText := ""
		bodySnippet := ""
		if resp != nil {
			statusCode = resp.StatusCode
			statusText = strings.TrimSpace(resp.Status)
			if resp.Body != nil {
				defer resp.Body.Close()
				if body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024)); readErr == nil {
					bodySnippet = strings.TrimSpace(string(body))
				}
			}
		}
		return nil, StageError{Stage: "tts", Detail: fmt.Sprintf("连接百度流式TTS失败: %v; status=%d; statusText=%s; body=%s", err, statusCode, statusText, bodySnippet)}
	}

	sess := &baiduStreamTTSSession{
		svc:          svc,
		meta:         fillAudioMetaDefaults(meta, svc.cfg),
		conn:         conn,
		onAudioChunk: onAudioChunk,
		createdAt:    time.Now(),
		startAckCh:   make(chan struct{}, 1),
		finishAckCh:  make(chan struct{}, 1),
		errCh:        make(chan error, 1),
	}
	go sess.readLoop()
	if err := sess.sendStart(ctx); err != nil {
		_ = sess.Close()
		return nil, err
	}
	return sess, nil
}

func fillAudioMetaDefaults(meta AudioMeta, cfg VoiceChatConfig) AudioMeta {
	if meta.SampleRate <= 0 {
		meta.SampleRate = cfg.Audio.SampleRate
	}
	if meta.Bits <= 0 {
		meta.Bits = cfg.Audio.Bits
	}
	if meta.Channels <= 0 {
		meta.Channels = cfg.Audio.Channels
	}
	return meta
}

func buildBaiduStreamingTTSURL(rawURL, token, voice string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", err
	}
	if u.Scheme == "https" {
		u.Scheme = "wss"
	}
	if u.Scheme == "http" {
		u.Scheme = "ws"
	}
	q := u.Query()
	if token != "" {
		q.Set("access_token", token)
	}
	if strings.TrimSpace(voice) != "" {
		q.Set("per", strings.TrimSpace(voice))
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (s *baiduStreamTTSSession) sendStart(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return StageError{Stage: "tts", Detail: "流式TTS会话已关闭"}
	}
	s.mu.Unlock()
	payload := map[string]interface{}{
		"type": "system.start",
		"payload": map[string]interface{}{
			"cuid": s.svc.cfg.TTS.CUID,
			"aue":  stringAtoiDefault(s.svc.cfg.TTS.AUE, 6),
			"spd":  stringAtoiDefault(s.svc.cfg.TTS.Speed, 5),
			"pit":  stringAtoiDefault(s.svc.cfg.TTS.Pitch, 5),
			"vol":  stringAtoiDefault(s.svc.cfg.TTS.Volume, 5),
			"lan":  strings.TrimSpace(s.svc.cfg.TTS.Language),
		},
	}
	if err := s.conn.WriteJSON(payload); err != nil {
		return StageError{Stage: "tts", Detail: "发送流式TTS system.start失败: " + err.Error()}
	}
	timeout := time.Duration(s.svc.cfg.TTS.StreamFinishTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	select {
	case <-s.startAckCh:
		glog.Infof(ctx, "[TTS响应] 百度流式TTS已启动。sampleRate=%d", s.meta.SampleRate)
		return nil
	case err := <-s.errCh:
		if err == nil {
			err = errors.New("百度流式TTS启动失败")
		}
		return StageError{Stage: "tts", Detail: err.Error()}
	case <-time.After(timeout):
		return StageError{Stage: "tts", Detail: "等待百度流式TTS启动超时"}
	}
}

func (s *baiduStreamTTSSession) WriteText(text string) error {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return StageError{Stage: "tts", Detail: "流式TTS会话已关闭"}
	}
	if s.finished {
		return StageError{Stage: "tts", Detail: "流式TTS会话已完成"}
	}
	glog.Infof(context.Background(), "[TTS请求] 发送流式文本。textLen=%d text=%q", len([]rune(trimmed)), truncateVoiceLogText(trimmed, 120))
	return s.conn.WriteJSON(map[string]interface{}{
		"type": "text",
		"payload": map[string]interface{}{
			"text": trimmed,
		},
	})
}

func (s *baiduStreamTTSSession) Finish(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	if s.finished {
		s.mu.Unlock()
		return nil
	}
	s.finished = true
	s.mu.Unlock()
	if err := s.conn.WriteJSON(map[string]interface{}{"type": "system.finish"}); err != nil {
		return StageError{Stage: "tts", Detail: "发送流式TTS system.finish失败: " + err.Error()}
	}
	timeout := time.Duration(s.svc.cfg.TTS.StreamFinishTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	select {
	case <-s.finishAckCh:
		glog.Infof(ctx, "[TTS响应] 百度流式TTS完成。frames=%d audioBytes=%d firstAudioDelayMs=%d",
			s.audioFrames, s.audioBytes, firstAudioDelayMs(s.createdAt, s.firstAudioAt))
		return nil
	case err := <-s.errCh:
		if err == nil {
			err = errors.New("百度流式TTS结束失败")
		}
		return StageError{Stage: "tts", Detail: err.Error()}
	case <-time.After(timeout):
		return StageError{Stage: "tts", Detail: "等待百度流式TTS完成超时"}
	}
}

func firstAudioDelayMs(start, first time.Time) int64 {
	if start.IsZero() || first.IsZero() {
		return 0
	}
	return first.Sub(start).Milliseconds()
}

func (s *baiduStreamTTSSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	_ = s.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(300*time.Millisecond))
	return s.conn.Close()
}

func (s *baiduStreamTTSSession) readLoop() {
	defer func() {
		_ = s.Close()
	}()
	for {
		mt, msg, err := s.conn.ReadMessage()
		if err != nil {
			if !isExpectedWSStreamClose(err) {
				s.pushErr(err)
			}
			return
		}
		switch mt {
		case websocket.BinaryMessage:
			s.handleAudioChunk(msg)
		case websocket.TextMessage:
			s.handleControlMessage(msg)
		}
	}
}

func (s *baiduStreamTTSSession) handleAudioChunk(msg []byte) {
	if len(msg) == 0 {
		return
	}
	if s.firstAudioAt.IsZero() {
		s.firstAudioAt = time.Now()
	}
	s.audioFrames++
	s.audioBytes += len(msg)
	if s.onAudioChunk != nil {
		if err := s.onAudioChunk(msg, s.meta); err != nil {
			s.pushErr(err)
		}
	}
}

func (s *baiduStreamTTSSession) handleControlMessage(msg []byte) {
	var obj map[string]interface{}
	if err := json.Unmarshal(msg, &obj); err != nil {
		s.pushErr(fmt.Errorf("解析百度流式TTS响应失败: %w", err))
		return
	}
	msgType := strings.TrimSpace(anyString(obj["type"]))
	code := anyInt(obj["code"])
	message := strings.TrimSpace(anyString(obj["message"]))
	if code != 0 {
		s.pushErr(fmt.Errorf("baidu stream tts code=%d message=%s", code, message))
		return
	}
	switch msgType {
	case "system.started":
		select {
		case s.startAckCh <- struct{}{}:
		default:
		}
	case "system.finished", "system.end", "system.completed":
		select {
		case s.finishAckCh <- struct{}{}:
		default:
		}
	case "system.error":
		s.pushErr(fmt.Errorf("baidu stream tts error: %s", message))
	}
}

func (s *baiduStreamTTSSession) pushErr(err error) {
	if err == nil {
		return
	}
	select {
	case s.errCh <- err:
	default:
	}
}

func stringAtoiDefault(value string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
