package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	voice "hello/internal/services/voice"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// registerVoiceAsrWS 注册仅听写（流式 ASR）的 WebSocket 入口，与对话 WS 分离。
func registerVoiceAsrWS(s *ghttp.Server) {
	s.BindHandler("/voice/asr/ws", voiceAsrWS)
}

// voiceAsrWS 处理实时听写：上行 PCM，下行 asr_partial/asr_final。
// 不注册 VoiceWSManager、不调用 LLM/TTS/UpdateLastTalk，避免与 /voice/chat/ws 互相踢连接或落对话库。
func voiceAsrWS(r *ghttp.Request) {
	ctx := r.Context()
	ws, err := r.WebSocket()
	if err != nil {
		r.Response.Status = 400
		r.Response.WriteJson(map[string]interface{}{
			"type":    "error",
			"code":    1,
			"stage":   "handshake",
			"message": fmt.Sprintf("WebSocket 握手失败: %v", err),
		})
		return
	}

	deviceNo := ""
	started := false
	meta := voice.AudioMeta{}
	var streamASR voice.StreamASRSession
	var audioBuffer bytes.Buffer
	defer func() {
		if streamASR != nil {
			_ = streamASR.Close()
		}
	}()

	streamStartAt := time.Time{}   // 整段听写 WS 会话起始（仅 start 时设置）
	utteranceStartAt := time.Time{} // 当前句/当前 ASR 连接起始，finalize 或新建 ASR 时重置
	chunkCount := 0
	streamASRBroken := false
	audioPassThroughLogged := false
	lastPartialText := ""
	latestTranscript := ""
	asrCallbackCount := 0
	lastASRAt := time.Time{}
	lastNoASRWarnChunk := 0

	preferLongerTranscript := func(current, candidate string) string {
		current = strings.TrimSpace(current)
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			return current
		}
		if current == "" {
			return candidate
		}
		if utf8.RuneCountInString(candidate) > utf8.RuneCountInString(current) {
			return candidate
		}
		return current
	}

	wsWriteMu := sync.Mutex{}
	safeWriteMessage := func(messageType int, data []byte) error {
		wsWriteMu.Lock()
		defer wsWriteMu.Unlock()
		return ws.WriteMessage(messageType, data)
	}
	safeWriteWSError := func(stage, detail string) {
		writeWSError(safeWriteMessage, stage, detail)
	}
	emitAsrPartial := func(text string) {
		payload, _ := json.Marshal(map[string]interface{}{
			"type": "asr_partial",
			"code": 0,
			"text": text,
		})
		_ = safeWriteMessage(1, payload)
	}
	emitAsrFinal := func(text, source string) {
		payload, _ := json.Marshal(map[string]interface{}{
			"type":   "asr_final",
			"code":   0,
			"text":   text,
			"source": source,
		})
		_ = safeWriteMessage(1, payload)
	}

	resetStreamBuffers := func() {
		audioBuffer.Reset()
		chunkCount = 0
		lastPartialText = ""
		latestTranscript = ""
	}

	var resetStreamASRUntilNextValid func()
	resetStreamASRUntilNextValid = func() {
		if streamASR != nil {
			_ = streamASR.Close()
			streamASR = nil
		}
		streamASRBroken = false
		asrCallbackCount = 0
		lastASRAt = time.Time{}
		lastNoASRWarnChunk = 0
	}

	// runAsrFinalize 对当前流式 ASR 执行 FINISH/commit，仅下发听写结果，不进入对话链路。
	runAsrFinalize := func(source string) {
		transcript := ""
		if streamASR != nil && !streamASRBroken {
			cctx, cancel := context.WithTimeout(ctx, wsStreamCommitTimeout)
			var tErr error
			transcript, tErr = streamASR.Commit(cctx)
			cancel()
			if tErr != nil {
				streamASRBroken = true
				_ = streamASR.Close()
				streamASR = nil
			}
		}
		transcript = strings.TrimSpace(transcript)
		if transcript == "" {
			transcript = strings.TrimSpace(latestTranscript)
		}
		if transcript != "" {
			latestTranscript = transcript
			lastPartialText = transcript
			emitAsrFinal(transcript, source)
			glog.Infof(ctx, "[听写WS] finalize。deviceNo=%s source=%s textLen=%d", deviceNo, source, utf8.RuneCountInString(transcript))
		} else {
			noResultPayload, _ := json.Marshal(map[string]interface{}{
				"type":    "asr_no_result",
				"code":    0,
				"message": "当前片段暂无有效听写文本",
			})
			_ = safeWriteMessage(1, noResultPayload)
		}
		resetStreamBuffers()
		resetStreamASRUntilNextValid()
		utteranceStartAt = time.Time{}
	}

	openStreamASR := func() error {
		if streamASR != nil {
			_ = streamASR.Close()
			streamASR = nil
		}
		streamASRBroken = false
		sess, sErr := voice.Voice().CreateStreamASRSession(ctx, meta,
			func(text string) {
				text = strings.TrimSpace(text)
				if text == "" || text == lastPartialText {
					return
				}
				asrCallbackCount++
				lastASRAt = time.Now()
				lastPartialText = text
				latestTranscript = preferLongerTranscript(latestTranscript, text)
				emitAsrPartial(text)
			},
			func(text string) {
				text = strings.TrimSpace(text)
				if text == "" {
					return
				}
				asrCallbackCount++
				lastASRAt = time.Now()
				latestTranscript = preferLongerTranscript(latestTranscript, text)
				lastPartialText = text
				emitAsrFinal(text, "asr_callback")
				glog.Infof(ctx, "[听写WS] ASR final 回调。deviceNo=%s textLen=%d", deviceNo, utf8.RuneCountInString(text))
				resetStreamBuffers()
				resetStreamASRUntilNextValid()
				utteranceStartAt = time.Time{}
			},
		)
		if sErr != nil {
			streamASRBroken = true
			return sErr
		}
		streamASR = sess
		// 新一句 ASR 计时从建连成功开始，避免沿用整段 WS 的 streamStartAt 触发误截句。
		utteranceStartAt = time.Now()
		return nil
	}

	tryAutoCommitWhenNoASRCallback := func() {
		if streamASR == nil || streamASRBroken {
			return
		}
		if asrCallbackCount > 0 || chunkCount < wsNoASRAutoCommitChunk {
			return
		}
		ref := utteranceStartAt
		if ref.IsZero() {
			ref = streamStartAt
		}
		if ref.IsZero() || time.Since(ref) < wsAutoCommitTimeout {
			return
		}
		if chunkCount-lastNoASRWarnChunk < wsNoASRAutoCommitChunk {
			return
		}
		lastNoASRWarnChunk = chunkCount
		runAsrFinalize("auto_commit")
	}

	for {
		msgType, msg, readErr := ws.ReadMessage()
		if readErr != nil {
			return
		}

		if msgType == 1 {
			if strings.EqualFold(strings.TrimSpace(string(msg)), "ping") {
				_ = safeWriteMessage(1, []byte("pong"))
				continue
			}

			typeName, err := parseControlType(msg)
			if err != nil {
				safeWriteWSError("bad_request", "控制消息格式错误，应为 JSON")
				continue
			}

			switch typeName {
			case "start":
				startMsg, vErr := parseStartMessage(msg)
				if vErr != nil {
					safeWriteWSError("validate", vErr.Error())
					continue
				}
				mode := strings.TrimSpace(strings.ToLower(startMsg.Mode))
				if mode != "" && mode != "stream" {
					safeWriteWSError("unsupported", "听写 WS 仅支持 stream 模式或省略 mode")
					continue
				}

				deviceNo = startMsg.DeviceNo
				meta = voice.AudioMeta{
					SampleRate: startMsg.SampleRate,
					Bits:       startMsg.Bits,
					Channels:   startMsg.Channels,
					Length:     startMsg.Length,
				}
				started = true
				audioBuffer.Reset()
				chunkCount = 0
				streamStartAt = time.Now()
				streamASRBroken = false
				lastPartialText = ""
				latestTranscript = ""
				asrCallbackCount = 0
				lastASRAt = time.Time{}
				lastNoASRWarnChunk = 0
				audioPassThroughLogged = false
				utteranceStartAt = time.Time{}
				resetStreamASRUntilNextValid()

				glog.Infof(ctx, "[听写WS] 会话启动。deviceNo=%s sampleRate=%d bits=%d channels=%d", deviceNo, meta.SampleRate, meta.Bits, meta.Channels)
				ack, _ := json.Marshal(map[string]interface{}{"type": "started", "code": 0, "mode": "stream"})
				_ = safeWriteMessage(1, ack)
				continue

			case "commit":
				// 可选「手动截句」：仅触发 ASR finalize 并返回 asr_final，不进入对话/TTS。
				if !started {
					safeWriteWSError("state", "请先发送 start")
					continue
				}
				if audioBuffer.Len() == 0 && streamASR == nil {
					safeWriteWSError("validate", "commit 前无音频数据")
					continue
				}
				runAsrFinalize("client")
				continue

			case "end":
				if !started {
					safeWriteWSError("state", "请先发送 start")
					continue
				}
				glog.Infof(ctx, "[听写WS] 收到 end。deviceNo=%s chunks=%d", deviceNo, chunkCount)
				if streamASR != nil && !streamASRBroken && audioBuffer.Len() > 0 {
					runAsrFinalize("end")
				}
				started = false
				resetStreamBuffers()
				resetStreamASRUntilNextValid()
				streamStartAt = time.Time{}
				utteranceStartAt = time.Time{}
				endPayload, _ := json.Marshal(map[string]interface{}{"type": "ended", "code": 0})
				_ = safeWriteMessage(1, endPayload)
				continue

			default:
				safeWriteWSError("unsupported", fmt.Sprintf("不支持的控制消息类型: %s", typeName))
				continue
			}
		}

		if msgType != 2 {
			continue
		}
		if !started {
			safeWriteWSError("state", "请先发送 start，再发送二进制音频")
			continue
		}
		if len(msg) == 0 {
			continue
		}

		effectiveChunk, _, _ := detectChunkSpeechWithReason(msg)
		if !audioPassThroughLogged {
			audioPassThroughLogged = true
			glog.Infof(ctx, "[听写WS] PCM(s16le) 透传。deviceNo=%s sampleRate=%d", deviceNo, meta.SampleRate)
		}
		_, _ = audioBuffer.Write(msg)
		chunkCount++

		if streamASR == nil && !streamASRBroken {
			if effectiveChunk {
				if sErr := openStreamASR(); sErr != nil {
					safeWriteWSError("stt", "流式 ASR 暂不可用："+sErr.Error())
				} else {
					glog.Infof(ctx, "[听写WS] 已建立流式 ASR。deviceNo=%s", deviceNo)
				}
			}
		}

		if streamASR != nil && !streamASRBroken {
			if wErr := streamASR.WriteAudio(msg); wErr != nil {
				safeWriteWSError("stt", "流式 ASR 写入失败")
				streamASRBroken = true
				_ = streamASR.Close()
				streamASR = nil
				utteranceStartAt = time.Time{}
			} else {
				// 先写入百度，再判断静音截句，避免「刚建连尚未送音频」就被 finalize。
				now := time.Now()
				sttSilence := time.Duration(0)
				if !lastASRAt.IsZero() {
					sttSilence = now.Sub(lastASRAt)
				}
				hasFirstSTT := !lastASRAt.IsZero()
				uttRef := utteranceStartAt
				if uttRef.IsZero() {
					uttRef = streamStartAt
				}
				noFirstSTTTimeout := !hasFirstSTT && !uttRef.IsZero() && now.Sub(uttRef) >= wsInitialNoASRGap
				sttTimeout := hasFirstSTT && sttSilence >= wsInterruptCommitGap

				if noFirstSTTTimeout || sttTimeout {
					runAsrFinalize("silence")
					continue
				}
				tryAutoCommitWhenNoASRCallback()
			}
		}
	}
}
