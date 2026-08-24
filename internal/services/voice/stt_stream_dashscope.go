// stt_stream_dashscope.go：百炼 DashScope 流式 ASR，供 /voice/chat/ws 远场对话使用。
package voice

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/glog"
	"github.com/gorilla/websocket"
)

var errDashScopeTaskNotReady = errors.New("dashscope task not started")

// dashscopeStreamASRSession 百炼实时 ASR WebSocket 会话（run-task / 二进制 PCM / finish-task）。
type dashscopeStreamASRSession struct {
	sttCfg    STTProfileConfig
	meta      AudioMeta
	conn      *websocket.Conn
	taskID    string
	onPartial func(string)
	onFinal   func(string)

	mu                sync.Mutex
	closed            bool
	taskReady         bool
	finishSent        bool
	latest            string
	final             string
	preBuf            []byte
	initOK            bool
	strippedWavHeader bool
	audioChunkCount   int
	audioSentBytes    int
	audioNonZeroBytes int
	pcmSampleCount    int
	pcmAbsSum         int64
	pcmPeakAbs        int
	taskReadyCh       chan struct{}
	finalCh           chan string
	errCh             chan error
}

func newDashScopeStreamASRSession(ctx context.Context, sttCfg STTProfileConfig, meta AudioMeta, onPartial func(text string), onFinal func(text string)) (StreamASRSession, error) {
	apiKey := resolveDashScopeAPIKey(sttCfg)
	if apiKey == "" {
		glog.Warningf(ctx, "[流式ASR][百炼] 初始化失败：API Key 未配置")
		return nil, StageError{Stage: "stt", Detail: "DashScope API Key 未配置（sttChat.apiKey / VOICE_DASHSCOPE_API_KEY / UCG_DASHSCOPE_API_KEY）"}
	}
	dialURL, err := buildDashScopeStreamWSURL(sttCfg)
	if err != nil {
		return nil, err
	}
	if meta.SampleRate <= 0 {
		meta.SampleRate = 16000
	}
	if meta.Bits <= 0 {
		meta.Bits = 16
	}
	if meta.Channels <= 0 {
		meta.Channels = 1
	}
	if meta.SampleRate <= 0 || meta.Bits != 16 || meta.Channels <= 0 {
		return nil, StageError{Stage: "stt", Detail: "流式ASR音频参数无效"}
	}

	taskID := newDashScopeTaskID()
	glog.Infof(ctx, "[流式ASR][百炼] 正在建立WS连接。model=%q url=%q taskID=%q sampleRate=%d speechNoiseThreshold=%.2f",
		strings.TrimSpace(sttCfg.Model), dialURL, taskID, meta.SampleRate, sttCfg.SpeechNoiseThreshold)

	header := http.Header{}
	header.Set("Authorization", "Bearer "+apiKey)
	dialer := websocket.Dialer{HandshakeTimeout: 8 * time.Second}
	conn, resp, err := dialer.DialContext(ctx, dialURL, header)
	if err != nil {
		detail := formatWSDialError("连接百炼流式ASR失败", dialURL, resp, err)
		glog.Warningf(ctx, "[流式ASR][百炼] WS握手失败。url=%q err=%v", dialURL, err)
		return nil, StageError{Stage: "stt", Detail: detail}
	}

	sess := &dashscopeStreamASRSession{
		sttCfg:      sttCfg,
		meta:        meta,
		conn:        conn,
		taskID:      taskID,
		onPartial:   onPartial,
		onFinal:     onFinal,
		taskReadyCh: make(chan struct{}, 1),
		finalCh:     make(chan string, 1),
		errCh:       make(chan error, 1),
	}
	go sess.readLoop()

	if err := sess.sendRunTask(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	waitCtx, cancel := context.WithTimeout(ctx, time.Duration(sttCfg.TimeoutSeconds)*time.Second)
	defer cancel()
	select {
	case <-sess.taskReadyCh:
		glog.Infof(ctx, "[流式ASR][百炼] 任务已启动。taskID=%q", taskID)
	case err := <-sess.errCh:
		_ = conn.Close()
		if err != nil {
			return nil, StageError{Stage: "stt", Detail: err.Error()}
		}
		return nil, StageError{Stage: "stt", Detail: "百炼 ASR 任务启动失败"}
	case <-waitCtx.Done():
		_ = conn.Close()
		return nil, StageError{Stage: "stt", Detail: "等待百炼 task-started 超时"}
	}
	return sess, nil
}

func newDashScopeTaskID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func (s *dashscopeStreamASRSession) sendRunTask() error {
	params := map[string]interface{}{
		"format":         s.sttCfg.Format,
		"sample_rate":    s.meta.SampleRate,
		"language_hints": []string{"zh"},
	}
	if s.sttCfg.SpeechNoiseThreshold != 0 {
		params["speech_noise_threshold"] = s.sttCfg.SpeechNoiseThreshold
	}
	payload := map[string]interface{}{
		"header": map[string]interface{}{
			"action":    "run-task",
			"task_id":   s.taskID,
			"streaming": "duplex",
		},
		"payload": map[string]interface{}{
			"task_group": "audio",
			"task":       "asr",
			"function":   "recognition",
			"model":      strings.TrimSpace(s.sttCfg.Model),
			"parameters": params,
			"input":      map[string]interface{}{},
		},
	}
	return s.conn.WriteJSON(payload)
}

func (s *dashscopeStreamASRSession) WriteAudio(chunk []byte) error {
	if len(chunk) == 0 {
		return nil
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return StageError{Stage: "stt", Detail: "流式ASR会话已关闭"}
	}
	if !s.taskReady {
		s.mu.Unlock()
		return StageError{Stage: "stt", Detail: errDashScopeTaskNotReady.Error()}
	}
	toSend, needMore := s.normalizeFirstAudioChunkLocked(chunk)
	if needMore || len(toSend) == 0 {
		s.mu.Unlock()
		return nil
	}
	err := s.conn.WriteMessage(websocket.BinaryMessage, toSend)
	if err != nil {
		s.mu.Unlock()
		return StageError{Stage: "stt", Detail: "流式ASR写音频失败: " + err.Error()}
	}
	s.audioChunkCount++
	s.audioSentBytes += len(toSend)
	s.audioNonZeroBytes += countNonZeroBytes(toSend)
	s.updatePCMStatsLocked(toSend)
	s.mu.Unlock()
	return nil
}

func (s *dashscopeStreamASRSession) Commit(ctx context.Context) (string, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return "", StageError{Stage: "stt", Detail: "流式ASR会话已关闭"}
	}
	if len(s.preBuf) > 0 {
		if err := s.conn.WriteMessage(websocket.BinaryMessage, s.preBuf); err != nil {
			s.mu.Unlock()
			return "", StageError{Stage: "stt", Detail: "流式ASR写音频失败: " + err.Error()}
		}
		s.audioChunkCount++
		s.audioSentBytes += len(s.preBuf)
		s.audioNonZeroBytes += countNonZeroBytes(s.preBuf)
		s.updatePCMStatsLocked(s.preBuf)
		s.preBuf = nil
		s.initOK = true
	}
	avgAbs := pcmAvgAbs(s.pcmAbsSum, s.pcmSampleCount)
	glog.Infof(ctx, "[流式ASR][百炼] commit前音频统计。taskID=%q chunks=%d sentBytes=%d avgAbs=%d peakAbs=%d",
		s.taskID, s.audioChunkCount, s.audioSentBytes, avgAbs, s.pcmPeakAbs)
	if avgAbs > 0 && avgAbs < 120 {
		glog.Warningf(ctx, "[流式ASR][百炼] 音频平均幅度偏低，远场识别可能受影响。avgAbs=%d peakAbs=%d", avgAbs, s.pcmPeakAbs)
	}
	latest := strings.TrimSpace(s.latest)
	if !s.finishSent {
		s.finishSent = true
		finishPayload := map[string]interface{}{
			"header": map[string]interface{}{
				"action":    "finish-task",
				"task_id":   s.taskID,
				"streaming": "duplex",
			},
			"payload": map[string]interface{}{
				"input": map[string]interface{}{},
			},
		}
		if err := s.conn.WriteJSON(finishPayload); err != nil {
			s.mu.Unlock()
			return "", StageError{Stage: "stt", Detail: "流式ASR发送finish-task失败: " + err.Error()}
		}
	}
	s.mu.Unlock()

	for {
		select {
		case text := <-s.finalCh:
			text = strings.TrimSpace(text)
			if text != "" {
				return text, nil
			}
		case err := <-s.errCh:
			if err != nil {
				if errors.Is(err, errStreamClosedAfterFinish) {
					if latest != "" {
						return latest, nil
					}
					return "", nil
				}
				if latest != "" {
					return latest, nil
				}
				return "", StageError{Stage: "stt", Detail: err.Error()}
			}
		case <-ctx.Done():
			if latest != "" {
				return latest, nil
			}
			return "", StageError{Stage: "stt", Detail: "流式ASR等待最终结果超时"}
		}
	}
}

func (s *dashscopeStreamASRSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	_ = s.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(300*time.Millisecond))
	return s.conn.Close()
}

func (s *dashscopeStreamASRSession) readLoop() {
	defer func() {
		_ = s.Close()
	}()
	for {
		mt, msg, err := s.conn.ReadMessage()
		if err != nil {
			if isExpectedWSStreamClose(err) {
				s.pushErr(errStreamClosedAfterFinish)
			} else {
				s.pushErr(err)
			}
			return
		}
		if mt == websocket.BinaryMessage {
			continue
		}
		if mt != websocket.TextMessage {
			continue
		}
		event, text, isFinal, parseErr := parseDashScopeStreamMessage(msg)
		if parseErr != nil {
			glog.Warningf(context.Background(), "[流式ASR][百炼] 解析响应失败: %v raw=%s", parseErr, truncateForLog(string(msg), 256))
			s.pushErr(parseErr)
			continue
		}
		switch event {
		case "task-started":
			s.mu.Lock()
			s.taskReady = true
			s.mu.Unlock()
			select {
			case s.taskReadyCh <- struct{}{}:
			default:
			}
		case "task-failed":
			s.pushErr(errors.New("dashscope task failed: " + truncateForLog(string(msg), 512)))
			return
		case "task-finished":
			s.mu.Lock()
			finalText := strings.TrimSpace(s.final)
			if finalText == "" {
				finalText = strings.TrimSpace(s.latest)
			}
			s.mu.Unlock()
			if finalText != "" {
				select {
				case s.finalCh <- finalText:
				default:
				}
			}
			s.pushErr(errStreamClosedAfterFinish)
			return
		case "result-generated":
			text = strings.TrimSpace(text)
			if text == "" {
				continue
			}
			s.mu.Lock()
			s.latest = text
			if isFinal {
				s.final = text
			}
			s.mu.Unlock()
			if isFinal {
				if s.onFinal != nil {
					s.onFinal(text)
				}
				select {
				case s.finalCh <- text:
				default:
				}
				continue
			}
			if s.onPartial != nil {
				s.onPartial(text)
			}
		}
	}
}

func parseDashScopeStreamMessage(raw []byte) (event string, text string, isFinal bool, err error) {
	var root map[string]interface{}
	if err = json.Unmarshal(raw, &root); err != nil {
		return "", "", false, err
	}
	header, _ := root["header"].(map[string]interface{})
	event = strings.TrimSpace(anyString(header["event"]))
	if event == "task-failed" {
		return event, "", false, fmt.Errorf("%s: %s", anyString(header["error_code"]), anyString(header["error_message"]))
	}
	payload, _ := root["payload"].(map[string]interface{})
	output, _ := payload["output"].(map[string]interface{})
	sentence, _ := output["sentence"].(map[string]interface{})
	text = strings.TrimSpace(anyString(sentence["text"]))
	if anyBool(sentence["heartbeat"]) {
		return event, "", false, nil
	}
	isFinal = anyBool(sentence["sentence_end"])
	return event, text, isFinal, nil
}

func (s *dashscopeStreamASRSession) pushErr(err error) {
	if err == nil {
		return
	}
	select {
	case s.errCh <- err:
	default:
	}
}

func (s *dashscopeStreamASRSession) normalizeFirstAudioChunkLocked(chunk []byte) ([]byte, bool) {
	if s.initOK {
		return chunk, false
	}
	s.preBuf = append(s.preBuf, chunk...)
	if len(s.preBuf) < 12 {
		return nil, true
	}
	isWave := string(s.preBuf[0:4]) == "RIFF" && string(s.preBuf[8:12]) == "WAVE"
	if isWave {
		pcm, stripped := stripWavDataChunkIfPresent(s.preBuf)
		if stripped {
			s.initOK = true
			s.strippedWavHeader = true
			s.preBuf = nil
			glog.Warningf(context.Background(), "[流式ASR][百炼] 检测到WAV封装，已自动剥离头部。pcmBytes=%d", len(pcm))
			return pcm, false
		}
		if len(s.preBuf) < 4096 {
			return nil, true
		}
	}
	out := s.preBuf
	s.preBuf = nil
	s.initOK = true
	return out, false
}

func (s *dashscopeStreamASRSession) updatePCMStatsLocked(data []byte) {
	if len(data) < 2 {
		return
	}
	limit := len(data) - (len(data) % 2)
	for i := 0; i < limit; i += 2 {
		v := int(int16(uint16(data[i]) | uint16(data[i+1])<<8))
		if v < 0 {
			v = -v
		}
		s.pcmSampleCount++
		s.pcmAbsSum += int64(v)
		if v > s.pcmPeakAbs {
			s.pcmPeakAbs = v
		}
	}
}

func formatWSDialError(prefix, dialURL string, resp *http.Response, err error) string {
	detail := prefix + ": " + err.Error()
	if resp != nil {
		detail += fmt.Sprintf("; status=%d", resp.StatusCode)
		if resp.Body != nil {
			defer resp.Body.Close()
			if body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024)); readErr == nil {
				if snippet := strings.TrimSpace(string(body)); snippet != "" {
					detail += "; body=" + snippet
				}
			}
		}
	}
	detail += "; url=" + dialURL
	return detail
}
