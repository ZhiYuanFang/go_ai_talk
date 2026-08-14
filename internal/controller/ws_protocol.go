package controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	voice "hello/internal/services/voice"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

const wsAudioChunkSize = 16 * 1024

type wsControlMessage struct {
	Type string `json:"type"`
}

type wsStartMessage struct {
	Type            string `json:"type"`
	DeviceNo        string `json:"deviceNo"`
	SampleRate      int    `json:"sampleRate"`
	Bits            int    `json:"bits"`
	Channels        int    `json:"channels"`
	Length          int    `json:"length"`
	Mode            string `json:"mode"`
	ProtocolVersion int    `json:"protocolVersion"`
	// InputModality 输入模态：audio（默认，PCM+ASR）或 text（JSON text 帧）；文模式仍须带齐音频元数据占位字段。
	InputModality string `json:"inputModality"`
	// OutputModality 输出模态：audio（默认，TTS）或 text（仅 thinking_delta/answer，不下发音频）。
	OutputModality string `json:"outputModality"`
}

// wsTextMessage 文输入上行帧（type=text）。
type wsTextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// normalizeIOModality 规范化 audio|text；空则用 def；非法返回 error。
func normalizeIOModality(raw, def string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == "" {
		s = def
	}
	if s != "audio" && s != "text" {
		return "", fmt.Errorf("模态须为 audio 或 text")
	}
	return s, nil
}

func parseControlType(msg []byte) (string, error) {
	trimmed := strings.TrimSpace(string(msg))
	if trimmed == "" {
		return "", fmt.Errorf("empty control message")
	}
	if strings.EqualFold(trimmed, "end") {
		return "end", nil
	}

	var ctrl wsControlMessage
	if err := json.Unmarshal([]byte(trimmed), &ctrl); err != nil {
		return "", err
	}
	return strings.ToLower(strings.TrimSpace(ctrl.Type)), nil
}

func parseStartMessage(msg []byte) (wsStartMessage, error) {
	var start wsStartMessage
	if err := json.Unmarshal(msg, &start); err != nil {
		return start, err
	}
	if strings.ToLower(strings.TrimSpace(start.Type)) != "start" {
		return start, fmt.Errorf("invalid start type")
	}
	if strings.TrimSpace(start.DeviceNo) == "" {
		return start, fmt.Errorf("deviceNo 不能为空")
	}
	// 文输入亦不放宽：sampleRate/bits/channels 仍必填（客户端可传横屏同款占位值）。
	if start.SampleRate <= 0 || start.Bits <= 0 || start.Channels <= 0 {
		return start, fmt.Errorf("start 消息中的音频参数无效")
	}
	inMod, err := normalizeIOModality(start.InputModality, "audio")
	if err != nil {
		return start, fmt.Errorf("inputModality: %w", err)
	}
	outMod, err := normalizeIOModality(start.OutputModality, "audio")
	if err != nil {
		return start, fmt.Errorf("outputModality: %w", err)
	}
	start.InputModality = inMod
	start.OutputModality = outMod
	return start, nil
}

// parseTextMessage 解析文输入帧；文本 trim 后不能为空。
func parseTextMessage(msg []byte) (wsTextMessage, error) {
	var tm wsTextMessage
	if err := json.Unmarshal(msg, &tm); err != nil {
		return tm, err
	}
	if strings.ToLower(strings.TrimSpace(tm.Type)) != "text" {
		return tm, fmt.Errorf("invalid text type")
	}
	tm.Text = strings.TrimSpace(tm.Text)
	if tm.Text == "" {
		return tm, fmt.Errorf("text 不能为空")
	}
	return tm, nil
}

func processVoiceBuffer(ctx context.Context, deviceNo string, meta voice.AudioMeta, audioRaw []byte) ([]byte, voice.AudioMeta, string, string, bool, bool, error) {
	if len(audioRaw) == 0 {
		return nil, meta, "", "", false, false, fmt.Errorf("empty audio")
	}
	glog.Infof(ctx, "[语音WS][关键节点] 收到整段音频，进入语音链路处理。deviceNo=%s audioBytes=%d", deviceNo, len(audioRaw))
	meta.Length = len(audioRaw)
	audioBase64 := base64.StdEncoding.EncodeToString(audioRaw)
	audio, outMeta, ask, answer, exit, finishTalk, handleErr := voice.Voice().HandleWithDialogue(ctx, deviceNo, meta, audioBase64)
	if handleErr != nil {
		return nil, meta, "", "", false, false, handleErr
	}
	glog.Infof(ctx, "[语音WS][关键节点] 语音链路处理完成。deviceNo=%s exit=%v askLen=%d answerLen=%d ttsBytes=%d", deviceNo, exit, utf8.RuneCountInString(strings.TrimSpace(ask)), utf8.RuneCountInString(strings.TrimSpace(answer)), len(audio))
	return audio, outMeta, ask, answer, exit, finishTalk, nil
}

func processVoiceTranscript(ctx context.Context, deviceNo string, meta voice.AudioMeta, transcript string) ([]byte, voice.AudioMeta, string, string, bool, bool, error) {
	if strings.TrimSpace(transcript) == "" {
		return nil, meta, "", "", false, false, fmt.Errorf("empty transcript")
	}
	glog.Infof(ctx, "[语音WS][关键节点] 转写完成，进入DeepSeek阶段。deviceNo=%s transcriptLen=%d", deviceNo, utf8.RuneCountInString(strings.TrimSpace(transcript)))
	audio, outMeta, ask, answer, exit, finishTalk, handleErr := voice.Voice().HandleWithTranscript(ctx, deviceNo, meta, transcript)
	if handleErr != nil {
		return nil, meta, "", "", false, false, handleErr
	}
	glog.Infof(ctx, "[语音WS][关键节点] DeepSeek+TTS处理完成。deviceNo=%s exit=%v askLen=%d answerLen=%d ttsBytes=%d", deviceNo, exit, utf8.RuneCountInString(strings.TrimSpace(ask)), utf8.RuneCountInString(strings.TrimSpace(answer)), len(audio))
	return audio, outMeta, ask, answer, exit, finishTalk, nil
}

func writeWSError(writeFn func(messageType int, data []byte) error, stage, detail string) {
	_ = stage
	writeWSErrorCode(writeFn, 1, detail)
}

func writeWSErrorCode(writeFn func(messageType int, data []byte) error, code int, detail string) {
	if writeFn == nil {
		return
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"type":    "error",
		"code":    code,
		"stage":   "quota",
		"message": detail,
	})
	_ = writeFn(1, payload)
}

func writeWSAudioChunks(ws *ghttp.WebSocket, audio []byte, meta voice.AudioMeta, finishTalk bool) error {
	if ws == nil {
		return fmt.Errorf("websocket is nil")
	}
	encoded := base64.StdEncoding.EncodeToString(audio)
	if encoded == "" {
		endPayload, _ := json.Marshal(map[string]interface{}{
			"type":        "audio_end",
			"code":        0,
			"exit":        false,
			"finish_talk": finishTalk,
		})
		return ws.WriteMessage(1, endPayload)
	}

	for start := 0; start < len(encoded); start += wsAudioChunkSize {
		end := start + wsAudioChunkSize
		if end > len(encoded) {
			end = len(encoded)
		}
		chunkPayload, _ := json.Marshal(map[string]interface{}{
			"type":       "audio_chunk",
			"audio":      encoded[start:end],
			"sampleRate": meta.SampleRate,
		})
		if err := ws.WriteMessage(1, chunkPayload); err != nil {
			return err
		}
	}

	endPayload, _ := json.Marshal(map[string]interface{}{
		"type":        "audio_end",
		"code":        0,
		"exit":        false,
		"finish_talk": finishTalk,
	})
	return ws.WriteMessage(1, endPayload)
}

type pcmSpeechStats struct {
	AvgAbs       int
	PeakAbs      int
	NonZeroRatio float64
}

func detectEffectiveSpeechPCM(raw []byte) (bool, pcmSpeechStats) {
	effective, stats, _ := detectEffectiveSpeechPCMWithReason(raw)
	return effective, stats
}

func detectEffectiveSpeechPCMWithReason(raw []byte) (bool, pcmSpeechStats, string) {
	stats := pcmSpeechStats{}
	if len(raw) < 1600 {
		return false, stats, "too_short"
	}
	limit := len(raw) - (len(raw) % 2)
	if limit <= 0 {
		return false, stats, "invalid_pcm_bytes"
	}
	var absSum int64
	nonZero := 0
	samples := limit / 2
	for i := 0; i < limit; i += 2 {
		v := int(int16(uint16(raw[i]) | uint16(raw[i+1])<<8))
		if v < 0 {
			v = -v
		}
		absSum += int64(v)
		if v > stats.PeakAbs {
			stats.PeakAbs = v
		}
		if v > 0 {
			nonZero++
		}
	}
	stats.AvgAbs = int(absSum / int64(samples))
	stats.NonZeroRatio = float64(nonZero) / float64(samples)

	if stats.NonZeroRatio < 0.02 {
		return false, stats, "low_non_zero_ratio"
	}
	if stats.AvgAbs < 220 && stats.PeakAbs < 1200 {
		return false, stats, "low_energy"
	}
	return true, stats, "ok"
}

// detectChunkSpeechWithReason 仅用于分片日志展示：
// 即使分片较短也会给出“有效/无效”的近似判断，不参与提交前过滤。
func detectChunkSpeechWithReason(raw []byte) (bool, pcmSpeechStats, string) {
	stats := pcmSpeechStats{}
	limit := len(raw) - (len(raw) % 2)
	if limit <= 0 {
		return false, stats, "invalid_pcm_bytes"
	}
	var absSum int64
	nonZero := 0
	samples := limit / 2
	for i := 0; i < limit; i += 2 {
		v := int(int16(uint16(raw[i]) | uint16(raw[i+1])<<8))
		if v < 0 {
			v = -v
		}
		absSum += int64(v)
		if v > stats.PeakAbs {
			stats.PeakAbs = v
		}
		if v > 0 {
			nonZero++
		}
	}
	stats.AvgAbs = int(absSum / int64(samples))
	stats.NonZeroRatio = float64(nonZero) / float64(samples)

	if stats.NonZeroRatio < 0.02 {
		if len(raw) < 1600 {
			return false, stats, "short_low_non_zero_ratio"
		}
		return false, stats, "low_non_zero_ratio"
	}
	if stats.AvgAbs < 220 && stats.PeakAbs < 1200 {
		if len(raw) < 1600 {
			return false, stats, "short_low_energy"
		}
		return false, stats, "low_energy"
	}
	if len(raw) < 1600 {
		return true, stats, "short_but_effective"
	}
	return true, stats, "ok"
}
