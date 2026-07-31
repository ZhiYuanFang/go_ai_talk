package voice

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/services/aimodel"

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

func suggestDurationMinutes(startSec, endSec int64) int {
	if startSec <= 0 || endSec <= 0 || endSec < startSec {
		return 0
	}
	return int(math.Round(float64(endSec-startSec) / 60))
}

// formatLocalDatetimeFromUnix 将 Unix 秒转为本地「YYYY-MM-DD HH:MM:SS」，供大模型上下文可读性。
func formatLocalDatetimeFromUnix(sec int64) string {
	if sec <= 0 {
		return ""
	}
	return time.Unix(sec, 0).In(time.Local).Format("2006-01-02 15:04:05")
}

// loadEventNameAndUnitByID 事件表 id -> 名称、单位（成长建议 history.type 用名称；amount 需带单位）。
func loadEventNameAndUnitByID(ctx context.Context) (names map[int64]string) {
	names = make(map[int64]string)
	events, err := DeviceAdmin().ListEvents(ctx)
	if err != nil || len(events) == 0 {
		return names
	}
	for _, e := range events {
		names[e.Id] = strings.TrimSpace(e.Name)
	}
	return names
}

// growthSuggestHistoryCutoff 成长建议只取最近 48 小时内的记录（滚动「两天」），Unix 秒下界。
func growthSuggestHistoryCutoff() int64 {
	return time.Now().Add(-48 * time.Hour).Unix()
}

func (s *VoiceService) buildGrowthSuggestHistory(ctx context.Context, deviceNo string) ([]map[string]interface{}, error) {
	cutoff := growthSuggestHistoryCutoff()
	rows, err := DeviceHistory().ListHistory(ctx, deviceNo)
	if err != nil {
		return nil, err
	}
	eventNames := loadEventNameAndUnitByID(ctx)
	out := make([]map[string]interface{}, 0, len(rows))
	for _, h := range rows {
		if h.StartTime < cutoff && h.EndTime < cutoff {
			continue
		}
		typeName := eventNames[h.EventId]
		if typeName == "" {
			typeName = strings.TrimSpace(h.EventName)
		}
		if typeName == "" {
			typeName = "未知事件"
		}
		note := strings.TrimSpace(h.Remark)
		amt := h.EventNumber
		if amt < 0 {
			amt = 0
		}
		start := formatLocalDatetimeFromUnix(h.StartTime)
		end := formatLocalDatetimeFromUnix(h.EndTime)
		item := map[string]interface{}{
			"type":         typeName,
			"start_time":   start,
			"end_time":     end,
			"amount_value": amt,
			"note":         note,
		}
		if dm := suggestDurationMinutes(h.StartTime, h.EndTime); dm > 0 {
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
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour).Unix()
	rows, err := DeviceHistory().ListHistory(ctx, deviceNo)
	if err != nil {
		glog.Warningf(ctx, "[上下文装配] 历史读取失败，触发降级。deviceNo=%s hours=%d err=%v", deviceNo, hours, err)
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(rows))
	skipped := 0
	cutoffAt := time.Unix(cutoff, 0).In(time.Local)
	for _, h := range rows {
		if h.StartTime < cutoff && h.EndTime < cutoff {
			skipped++
			continue
		}
		start := formatLocalDatetimeFromUnix(h.StartTime)
		end := formatLocalDatetimeFromUnix(h.EndTime)
		out = append(out, map[string]interface{}{
			"event_name":   strings.TrimSpace(h.EventName),
			"start_time":   start,
			"end_time":     end,
			"event_number": h.EventNumber,
			"remark":       strings.TrimSpace(h.Remark),
		})
	}
	// 记录上下文读取命中/回源后的窗口统计，便于定位历史窗口偏差问题。
	glog.Debugf(ctx, "[上下文装配] 历史窗口统计。deviceNo=%s hours=%d total=%d kept=%d skipped=%d cutoff_unix=%d", deviceNo, hours, len(rows), len(out), skipped, cutoff)
	for _, item := range out {
		start := strings.TrimSpace(g.NewVar(item["start_time"]).String())
		end := strings.TrimSpace(g.NewVar(item["end_time"]).String())
		startAt, sErr := time.ParseInLocation("2006-01-02 15:04:05", start, time.Local)
		endAt, eErr := time.ParseInLocation("2006-01-02 15:04:05", end, time.Local)
		if sErr != nil || eErr != nil {
			continue
		}
		if startAt.Before(cutoffAt) && endAt.Before(cutoffAt) {
			glog.Warningf(ctx, "[上下文装配] 历史窗口校验异常。deviceNo=%s cutoff_unix=%d start=%s end=%s", deviceNo, cutoff, start, end)
		}
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

// callDeepSeekGrowthSuggestion 成长建议：经 LaneVoiceUnderstanding 调用上游。
// 独立 AnalyzeIntent，不得附带喂养澄清 conversation_id（不调用 pendingConversationID）。
// 若外层已标记 voice_ai degraded，则强制种子智谱。
func (s *VoiceService) callDeepSeekGrowthSuggestion(ctx context.Context, deviceNo string) (string, error) {
	// 调用 Python 微服务获取成长建议，Python 不可用时直接返回错误，由上层返回降级提示语。
	if vuProfile, vuErr := loadVoiceUnderstandingProfile(ctx); vuErr == nil {
		pythonClient := PythonAIClientFromCfg()
		pythonResp, pythonErr := pythonClient.AnalyzeIntent(ctx, &AnalyzeIntentRequest{
			Text:     "成长建议",
			DeviceNo: deviceNo,
			Model: PythonModelCfg{
				Provider:    string(vuProfile.Provider),
				Name:        vuProfile.Model,
				MaxInFlight: vuProfile.MaxInFlight,
			},
		})
		if pythonErr == nil && pythonResp != nil && pythonResp.Content != "" {
			reply := strings.TrimSpace(pythonResp.Content)
			if reply != "" {
				_, insertErr := dao.Suggest.Ctx(ctx).Data(g.Map{
					dao.Suggest.Columns().DeviceNo: deviceNo,
					dao.Suggest.Columns().Suggest:  reply,
					dao.Suggest.Columns().Time:     nowUnixSec(),
				}).Insert()
				if insertErr != nil {
					glog.Warningf(ctx, "insert suggest failed: %v", insertErr)
				}
				return reply, nil
			}
		}
		if pythonErr != nil {
			glog.Warningf(ctx, "[Python AI] 成长建议调用失败。deviceNo=%s err=%v", deviceNo, pythonErr)
			return "", pythonErr
		}
	}
	return "", errors.New("成长建议配置缺失")
}

func mapVoiceChatMessages(messages []map[string]string) []aimodel.Message {
	out := make([]aimodel.Message, 0, len(messages))
	for _, m := range messages {
		out = append(out, aimodel.Message{Role: m["role"], Content: m["content"]})
	}
	return out
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
	// 历史问答：独立 AnalyzeIntent，不得附带喂养澄清 conversation_id。
	// 调用 Python 微服务进行历史问答；degraded 时强制种子智谱。
	if vuProfile, vuErr := loadVoiceUnderstandingProfile(ctx); vuErr == nil {
		pythonClient := PythonAIClientFromCfg()
		pythonResp, pythonErr := pythonClient.AnalyzeIntent(ctx, &AnalyzeIntentRequest{
			Text:     transcript,
			DeviceNo: deviceNo,
			Model: PythonModelCfg{
				Provider:    string(vuProfile.Provider),
				Name:        vuProfile.Model,
				MaxInFlight: vuProfile.MaxInFlight,
			},
		})
		if pythonErr == nil && pythonResp != nil && pythonResp.Content != "" {
			reply := strings.TrimSpace(pythonResp.Content)
			if reply != "" {
				return reply, nil
			}
		}
		if pythonErr != nil {
			glog.Warningf(ctx, "[Python AI] 历史问答调用失败。deviceNo=%s err=%v", deviceNo, pythonErr)
			return "", pythonErr
		}
	}
	return "", errors.New("历史问答配置缺失")
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
