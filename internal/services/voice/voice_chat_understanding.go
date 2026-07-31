package voice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"hello/internal/model/entity"
	"hello/internal/services/aimodel"
	contracts "hello/internal/services/contracts"
	"hello/internal/services/device"

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
	Events         []IntentEvent `json:"events"`          // 多事件列表（当 action 为 multi 时使用）
}

// historyRowEventName 写入 history.event_name：始终用事件主档标准名；displayHint 为模型/用户说法或 extra 命中词，仅作主档名为空时的回退。
// mergeDeepSeekEventTypeJSON 从模型 JSON 读取 event_type（蛇形字段），写入 entity.Event。
func mergeDeepSeekEventTypeJSON(trimmed string, ev *entity.Event) {
	if ev == nil || trimmed == "" {
		return
	}
	var aux struct {
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal([]byte(trimmed), &aux); err == nil && strings.TrimSpace(aux.EventType) != "" {
		ev.EventType = aux.EventType
	}
}

func mergeDeepSeekEventUnitJSON(trimmed string, ev *entity.Event) {
	if ev == nil || trimmed == "" {
		return
	}
	var aux struct {
		EventUnit string `json:"event_unit"`
	}
	if err := json.Unmarshal([]byte(trimmed), &aux); err == nil {
		ev.Unit = strings.TrimSpace(aux.EventUnit)
	}
}

func historyRowEventName(ev entity.Event, displayHint string) string {
	if n := strings.TrimSpace(ev.Name); n != "" {
		return n
	}
	return strings.TrimSpace(displayHint)
}

// historyRowEventUnit 写入 history.event_unit 时使用事件主档单位。
func historyRowEventUnit(ev entity.Event) string {
	return strings.TrimSpace(ev.Unit)
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
	Events               []entity.Event
}

// prepareChatPreamble 对话前置编排（流式/非流式入口共享）：
// 1) 文本规范化与长度校验
// 2) pending child：仅在直接子节点中匹配并落库（Go 本域事件树逻辑）
// 澄清续聊不再在此短路：conversation_id 由统一 AnalyzeIntent / Stream 携带，自由文本由 Python 解析。
// Args:
//   - deviceNo: 设备号（pending child 按设备隔离）
//   - transcript: 用户本轮原文/转写
//
// Returns: Continue=false 时 Result/Err 可直接返回；Continue=true 时携带 NormalizedTranscript 与事件主档。
// Side Effects: 可能继续子事件落库、insertQa。
func (s *VoiceService) prepareChatPreamble(ctx context.Context, deviceNo, transcript string) chatPreambleOutcome {
	normalizedTranscript, err := s.normalizeAndValidateChatText(transcript)
	if err != nil {
		return chatPreambleOutcome{Err: err}
	}

	events := []entity.Event{}
	events, _ = DeviceAdmin().ListEvents(ctx)

	// 待选子事件：上一轮命中非叶子父节点后，本轮仅在直接子节点中匹配并落库
	if pending, ok := s.getPendingChildEvent(deviceNo); ok {
		reply, exit, finishTalk, pendErr := s.continuePendingChildEvent(ctx, deviceNo, normalizedTranscript, events, pending)
		if pendErr != nil {
			return chatPreambleOutcome{
				Result: chatResult{
					Reply:      "我暂时没理解清楚，请再说一次",
					Ask:        normalizedTranscript,
					FinishTalk: false,
				},
				Err: pendErr,
			}
		}
		s.insertQa(ctx, normalizedTranscript, reply)
		return chatPreambleOutcome{
			Result: chatResult{
				Reply:      reply,
				Ask:        normalizedTranscript,
				Exit:       exit,
				FinishTalk: finishTalk,
			},
		}
	}

	return chatPreambleOutcome{
		Continue:             true,
		NormalizedTranscript: normalizedTranscript,
		Events:               events,
	}
}

// pythonIntentLandPlan 按 Python IntentResponse 两轴映射后的落地计划。
// 权威：target_type=feeding|history|suggest|conversation|exit；
// 喂养 action=start|end|one|multi|disambiguate（与兄弟仓 schema 对齐）。
type pythonIntentLandPlan struct {
	ReplyOnly     bool          // true：仅自然语言，不进入 handleUnifiedIntentAction / AddHistory
	Action        entity.Action // ReplyOnly=false 时传给动作执行器
	UnknownDomain bool          // target_type 未识别，调用方应打 Warning
}

// mapPythonIntentToLandPlan 将 Python 意图两轴映射为 Go 落地计划。
// 业务说明：
//   - 领域看 target_type，喂养动作看 action；禁止用 ParseActionTargetType(target_type) 驱动 CRUD
//   - history→search、suggest→suggest、exit→exit；feeding 的 start|end|one|multi 进动作器
//   - conversation/空/未知领域、feeding+disambiguate/未知喂养 action → 仅回复
//
// Args:
//   - intent: 已映射的统一意图
//   - normalizedTranscript: 已规范化用户文本（Action.Name 最终回退）
//
// Returns: 落地计划（ReplyOnly 或可执行 Action）
func mapPythonIntentToLandPlan(intent deepSeekUnifiedIntent, normalizedTranscript string) pythonIntentLandPlan {
	tt := strings.TrimSpace(strings.ToLower(intent.TargetType))
	act := strings.TrimSpace(strings.ToLower(intent.Action))

	// Action.Name：展示/日志用，与 CRUD 开关（TargetType）分离
	name := strings.TrimSpace(intent.ActionName)
	if name == "" {
		name = strings.TrimSpace(intent.Action)
	}
	if name == "" {
		name = strings.TrimSpace(intent.EventName)
	}
	if name == "" {
		name = normalizedTranscript
	}

	switch tt {
	case "feeding":
		switch act {
		case "start", "end", "one":
			// 喂养 CRUD：TargetType 取 action，而非 target_type
			return pythonIntentLandPlan{
				Action: entity.Action{
					Name:       name,
					TargetType: ParseActionTargetType(act).String(),
				},
			}
		case "multi":
			// 多事件：handleUnifiedIntentAction 内按 intent.Action==multi 分流；TargetType 占位即可
			return pythonIntentLandPlan{
				Action: entity.Action{
					Name:       name,
					TargetType: ActionTargetTypeOne.String(),
				},
			}
		default:
			return pythonIntentLandPlan{ReplyOnly: true}
		}
	case "history":
		return pythonIntentLandPlan{
			Action: entity.Action{
				Name:       name,
				TargetType: ActionTargetTypeSearch.String(),
			},
		}
	case "suggest":
		// 领域以 target_type=suggest 为准；不依赖 Python action=suggestion 字符串
		return pythonIntentLandPlan{
			Action: entity.Action{
				Name:       name,
				TargetType: ActionTargetTypeSuggest.String(),
			},
		}
	case "exit":
		return pythonIntentLandPlan{
			Action: entity.Action{
				Name:       name,
				TargetType: ActionTargetTypeExit.String(),
			},
		}
	case "conversation", "":
		return pythonIntentLandPlan{ReplyOnly: true}
	default:
		return pythonIntentLandPlan{ReplyOnly: true, UnknownDomain: true}
	}
}

// applyUnifiedIntentResult 将已映射的统一意图落地为澄清 / 闲聊 / 动作执行。
// 供非流式 chatWithResult 与流式 HandleTranscriptForIntentStream 共用，保证行为矩阵单一真源。
// 业务分支：NeedConfirm → 仅存 conversation_id 并透传自然语言；否则清 cid → 按 Python 两轴映射再 CRUD 或仅回复。
// Args:
//   - normalizedTranscript: 已规范化用户文本
//   - events: 设备事件主档（经 DeviceAdmin）
//   - intent: 已由 mapPythonRespToIntent 等映射完成的结构
//
// Returns: chatResult 与业务错误
// Side Effects: 可能 set/clear pendingConfirm cid、insertQa、经 handleUnifiedIntentAction 写事件
func (s *VoiceService) applyUnifiedIntentResult(ctx context.Context, deviceNo, normalizedTranscript string, events []entity.Event, intent deepSeekUnifiedIntent) (chatResult, error) {
	// NeedConfirm：只保存 conversation_id，透传 Python 自然语言，不落库
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
		glog.Infof(ctx, "[Clarify] need_confirm，透传话术并等待续聊。deviceNo=%s conversation_id=%s", deviceNo, intent.ConversationID)
		s.insertQa(ctx, normalizedTranscript, reply)
		return chatResult{
			Reply:      reply,
			Ask:        normalizedTranscript,
			FinishTalk: false,
		}, nil
	}

	// 非确认结果：清除本地 cid，再按 Python 两轴落地
	pendingConfirmState.clear(deviceNo)

	plan := mapPythonIntentToLandPlan(intent, normalizedTranscript)
	if plan.UnknownDomain {
		glog.Warningf(ctx, "[IntentLand] 未知 target_type，按 conversation 仅回复。deviceNo=%s target_type=%s action=%s",
			deviceNo, intent.TargetType, intent.Action)
	}
	if plan.ReplyOnly {
		reply := strings.TrimSpace(intent.Reply)
		if reply == "" {
			reply = "我明白了，请再具体一点，我马上帮你处理。"
		}
		s.insertQa(ctx, normalizedTranscript, reply)
		return chatResult{
			Reply:      reply,
			Ask:        normalizedTranscript,
			FinishTalk: !intent.NeedUserReply,
		}, nil
	}

	glog.Infof(ctx, "[IntentLand] Python 两轴映射命中动作。deviceNo=%s target_type=%s action=%s go_target=%s name=%s",
		deviceNo, intent.TargetType, intent.Action, plan.Action.TargetType, plan.Action.Name)
	finalReply, exit, finishTalk, err := s.handleUnifiedIntentAction(ctx, deviceNo, normalizedTranscript, plan.Action, events, intent)
	if err == nil {
		s.insertQa(ctx, normalizedTranscript, finalReply)
	}
	return chatResult{
		Reply:      finalReply,
		Ask:        normalizedTranscript,
		Exit:       exit,
		FinishTalk: finishTalk,
	}, err
}

// chatWithResult 对话核心流程（非流式意图分析入口）：
// - 前置：pending child（prepareChatPreamble）；澄清 cid 由本函数内 AnalyzeIntent 携带
// - 常规：非流式 callDeepSeekUnifiedIntent → applyUnifiedIntentResult 落库/回复
// - 额度：发起前有有效澄清 cid 则免计（跳过 guard/consume）；冷启动含首次 need_confirm 仍计次
// TTS / HandleTranscriptForStreaming 等仍走本函数；流式入口禁止再调用本函数以免二次 AnalyzeIntent。
func (s *VoiceService) chatWithResult(ctx context.Context, deviceNo, transcript string) (chatResult, error) {
	pre := s.prepareChatPreamble(ctx, deviceNo, transcript)
	if !pre.Continue {
		return pre.Result, pre.Err
	}
	normalizedTranscript := pre.NormalizedTranscript
	events := pre.Events

	glog.Infof(ctx, "问题=%q", normalizedTranscript)

	// 澄清续聊判定须在意图调用前读取本地 cid（成功落地后可能 clear）
	clarifyResume := pendingConversationID(deviceNo) != ""
	var wxID int64
	if clarifyResume {
		// 免计：额度用尽时仍可完成澄清，避免半途卡死
		glog.Infof(ctx, "[Clarify] 续聊免计 AI 额度。deviceNo=%s conversation_id=%s", deviceNo, pendingConversationID(deviceNo))
	} else {
		var qErr error
		wxID, ctx, qErr = s.guardVoiceAIQuota(ctx, deviceNo)
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
	// 仅冷启动成功计次；澄清续聊不 consume
	if !clarifyResume {
		s.consumeVoiceAIQuotaOnSuccess(ctx, wxID)
	}

	return s.applyUnifiedIntentResult(ctx, deviceNo, normalizedTranscript, events, intent)
}

func (s *VoiceService) callDeepSeekUnifiedIntent(ctx context.Context, deviceNo, transcript string) (deepSeekUnifiedIntent, error) {
	// 调用 Python 微服务进行意图分析；若有本地澄清 cid 则写入请求以续聊（调用失败不 clear cid）。
	if vuProfile, vuErr := aimodel.LoadProfile(ctx, aimodel.LaneVoiceUnderstanding); vuErr == nil {
		pythonClient := PythonAIClientFromCfg()
		req := &AnalyzeIntentRequest{
			Text:     transcript,
			DeviceNo: deviceNo,
			Model: PythonModelCfg{
				Provider:    string(vuProfile.Provider),
				Name:        vuProfile.Model,
				MaxInFlight: vuProfile.MaxInFlight,
			},
			ConversationID: pendingConversationID(deviceNo),
		}
		pythonResp, pythonErr := pythonClient.AnalyzeIntent(ctx, req)
		if pythonErr == nil && pythonResp != nil {
			intent := mapPythonRespToIntent(pythonResp)
			glog.Debugf(ctx, "[Python AI] 意图分析成功。deviceNo=%s target_type=%s action=%s need_confirm=%v conversation_id=%s",
				deviceNo, intent.TargetType, intent.Action, intent.NeedConfirm, intent.ConversationID)
			return intent, nil
		}
		if pythonErr != nil {
			glog.Warningf(ctx, "[Python AI] 意图分析调用失败。deviceNo=%s err=%v", deviceNo, pythonErr)
			return deepSeekUnifiedIntent{}, pythonErr
		}
	}
	return deepSeekUnifiedIntent{}, errors.New("意图分析配置缺失")
}

// mapPythonRespToIntent 将 Python 侧 AnalyzeIntentResponse 映射为 deepSeekUnifiedIntent
// 供非流式和流式意图分析复用，避免重复代码
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

// callDeepSeekUnifiedIntentStream 调用流式 Python 意图分析接口，同时将结果通过回调推送给调用方。
// 与非流式一致：本地有澄清 cid 时写入请求；调用失败不 clear cid。
func (s *VoiceService) callDeepSeekUnifiedIntentStream(ctx context.Context, deviceNo, transcript string, cb *contracts.IntentStreamCallback) (*AnalyzeIntentStreamResponse, error) {
	if vuProfile, vuErr := aimodel.LoadProfile(ctx, aimodel.LaneVoiceUnderstanding); vuErr == nil {
		pythonClient := PythonAIClientFromCfg()
		// 构造流式回调：将 Python 侧的流式事件转发给上层调用方
		streamCb := &AnalyzeIntentStreamCallback{}
		if cb != nil {
			if cb.OnThinking != nil {
				streamCb.OnThinking = func(delta string) error {
					return cb.OnThinking(delta)
				}
			}
			if cb.OnAnswer != nil {
				streamCb.OnAnswer = func(delta string) error {
					return cb.OnAnswer(delta)
				}
			}
		}
		// 调用流式 Python 接口（可附带 conversation_id 续聊）
		streamRes, streamErr := pythonClient.AnalyzeIntentStream(ctx, &AnalyzeIntentStreamRequest{
			Text:     transcript,
			DeviceNo: deviceNo,
			Model: PythonModelCfg{
				Provider:    string(vuProfile.Provider),
				Name:        vuProfile.Model,
				MaxInFlight: vuProfile.MaxInFlight,
			},
			ConversationID: pendingConversationID(deviceNo),
		}, streamCb)
		if streamErr != nil {
			glog.Warningf(ctx, "[Python AI] 流式意图分析调用失败。deviceNo=%s err=%v", deviceNo, streamErr)
			return nil, streamErr
		}
		// Result 可能为空：上层流式落地路径须降级，禁止回退非流式 AnalyzeIntent
		if streamRes != nil && streamRes.Result != nil {
			glog.Debugf(ctx, "[Python AI] 流式意图分析成功。deviceNo=%s target_type=%s action=%s need_confirm=%v",
				deviceNo, streamRes.Result.TargetType, streamRes.Result.Action, streamRes.Result.NeedConfirm)
		} else {
			glog.Warningf(ctx, "[Python AI] 流式意图分析返回空 Result。deviceNo=%s", deviceNo)
		}
		return streamRes, nil
	}
	return nil, errors.New("意图分析配置缺失")
}

func (s *VoiceService) handleUnifiedIntentAction(ctx context.Context, deviceNo, normalizedTranscript string, action entity.Action, events []entity.Event, intent deepSeekUnifiedIntent) (finalReply string, exit bool, finishTalk bool, err error) {
	// 统一意图动作执行器：对 suggest/search/exit 直接返回，其余动作落到事件写库流程。

	// 多事件场景：遍历 Events 列表，逐个处理每个事件
	if intent.Action == "multi" && len(intent.Events) > 0 {
		return s.handleMultiEventIntent(ctx, deviceNo, normalizedTranscript, events, intent)
	}

	nowTime := time.Now().Unix()
	switch action.TargetType {
	case ActionTargetTypeSuggest.String():
		reply := strings.TrimSpace(intent.Reply)
		if reply == "" {
			reply = "今天建议保持规律作息，按需喂养，关注精神状态与排便情况。"
		}
		return strings.TrimSpace(reply), false, true, nil
	case ActionTargetTypeSearch.String():
		reply := strings.TrimSpace(intent.Reply)
		if reply == "" {
			reply = "我暂时没找到明确的历史记录结论，请你换个问法再试试。"
		}
		return strings.TrimSpace(reply), false, true, nil
	case ActionTargetTypeExit.String():
		return "好的，再见", true, false, nil
	}

	event, targetName, pendingReply, ok, err := s.resolveEventForAction(ctx, deviceNo, normalizedTranscript, events, action.TargetType, &intent)
	if err != nil {
		return "我听不懂你说的事件,请用具体的名称告诉我", false, false, err
	}
	if pendingReply != "" {
		return pendingReply, false, false, nil
	}
	if !ok {
		return "我听不懂你说的事件,请用具体的名称告诉我", false, false, errors.New("未识别事件")
	}

	switch action.TargetType {
	case ActionTargetTypeStart.String():
		_, err = DeviceHistory().AddHistory(ctx, entity.History{
			DeviceNo:  deviceNo,
			EventId:   event.Id,
			EventName: historyRowEventName(event, targetName),
			EventUnit: historyRowEventUnit(event),
			StartTime: nowTime,
			Remark:    normalizedTranscript,
		})
		if err != nil {
			return "记录失败,请重试", false, true, err
		}
		return fmt.Sprintf("好的，已记录%s开始", targetName), false, true, nil
	case ActionTargetTypeEnd.String():
		reply, err := applyVoiceEventEndHistory(ctx, deviceNo, event, targetName, normalizedTranscript, nowTime)
		return reply, false, true, err
	case ActionTargetTypeOne.String():
		quantity := intent.Quantity
		if strings.EqualFold(strings.TrimSpace(intent.EventType), device.EventTypeNumber) && quantity <= 0 {
			return "请问 " + action.Name + " " + targetName + " 的数量是" + quantityKeyword, false, false, nil
		}
		eventNumber := int64(1)
		if quantity > 0 {
			eventNumber = int64(quantity)
		}
		_, err = DeviceHistory().AddHistory(ctx, entity.History{
			DeviceNo:    deviceNo,
			EventId:     event.Id,
			EventName:   historyRowEventName(event, targetName),
			EventUnit:   historyRowEventUnit(event),
			EventNumber: eventNumber,
			StartTime:   nowTime,
			EndTime:     nowTime,
			Remark:      normalizedTranscript,
		})
		if err != nil {
			return "记录事件失败,请重试", false, true, err
		}
		if quantity > 0 {
			return fmt.Sprintf("好的，已记录 %s %d", targetName, quantity), false, true, nil
		}
		return fmt.Sprintf("好的，已记录 %s", targetName), false, true, nil
	default:
		return "我没有理解你的意思", false, false, nil
	}
}

// handleMultiEventIntent 处理多事件场景
// 遍历 intent.Events，逐个匹配事件并落库
func (s *VoiceService) handleMultiEventIntent(ctx context.Context, deviceNo, normalizedTranscript string, events []entity.Event, intent deepSeekUnifiedIntent) (finalReply string, exit bool, finishTalk bool, err error) {
	nowTime := time.Now().Unix()
	var replies []string

	for _, eventItem := range intent.Events {
		action := entity.Action{
			Name:       eventItem.Action,
			TargetType: ParseActionTargetType(eventItem.Action).String(),
		}
		if action.Name == "" {
			action.Name = eventItem.Action
		}

		// 根据动作类型处理每个事件
		switch eventItem.Action {
		case ActionTargetTypeStart.String():
			// 查找匹配的事件
			event, targetName, pendingReply, ok, err := s.resolveEventForAction(ctx, deviceNo, normalizedTranscript, events, action.TargetType, nil)
			if err != nil || pendingReply != "" || !ok {
				continue
			}
			_, err = DeviceHistory().AddHistory(ctx, entity.History{
				DeviceNo:  deviceNo,
				EventId:   event.Id,
				EventName: historyRowEventName(event, targetName),
				EventUnit: historyRowEventUnit(event),
				StartTime: nowTime,
				Remark:    normalizedTranscript,
			})
			if err == nil {
				replies = append(replies, fmt.Sprintf("已记录%s开始", targetName))
			}
		case ActionTargetTypeEnd.String():
			event, targetName, pendingReply, ok, err := s.resolveEventForAction(ctx, deviceNo, normalizedTranscript, events, action.TargetType, nil)
			if err != nil || pendingReply != "" || !ok {
				continue
			}
			reply, err := applyVoiceEventEndHistory(ctx, deviceNo, event, targetName, normalizedTranscript, nowTime)
			if err == nil && reply != "" {
				replies = append(replies, reply)
			}
		case ActionTargetTypeOne.String():
			event, targetName, pendingReply, ok, err := s.resolveEventForAction(ctx, deviceNo, normalizedTranscript, events, action.TargetType, nil)
			if err != nil || pendingReply != "" || !ok {
				continue
			}
			_, err = DeviceHistory().AddHistory(ctx, entity.History{
				DeviceNo:    deviceNo,
				EventId:     event.Id,
				EventName:   historyRowEventName(event, targetName),
				EventUnit:   historyRowEventUnit(event),
				EventNumber: 1,
				StartTime:   nowTime,
				EndTime:     nowTime,
				Remark:      normalizedTranscript,
			})
			if err == nil {
				replies = append(replies, fmt.Sprintf("已记录%s", targetName))
			}
		}
	}

	if len(replies) > 0 {
		return fmt.Sprintf("好的，%s", strings.Join(replies, "，")), false, true, nil
	}
	return "我听不懂你说的事件,请用具体的名称告诉我", false, false, errors.New("未识别事件")
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
	// 兼容模型返回包裹格式：{"role":"assistant","content":"{...intent...}"}
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

// matchEventByName 使用宽松规则将模型输出事件名映射到库内事件。
func (s *VoiceService) matchEventByName(events []eventInfo, eventName string) (eventInfo, bool) {
	needle := strings.TrimSpace(strings.ToLower(eventName))
	if needle == "" {
		return eventInfo{}, false
	}
	for _, ev := range events {
		name := strings.TrimSpace(strings.ToLower(ev.Name))
		if name == needle || strings.Contains(name, needle) || strings.Contains(needle, name) {
			return ev, true
		}
	}
	return eventInfo{}, false
}

func (s *VoiceService) setPendingQuantity(deviceNo string, state pendingQuantityState) {
	// 记录“该设备下一轮需要补充数量”的状态。
	// 这个状态是内存态，服务重启后会丢失（符合当前设计预期）。
	s.pendingQuantityMu.Lock()
	defer s.pendingQuantityMu.Unlock()
	if strings.TrimSpace(deviceNo) == "" {
		return
	}
	s.pendingQuantity[deviceNo] = state
}

func (s *VoiceService) clearPendingQuantity(deviceNo string) {
	// 清理待补量词状态：
	// - 已成功补录数量后
	// - 或当前流程判断不再需要补录时
	s.pendingQuantityMu.Lock()
	defer s.pendingQuantityMu.Unlock()
	delete(s.pendingQuantity, deviceNo)
}

func (s *VoiceService) getPendingQuantity(deviceNo string) (pendingQuantityState, bool) {
	// 读取指定设备是否存在待补量词上下文。
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
