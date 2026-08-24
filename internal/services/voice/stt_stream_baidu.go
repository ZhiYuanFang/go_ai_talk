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

var errStreamClosedAfterFinish = errors.New("stream closed after finish")

type baiduStreamASRSession struct {
	svc       *VoiceService
	sttCfg    STTProfileConfig
	meta      AudioMeta
	sn        string
	conn      *websocket.Conn
	onPartial func(string)
	onFinal   func(string)

	mu                sync.Mutex
	closed            bool
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
	finalCh           chan string
	errCh             chan error
}

func newBaiduStreamASRSession(ctx context.Context, svc *VoiceService, sttCfg STTProfileConfig, meta AudioMeta, onPartial func(text string), onFinal func(text string)) (StreamASRSession, error) {
	if svc == nil {
		glog.Warningf(ctx, "[流式ASR] 初始化失败：voice service 为空")
		return nil, StageError{Stage: "stt", Detail: "voice service 为空"}
	}
	originalMeta := meta
	if meta.SampleRate <= 0 {
		meta.SampleRate = svc.cfg.Audio.SampleRate
	}
	if meta.Bits <= 0 {
		meta.Bits = svc.cfg.Audio.Bits
	}
	if meta.Channels <= 0 {
		meta.Channels = svc.cfg.Audio.Channels
	}
	glog.Infof(ctx, "[流式ASR] 初始化参数。sampleRate=%d bits=%d channels=%d (原始: sampleRate=%d bits=%d channels=%d) streamEndpoint=%q tokenEndpoint=%q cuid=%q devPid=%d format=%q timeoutSec=%d apiKeySet=%v apiSecretSet=%v",
		meta.SampleRate,
		meta.Bits,
		meta.Channels,
		originalMeta.SampleRate,
		originalMeta.Bits,
		originalMeta.Channels,
		strings.TrimSpace(sttCfg.StreamEndpoint),
		strings.TrimSpace(sttCfg.TokenEndpoint),
		strings.TrimSpace(sttCfg.CUID),
		sttCfg.DevPID,
		strings.TrimSpace(sttCfg.Format),
		sttCfg.TimeoutSeconds,
		strings.TrimSpace(sttCfg.APIKey) != "",
		strings.TrimSpace(sttCfg.APISecret) != "",
	)
	if meta.SampleRate <= 0 || meta.Bits != 16 || meta.Channels <= 0 {
		glog.Warningf(ctx, "[流式ASR] 初始化失败：音频参数无效。sampleRate=%d bits=%d channels=%d", meta.SampleRate, meta.Bits, meta.Channels)
		return nil, StageError{Stage: "stt", Detail: "流式ASR音频参数无效"}
	}
	if meta.SampleRate != 8000 && meta.SampleRate != 16000 {
		glog.Warningf(ctx, "[流式ASR] 音频采样率非常规。sampleRate=%d（建议 8000/16000）", meta.SampleRate)
	}

	token, err := svc.getBaiduAccessToken(ctx, &svc.sttToken, sttCfg.APIKey, sttCfg.APISecret, sttCfg.TokenEndpoint, sttCfg.TimeoutSeconds)
	if err != nil {
		glog.Warningf(ctx, "[流式ASR] 初始化失败：获取token失败。tokenEndpoint=%q timeoutSec=%d apiKeySet=%v apiSecretSet=%v err=%v", strings.TrimSpace(sttCfg.TokenEndpoint), sttCfg.TimeoutSeconds, strings.TrimSpace(sttCfg.APIKey) != "", strings.TrimSpace(sttCfg.APISecret) != "", err)
		return nil, StageError{Stage: "stt", Detail: err.Error()}
	}

	rawSN := strings.TrimSpace(sttCfg.SN)
	sn, snFromConfig := normalizeBaiduStreamSN(rawSN)
	if !snFromConfig && rawSN != "" {
		glog.Warningf(ctx, "[流式ASR] 配置sn未通过校验，已改为动态唯一sn。rawSN=%q generatedSN=%q", rawSN, sn)
	}
	if rawSN == "" {
		glog.Infof(ctx, "[流式ASR] 未配置sn，已使用动态唯一sn。generatedSN=%q", sn)
	}

	dialURL, err := buildBaiduStreamURL(sttCfg.StreamEndpoint, token, sttCfg.CUID, sn, meta.SampleRate, sttCfg.DevPID)
	if err != nil {
		glog.Warningf(ctx, "[流式ASR] 初始化失败：构建WS地址失败。streamEndpoint=%q cuid=%q sampleRate=%d devPid=%d err=%v", strings.TrimSpace(sttCfg.StreamEndpoint), strings.TrimSpace(sttCfg.CUID), meta.SampleRate, sttCfg.DevPID, err)
		return nil, StageError{Stage: "stt", Detail: err.Error()}
	}
	glog.Infof(ctx, "[流式ASR] 正在建立WS连接。sn=%q url=%q", sn, dialURL)

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
		detail := "连接百度流式ASR失败: " + err.Error()
		if statusCode > 0 {
			detail += fmt.Sprintf("; status=%d", statusCode)
		}
		if statusText != "" {
			detail += fmt.Sprintf("; statusText=%s", statusText)
		}
		if bodySnippet != "" {
			detail += fmt.Sprintf("; body=%s", bodySnippet)
		}

		glog.Warningf(ctx, "[流式ASR] 初始化失败：WS握手失败。status=%d statusText=%q bodySnippet=%q url=%q err=%v", statusCode, statusText, bodySnippet, dialURL, err)
		// 末尾再拼一个请求的URL，方便排查一些特殊错误
		detail += fmt.Sprintf("; url=%s", dialURL)
		return nil, StageError{Stage: "stt", Detail: detail}
	}

	sess := &baiduStreamASRSession{
		svc:       svc,
		sttCfg:    sttCfg,
		meta:      meta,
		sn:        sn,
		conn:      conn,
		onPartial: onPartial,
		onFinal:   onFinal,
		finalCh:   make(chan string, 1),
		errCh:     make(chan error, 1),
	}

	if err := sess.sendStart(token); err != nil {
		glog.Warningf(ctx, "[流式ASR] 初始化失败：发送START失败。url=%q err=%v", dialURL, err)
		_ = conn.Close()
		return nil, err
	}
	glog.Infof(ctx, "[流式ASR] 初始化成功：会话已建立。sampleRate=%d bits=%d channels=%d", meta.SampleRate, meta.Bits, meta.Channels)

	go sess.readLoop()
	return sess, nil
}

func buildBaiduStreamURL(rawURL, token, cuid, sn string, sampleRate, devPid int) (string, error) {
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
		if q.Get("token") == "" {
			q.Set("token", token)
		}
		if q.Get("access_token") == "" {
			q.Set("access_token", token)
		}
	}
	if cuid != "" && q.Get("cuid") == "" {
		q.Set("cuid", cuid)
	}
	if sn != "" && q.Get("sn") == "" {
		q.Set("sn", sn)
	}
	if sampleRate > 0 && q.Get("sample") == "" {
		q.Set("sample", strconv.Itoa(sampleRate))
	}
	if devPid > 0 && q.Get("dev_pid") == "" {
		q.Set("dev_pid", strconv.Itoa(devPid))
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func buildBaiduStreamSN(cuid string) string {
	_ = cuid
	ms := strconv.FormatInt(time.Now().UnixMilli(), 10)
	rnd := fmt.Sprintf("%03d", time.Now().UnixNano()%1000)
	return ms + rnd
}

func normalizeBaiduStreamSN(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	base := digitsOnly(raw)
	validBase := base != "" && base == raw && len(raw) >= 8 && len(raw) <= 20
	if len(base) > 4 {
		base = base[len(base)-4:]
	}
	sn := base + buildBaiduStreamSN("")
	if len(sn) > 20 {
		sn = sn[len(sn)-20:]
	}
	if len(sn) < 8 {
		sn = buildBaiduStreamSN("")
	}
	return sn, validBase
}

func digitsOnly(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func (s *baiduStreamASRSession) sendStart(token string) error {
	payload := map[string]interface{}{
		"type": "START",
		"data": map[string]interface{}{
			"token":        token,
			"access_token": token,
			"cuid":         s.sttCfg.CUID,
			"format":       s.sttCfg.Format,
			"dev_pid":      s.sttCfg.DevPID,
			"sample":       s.meta.SampleRate,
			"sample_rate":  s.meta.SampleRate,
			"rate":         s.meta.SampleRate,
			"channel":      s.meta.Channels,
			"model":        s.sttCfg.Model,
		},
	}
	if err := s.conn.WriteJSON(payload); err != nil {
		return StageError{Stage: "stt", Detail: "发送流式ASR START失败: " + err.Error()}
	}
	return nil
}

func (s *baiduStreamASRSession) WriteAudio(chunk []byte) error {
	if len(chunk) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return StageError{Stage: "stt", Detail: "流式ASR会话已关闭"}
	}
	toSend, needMore := s.normalizeFirstAudioChunkLocked(chunk)
	if needMore || len(toSend) == 0 {
		return nil
	}
	if err := s.conn.WriteMessage(websocket.BinaryMessage, toSend); err != nil {
		return StageError{Stage: "stt", Detail: "流式ASR写音频失败: " + err.Error()}
	}
	s.audioChunkCount++
	s.audioSentBytes += len(toSend)
	s.audioNonZeroBytes += countNonZeroBytes(toSend)
	s.updatePCMStatsLocked(toSend)
	return nil
}

func (s *baiduStreamASRSession) Commit(ctx context.Context) (string, error) {
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
	glog.Infof(ctx, "[流式ASR] commit前音频统计。sampleRate=%d chunks=%d sentBytes=%d nonZeroBytes=%d nonZeroRatio=%.4f avgAbs=%d peakAbs=%d strippedWavHeader=%v",
		s.meta.SampleRate,
		s.audioChunkCount,
		s.audioSentBytes,
		s.audioNonZeroBytes,
		nonZeroRatio(s.audioNonZeroBytes, s.audioSentBytes),
		avgAbs,
		s.pcmPeakAbs,
		s.strippedWavHeader,
	)
	if avgAbs > 0 && avgAbs < 120 {
		glog.Warningf(ctx, "[流式ASR] 音频平均幅度偏低，可能被判定为无有效语音。avgAbs=%d peakAbs=%d sampleRate=%d", avgAbs, s.pcmPeakAbs, s.meta.SampleRate)
	}
	err := s.conn.WriteJSON(map[string]interface{}{"type": "FINISH"})
	latest := strings.TrimSpace(s.latest)
	s.mu.Unlock()
	if err != nil {
		return "", StageError{Stage: "stt", Detail: "流式ASR发送FINISH失败: " + err.Error()}
	}

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

func (s *baiduStreamASRSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	_ = s.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(300*time.Millisecond))
	return s.conn.Close()
}

func (s *baiduStreamASRSession) readLoop() {
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
		if mt != websocket.TextMessage {
			continue
		}
		messageType, text, isFinal, parseErr := parseBaiduStreamMessage(msg)
		if parseErr != nil {
			s.mu.Lock()
			sentBytes := s.audioSentBytes
			nonZeroBytes := s.audioNonZeroBytes
			chunks := s.audioChunkCount
			stripped := s.strippedWavHeader
			avgAbs := pcmAvgAbs(s.pcmAbsSum, s.pcmSampleCount)
			peakAbs := s.pcmPeakAbs
			s.mu.Unlock()
			glog.Warningf(context.Background(), "[流式ASR] 解析响应失败: %v; raw=%s; chunks=%d sentBytes=%d nonZeroBytes=%d nonZeroRatio=%.4f avgAbs=%d peakAbs=%d strippedWavHeader=%v",
				parseErr,
				truncateForLog(string(msg), 256),
				chunks,
				sentBytes,
				nonZeroBytes,
				nonZeroRatio(nonZeroBytes, sentBytes),
				avgAbs,
				peakAbs,
				stripped,
			)
			s.pushErr(parseErr)
			continue
		}
		if strings.Contains(strings.ToUpper(messageType), "ERROR") {
			s.pushErr(errors.New(string(msg)))
			continue
		}
		text = strings.TrimSpace(text)
		if text == "" {
			if isFinal || strings.Contains(strings.ToUpper(messageType), "FIN") {
				glog.Infof(context.Background(), "[流式ASR] 收到FIN但无有效转写文本，等待上层走整段回退。raw=%s", truncateForLog(string(msg), 256))
			}
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

func isExpectedWSStreamClose(err error) bool {
	if err == nil {
		return false
	}
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
		return true
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	if strings.Contains(msg, "close 1005") || strings.Contains(msg, "no status") {
		return true
	}
	return false
}

func (s *baiduStreamASRSession) pushErr(err error) {
	if err == nil {
		return
	}
	select {
	case s.errCh <- err:
	default:
	}
}

func (s *baiduStreamASRSession) normalizeFirstAudioChunkLocked(chunk []byte) ([]byte, bool) {
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
			glog.Warningf(context.Background(), "[流式ASR] 检测到WAV封装，已自动剥离头部后送入流式ASR。pcmBytes=%d", len(pcm))
			return pcm, false
		}
		if len(s.preBuf) < 4096 {
			return nil, true
		}
		glog.Warningf(context.Background(), "[流式ASR] 检测到WAV头但未解析到data chunk，已按原始字节透传。bufBytes=%d", len(s.preBuf))
	}
	out := s.preBuf
	s.preBuf = nil
	s.initOK = true
	return out, false
}

func countNonZeroBytes(data []byte) int {
	count := 0
	for _, b := range data {
		if b != 0 {
			count++
		}
	}
	return count
}

func (s *baiduStreamASRSession) updatePCMStatsLocked(data []byte) {
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

func pcmAvgAbs(absSum int64, sampleCount int) int {
	if sampleCount <= 0 {
		return 0
	}
	return int(absSum / int64(sampleCount))
}

func nonZeroRatio(nonZero, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(nonZero) / float64(total)
}

func truncateForLog(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func parseBaiduStreamMessage(raw []byte) (msgType string, text string, isFinal bool, err error) {
	var obj map[string]interface{}
	if err = json.Unmarshal(raw, &obj); err != nil {
		return "", "", false, err
	}
	msgType = strings.TrimSpace(strings.ToUpper(anyString(obj["type"])))
	if msgType == "" {
		msgType = strings.TrimSpace(strings.ToUpper(anyString(obj["message_type"])))
	}
	if errno := anyInt(obj["err_no"]); errno != 0 {
		if errno == -3101 {
			// 百度在长时间未继续送音频时会返回 wait audio over time。
			// 该场景可视为一次“无文本结束”，不应中断整个流式会话处理链路。
			if msgType == "" {
				msgType = "FIN_TEXT"
			}
			return msgType, "", true, nil
		}
		return "ERROR", "", false, fmt.Errorf("baidu stream err_no=%d err_msg=%s", errno, anyString(obj["err_msg"]))
	}
	if e := anyString(obj["error"]); strings.TrimSpace(e) != "" {
		return "ERROR", "", false, errors.New(e)
	}

	isFinal = anyBool(obj["is_final"]) || anyBool(obj["final"]) || strings.Contains(msgType, "FINAL") || strings.Contains(msgType, "FIN")
	text = firstText(obj)
	if text == "" {
		if data, ok := obj["data"].(map[string]interface{}); ok {
			text = firstText(data)
			if !isFinal {
				isFinal = anyBool(data["is_final"]) || anyBool(data["final"])
			}
		}
	}
	return msgType, text, isFinal, nil
}

func firstText(obj map[string]interface{}) string {
	for _, key := range []string{"result", "best_result", "final_result", "text", "content", "sentence"} {
		if v, ok := obj[key]; ok {
			if text := anyText(v); text != "" {
				return text
			}
		}
	}
	return ""
}

func anyText(v interface{}) string {
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val)
	case []interface{}:
		parts := make([]string, 0, len(val))
		for _, item := range val {
			s := strings.TrimSpace(anyString(item))
			if s != "" {
				parts = append(parts, s)
			}
		}
		return strings.TrimSpace(strings.Join(parts, ""))
	case map[string]interface{}:
		return strings.TrimSpace(firstText(val))
	default:
		return strings.TrimSpace(anyString(v))
	}
}

func anyString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case json.Number:
		return val.String()
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func anyInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case json.Number:
		i, _ := val.Int64()
		return int(i)
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(val))
		return i
	default:
		return 0
	}
}

func anyBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return strings.EqualFold(strings.TrimSpace(val), "true") || strings.TrimSpace(val) == "1"
	case int:
		return val != 0
	case float64:
		return val != 0
	default:
		return false
	}
}
