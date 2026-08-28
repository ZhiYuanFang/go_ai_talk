package voice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"hello/internal/services/aimodel"
	contracts "hello/internal/services/contracts"

	"github.com/gogf/gf/v2/os/glog"
)

type deepSeekUnifiedIntent struct {
	Action         string        `json:"action"`
	ActionName     string        `json:"action_name"`
	EventName      string        `json:"event_name"`
	ExtraEvent     string        `json:"extra_event_name"`
	EventId        string        `json:"event_id"` // 事件ID（Python 向量匹配返回）
	EventType      string        `json:"event_type"`
	EventUnit      string        `json:"event_unit"`
	Quantity       int           `json:"quantity"`
	IsNewEvent     bool          `json:"is_new_event"` // 是否为新事件（Python 返回）
	Reply          string        `json:"reply"`
	NeedUserReply  bool          `json:"need_user_reply"`
	TargetType     string        `json:"target_type"`
	NeedConfirm    bool          `json:"need_confirm"`    // 是否需要用户澄清（同一 /intent + conversation_id 续聊）
	ConfirmMessage string        `json:"confirm_message"` // 澄清话术（Python 生成，Go 原样透传）
	ConversationID string        `json:"conversation_id"` // 会话 ID；need_confirm 时保存，下一轮 intent 请求带回
	Events         []IntentEvent `json:"events"`          // 多事件列表（Python 返回；Go 不解析子项 op）
}

type chatResult struct {
	Reply      string
	Ask        string
	Exit       bool
	FinishTalk bool
}

// normalizeAndValidateChatText 统一清理并校验转写文本长度。
func (s *VoiceService) normalizeAndValidateChatText(text string) (string, error) {
	normalized := strings.TrimSpace(text)
	if normalized == "" {
		return "", StageError{Stage: "chat", Detail: "文本为空，无法进行聊天"}
	}

	if utf8.RuneCountInString(normalized) < s.cfg.DeepSeek.MinTextLength {
		return "", StageError{Stage: "chat", Detail: fmt.Sprintf("文本长度不能小于%d", s.cfg.DeepSeek.MinTextLength)}
	}

	return normalized, nil
}

// chatPreambleOutcome 对话前置处理结果。
// Continue=true 表示需继续发起意图分析；false 表示已完整处理（含校验失败），调用方应直接返回 Result/Err。
type chatPreambleOutcome struct {
	Continue             bool
	Result               chatResult
	Err                  error
	NormalizedTranscript string
}

// prepareChatPreamble 对话前置编排（流式/非流式入口共享）：
// 1) 文本规范化与长度校验
// 澄清续聊不在此短路：conversation_id 由统一 AnalyzeIntent / Stream 携带，自由文本与喂养落库均由 Python 意图图完成。
func (s *VoiceService) prepareChatPreamble(ctx context.Context, deviceNo, transcript string) chatPreambleOutcome {
	_ = ctx
	_ = deviceNo
	normalizedTranscript, err := s.normalizeAndValidateChatText(transcript)
	if err != nil {
		return chatPreambleOutcome{Err: err}
	}
	return chatPreambleOutcome{
		Continue:             true,
		NormalizedTranscript: normalizedTranscript,
	}
}

// applyUnifiedIntentResult 只透传 Python 结果：确认话术 / content / 退出。
// 喂养 history 写库由 Python 在意图分析阶段经 history event/batch 完成；Go 不再二次写库或匹配事件。
func (s *VoiceService) applyUnifiedIntentResult(ctx context.Context, deviceNo, normalizedTranscript string, intent deepSeekUnifiedIntent) (chatResult, error) {
	if intent.NeedConfirm {
		pendingConfirmState.set(deviceNo, &pendingConfirmEntry{
			ConversationID: intent.ConversationID,
			EventName:      intent.EventName,
			Action:         intent.Action,
			CreatedAt:      time.Now(),
		})
		reply := strings.TrimSpace(intent.ConfirmMessage)
		if reply == "" {
			reply = strings.TrimSpace(intent.Reply)
		}
		if reply == "" {
			reply = "请确认是否执行该操作？"
		}
		glog.Infof(ctx, "[Clarify] need_confirm，透传话术。deviceNo=%s conversation_id=%s", deviceNo, intent.ConversationID)
		s.insertQa(ctx, normalizedTranscript, reply)
		return chatResult{Reply: reply, Ask: normalizedTranscript, FinishTalk: false}, nil
	}

	pendingConfirmState.clear(deviceNo)
	if strings.EqualFold(strings.TrimSpace(intent.TargetType), "exit") || strings.EqualFold(strings.TrimSpace(intent.Action), "exit") {
		reply := strings.TrimSpace(intent.Reply)
		if reply == "" {
			reply = "好的，再见"
		}
		s.insertQa(ctx, normalizedTranscript, reply)
		return chatResult{Reply: reply, Ask: normalizedTranscript, Exit: true, FinishTalk: false}, nil
	}

	reply := strings.TrimSpace(intent.Reply)
	if reply == "" {
		reply = "我明白了，请再具体一点。"
	}
	s.insertQa(ctx, normalizedTranscript, reply)
	return chatResult{
		Reply:      reply,
		Ask:        normalizedTranscript,
		FinishTalk: !intent.NeedUserReply,
	}, nil
}

// chatWithResult 对话核心流程（非流式意图分析入口）：
// - 前置：prepareChatPreamble 文本校验；澄清 cid 由 AnalyzeIntent 携带
// - 常规：非流式 callDeepSeekUnifiedIntent → applyUnifiedIntentResult 透传回复（落库在 Python）
// - 额度：澄清续聊免计；冷启动额度内计次；用尽走智谱降速且不计次
func (s *VoiceService) chatWithResult(ctx context.Context, deviceNo, transcript string) (chatResult, error) {
	pre := s.prepareChatPreamble(ctx, deviceNo, transcript)
	if !pre.Continue {
		return pre.Result, pre.Err
	}
	normalizedTranscript := pre.NormalizedTranscript

	glog.Infof(ctx, "问题=%q", normalizedTranscript)

	clarifyResume := pendingConversationID(deviceNo) != ""
	var wxID int64
	var shouldConsume bool
	if clarifyResume {
		glog.Infof(ctx, "[Clarify] 续聊免计 AI 额度。deviceNo=%s conversation_id=%s", deviceNo, pendingConversationID(deviceNo))
	} else {
		var qErr error
		wxID, shouldConsume, ctx, qErr = s.guardVoiceAIQuota(ctx, deviceNo)
		if qErr != nil {
			return chatResult{}, qErr
		}
	}

	intent, err := s.callDeepSeekUnifiedIntent(ctx, deviceNo, normalizedTranscript)
	if err != nil {
		return chatResult{
			Reply:      "AI 服务暂时不可用，请稍后再试",
			Ask:        normalizedTranscript,
			Exit:       false,
			FinishTalk: false,
		}, err
	}
	if !clarifyResume && shouldConsume {
		s.consumeVoiceAIQuotaOnSuccess(ctx, wxID)
	}

	return s.applyUnifiedIntentResult(ctx, deviceNo, normalizedTranscript, intent)
}

func (s *VoiceService) callDeepSeekUnifiedIntent(ctx context.Context, deviceNo, transcript string) (deepSeekUnifiedIntent, error) {
	wxID := VoiceWxIDFromCtx(ctx)
	if wxID <= 0 {
		if id, e := VoiceWxIDFromRequest(ctx, deviceNo); e == nil {
			wxID = id
		}
	}
	ent, runtime, modelCfg, _ := resolveVoiceUnderstandingModel(ctx, wxID)
	if modelCfg != nil {
		rel, acqErr := aimodel.Acquire(ctx, runtime)
		if acqErr != nil {
			return deepSeekUnifiedIntent{}, acqErr
		}
		defer rel()
	}
	pythonClient := PythonAIClientFromCfg()
	req := &AnalyzeIntentRequest{
		Text:           transcript,
		DeviceNo:       deviceNo,
		Model:          modelCfg,
		ConversationID: pendingConversationID(deviceNo),
	}
	pythonResp, pythonErr := pythonClient.AnalyzeIntent(ctx, req)
	if pythonErr == nil && pythonResp != nil {
		intent := mapPythonRespToIntent(pythonResp)
		glog.Debugf(ctx, "[Python AI] 意图分析成功。deviceNo=%s target_type=%s action=%s need_confirm=%v premium=%v vip=%v",
			deviceNo, intent.TargetType, intent.Action, intent.NeedConfirm, ent.Premium, ent.VIP)
		return intent, nil
	}
	if pythonErr != nil {
		glog.Warningf(ctx, "[Python AI] 意图分析调用失败。deviceNo=%s err=%v", deviceNo, pythonErr)
		return deepSeekUnifiedIntent{}, pythonErr
	}
	return deepSeekUnifiedIntent{}, errors.New("意图分析无响应")
}

// mapPythonRespToIntent 将 Python 侧 AnalyzeIntentResponse 映射为 deepSeekUnifiedIntent。
func mapPythonRespToIntent(pythonResp *AnalyzeIntentResponse) deepSeekUnifiedIntent {
	intent := deepSeekUnifiedIntent{
		TargetType:     strings.TrimSpace(strings.ToLower(pythonResp.TargetType)),
		Action:         strings.TrimSpace(pythonResp.Action),
		EventName:      strings.TrimSpace(pythonResp.EventName),
		EventId:        strings.TrimSpace(pythonResp.EventId),
		EventType:      strings.TrimSpace(pythonResp.EventType),
		EventUnit:      strings.TrimSpace(pythonResp.EventUnit),
		IsNewEvent:     pythonResp.IsNewEvent,
		Reply:          sanitizeModelReplyText(pythonResp.Content),
		NeedUserReply:  true,
		NeedConfirm:    pythonResp.NeedConfirm,
		ConfirmMessage: strings.TrimSpace(pythonResp.ConfirmMessage),
		ConversationID: strings.TrimSpace(pythonResp.ConversationID),
		Events:         pythonResp.Events,
	}
	if pythonResp.Quantity != nil {
		intent.Quantity = *pythonResp.Quantity
	}
	if intent.TargetType == "conversation" || intent.TargetType == "" {
		intent.Reply = sanitizeModelReplyText(intent.Reply)
	}
	return intent
}

// callDeepSeekUnifiedIntentStream 调用流式 Python 意图分析接口。
func (s *VoiceService) callDeepSeekUnifiedIntentStream(ctx context.Context, deviceNo, transcript string, cb *contracts.IntentStreamCallback) (*AnalyzeIntentStreamResponse, error) {
	wxID := VoiceWxIDFromCtx(ctx)
	if wxID <= 0 {
		if id, e := VoiceWxIDFromRequest(ctx, deviceNo); e == nil {
			wxID = id
		}
	}
	ent, runtime, modelCfg, _ := resolveVoiceUnderstandingModel(ctx, wxID)
	if modelCfg != nil {
		rel, acqErr := aimodel.Acquire(ctx, runtime)
		if acqErr != nil {
			return nil, acqErr
		}
		defer rel()
	}
	pythonClient := PythonAIClientFromCfg()
	streamCb := &AnalyzeIntentStreamCallback{}
	if cb != nil && cb.OnThinking != nil {
		streamCb.OnThinking = func(delta string) error {
			return cb.OnThinking(delta)
		}
	}
	streamRes, streamErr := pythonClient.AnalyzeIntentStream(ctx, &AnalyzeIntentStreamRequest{
		Text:           transcript,
		DeviceNo:       deviceNo,
		Model:          modelCfg,
		ConversationID: pendingConversationID(deviceNo),
	}, streamCb)
	if streamErr != nil {
		glog.Warningf(ctx, "[Python AI] 流式意图分析调用失败。deviceNo=%s err=%v", deviceNo, streamErr)
		return nil, streamErr
	}
	if streamRes != nil && streamRes.Result != nil {
		glog.Debugf(ctx, "[Python AI] 流式意图分析成功。deviceNo=%s target_type=%s action=%s need_confirm=%v premium=%v",
			deviceNo, streamRes.Result.TargetType, streamRes.Result.Action, streamRes.Result.NeedConfirm, ent.Premium)
	} else {
		glog.Warningf(ctx, "[Python AI] 流式意图分析返回空 Result。deviceNo=%s", deviceNo)
	}
	return streamRes, nil
}

// parseEventIntentFromReply 从模型回复中提取结构化 JSON 意图。
func parseEventIntentFromReply(reply string) (eventIntentResult, bool) {
	intent := eventIntentResult{}
	trimmed := strings.TrimSpace(reply)
	if trimmed == "" {
		return intent, false
	}
	trimmed = normalizeIntentCandidateText(trimmed)
	if trimmed == "" {
		return intent, false
	}
	if err := json.Unmarshal([]byte(trimmed), &intent); err == nil {
		intent.Action = strings.ToLower(strings.TrimSpace(intent.Action))
		intent.EventName = strings.TrimSpace(intent.EventName)
		intent.ActionKeyWord = strings.TrimSpace(intent.ActionKeyWord)
		intent.Remark = strings.TrimSpace(intent.Remark)
		intent.Reply = sanitizeModelReplyText(intent.Reply)
		intent.Reason = strings.TrimSpace(intent.Reason)
		if intent.Action == "" {
			intent.Action = "none"
		}
		return intent, true
	}
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		segment := trimmed[start : end+1]
		if err := json.Unmarshal([]byte(segment), &intent); err == nil {
			intent.Action = strings.ToLower(strings.TrimSpace(intent.Action))
			intent.EventName = strings.TrimSpace(intent.EventName)
			intent.ActionKeyWord = strings.TrimSpace(intent.ActionKeyWord)
			intent.Remark = strings.TrimSpace(intent.Remark)
			intent.Reply = sanitizeModelReplyText(intent.Reply)
			intent.Reason = strings.TrimSpace(intent.Reason)
			if intent.Action == "" {
				intent.Action = "none"
			}
			return intent, true
		}
	}
	return eventIntentResult{}, false
}

func normalizeIntentCandidateText(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(strings.TrimSpace(trimmed), "```")
	}
	var wrapper map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &wrapper); err == nil {
		if content, ok := wrapper["content"].(string); ok {
			content = strings.TrimSpace(content)
			if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
				return content
			}
		}
	}
	return trimmed
}

func (s *VoiceService) setPendingQuantity(deviceNo string, state pendingQuantityState) {
	s.pendingQuantityMu.Lock()
	defer s.pendingQuantityMu.Unlock()
	if strings.TrimSpace(deviceNo) == "" {
		return
	}
	s.pendingQuantity[deviceNo] = state
}

func (s *VoiceService) clearPendingQuantity(deviceNo string) {
	s.pendingQuantityMu.Lock()
	defer s.pendingQuantityMu.Unlock()
	delete(s.pendingQuantity, deviceNo)
}

func (s *VoiceService) getPendingQuantity(deviceNo string) (pendingQuantityState, bool) {
	s.pendingQuantityMu.Lock()
	defer s.pendingQuantityMu.Unlock()
	state, ok := s.pendingQuantity[deviceNo]
	return state, ok
}

func (s *VoiceService) hasPendingQuantity(deviceNo string) bool {
	_, ok := s.getPendingQuantity(deviceNo)
	return ok
}

// appendRemark 合并备注内容，避免覆盖已有备注。
func appendRemark(existing, extra string) string {
	existing = strings.TrimSpace(existing)
	extra = strings.TrimSpace(extra)
	if existing == "" {
		return extra
	}
	if extra == "" {
		return existing
	}
	return existing + "；" + extra
}

// isGrowthSuggestionIntent 判断用户是否在询问成长建议。
func (s *VoiceService) isGrowthSuggestionIntent(text string) bool {
	keys := []string{"成长建议", "生长建议", "发育建议", "育儿建议", "喂养建议", "孩子建议", "建议"}
	for _, key := range keys {
		if strings.Contains(text, key) {
			return true
		}
	}
	return false
}
