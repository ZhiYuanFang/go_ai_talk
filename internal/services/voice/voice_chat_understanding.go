package voice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"hello/internal/model/entity"
	"hello/internal/services/device"

	"github.com/gogf/gf/v2/os/glog"
)

type deepSeekUnifiedIntent struct {
	Action        string `json:"action"`
	ActionName    string `json:"action_name"`
	EventName     string `json:"event_name"`
	ExtraEvent    string `json:"extra_event_name"`
	EventType     string `json:"event_type"`
	EventUnit     string `json:"event_unit"`
	Quantity      int    `json:"quantity"`
	Reply         string `json:"reply"`
	NeedUserReply bool   `json:"need_user_reply"`
	TargetType    string `json:"target_type"`
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
	Reply            string
	Ask              string
	Mode             string
	Exit             bool
	FinishTalk       bool
	NeedCasualStream bool
}

// detectChatModeByTranscript 用于「显式切换命令」场景下，从话术里推断目标模式。
// 命中「母婴」→ 母婴；否则→ 闲聊（例如「切换到闲聊」不含母婴关键词时仍应落到闲聊）。
func detectChatModeByTranscript(text string) string {
	if strings.Contains(strings.TrimSpace(text), "母婴") {
		return ChatModeMaternity.String()
	}
	return ChatModeCasual.String()
}

func isModeSwitchCommand(text string) bool {
	// 如果text包含切换到或者相近意思
	normalized := strings.TrimSpace(text)
	if normalized == "" {
		return false
	}
	commandKeywords := []string{
		"切换到",
		"改为",
		"变成",
		"设置为",
	}
	for _, kw := range commandKeywords {
		if strings.Contains(normalized, kw) {
			return true
		}
	}
	return false
}

func isModeQueryCommand(text string) bool {
	normalized := strings.TrimSpace(text)
	if normalized == "" {
		return false
	}
	queryKeywords := []string{
		"当前是什么模式",
		"现在是什么模式",
		"当前模式",
		"现在模式",
		"什么模式",
	}
	for _, kw := range queryKeywords {
		if strings.Contains(normalized, kw) {
			return true
		}
	}
	return false
}

func chatModeDisplayName(mode string) string {
	if ParseChatMode(mode) == ChatModeMaternity {
		return "母婴"
	}
	return "闲聊"
}

func (s *VoiceService) setDeviceChatMode(deviceNo, mode string) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return
	}
	s.modeMu.Lock()
	defer s.modeMu.Unlock()
	s.deviceModes[deviceNo] = mode
}

func (s *VoiceService) getDeviceChatMode(deviceNo string) string {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return ""
	}
	s.modeMu.Lock()
	defer s.modeMu.Unlock()
	return strings.TrimSpace(s.deviceModes[deviceNo])
}

func (s *VoiceService) resolveChatMode(deviceNo, text string) string {
	normalized := strings.TrimSpace(text)
	// 显式切换命令优先级最高，其次复用设备已有模式，最后默认母婴（新设备/无会话内记忆）。
	if isModeSwitchCommand(normalized) {
		return detectChatModeByTranscript(normalized)
	}
	if mode := s.getDeviceChatMode(deviceNo); mode != "" {
		return mode
	}
	// 新账号或本进程内尚未写入模式时默认走母婴喂养流程；闲聊需用户显式切换话术触发上一分支。
	return ChatModeMaternity.String()
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

// chatWithResult 对话核心流程：
// - 处理退出意图
// - 处理待补量词
// - 处理“成长建议”特殊意图
// - 调用 DeepSeek 进行结构化事件识别并落库
func (s *VoiceService) chatWithResult(ctx context.Context, deviceNo, transcript string, directCasual bool) (chatResult, error) {
	normalizedTranscript, err := s.normalizeAndValidateChatText(transcript)
	if err != nil {
		return chatResult{}, err
	}
	if isModeQueryCommand(normalizedTranscript) {
		currentMode := s.getDeviceChatMode(deviceNo)
		if currentMode == "" {
			currentMode = ChatModeMaternity.String()
		}
		reply := fmt.Sprintf("当前是%s模式", chatModeDisplayName(currentMode))
		s.insertQa(ctx, normalizedTranscript, reply)
		return chatResult{Reply: reply, Ask: normalizedTranscript, Mode: currentMode, FinishTalk: true}, nil
	}
	resolvedMode := s.resolveChatMode(deviceNo, normalizedTranscript)
	if isModeSwitchCommand(normalizedTranscript) {
		s.setDeviceChatMode(deviceNo, resolvedMode)
		reply := fmt.Sprintf("已切换到%s模式", chatModeDisplayName(resolvedMode))
		s.insertQa(ctx, normalizedTranscript, reply)
		return chatResult{Reply: reply, Ask: normalizedTranscript, Mode: resolvedMode, FinishTalk: true}, nil
	}
	if ParseChatMode(resolvedMode) == ChatModeCasual {
		if !directCasual {
			wxID, qCtx, qErr := s.guardVoiceAIQuota(ctx, deviceNo)
			if qErr != nil {
				return chatResult{}, qErr
			}
			ctx = qCtx
			_ = wxID // casual 流式成功后在 WS 层 consume
			return chatResult{
				Ask:              normalizedTranscript,
				Mode:             resolvedMode,
				FinishTalk:       true,
				NeedCasualStream: true,
			}, nil
		}
		wxID, qCtx, qErr := s.guardVoiceAIQuota(ctx, deviceNo)
		if qErr != nil {
			return chatResult{}, qErr
		}
		ctx = qCtx
		reply, chatErr := s.callDeepSeekDirectReply(ctx, deviceNo, normalizedTranscript)
		if chatErr != nil {
			return chatResult{
				Reply:      "我暂时没理解清楚你的意思，请再说一次",
				Ask:        normalizedTranscript,
				Mode:       resolvedMode,
				FinishTalk: false,
			}, chatErr
		}
		s.consumeVoiceAIQuotaOnSuccess(ctx, wxID)
		s.insertQa(ctx, normalizedTranscript, reply)
		return chatResult{
			Reply:      reply,
			Ask:        normalizedTranscript,
			Mode:       resolvedMode,
			FinishTalk: false,
		}, nil
	}
	events := []entity.Event{}
	events, _ = DeviceAdmin().ListEvents(ctx)

	// 待选子事件：上一轮命中非叶子父节点后，本轮仅在直接子节点中匹配并落库。
	if pending, ok := s.getPendingChildEvent(deviceNo); ok {
		reply, exit, finishTalk, pendErr := s.continuePendingChildEvent(ctx, deviceNo, normalizedTranscript, events, pending)
		if pendErr != nil {
			return chatResult{
				Reply:      "我暂时没理解清楚，请再说一次",
				Ask:        normalizedTranscript,
				Mode:       resolvedMode,
				FinishTalk: false,
			}, pendErr
		}
		s.insertQa(ctx, normalizedTranscript, reply)
		return chatResult{
			Reply:      reply,
			Ask:        normalizedTranscript,
			Mode:       resolvedMode,
			Exit:       exit,
			FinishTalk: finishTalk,
		}, nil
	}

	// 取设备域动作词典（经 DeviceAdmin HTTP → device-service）
	actions := []entity.Action{}
	adminActions, _ := DeviceAdmin().ListActionsForAdmin(ctx)
	for _, action := range adminActions {
		actions = append(actions, entity.Action{
			Id:         action.Id,
			Name:       action.Name,
			TargetType: action.TargetType,
		})
	}
	exit := false

	// 获取上一次的对话缓存中，我回答的最后一条记录
	now := time.Now()
	lastUserMessage := s.getLastUserMessageFromSession(ctx, deviceNo, now)

	// 上一次的对话缓存中，我回答的最后一条记录，是否包含"多少"关键词
	mayReplayQuantity := false
	if strings.Contains(lastUserMessage, quantityKeyword) {
		mayReplayQuantity = true
	}

	// 获取这一次对话中的数量
	number, ok := extractNumberFromText(normalizedTranscript)
	if ok {
		// 上一次对话中如果包含"多少"关键词，则需要判断是要将上一次对话中的"多少"改为这一次的会话内容，然后走下面的逻辑
		if mayReplayQuantity {
			normalizedTranscript = strings.Replace(lastUserMessage, quantityKeyword, "? "+strconv.Itoa(number)+"。", 1)
			// 日志打印
			glog.Infof(ctx, "上一次对话中包含\"多少\"关键词，将\"多少\"改为这一次的会话内容。lastUserMessage=%q normalizedTranscript=%q", lastUserMessage, number)
		}
	}

	// 打印normalizedTranscript
	glog.Infof(ctx, "问题=%q", normalizedTranscript)

	// 先将动作按名称长度从长到短排序
	sort.Slice(actions, func(i, j int) bool {
		return len(actions[i].Name) > len(actions[j].Name)
	})
	// 判断文本是否包含预设动作关键词
	for _, action := range actions {
		if strings.Contains(normalizedTranscript, action.Name) {
			// 打印日志命中动作
			glog.Infof(ctx, "命中动作: %s", action.Name)
			finalReply, exit, finishTalk, err := s.handleActionRecord(ctx, deviceNo, normalizedTranscript, action, events)
			if err != nil {
				// 处理动作失败,可能动作解析错误,尝试解析出新的动作,再走命中事件流程
				continue
			}
			// 往QA里录入问题和答案
			s.insertQa(ctx, normalizedTranscript, finalReply)
			return chatResult{
				Reply:      finalReply,
				Ask:        normalizedTranscript,
				Mode:       resolvedMode,
				Exit:       exit,
				FinishTalk: finishTalk,
			}, err
		}
	}
	// 没有命中预设动作: 单次请求 DeepSeek 返回全量结构，再统一处理。
	wxID, qCtx, qErr := s.guardVoiceAIQuota(ctx, deviceNo)
	if qErr != nil {
		return chatResult{}, qErr
	}
	ctx = qCtx
	intent, err := s.callDeepSeekUnifiedIntent(ctx, deviceNo, normalizedTranscript)
	if err != nil {
		return chatResult{
			Reply:      "我暂时没理解清楚你的意思，请再说一次",
			Ask:        normalizedTranscript,
			Mode:       resolvedMode,
			Exit:       exit,
			FinishTalk: false,
		}, err
	}
	s.consumeVoiceAIQuotaOnSuccess(ctx, wxID)

	if ParseActionTargetType(intent.TargetType) == ActionTargetTypeConversation || strings.TrimSpace(intent.TargetType) == "" {
		reply := strings.TrimSpace(intent.Reply)
		if reply == "" {
			reply = "我明白了，请再具体一点，我马上帮你处理。"
		}
		s.insertQa(ctx, normalizedTranscript, reply)
		return chatResult{
			Reply:      reply,
			Ask:        normalizedTranscript,
			Mode:       resolvedMode,
			FinishTalk: !intent.NeedUserReply,
		}, nil
	}

	action := entity.Action{
		Name:       strings.TrimSpace(intent.ActionName),
		TargetType: ParseActionTargetType(intent.TargetType).String(),
	}
	if action.Name == "" {
		action.Name = strings.TrimSpace(intent.Action)
	}
	if action.Name == "" {
		action.Name = normalizedTranscript
	}

	glog.Infof(ctx, "未命中预设动作，DeepSeek 单次结构化返回命中动作: %s", action.Name)
	finalReply, exit, finishTalk, err := s.handleUnifiedIntentAction(ctx, deviceNo, normalizedTranscript, action, events, intent)
	if err == nil {
		s.insertQa(ctx, normalizedTranscript, finalReply)
	}
	return chatResult{
		Reply:      finalReply,
		Ask:        normalizedTranscript,
		Mode:       resolvedMode,
		Exit:       exit,
		FinishTalk: finishTalk,
	}, err
}

func (s *VoiceService) callDeepSeekUnifiedIntent(ctx context.Context, deviceNo, transcript string) (deepSeekUnifiedIntent, error) {
	// 单次请求返回完整意图结构，减少多轮模型调用带来的延迟和不一致。
	history, _ := s.buildRecentHistory(ctx, deviceNo, 12)
	historyBytes, _ := json.Marshal(history)
	birthday, gender := s.loadDeviceProfile(ctx, deviceNo)
	prompt := fmt.Sprintf("用户输入=%s。用户宝宝信息={\"birthday\":\"%s\",\"gender\":\"%s\"}。最近12小时历史记录=%s。请仅输出JSON。", transcript, birthday, gender, string(historyBytes))
	systemMessage := fmt.Sprintf(`你是母婴语音助手的意图中枢。你需要一次性输出完整结构化结果。
目标类型仅允许: %s,%s,%s,%s,%s,%s,%s
含义:
- %s: 开始记录计时事件
- %s: 结束记录计时事件
- %s: 记录一次性事件
- %s: 退出对话
- %s: 成长建议
- %s: 历史搜索问答
- %s: 普通对话
输出JSON格式:
{
  "target_type":"...",
  "action":"动作关键词",
  "action_name":"动作展示名",
  "event_name":"事件名(可空)",
  "extra_event_name":"事件别名(可空)",
  "event_type":"number|time|one",
  "event_unit":"计数单位(可空，如 ml、次；仅新建 number 事件时填写)",
  "quantity":0,
  "reply":"给用户的回复(可空)",
  "need_user_reply":true/false
}
要求:
1) 不要输出解释，不要markdown。
2) 若是对话类，target_type=%s 并尽量给出 reply。
3) quantity 无法确定时给 0。
4) 仅当需要新建事件名时填写 event_type：计数语义 number，持续计时 time，一次性记录 one；匹配已有事件名时不要填 event_type。
5) 新建 number 事件时 SHOULD 填写 event_unit（如奶量 ml、次数 次）；无法确定可留空。
6) 当 target_type=%s 或 target_type=%s 时，你必须直接给出可播报给用户的最终 reply，不要输出“正在查询”这类中间态。`,
		ActionTargetTypeStart, ActionTargetTypeEnd, ActionTargetTypeOne, ActionTargetTypeExit, ActionTargetTypeSuggest, ActionTargetTypeSearch, ActionTargetTypeConversation,
		ActionTargetTypeStart, ActionTargetTypeEnd, ActionTargetTypeOne, ActionTargetTypeExit, ActionTargetTypeSuggest, ActionTargetTypeSearch, ActionTargetTypeConversation,
		ActionTargetTypeConversation, ActionTargetTypeSearch, ActionTargetTypeSuggest)

	raw, _, _, err := s.callDeepSeekRaw(ctx, deviceNo, prompt, 5, systemMessage)
	if err != nil {
		return deepSeekUnifiedIntent{}, err
	}
	trimmed := normalizeIntentCandidateText(raw)
	var intent deepSeekUnifiedIntent
	if err := json.Unmarshal([]byte(trimmed), &intent); err != nil {
		start := strings.Index(trimmed, "{")
		end := strings.LastIndex(trimmed, "}")
		if start >= 0 && end > start {
			if nestedErr := json.Unmarshal([]byte(trimmed[start:end+1]), &intent); nestedErr != nil {
				return deepSeekUnifiedIntent{}, err
			}
		} else {
			return deepSeekUnifiedIntent{}, err
		}
	}
	intent.TargetType = strings.TrimSpace(strings.ToLower(intent.TargetType))
	intent.Action = strings.TrimSpace(intent.Action)
	intent.ActionName = strings.TrimSpace(intent.ActionName)
	intent.EventName = strings.TrimSpace(intent.EventName)
	intent.ExtraEvent = strings.TrimSpace(intent.ExtraEvent)
	intent.Reply = sanitizeModelReplyText(intent.Reply)
	return intent, nil
}

func (s *VoiceService) handleUnifiedIntentAction(ctx context.Context, deviceNo, normalizedTranscript string, action entity.Action, events []entity.Event, intent deepSeekUnifiedIntent) (finalReply string, exit bool, finishTalk bool, err error) {
	// 统一意图动作执行器：对 suggest/search/exit 直接返回，其余动作落到事件写库流程。
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
		if quantity <= 0 {
			if q, ok := extractNumberFromText(normalizedTranscript); ok {
				quantity = q
			}
		}
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

func (s *VoiceService) resolveEventFromUnifiedIntent(ctx context.Context, deviceNo, normalizedTranscript string, events []entity.Event, intent deepSeekUnifiedIntent) (entity.Event, string, bool) {
	event, target, pending, ok, err := s.resolveEventForAction(ctx, deviceNo, normalizedTranscript, events, "", &intent)
	if err != nil || pending != "" || !ok {
		return entity.Event{}, "", false
	}
	return event, target, true
}

// 根据动作，判断后续逻辑
func (s *VoiceService) handleActionRecord(ctx context.Context, deviceNo string, normalizedTranscript string, action entity.Action, events []entity.Event) (finalReply string, exit bool, finishTalk bool, err error) {
	// 根据不同的action做出不同的处理
	nowTime := time.Now().Unix()
	switch action.TargetType {
	case "start": //开始记录计时动作
		event, targetName, pendingReply, ok, err := s.resolveEventForAction(ctx, deviceNo, normalizedTranscript, events, action.TargetType, nil)
		if err != nil {
			glog.Warningf(ctx, "调用 DeepSeek 进行实体抽取失败，deviceNo=%s transcript=%q err=%v", deviceNo, normalizedTranscript, err)
			return "我听不懂你说的事件,请用具体的名称告诉我", false, false, err
		}
		if pendingReply != "" {
			return pendingReply, false, false, nil
		}
		if !ok {
			return "我听不懂你说的事件,请用具体的名称告诉我", false, false, errors.New("未识别事件")
		}
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
		finalReply = fmt.Sprintf("好的，已记录%s开始", targetName)
		finishTalk = true
		return finalReply, false, finishTalk, nil
	case "end": //结束记录计时动作，自动补结束时间
		event, targetName, pendingReply, ok, err := s.resolveEventForAction(ctx, deviceNo, normalizedTranscript, events, action.TargetType, nil)
		if err != nil {
			glog.Warningf(ctx, "调用 DeepSeek 进行实体抽取失败，deviceNo=%s transcript=%q err=%v", deviceNo, normalizedTranscript, err)
			return "我听不懂你说的事件,请用具体的名称告诉我", false, false, err
		}
		if pendingReply != "" {
			return pendingReply, false, false, nil
		}
		if !ok {
			return "我听不懂你说的事件,请用具体的名称告诉我", false, false, errors.New("未识别事件")
		}

		finalReply, err = applyVoiceEventEndHistory(ctx, deviceNo, event, targetName, normalizedTranscript, nowTime)
		finishTalk = true
		return finalReply, false, finishTalk, err
	case "one": //记录一次性动作，记录一次
		event, targetName, pendingReply, ok, err := s.resolveEventForAction(ctx, deviceNo, normalizedTranscript, events, action.TargetType, nil)
		if err != nil {
			glog.Warningf(ctx, "调用 DeepSeek 进行实体抽取失败，deviceNo=%s transcript=%q err=%v", deviceNo, normalizedTranscript, err)
			return "我听不懂你说的事件,请用具体的名称告诉我", false, false, err
		}
		if pendingReply != "" {
			return pendingReply, false, false, nil
		}
		if !ok {
			return "我听不懂你说的事件,请用具体的名称告诉我", false, false, errors.New("未识别事件")
		}
		eventNumber := int64(1)
		if quantity, ok := extractNumberFromText(normalizedTranscript); ok && quantity > 0 {
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
		if eventNumber > 1 {
			finalReply = fmt.Sprintf("好的，已记录 %s %d", targetName, eventNumber)
		} else {
			finalReply = fmt.Sprintf("好的，已记录 %s", targetName)
		}
		finishTalk = true
		return finalReply, false, finishTalk, nil
	case "suggest": //成长建议动作
		wxID, qCtx, qErr := s.guardVoiceAIQuota(ctx, deviceNo)
		if qErr != nil {
			return "", false, true, qErr
		}
		ctx = qCtx
		reply, handleErr := s.callDeepSeekGrowthSuggestion(ctx, deviceNo)
		if handleErr != nil {
			return "获取成长建议失败,请重试", false, true, handleErr
		}
		s.consumeVoiceAIQuotaOnSuccess(ctx, wxID)
		finalReply = strings.TrimSpace(reply)
		finishTalk = true
		return finalReply, false, finishTalk, nil
	case "search": //搜索动作
		wxID, qCtx, qErr := s.guardVoiceAIQuota(ctx, deviceNo)
		if qErr != nil {
			return "", false, true, qErr
		}
		ctx = qCtx
		reply, handleErr := s.callDeepSeekHistoryReply(ctx, deviceNo, normalizedTranscript, 12)
		if handleErr != nil {
			return "获取历史记录失败,请重试", false, true, handleErr
		}
		s.consumeVoiceAIQuotaOnSuccess(ctx, wxID)
		finalReply = strings.TrimSpace(reply)
		finishTalk = true
		return finalReply, false, finishTalk, nil
	case "exit": //退出动作
		return "好的，再见", true, false, nil
	default:
		return "我没有理解你的意思", false, false, nil
	}
}

// 根据文本,请求deepSeek分析文案中的动作是什么,并判断该动作的目标类型(ActionTargetTypeStart,ActionTargetTypeEnd,ActionTargetTypeOne,ActionTargetTypeExit,ActionTargetTypeSuggest,ActionTargetTypeSearch),输出JSON:{"action":"动作名称","target_type":"目标类型"}
func (s *VoiceService) callDeepSeekActionExtract(ctx context.Context, deviceNo, transcript string) (entity.Action, error) {
	prompt := fmt.Sprintf("输入：%s", transcript)
	systemMessage := fmt.Sprintf(
		`你是一个主要记录母婴喂养的助手且具备精准的动作提取能力，严格按指定JSON格式输出，不添加任何解释。

动作名称提取：从输入文本中提取代表性的连续文案,至少两个字。

目标类型选择：
- %s(%s)：开始计时
- %s(%s)：结束计时
- %s(%s)：一次性记录
- %s(%s)：退出
- %s(%s)：成长建议
- %s(%s)：搜索
- %s(%s)：对话

特别规则：
1. 睡眠事件：睡着→%s，睡醒→%s
2. 只有动作无事件→%s或%s
3. 关于孩子的问题→%s
4. 关于历史数据的问题→%s

输出格式：{"name":"动作名称","targetType":"目标类型"}
无法判断时：{"name":"","targetType":""}`,
		ActionTargetTypeStart, ActionTargetTypeChinese(ActionTargetTypeStart),
		ActionTargetTypeEnd, ActionTargetTypeChinese(ActionTargetTypeEnd),
		ActionTargetTypeOne, ActionTargetTypeChinese(ActionTargetTypeOne),
		ActionTargetTypeExit, ActionTargetTypeChinese(ActionTargetTypeExit),
		ActionTargetTypeSuggest, ActionTargetTypeChinese(ActionTargetTypeSuggest),
		ActionTargetTypeSearch, ActionTargetTypeChinese(ActionTargetTypeSearch),
		ActionTargetTypeConversation, ActionTargetTypeChinese(ActionTargetTypeConversation),
		ActionTargetTypeStart, ActionTargetTypeEnd,
		ActionTargetTypeConversation, ActionTargetTypeExit,
		ActionTargetTypeSuggest, ActionTargetTypeSearch)
	raw, _, _, err := s.callDeepSeekRaw(ctx, deviceNo, prompt, 0, systemMessage)
	if err != nil {
		return entity.Action{}, err
	}
	parsed := entity.Action{}
	err = json.Unmarshal([]byte(raw), &parsed)
	if err != nil {
		return entity.Action{}, err
	}
	if ParseActionTargetType(parsed.TargetType) == ActionTargetTypeConversation || strings.TrimSpace(parsed.TargetType) == "" {
		return entity.Action{}, errors.New("对话动作不需要落库")
	} else {
		// 如果动作名为空,则为输入的文本值
		if parsed.Name == "" {
			parsed.Name = transcript
		}

		// 将动作落库,保证动作名称唯一
		actions, aErr := DeviceAdmin().ListActionsForAdmin(ctx)
		if aErr == nil {
			for _, action := range actions {
				if strings.EqualFold(strings.TrimSpace(action.Name), strings.TrimSpace(parsed.Name)) {
					return entity.Action{}, errors.New("动作名称已存在")
				}
			}
		}
		if insErr := DeviceAdmin().InsertVoiceActionRecord(ctx, parsed.Name, parsed.TargetType); insErr != nil {
			return entity.Action{}, insErr
		}
		return parsed, nil
	}
}

// extractEventFromText 提取文本中的事件对象（全树、子节点优先）。
func (s *VoiceService) extractEventFromText(ctx context.Context, normalizedTranscript string, events []entity.Event) (entity.Event, string, bool) {
	idx := buildEventTreeIndex(events)
	return extractEventFromCandidates(ctx, normalizedTranscript, events, idx)
}

// 提取文本中的数量值
func extractNumberFromText(text string) (int, bool) {

	// 把text中的一、二、三、四、五、六、七、八、九转换为1、2、3、4、5、6、7、8、9
	text = strings.ReplaceAll(text, "一", "1")
	text = strings.ReplaceAll(text, "二", "2")
	text = strings.ReplaceAll(text, "三", "3")
	text = strings.ReplaceAll(text, "四", "4")
	text = strings.ReplaceAll(text, "五", "5")
	text = strings.ReplaceAll(text, "六", "6")
	text = strings.ReplaceAll(text, "七", "7")
	text = strings.ReplaceAll(text, "八", "8")
	text = strings.ReplaceAll(text, "九", "9")

	text = strings.TrimSpace(text)
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(text)
	if match == "" {
		return 0, false
	}
	value, err := strconv.Atoi(match)
	if err != nil {
		return 0, false
	}
	return value, true
}

// hasSignificantOverlap 判断两个文本是否有显著的交集（至少两个连续字符）。考虑到用户说D3的情况
func hasSignificantOverlap(text, keyword string) bool {
	textRunes := []rune(text)
	keywordRunes := []rune(keyword)
	if len(textRunes) < 2 || len(keywordRunes) < 2 {
		return false
	}
	for i := 0; i < len(textRunes)-1; i++ {
		for j := 0; j < len(keywordRunes)-1; j++ {
			if textRunes[i] == keywordRunes[j] && textRunes[i+1] == keywordRunes[j+1] {
				return true
			}
		}
	}
	return false
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

func parseGeneralChatResult(reply string) (generalChatResult, bool) {
	var out generalChatResult
	trimmed := normalizeIntentCandidateText(reply)
	if strings.TrimSpace(trimmed) == "" {
		return out, false
	}
	if err := json.Unmarshal([]byte(trimmed), &out); err == nil {
		out.Reply = sanitizeModelReplyText(out.Reply)
		return out, true
	}
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(trimmed[start:end+1]), &out); err == nil {
			out.Reply = sanitizeModelReplyText(out.Reply)
			return out, true
		}
	}
	return out, false
}

// 普通对话
func (s *VoiceService) handleIntentGeneral(ctx context.Context, deviceNo, transcript string) (string, bool, error) {
	// 将用户的生日和性别作为宝宝的信息上下文输入，辅助模型判断是否需要回复以及回复内容
	birthday, gender := s.loadDeviceProfile(ctx, deviceNo)
	systemMessage := fmt.Sprintf("用户的宝宝出生日期是%s，性别是%s。你还是一个语音助手，可以解答任何问题,回应内容控制在20字以内。", birthday, gender)
	prompt := fmt.Sprintf("用户输入=%s。请仅输出JSON：{\"reply\":\"回复内容\",\"need_user_reply\":true|false}。", transcript)
	raw, reply, _, err := s.callDeepSeekRaw(ctx, deviceNo, prompt, 5, systemMessage)
	if err != nil {
		return "", false, err
	}
	if parsed, ok := parseGeneralChatResult(raw); ok {
		if s.cfg.DebugLog {
			parsedJSON, _ := json.Marshal(parsed)
			glog.Infof(ctx, "[思考过程] 其它问答结构化解析成功。deviceNo=%s parsed=%s", deviceNo, string(parsedJSON))
		}
		if parsed.Reply == "" {
			parsed.Reply = reply
		}
		return strings.TrimSpace(parsed.Reply), !parsed.NeedUserReply, nil
	}
	if s.cfg.DebugLog {
		glog.Warningf(ctx, "[思考过程] 其它问答结构化解析失败，使用文本回复。deviceNo=%s raw=%s", deviceNo, truncateVoiceLogText(raw, 800))
	}
	return strings.TrimSpace(reply), true, nil
}

func (s *VoiceService) callDeepSeekEntityExtract(ctx context.Context, deviceNo, transcript string) (entity.Event, string, error) {
	out := entity.Event{}

	// DeepSeek 分析事件名；新建事件时需输出 event_type（number|time|one）。
	// 将数据库中的事件名称拼接起来,用逗号分隔,然后告诉deepseek,事件名称有:xxx,xxx,xxx
	eventList := []entity.Event{}
	eventList, _ = DeviceAdmin().ListEvents(ctx)
	eventNamesStr := ""
	for _, event := range eventList {
		eventNamesStr += event.Name + ","
	}
	systemMessage := fmt.Sprintf(`你是一个精准的事件提取器，严格输出JSON。

事件列表（含 parent_id 层级，子事件名更具体）：%s

特别规则：
1. 扩展词从文本提取连续文案
2. 吃奶事件：如无法区分母乳/配方奶，输出{"name":"","extraNames":""}
3. 用户仅提及父事件且未说明子类型时，可返回该父事件名（是否追问子类型由系统判定）
4. 不是想记录事件:输出{"name":"","extraNames":""}

输出规则：
1. 匹配事件列表 → {"name":"原事件名","extraNames":"扩展词"}
2. 不匹配但可识别新事件 → {"name":"新事件名","extraNames":"","event_type":"number|time|one","event_unit":"单位(仅 number)"}
   - 计数/数量语义 → number，并尽量给出 event_unit（如 ml、次）；持续开始结束 → time；一次性完成 → one
3. 无法确定 → {"name":""}`, eventNamesStr)

	prompt := fmt.Sprintf("输入=%s。按规则分析并输出JSON。", transcript)
	// 你需要从事件列表中,查看是否有符合的事件类型,如果有则直接返回列表中的事件类型,如果没有则需要从文本中提取事件名称。
	raw, _, _, err := s.callDeepSeekRaw(ctx, deviceNo, prompt, 0, systemMessage)
	if err != nil {
		return out, "", err
	}

	parsed := entity.Event{}
	trimmed := normalizeIntentCandidateText(raw)
	if unmarshalErr := json.Unmarshal([]byte(trimmed), &parsed); unmarshalErr != nil {
		start := strings.Index(trimmed, "{")
		end := strings.LastIndex(trimmed, "}")
		if start >= 0 && end > start {
			if nestedErr := json.Unmarshal([]byte(trimmed[start:end+1]), &parsed); nestedErr != nil {
				return out, "", fmt.Errorf("解析实体抽取结果失败: %w", unmarshalErr)
			}
		} else {
			return out, "", fmt.Errorf("解析实体抽取结果失败: %w", unmarshalErr)
		}
	}

	name := strings.TrimSpace(parsed.Name)
	if name == "" {
		name = strings.TrimSpace(parsed.Name)
	}
	if name == "" {
		return out, "", errors.New("未抽取到事件名称")
	}

	out.Name = name
	out.ExtraNames = parsed.ExtraNames
	mergeDeepSeekEventTypeJSON(trimmed, &parsed)
	mergeDeepSeekEventUnitJSON(trimmed, &parsed)
	out.EventType = device.NormalizeEventType(parsed.EventType)
	out.Unit = strings.TrimSpace(parsed.Unit)
	return DeviceAdmin().ApplyDeepSeekEventExtractPersistence(ctx, out)
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
