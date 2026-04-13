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

	"hello/internal/service"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	wsSlowReceiveThreshold = 5 * time.Second
	wsSlowProcessThreshold = 8 * time.Second
	wsStreamCommitTimeout  = 8 * time.Second
	wsInterruptCommitGap   = 1 * time.Second
	wsInterruptJudgeLogGap = 500 * time.Millisecond
	wsAutoCommitTimeout    = 4 * time.Second
	wsNoASRAutoCommitChunk = 10
	wsAudioChunkLogEvery   = 10
)

func registerVoiceChatWS(s *ghttp.Server) {
	s.BindHandler("/voice/chat/ws", voiceChatWS)
}

func voiceChatWS(r *ghttp.Request) {
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

	// 允许并忽略自定义头 X-Device-No；真实会话参数来自首条 start 文本帧。
	_ = r.Header.Get("X-Device-No")

	deviceNo := ""
	started := false
	meta := service.AudioMeta{}
	var streamASR service.StreamASRSession
	var audioBuffer bytes.Buffer
	defer func() {
		if streamASR != nil {
			_ = streamASR.Close()
		}
		if deviceNo != "" {
			service.VoiceWSManager().Unregister(deviceNo, ws)
		}
	}()
	streamStartAt := time.Time{}
	chunkCount := 0
	streamMode := false
	streamASRBroken := false
	audioPassThroughLogged := false
	lastPartialText := ""
	latestTranscript := ""
	lastInterruptJudgeLogAt := time.Time{}
	waitEndAfterCommit := false
	dropAudioAfterInterrupt := false
	droppedAudioChunks := 0
	asrCallbackCount := 0
	lastASRAt := time.Time{}
	lastNoASRWarnChunk := 0
	realtimeDebounce, realtimeMinRunes := service.Voice().StreamRealtimeOptions()
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
	safeWriteAudioChunks := func(audio []byte, outMeta service.AudioMeta, finishTalk bool) error {
		wsWriteMu.Lock()
		defer wsWriteMu.Unlock()
		return writeWSAudioChunks(ws, audio, outMeta, finishTalk)
	}
	resetRealtimeState := func(clearProcessed bool) {}
	var resetStreamASRUntilNextValid func()
	resetStreamBuffers := func() {
		audioBuffer.Reset()
		chunkCount = 0
		lastPartialText = ""
		latestTranscript = ""
		lastInterruptJudgeLogAt = time.Time{}
		dropAudioAfterInterrupt = false
		droppedAudioChunks = 0
	}
	runStreamCommit := func(trigger string) {
		// processStartAt := time.Now()
		transcript := ""
		if streamASR != nil && !streamASRBroken {
			cctx, cancel := context.WithTimeout(ctx, wsStreamCommitTimeout)
			var tErr error
			transcript, tErr = streamASR.Commit(cctx)
			cancel()
			if tErr != nil {
				// glog.Warningf(ctx, "[语音WS] 流式ASR commit失败，降级为整段音频处理。deviceNo=%s err=%v", deviceNo, tErr)
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
		}
		if transcript != "" {
			glog.Infof(ctx, "[语音WS][关键节点] commit获得转写。deviceNo=%s textLen=%d", deviceNo, utf8.RuneCountInString(transcript))
			finalPayload, _ := json.Marshal(map[string]interface{}{
				"type": "asr_final",
				"code": 0,
				"text": transcript,
			})
			_ = safeWriteMessage(1, finalPayload)
		}
		if trigger == "interrupt" {
			dropAudioAfterInterrupt = true
			droppedAudioChunks = 0
			interruptPayload, _ := json.Marshal(map[string]interface{}{
				"type": "interrupt_commit",
				"code": 0,
			})
			_ = safeWriteMessage(1, interruptPayload)
		}
		if transcript == "" {
			// glog.Infof(ctx, "[语音WS] commit未拿到有效转写，暂不回退整段识别，继续接收后续音频。deviceNo=%s recvBytes=%d", deviceNo, audioBuffer.Len())
			noResultPayload, _ := json.Marshal(map[string]interface{}{
				"type":    "asr_no_result",
				"code":    0,
				"message": "当前片段暂无有效转写，继续接收音频",
			})
			_ = safeWriteMessage(1, noResultPayload)

			resetStreamBuffers()
			resetStreamASRUntilNextValid()
			resetRealtimeState(true)
			return
		}
		var (
			audio      []byte
			outMeta    service.AudioMeta
			ask        string
			answer     string
			exit       bool
			finishTalk bool
			pErr       error
		)
		audio, outMeta, ask, answer, exit, finishTalk, pErr = processVoiceTranscript(ctx, deviceNo, meta, transcript)
		if pErr != nil {
			// if stageErr, ok := pErr.(service.StageError); ok {
			// glog.Warningf(ctx, "[语音WS] commit处理失败。deviceNo=%s stage=%s detail=%s 处理耗时=%s", deviceNo, stageErr.Stage, stageErr.Detail, time.Since(processStartAt))
			// } else {
			// glog.Warningf(ctx, "[语音WS] commit处理失败。deviceNo=%s error=%v 处理耗时=%s", deviceNo, pErr, time.Since(processStartAt))
			// }
			safeWriteWSError("service", pErr.Error())
			resetStreamBuffers()
			resetStreamASRUntilNextValid()
			resetRealtimeState(true)
			return
		}
		if exit {
			// glog.Infof(ctx, "[语音WS][关键节点] commit命中退出意图。deviceNo=%s", deviceNo)
			exitPayload, _ := json.Marshal(map[string]interface{}{
				"type": "exit",
				"code": 0,
				"exit": true,
			})
			_ = safeWriteMessage(1, exitPayload)
			resetStreamBuffers()
			resetStreamASRUntilNextValid()
			resetRealtimeState(true)
			return
		}

		// glog.Infof(ctx, "[语音WS][关键节点] 进入音频下发阶段（commit）。deviceNo=%s sampleRate=%d ttsBytes=%d", deviceNo, outMeta.SampleRate, len(audio))
		sendErr := safeWriteAudioChunks(audio, outMeta, finishTalk)
		if sendErr != nil {
			safeWriteWSError("service", sendErr.Error())
			resetStreamBuffers()
			resetStreamASRUntilNextValid()
			resetRealtimeState(true)
			return
		}
		if talkErr := service.DeviceAdmin().UpdateLastTalk(ctx, deviceNo, ask, answer); talkErr != nil {
			// glog.Warningf(ctx, "[语音WS] 对话记录落库失败。deviceNo=%s error=%v", deviceNo, talkErr)
			safeWriteWSError("service", talkErr.Error())
			resetStreamBuffers()
			resetStreamASRUntilNextValid()
			resetRealtimeState(true)
			return
		}
		// glog.Infof(ctx, "[语音WS][关键节点] commit处理完成并已下发音频。deviceNo=%s askLen=%d answerLen=%d ttsBytes=%d cost=%s", deviceNo, utf8.RuneCountInString(strings.TrimSpace(ask)), utf8.RuneCountInString(strings.TrimSpace(answer)), len(audio), time.Since(processStartAt))
		waitEndAfterCommit = true
		// glog.Infof(ctx, "[语音WS][关键节点] 已进入等待end状态。deviceNo=%s reason=commit_finished", deviceNo)
		resetStreamBuffers()
		resetStreamASRUntilNextValid()
		resetRealtimeState(true)
	}
	triggerRealtimeTranslate := func(_ string, text string) {
		if !streamMode {
			// glog.Infof(ctx, "[语音WS] 跳过实时处理：当前不是stream模式。deviceNo=%s source=%s", deviceNo, source)
			return
		}
		text = strings.TrimSpace(text)
		runes := utf8.RuneCountInString(text)
		if text == "" {
			// glog.Infof(ctx, "[语音WS] 跳过实时处理：ASR文本为空。deviceNo=%s source=%s", deviceNo, source)
			return
		}
		if runes < realtimeMinRunes {
			// glog.Infof(ctx, "[语音WS] 跳过实时处理：文本长度不足。deviceNo=%s source=%s runes=%d minRunes=%d text=%q", deviceNo, source, runes, realtimeMinRunes, text)
			return
		}
		// glog.Infof(ctx, "[语音WS][关键节点] 已缓存实时转写，等待commit后执行DeepSeek+TTS。deviceNo=%s source=%s textLen=%d", deviceNo, source, runes)
	}
	openStreamASR := func() error {
		if !streamMode {
			return nil
		}
		if streamASR != nil {
			_ = streamASR.Close()
			streamASR = nil
		}
		streamASRBroken = false
		sess, sErr := service.Voice().CreateStreamASRSession(ctx, meta,
			func(text string) {
				text = strings.TrimSpace(text)
				if text == "" || text == lastPartialText {
					return
				}
				asrCallbackCount++
				lastASRAt = time.Now()
				// glog.Infof(ctx, "[语音WS] 收到ASR中间结果回调。deviceNo=%s text=%q textLen=%d callbackCount=%d", deviceNo, text, utf8.RuneCountInString(text), asrCallbackCount)
				lastPartialText = text
				latestTranscript = preferLongerTranscript(latestTranscript, text)
				partialPayload, _ := json.Marshal(map[string]interface{}{
					"type": "asr_partial",
					"code": 0,
					"text": text,
				})
				_ = safeWriteMessage(1, partialPayload)
				triggerRealtimeTranslate("asr_partial", text)
			},
			func(text string) {
				text = strings.TrimSpace(text)
				if text == "" {
					return
				}
				asrCallbackCount++
				lastASRAt = time.Now()
				glog.Infof(ctx, "[语音WS] 收到ASR最终结果回调。deviceNo=%s text=%q textLen=%d callbackCount=%d", deviceNo, text, utf8.RuneCountInString(text), asrCallbackCount)
				latestTranscript = preferLongerTranscript(latestTranscript, text)
				if text == lastPartialText {
					return
				}
				lastPartialText = text
				partialPayload, _ := json.Marshal(map[string]interface{}{
					"type": "asr_partial",
					"code": 0,
					"text": text,
				})
				_ = safeWriteMessage(1, partialPayload)
				triggerRealtimeTranslate("asr_final", text)
			},
		)
		if sErr != nil {
			streamASRBroken = true
			return sErr
		}
		streamASR = sess
		return nil
	}
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
	tryAutoCommitWhenNoASRCallback := func() {
		if !streamMode || streamASR == nil || streamASRBroken {
			return
		}
		if asrCallbackCount > 0 || chunkCount < wsNoASRAutoCommitChunk {
			return
		}
		if time.Since(streamStartAt) < wsAutoCommitTimeout {
			return
		}
		if chunkCount-lastNoASRWarnChunk < wsNoASRAutoCommitChunk {
			return
		}
		lastNoASRWarnChunk = chunkCount

		// glog.Warningf(ctx, "[语音WS] 触发无回调自动commit兜底。deviceNo=%s chunks=%d recvBytes=%d 目的=尝试获取ASR文本并继续实时翻译", deviceNo, chunkCount, audioBuffer.Len())
		cctx, cancel := context.WithTimeout(ctx, wsAutoCommitTimeout)
		transcript, cErr := streamASR.Commit(cctx)
		cancel()
		if cErr != nil {
			// glog.Warningf(ctx, "[语音WS] 自动commit失败，关闭当前流式ASR并等待下个有效音频再建连。deviceNo=%s err=%v", deviceNo, cErr)
			resetStreamASRUntilNextValid()
			resetStreamBuffers()
			return
		}

		transcript = strings.TrimSpace(transcript)
		if transcript == "" {
			// glog.Warningf(ctx, "[语音WS] 自动commit未得到有效文本，关闭当前流式ASR并等待下个有效音频再建连。deviceNo=%s", deviceNo)
			resetStreamBuffers()
			resetStreamASRUntilNextValid()
			return
		}

		latestTranscript = transcript
		lastPartialText = transcript
		// glog.Infof(ctx, "[语音WS] 自动commit得到ASR文本。deviceNo=%s transcript=%q", deviceNo, transcript)
		finalPayload, _ := json.Marshal(map[string]interface{}{
			"type":   "asr_final",
			"code":   0,
			"text":   transcript,
			"source": "auto_commit",
		})
		_ = safeWriteMessage(1, finalPayload)
		triggerRealtimeTranslate("auto_commit", transcript)

		resetStreamBuffers()
		resetStreamASRUntilNextValid()
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

				if deviceNo != "" && deviceNo != startMsg.DeviceNo {
					service.VoiceWSManager().Unregister(deviceNo, ws)
				}
				deviceNo = startMsg.DeviceNo
				meta = service.AudioMeta{
					SampleRate: startMsg.SampleRate,
					Bits:       startMsg.Bits,
					Channels:   startMsg.Channels,
					Length:     startMsg.Length,
				}
				started = true
				streamMode = strings.EqualFold(strings.TrimSpace(startMsg.Mode), "stream")
				audioBuffer.Reset()
				chunkCount = 0
				streamStartAt = time.Now()
				streamASRBroken = false
				lastPartialText = ""
				latestTranscript = ""
				lastInterruptJudgeLogAt = time.Time{}
				waitEndAfterCommit = false
				dropAudioAfterInterrupt = false
				droppedAudioChunks = 0
				asrCallbackCount = 0
				lastASRAt = time.Time{}
				lastNoASRWarnChunk = 0
				resetRealtimeState(true)
				if streamMode {
					resetStreamASRUntilNextValid()
					glog.Infof(ctx, "[语音WS][关键节点] stream会话启动。deviceNo=%s", deviceNo)
					glog.Infof(ctx, "[语音WS] 流式模式已启动，等待首个有效音频分片后再连接百度ASR。deviceNo=%s sampleRate=%d bits=%d channels=%d", deviceNo, meta.SampleRate, meta.Bits, meta.Channels)
				} else if streamASR != nil {
					_ = streamASR.Close()
					streamASR = nil
				}

				glog.Infof(ctx, "[语音WS] 已收到 start，开始接收音频。deviceNo=%s sampleRate=%d bits=%d channels=%d declaredLen=%d",
					deviceNo,
					meta.SampleRate,
					meta.Bits,
					meta.Channels,
					meta.Length,
				)

				replaced := service.VoiceWSManager().Register(deviceNo, ws)
				if replaced != nil && replaced != ws {
					_ = replaced.Close()
				}
				ack, _ := json.Marshal(map[string]interface{}{"type": "started", "code": 0, "mode": func() string {
					if streamMode {
						return "stream"
					}
					return "legacy"
				}()})
				if streamMode {
					glog.Infof(ctx, "[语音WS] 实时翻译配置已生效。deviceNo=%s debounce=%s minRunes=%d", deviceNo, realtimeDebounce, realtimeMinRunes)
				}
				_ = safeWriteMessage(1, ack)
				continue

			case "commit":
				if !started {
					safeWriteWSError("state", "请先发送 start")
					continue
				}
				if !streamMode {
					safeWriteWSError("unsupported", "当前会话不是流式模式，请使用 end")
					continue
				}
				if waitEndAfterCommit {
					glog.Infof(ctx, "[语音WS] 忽略commit：上一轮已完成，等待前端end开启下一轮。deviceNo=%s", deviceNo)
					continue
				}
				if dropAudioAfterInterrupt {
					glog.Infof(ctx, "[语音WS] 忽略重复commit：当前处于interrupt后丢弃窗口，等待end/start。deviceNo=%s chunks=%d recvBytes=%d droppedChunks=%d", deviceNo, chunkCount, audioBuffer.Len(), droppedAudioChunks)
					continue
				}
				if audioBuffer.Len() == 0 {
					safeWriteWSError("validate", "audio chunk empty before commit")
					continue
				}
				glog.Infof(ctx, "[语音WS][关键节点] 收到commit。deviceNo=%s chunks=%d recvBytes=%d asrCallbacks=%d", deviceNo, chunkCount, audioBuffer.Len(), asrCallbackCount)
				runStreamCommit("client")
				continue

			case "end":
				if !started {
					safeWriteWSError("state", "请先发送 start")
					continue
				}
				if streamMode {
					glog.Infof(ctx, "[语音WS][关键节点] 收到end并关闭stream会话。deviceNo=%s chunks=%d recvBytes=%d", deviceNo, chunkCount, audioBuffer.Len())
					started = false
					audioBuffer.Reset()
					chunkCount = 0
					streamASRBroken = false
					lastPartialText = ""
					latestTranscript = ""
					lastInterruptJudgeLogAt = time.Time{}
					waitEndAfterCommit = false
					dropAudioAfterInterrupt = false
					droppedAudioChunks = 0
					resetRealtimeState(true)
					if streamASR != nil {
						_ = streamASR.Close()
						streamASR = nil
					}
					streamStartAt = time.Time{}
					endPayload, _ := json.Marshal(map[string]interface{}{"type": "ended", "code": 0})
					_ = safeWriteMessage(1, endPayload)
					continue
				}
				if audioBuffer.Len() == 0 {
					safeWriteWSError("validate", "audio chunk empty before end")
					continue
				}

				receiveCost := time.Since(streamStartAt)
				if receiveCost >= wsSlowReceiveThreshold {
					// glog.Warningf(ctx, "[语音WS] 音频接收偏慢。deviceNo=%s chunks=%d recvBytes=%d 接收耗时=%s", deviceNo, chunkCount, audioBuffer.Len(), receiveCost)
				}

				// processStartAt := time.Now()
				effective, stats := detectEffectiveSpeechPCM(audioBuffer.Bytes())
				if !effective {
					glog.Infof(ctx, "[语音WS] end音频判定为无效语音，跳过整段识别。deviceNo=%s recvBytes=%d avgAbs=%d peakAbs=%d nonZeroRatio=%.4f", deviceNo, audioBuffer.Len(), stats.AvgAbs, stats.PeakAbs, stats.NonZeroRatio)
					noResultPayload, _ := json.Marshal(map[string]interface{}{
						"type":    "asr_no_result",
						"code":    0,
						"message": "当前片段无有效语音，已跳过识别",
					})
					_ = safeWriteMessage(1, noResultPayload)
					audioBuffer.Reset()
					chunkCount = 0
					streamStartAt = time.Time{}
					continue
				}
				audio, outMeta, ask, answer, exit, finishTalk, pErr := processVoiceBuffer(ctx, deviceNo, meta, audioBuffer.Bytes())
				if pErr != nil {
					// if stageErr, ok := pErr.(service.StageError); ok {
					// 	glog.Warningf(ctx, "[语音WS] 音频处理失败。deviceNo=%s stage=%s detail=%s 处理耗时=%s", deviceNo, stageErr.Stage, stageErr.Detail, time.Since(processStartAt))
					// } else {
					// 	glog.Warningf(ctx, "[语音WS] 音频处理失败。deviceNo=%s error=%v 处理耗时=%s", deviceNo, pErr, time.Since(processStartAt))
					// }
					safeWriteWSError("service", pErr.Error())
					audioBuffer.Reset()
					chunkCount = 0
					streamStartAt = time.Time{}
					continue
				}
				if exit {
					// processCost := time.Since(processStartAt)
					// if processCost >= wsSlowProcessThreshold {
					// 	glog.Warningf(ctx, "[语音WS] 退出意图处理偏慢。deviceNo=%s 处理耗时=%s", deviceNo, processCost)
					// }
					// 发送退出事件给前端，前端应进入待唤醒状态
					exitPayload, _ := json.Marshal(map[string]interface{}{
						"type": "exit",
						"code": 0,
						"exit": true,
					})
					_ = safeWriteMessage(1, exitPayload)
					audioBuffer.Reset()
					chunkCount = 0
					streamStartAt = time.Time{}
					continue
				}

				// glog.Infof(ctx, "[语音WS][关键节点] 进入音频下发阶段（legacy end）。deviceNo=%s sampleRate=%d ttsBytes=%d", deviceNo, outMeta.SampleRate, len(audio))
				sendErr := safeWriteAudioChunks(audio, outMeta, finishTalk)
				if sendErr != nil {
					safeWriteWSError("service", sendErr.Error())
					audioBuffer.Reset()
					continue
				}
				if talkErr := service.DeviceAdmin().UpdateLastTalk(ctx, deviceNo, ask, answer); talkErr != nil {
					glog.Warningf(ctx, "[语音WS] 对话记录落库失败。deviceNo=%s error=%v", deviceNo, talkErr)
					safeWriteWSError("service", talkErr.Error())
					audioBuffer.Reset()
					chunkCount = 0
					streamStartAt = time.Time{}
					continue
				}
				// processCost := time.Since(processStartAt)
				// if processCost >= wsSlowProcessThreshold {
				// glog.Warningf(ctx, "[语音WS] 语音处理偏慢。deviceNo=%s sampleRate=%d bits=%d channels=%d audioBytes=%d 处理耗时=%s",
				// 	deviceNo,
				// 	outMeta.SampleRate,
				// 	outMeta.Bits,
				// 	outMeta.Channels,
				// 	len(audio),
				// 	processCost,
				// )
				// }
				audioBuffer.Reset()
				chunkCount = 0
				streamStartAt = time.Time{}
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
		if streamMode && waitEndAfterCommit {
			if chunkCount == 0 || chunkCount%10 == 0 {
				glog.Infof(ctx, "[语音WS][关键节点] 已完成commit并等待end，丢弃本轮后续音频。deviceNo=%s recvBytes=%d", deviceNo, len(msg))
			}
			continue
		}
		if streamMode && dropAudioAfterInterrupt {
			droppedAudioChunks++
			if droppedAudioChunks == 1 || droppedAudioChunks%10 == 0 {
				glog.Infof(ctx, "[语音WS][关键节点] 已触发主动commit中断，丢弃后续音频分片。deviceNo=%s droppedChunks=%d droppedBytes=%d", deviceNo, droppedAudioChunks, len(msg))
			}
			continue
		}

		if len(msg) == 0 {
			continue
		}
		effectiveChunk, _, _ := detectChunkSpeechWithReason(msg)
		if effectiveChunk {
			now := time.Now()
			if streamMode && streamASR != nil && !streamASRBroken && (lastInterruptJudgeLogAt.IsZero() || now.Sub(lastInterruptJudgeLogAt) >= wsInterruptJudgeLogGap) {
				lastInterruptJudgeLogAt = now
				glog.Infof(ctx, "[语音WS][关键节点] 当前未中断：本分片判定为有效音频，静默计时已重置。deviceNo=%s chunks=%d recvBytes=%d", deviceNo, chunkCount, audioBuffer.Len())
			}
		}
		if !audioPassThroughLogged {
			audioPassThroughLogged = true
			glog.Infof(ctx, "[语音WS] 按裸PCM(s16le)透传音频流。deviceNo=%s sampleRate=%d bits=%d channels=%d", deviceNo, meta.SampleRate, meta.Bits, meta.Channels)
		}
		_, _ = audioBuffer.Write(msg)
		chunkCount++
		if streamMode && (chunkCount == 1 || chunkCount%wsAudioChunkLogEvery == 0) {
			glog.Infof(ctx, "[语音WS][关键节点] stream音频接收心跳。deviceNo=%s chunks=%d recvBytes=%d asrConnected=%v asrBroken=%v callbacks=%d",
				deviceNo,
				chunkCount,
				audioBuffer.Len(),
				streamASR != nil,
				streamASRBroken,
				asrCallbackCount,
			)
		}
		if streamMode && streamASR == nil && !streamASRBroken {
			if effectiveChunk {
				if sErr := openStreamASR(); sErr != nil {
					// glog.Warningf(ctx, "[语音WS] 首个有效音频触发流式ASR建连失败，继续等待后续有效音频。deviceNo=%s err=%v", deviceNo, sErr)
					safeWriteWSError("service", "流式ASR暂不可用，稍后将自动重试")
				} else {
					glog.Infof(ctx, "[语音WS] 检测到首个有效音频，已建立流式ASR会话。deviceNo=%s", deviceNo)
				}
			} else if chunkCount == 1 || chunkCount%20 == 0 {
				// glog.Infof(ctx, "[语音WS] 尚未检测到有效音频，暂不连接百度ASR。deviceNo=%s chunks=%d recvBytes=%d", deviceNo, chunkCount, audioBuffer.Len())
			}
		}
		if streamMode && streamASR != nil && !streamASRBroken {
			if asrCallbackCount == 0 && chunkCount-lastNoASRWarnChunk >= 10 {
				lastNoASRWarnChunk = chunkCount
				// glog.Warningf(ctx, "[语音WS] 尚未收到ASR文本回调。deviceNo=%s chunks=%d recvBytes=%d 已接收时长=%s 提示=请确认客户端发送16bit小端PCM(s16le)且采样率与start一致；部分ASR服务在未收到FINISH(commit)前可能不返回最终文本。", deviceNo, chunkCount, audioBuffer.Len(), time.Since(streamStartAt))
			} else if asrCallbackCount > 0 && chunkCount%20 == 0 {
				// since := "-"
				// if !lastASRAt.IsZero() {
				// 	since = time.Since(lastASRAt).String()
				// }
				// glog.Infof(ctx, "[语音WS] ASR回调心跳。deviceNo=%s callbackCount=%d 距上次回调=%s chunks=%d", deviceNo, asrCallbackCount, since, chunkCount)
			}
		}
		if streamMode && streamASR != nil && !streamASRBroken {
			now := time.Now()
			sttSilence := time.Duration(0)
			if !lastASRAt.IsZero() {
				sttSilence = now.Sub(lastASRAt)
			}

			hasFirstSTT := !lastASRAt.IsZero()
			sttTimeout := hasFirstSTT && sttSilence >= wsInterruptCommitGap

			if sttTimeout {
				// glog.Warningf(ctx, "[语音WS] 触发主动commit中断。deviceNo=%s reason=stt_callback_timeout sttSilence=%s threshold=%s chunks=%d recvBytes=%d", deviceNo, sttSilence, wsInterruptCommitGap, chunkCount, audioBuffer.Len())
				runStreamCommit("interrupt")
				continue
			}

			if lastInterruptJudgeLogAt.IsZero() || now.Sub(lastInterruptJudgeLogAt) >= wsInterruptJudgeLogGap {
				lastInterruptJudgeLogAt = now
				if !hasFirstSTT {
					// glog.Infof(ctx, "[语音WS][关键节点] 当前未中断：尚未收到首个有效STT回调，静默计时未开始。deviceNo=%s threshold=%s chunks=%d recvBytes=%d", deviceNo, wsInterruptCommitGap, chunkCount, audioBuffer.Len())
				} else {
					sttRemain := wsInterruptCommitGap - sttSilence
					if sttRemain < 0 {
						sttRemain = 0
					}
					// glog.Infof(ctx, "[语音WS][关键节点] 当前未中断：STT回调静默未超时。deviceNo=%s sttSilence=%s sttRemain=%s threshold=%s chunks=%d recvBytes=%d", deviceNo, sttSilence, sttRemain, wsInterruptCommitGap, chunkCount, audioBuffer.Len())
				}
			}
		}

		if streamMode && streamASR != nil {
			if wErr := streamASR.WriteAudio(msg); wErr != nil {
				// glog.Warningf(ctx, "[语音WS] 流式ASR写入失败，降级为commit整段识别。deviceNo=%s err=%v", deviceNo, wErr)
				safeWriteWSError("service", "流式ASR写入失败，已自动降级为commit整段识别")
				streamASRBroken = true
				_ = streamASR.Close()
				streamASR = nil
			} else {
				tryAutoCommitWhenNoASRCallback()
			}
		}
	}
}
