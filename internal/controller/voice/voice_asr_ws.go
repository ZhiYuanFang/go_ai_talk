package voicectrl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	voice "hello/internal/services/voice"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// registerVoiceAsrWS 注册仅听写（流式 ASR）的 WebSocket 入口，与对话 WS 分离。
func RegisterVoiceAsrWS(s *ghttp.Server) {
	s.BindHandler("/voice/asr/ws", voiceAsrWS)
}

// voiceAsrWS 处理实时听写：上行 PCM，下行 asr_partial / asr_final。
//
// 与 /voice/chat/ws 的差异：听写线不做服务端静音截句（无 silence/auto_commit），
// 引擎 onFinal（如百炼 sentence_end）仅再发 asr_partial，不作为业务 asr_final、不关 ASR；
// 业务定稿 asr_final 仅由前端 commit/end 触发。
//
// 下行 text 为「本次 start…commit 会话全文」：引擎分句后在服务端拼接，供 Flutter 听写预览/定稿，
// 避免静音后 partial 只带末句导致客户端「变成 B、松手变 A」。
// 不注册 VoiceWSManager，不调用 LLM/TTS/UpdateLastTalk。
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

	chunkCount := 0
	streamASRBroken := false
	audioPassThroughLogged := false
	// lastEmittedFull：上次已下发的会话全文，用于去重
	lastEmittedFull := ""
	// sessionCommitted：引擎句末已锁定的全文；sessionUnstable：当前句未锁定草稿
	sessionCommitted := ""
	sessionUnstable := ""
	// lastCommittedFrag：上一句定稿原文，用于识别引擎重复推送（避免 HasSuffix 误伤短新句）
	lastCommittedFrag := ""

	// joinTranscript 中文听写直接拼接（段已各自 trim）
	joinTranscript := func(a, b string) string {
		a = strings.TrimSpace(a)
		b = strings.TrimSpace(b)
		if a == "" {
			return b
		}
		if b == "" {
			return a
		}
		return a + b
	}
	// preferLongerSame 同句伸长：取 rune 更长者
	preferLongerSame := func(a, b string) string {
		a = strings.TrimSpace(a)
		b = strings.TrimSpace(b)
		if a == "" {
			return b
		}
		if b == "" {
			return a
		}
		if utf8.RuneCountInString(b) > utf8.RuneCountInString(a) {
			return b
		}
		return a
	}
	// sessionFull 当前会话草稿全文（committed 与 unstable 同句重叠时不重复拼）
	sessionFull := func() string {
		c := strings.TrimSpace(sessionCommitted)
		u := strings.TrimSpace(sessionUnstable)
		if u == "" {
			return c
		}
		if c == "" {
			return u
		}
		// 句末后引擎若再推同句 partial → 只回 committed
		if u == c || (lastCommittedFrag != "" && u == lastCommittedFrag) {
			return c
		}
		// 引擎 partial 已带全文前缀
		if strings.HasPrefix(u, c) {
			return u
		}
		return joinTranscript(c, u)
	}
	// appendCommittedFrag 将一句定稿并入 committed（已在末尾则跳过，防 AA）
	appendCommittedFrag := func(frag string) {
		frag = strings.TrimSpace(frag)
		if frag == "" {
			return
		}
		c := strings.TrimSpace(sessionCommitted)
		if c == "" {
			sessionCommitted = frag
			return
		}
		if frag == c || strings.HasSuffix(c, frag) {
			return
		}
		// 引擎句末 text 已是含前缀的全文
		if strings.HasPrefix(frag, c) {
			sessionCommitted = frag
			return
		}
		sessionCommitted = joinTranscript(c, frag)
	}
	// mergeFinalizeText 将 Commit 返回与 session 全文去重合并
	mergeFinalizeText := func(commitText string) string {
		pending := sessionFull()
		commitText = strings.TrimSpace(commitText)
		if pending == "" {
			return commitText
		}
		if commitText == "" {
			return pending
		}
		if commitText == pending || strings.HasSuffix(pending, commitText) {
			return pending
		}
		if strings.HasPrefix(commitText, pending) {
			return commitText
		}
		if strings.Contains(pending, commitText) {
			return pending
		}
		if strings.Contains(commitText, pending) {
			return commitText
		}
		return joinTranscript(pending, commitText)
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
		text = strings.TrimSpace(text)
		if text == "" || text == lastEmittedFull {
			return
		}
		lastEmittedFull = text
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
		lastEmittedFull = ""
		sessionCommitted = ""
		sessionUnstable = ""
		lastCommittedFrag = ""
	}

	var resetStreamASRUntilNextValid func()
	resetStreamASRUntilNextValid = func() {
		if streamASR != nil {
			_ = streamASR.Close()
			streamASR = nil
		}
		streamASRBroken = false
	}

	// runAsrFinalize 仅由前端 commit/end 调用：对当前流式 ASR 执行 Commit（百炼 finish-task / 百度 FINISH）并下发会话全文 asr_final。
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
		full := mergeFinalizeText(transcript)
		if full != "" {
			emitAsrFinal(full, source)
			glog.Infof(ctx, "[听写WS] finalize。deviceNo=%s source=%s textLen=%d", deviceNo, source, utf8.RuneCountInString(full))
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
	}

	openStreamASR := func() error {
		if streamASR != nil {
			_ = streamASR.Close()
			streamASR = nil
		}
		streamASRBroken = false
		sess, sErr := voice.Voice().CreateStreamASRSession(ctx, voice.STTProfileDictation, meta,
			func(text string) {
				// 当前句中间结果：刷新 unstable，下行会话全文（与已锁定句去重）
				text = strings.TrimSpace(text)
				if text == "" {
					return
				}
				c := strings.TrimSpace(sessionCommitted)
				// 与已锁定全文相同，或引擎重复推送上一句定稿 → 忽略
				if c != "" && (text == c || (lastCommittedFrag != "" && text == lastCommittedFrag)) {
					return
				}
				// 引擎若直接推「已锁定前缀 + 新内容」的全文，只把后缀放入 unstable
				if c != "" && strings.HasPrefix(text, c) {
					sessionUnstable = strings.TrimSpace(strings.TrimPrefix(text, c))
					emitAsrPartial(sessionFull())
					return
				}
				sessionUnstable = text
				emitAsrPartial(sessionFull())
			},
			// 引擎 onFinal（句末）：锁定进 committed（去重），仍只发 asr_partial（全文），不发 asr_final、不关 ASR。
			func(text string) {
				text = strings.TrimSpace(text)
				if text == "" {
					return
				}
				// 本句 frag：unstable 与 onFinal text 取更长（同句）
				frag := preferLongerSame(sessionUnstable, text)
				sessionUnstable = ""
				beforeLen := utf8.RuneCountInString(strings.TrimSpace(sessionCommitted))
				appendCommittedFrag(frag)
				lastCommittedFrag = frag
				emitAsrPartial(strings.TrimSpace(sessionCommitted))
				afterLen := utf8.RuneCountInString(strings.TrimSpace(sessionCommitted))
				glog.Infof(ctx, "[听写WS] 引擎句末并入会话全文。deviceNo=%s fragLen=%d fullLen=%d appended=%v",
					deviceNo, utf8.RuneCountInString(frag), afterLen, afterLen > beforeLen)
			},
		)
		if sErr != nil {
			streamASRBroken = true
			return sErr
		}
		streamASR = sess
		return nil
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
				streamASRBroken = false
				lastEmittedFull = ""
				sessionCommitted = ""
				sessionUnstable = ""
				lastCommittedFrag = ""
				audioPassThroughLogged = false
				resetStreamASRUntilNextValid()

				glog.Infof(ctx, "[听写WS] 会话启动。deviceNo=%s sampleRate=%d bits=%d channels=%d", deviceNo, meta.SampleRate, meta.Bits, meta.Channels)
				ack, _ := json.Marshal(map[string]interface{}{"type": "started", "code": 0, "mode": "stream"})
				_ = safeWriteMessage(1, ack)
				continue

			case "commit":
				// 一句听写结束：唯一由前端主动截句的常规路径（松手/点完成时发送）。
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
			}
		}
	}
}
