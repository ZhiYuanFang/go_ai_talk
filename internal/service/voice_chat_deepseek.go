package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const growthSuggestUserPrompt = "你是育儿助手。请根据提供的育儿历史记录，给出今日成长建议（100字以内，实用、温和）。"
const growthSuggestDataHint = "下列 JSON 含 child_info 与 history。你必须严格依据这些记录作答，不要编造未出现的情节。\n输出为一段纯中文正文，不要 Markdown、不要输出 JSON。"

// growthSuggestPayload 成长建议专用 DeepSeek 请求体（含 child_info / history）。
type growthSuggestPayload struct {
	Messages  []growthSuggestMessage   `json:"messages"`
	Model     string                   `json:"model"`
	Stream    bool                     `json:"stream"`
	ChildInfo growthSuggestChildInfo   `json:"child_info"`
	History   []map[string]interface{} `json:"history"`
}

type growthSuggestMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type growthSuggestChildInfo struct {
	Birthday string `json:"birthday"`
	Gender   string `json:"gender"`
}

func defaultSystemPromptForSuggest(cfg VoiceChatConfig) string {
	p := strings.TrimSpace(cfg.DeepSeek.SystemPrompt)
	if p == "" {
		return "你是语音助手。"
	}
	return p
}

func birthdayForSuggestAPI(birthday string) string {
	b := strings.TrimSpace(birthday)
	if b == "" || b == "未设置" {
		return ""
	}
	return b
}

func suggestDurationMinutes(startStr, endStr string) int {
	startAt, ok1 := parseDBTime(strings.TrimSpace(startStr))
	endAt, ok2 := parseDBTime(strings.TrimSpace(endStr))
	if !ok1 || !ok2 || endAt.Before(startAt) {
		return 0
	}
	sec := endAt.Sub(startAt).Seconds()
	if sec < 0 {
		return 0
	}
	return int(math.Round(sec / 60))
}

// loadEventNameAndUnitByID 事件表 id -> 名称、单位（成长建议 history.type 用名称；amount 需带单位）。
func loadEventNameAndUnitByID(ctx context.Context) (names map[int64]string) {
	names = make(map[int64]string)
	var events []entity.Event
	err := dao.Event.Ctx(ctx).Fields(dao.Event.Columns().Id, dao.Event.Columns().Name, dao.Event.Columns().NeedQuantity).Scan(&events)
	if err != nil || len(events) == 0 {
		return names
	}
	for _, e := range events {
		names[e.Id] = strings.TrimSpace(e.Name)
	}
	return names
}

// growthSuggestHistoryCutoff 成长建议只取最近 48 小时内的记录（滚动「两天」）。
func growthSuggestHistoryCutoff() string {
	return time.Now().Add(-48 * time.Hour).Format("2006-01-02 15:04:05")
}

func (s *VoiceService) buildGrowthSuggestHistory(ctx context.Context, deviceNo string) ([]map[string]interface{}, error) {
	cutoff := growthSuggestHistoryCutoff()
	startCol := dao.History.Columns().StartTime
	endCol := dao.History.Columns().EndTime
	// 开始或结束时间任一落在窗口内即纳入（跨天会话也能保留）
	whereOverlap := fmt.Sprintf("(%s >= ? OR %s >= ?)", startCol, endCol)

	var rows []entity.History
	err := dao.History.Ctx(ctx).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		Where(whereOverlap, cutoff, cutoff).
		OrderDesc(dao.History.Columns().Id).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	eventNames := loadEventNameAndUnitByID(ctx)
	out := make([]map[string]interface{}, 0, len(rows))
	for _, h := range rows {
		typeName := eventNames[h.EventId]
		if typeName == "" {
			typeName = strings.TrimSpace(h.EventName)
		}
		if typeName == "" {
			typeName = "未知事件"
		}
		start := strings.TrimSpace(h.StartTime)
		end := strings.TrimSpace(h.EndTime)
		note := strings.TrimSpace(h.Remark)
		amt := h.EventNumber
		if amt < 0 {
			amt = 0
		}
		item := map[string]interface{}{
			"type":         typeName,
			"start_time":   start,
			"end_time":     end,
			"amount_value": amt,
			"note":         note,
		}
		if dm := suggestDurationMinutes(start, end); dm > 0 {
			item["duration_minutes"] = dm
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *VoiceService) buildRecentHistory(ctx context.Context, deviceNo string, hours int) ([]map[string]interface{}, error) {
	if hours <= 0 {
		hours = 12
	}
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour).Format("2006-01-02 15:04:05")
	startCol := dao.History.Columns().StartTime
	endCol := dao.History.Columns().EndTime
	whereOverlap := fmt.Sprintf("(%s >= ? OR %s >= ?)", startCol, endCol)
	var rows []entity.History
	err := dao.History.Ctx(ctx).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		Where(whereOverlap, cutoff, cutoff).
		OrderDesc(dao.History.Columns().Id).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, h := range rows {
		out = append(out, map[string]interface{}{
			"event_name":   strings.TrimSpace(h.EventName),
			"start_time":   strings.TrimSpace(h.StartTime),
			"end_time":     strings.TrimSpace(h.EndTime),
			"event_number": h.EventNumber,
			"remark":       strings.TrimSpace(h.Remark),
		})
	}
	return out, nil
}

func buildGrowthSuggestUserContent(child growthSuggestChildInfo, history []map[string]interface{}) (string, error) {
	ctxBlock := map[string]interface{}{
		"child_info": child,
		"history":    history,
	}
	ctxBytes, err := json.Marshal(ctxBlock)
	if err != nil {
		return "", err
	}
	// 标准 Chat Completions 只消费 messages；顶层 child_info/history 会被忽略，必须把数据写进 user.content。
	return growthSuggestUserPrompt + "\n\n" + growthSuggestDataHint + "\n\n" + string(ctxBytes), nil
}

func (s *VoiceService) buildGrowthSuggestPayload(ctx context.Context, deviceNo string) (growthSuggestPayload, error) {
	birthday, gender := s.loadDeviceProfile(ctx, deviceNo)
	history, err := s.buildGrowthSuggestHistory(ctx, deviceNo)
	if err != nil {
		return growthSuggestPayload{}, err
	}
	child := growthSuggestChildInfo{
		Birthday: birthdayForSuggestAPI(birthday),
		Gender:   gender,
	}
	userContent, err := buildGrowthSuggestUserContent(child, history)
	if err != nil {
		return growthSuggestPayload{}, err
	}
	return growthSuggestPayload{
		Messages: []growthSuggestMessage{
			{Content: "您是育儿助手，主要帮助家长根据历史记录提供成长建议。", Role: "system"},
			{Content: userContent, Role: "user"},
		},
		Model:     s.cfg.DeepSeek.Model,
		Stream:    false,
		ChildInfo: child,
		History:   history,
	}, nil
}

// callDeepSeekGrowthSuggestion 成长建议：按结构化 child_info + history 请求 DeepSeek。
func (s *VoiceService) callDeepSeekGrowthSuggestion(ctx context.Context, deviceNo string) (string, error) {
	if s.cfg.DeepSeek.Endpoint == "" {
		return "", StageError{Stage: "chat", Detail: "DeepSeek endpoint 未配置"}
	}
	payload, err := s.buildGrowthSuggestPayload(ctx, deviceNo)
	if err != nil {
		return "", err
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	if s.cfg.DebugLog {
		glog.Infof(ctx, "[大模型请求] 发送 DeepSeek 请求（成长建议）。deviceNo=%s 请求体=%s", deviceNo, string(bodyBytes))
	}
	cctx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.DeepSeek.TimeoutSeconds)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(cctx, http.MethodPost, s.cfg.DeepSeek.Endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.DeepSeek.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.DeepSeek.APIKey)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	if s.cfg.DebugLog {
		glog.Infof(ctx, "[大模型响应] 收到 DeepSeek 原始响应（成长建议）。deviceNo=%s 响应体=%s", deviceNo, string(body))
	}
	rawContent, replyNormalized, _, err := extractChatReplyRaw(body)
	if err != nil {
		return "", err
	}
	reply := pickGrowthSuggestionDisplayText(rawContent, replyNormalized)
	if s.cfg.DebugLog {
		glog.Infof(ctx, "[大模型解析] 解析回复完成（成长建议）。deviceNo=%s 回复文本=%s", deviceNo, reply)
	}

	// 将成长建议回复存入数据库，便于后续查询和分析。
	_, insertErr := dao.Suggest.Ctx(ctx).Data(g.Map{
		dao.Suggest.Columns().DeviceNo: deviceNo,
		dao.Suggest.Columns().Suggest:  reply,
		dao.Suggest.Columns().Time:     nowText(),
	}).Insert()
	if insertErr != nil {
		glog.Warningf(ctx, "insert suggest failed: %v", insertErr)
	}

	return reply, nil
}

func (s *VoiceService) callDeepSeekRaw(ctx context.Context, deviceNo, prompt string, historyLimit int, systemMessage ...string) (rawContent string, reply string, exit bool, err error) {
	if s.cfg.DeepSeek.Endpoint == "" {
		return "", "", false, StageError{Stage: "chat", Detail: "DeepSeek endpoint 未配置"}
	}
	messages := s.buildChatMessagesWithLimit(deviceNo, prompt, historyLimit, systemMessage...)
	payload := map[string]interface{}{
		"model":    s.cfg.DeepSeek.Model,
		"messages": messages,
		"stream":   false,
	}
	bodyBytes, _ := json.Marshal(payload)
	if s.cfg.DebugLog {
		glog.Infof(ctx, "[大模型请求] 发送 DeepSeek 请求（统一调用）。deviceNo=%s historyLimit=%d 请求体=%s", deviceNo, historyLimit, string(bodyBytes))
	}
	cctx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.DeepSeek.TimeoutSeconds)*time.Second)
	defer cancel()

	req, reqErr := http.NewRequestWithContext(cctx, http.MethodPost, s.cfg.DeepSeek.Endpoint, bytes.NewReader(bodyBytes))
	if reqErr != nil {
		return "", "", false, reqErr
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.DeepSeek.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.DeepSeek.APIKey)
	}
	resp, doErr := s.httpClient.Do(req)
	if doErr != nil {
		return "", "", false, doErr
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", "", false, readErr
	}
	if resp.StatusCode >= 300 {
		return "", "", false, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	if s.cfg.DebugLog {
		glog.Infof(ctx, "[大模型响应] 收到 DeepSeek 原始响应（统一调用）。deviceNo=%s 响应体=%s", deviceNo, string(body))
	}
	return extractChatReplyRaw(body)
}

func (s *VoiceService) callDeepSeekDirectReply(ctx context.Context, deviceNo, transcript string) (string, error) {
	systemMessage := "你是闲聊助手，请直接回答用户问题，语言自然简洁。不要使用表情符号或特殊颜文字。"
	_, reply, _, err := s.callDeepSeekRaw(ctx, deviceNo, transcript, 6, systemMessage)
	if err != nil {
		return "", err
	}
	reply = sanitizeModelReplyText(reply)
	if reply == "" {
		reply = "我在，请继续说。"
	}
	return reply, nil
}

func (s *VoiceService) streamCasualReplyWithBaiduTTS(
	ctx context.Context,
	deviceNo string,
	meta AudioMeta,
	transcript string,
	onTextDelta func(text string) error,
	onAudioChunk func(audio []byte, meta AudioMeta, seq int) error,
) (string, error) {
	if strings.ToLower(strings.TrimSpace(s.cfg.TTS.Provider)) != "baidu" {
		return "", StageError{Stage: "tts", Detail: "闲聊流式音频下发仅支持百度TTS"}
	}
	messages := s.buildChatMessagesWithLimit(deviceNo, transcript, 6, "你是闲聊助手，请直接回答用户问题，语言自然简洁。不要使用表情符号或特殊颜文字。")
	payload := map[string]interface{}{
		"model":    s.cfg.DeepSeek.Model,
		"messages": messages,
		"stream":   true,
	}
	bodyBytes, _ := json.Marshal(payload)
	cctx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.DeepSeek.TimeoutSeconds)*time.Second)
	defer cancel()

	req, reqErr := http.NewRequestWithContext(cctx, http.MethodPost, s.cfg.DeepSeek.Endpoint, bytes.NewReader(bodyBytes))
	if reqErr != nil {
		return "", reqErr
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.DeepSeek.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.DeepSeek.APIKey)
	}
	resp, doErr := s.httpClient.Do(req)
	if doErr != nil {
		return "", doErr
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var fullReply strings.Builder
	var sentenceBuf strings.Builder
	seq := 0
	if !s.cfg.TTS.StreamEnabled {
		return "", StageError{Stage: "tts", Detail: "未启用百度流式TTS（tts.streamEnabled=false）"}
	}
	ttsSession, sessionErr := s.CreateStreamTTSSession(ctx, meta, func(audio []byte, chunkMeta AudioMeta) error {
		seq++
		if onAudioChunk != nil {
			return onAudioChunk(audio, chunkMeta, seq)
		}
		return nil
	})
	if sessionErr != nil {
		return "", sessionErr
	}
	defer ttsSession.Close()
	flushSentence := func(text string) error {
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			return nil
		}
		return ttsSession.WriteText(trimmed)
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		event := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if event == "" {
			continue
		}
		if event == "[DONE]" {
			break
		}
		chunk, chunkErr := extractChatStreamChunk(event)
		if chunkErr != nil {
			return "", chunkErr
		}
		chunk = sanitizeModelReplyText(chunk)
		if chunk == "" {
			continue
		}
		fullReply.WriteString(chunk)
		if onTextDelta != nil {
			if cbErr := onTextDelta(chunk); cbErr != nil {
				return "", cbErr
			}
		}
		sentenceBuf.WriteString(chunk)
		parts := splitBySentence(sentenceBuf.String())
		if len(parts) == 0 {
			continue
		}
		for i := 0; i < len(parts)-1; i++ {
			if err := flushSentence(parts[i]); err != nil {
				return "", err
			}
		}
		sentenceBuf.Reset()
		sentenceBuf.WriteString(parts[len(parts)-1])
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if err := flushSentence(sentenceBuf.String()); err != nil {
		return "", err
	}
	if err := ttsSession.Finish(ctx); err != nil {
		return "", err
	}
	reply := sanitizeModelReplyText(fullReply.String())
	if reply == "" {
		reply = "我在，请继续说。"
	}
	return reply, nil
}

func splitBySentence(input string) []string {
	text := strings.TrimSpace(input)
	if text == "" {
		return nil
	}
	cutRunes := map[rune]bool{'。': true, '！': true, '？': true, '.': true, '!': true, '?': true, '\n': true}
	var out []string
	var buf []rune
	for _, r := range []rune(text) {
		buf = append(buf, r)
		if cutRunes[r] {
			out = append(out, strings.TrimSpace(string(buf)))
			buf = buf[:0]
		}
	}
	out = append(out, strings.TrimSpace(string(buf)))
	return out
}

func (s *VoiceService) callDeepSeekHistoryReply(ctx context.Context, deviceNo, transcript string, hours int) (string, error) {
	if hours <= 0 {
		hours = 12
	}
	history, err := s.buildRecentHistory(ctx, deviceNo, hours)
	if err != nil {
		return "", err
	}
	historyBytes, _ := json.Marshal(history)
	prompt := fmt.Sprintf("用户输入=%s。请基于最近%d小时记录回答。记录=%s。仅输出JSON：{\"reply\":\"回复内容\"}。", transcript, hours, string(historyBytes))
	systemMessage := "您是育儿助手，主要帮助家长根据历史事件回应用户提问。"
	raw, reply, _, callErr := s.callDeepSeekRaw(ctx, deviceNo, prompt, 5, systemMessage)
	if callErr != nil {
		return "", callErr
	}
	if parsed, ok := parseGeneralChatResult(raw); ok && strings.TrimSpace(parsed.Reply) != "" {
		if s.cfg.DebugLog {
			parsedJSON, _ := json.Marshal(parsed)
			glog.Infof(ctx, "[思考过程] 历史问答结构化解析成功。deviceNo=%s parsed=%s", deviceNo, string(parsedJSON))
		}
		return parsed.Reply, nil
	}
	if s.cfg.DebugLog {
		glog.Warningf(ctx, "[思考过程] 历史问答结构化解析失败，使用文本回复。deviceNo=%s raw=%s", deviceNo, truncateVoiceLogText(raw, 800))
	}
	return strings.TrimSpace(reply), nil
}

// pickGrowthSuggestionDisplayText 成长建议：优先取模型 JSON 内的 reply 字段，避免把整段外层 JSON 写入数据库。
func pickGrowthSuggestionDisplayText(rawContent, replyNormalized string) string {
	if t := extractReplyFieldFromJSONText(rawContent); t != "" {
		return t
	}
	if t := extractReplyFieldFromJSONText(replyNormalized); t != "" {
		return t
	}
	rn := strings.TrimSpace(replyNormalized)
	rc := strings.TrimSpace(rawContent)
	if rn != "" && !strings.HasPrefix(rn, "{") {
		return rn
	}
	if rc != "" && !strings.HasPrefix(rc, "{") {
		return rc
	}
	if rn != "" {
		return rn
	}
	return rc
}

// extractChatReply 从 DeepSeek 风格返回中提取回复文本和退出标记。
func extractChatReply(body []byte) (string, bool, error) {
	_, reply, exit, err := extractChatReplyRaw(body)
	return reply, exit, err
}

// extractChatReplyRaw 返回原始 content 与标准化 reply（避免上层丢失 JSON 结构）。
func extractChatReplyRaw(body []byte) (rawContent string, reply string, exit bool, err error) {
	rawContent, exit, err = extractChatRawContent(body)
	if err != nil || exit {
		return rawContent, "", exit, err
	}
	reply, modelExit := normalizeReplyAndDetectExit(rawContent)
	return rawContent, reply, modelExit, nil
}
