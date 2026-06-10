package voice

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hello/internal/model/entity"
)

// pendingChildEventState 记录待用户细分子事件的上下文（进程内存，不持久化）。
type pendingChildEventState struct {
	ParentEventId int64
	ActionTarget  string
	ParentName    string
}

// eventLeafResolveResult 事件解析结果：叶子可落库，非叶子需追问。
type eventLeafResolveResult struct {
	Event                  entity.Event
	TargetName             string
	OK                     bool
	NeedChildDisambiguation bool
	DisambiguationReply    string
}

func (s *VoiceService) setPendingChildEvent(deviceNo string, state pendingChildEventState) {
	s.pendingChildMu.Lock()
	defer s.pendingChildMu.Unlock()
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return
	}
	s.pendingChild[deviceNo] = state
}

func (s *VoiceService) clearPendingChildEvent(deviceNo string) {
	s.pendingChildMu.Lock()
	defer s.pendingChildMu.Unlock()
	delete(s.pendingChild, strings.TrimSpace(deviceNo))
}

func (s *VoiceService) getPendingChildEvent(deviceNo string) (pendingChildEventState, bool) {
	s.pendingChildMu.Lock()
	defer s.pendingChildMu.Unlock()
	state, ok := s.pendingChild[strings.TrimSpace(deviceNo)]
	return state, ok
}

func (s *VoiceService) wrapLeafResult(ev entity.Event, target string, idx eventTreeIndex) eventLeafResolveResult {
	if idx.hasChildren(ev.Id) {
		children := idx.childrenByParent[ev.Id]
		return eventLeafResolveResult{
			Event:                   ev,
			OK:                      true,
			NeedChildDisambiguation: true,
			DisambiguationReply:     formatChildDisambiguationReply(ev.Name, children),
		}
	}
	if target == "" {
		target = ev.Name
	}
	return eventLeafResolveResult{Event: ev, TargetName: target, OK: true}
}

// resolveEventLeaf 在全树或指定候选中解析事件；非叶子返回追问语义。
func (s *VoiceService) resolveEventLeaf(ctx context.Context, normalizedTranscript string, events []entity.Event, candidates []entity.Event, idx eventTreeIndex) eventLeafResolveResult {
	if len(candidates) == 0 {
		candidates = events
	}
	if ev, target, ok := extractEventFromCandidates(ctx, normalizedTranscript, candidates, idx); ok {
		return s.wrapLeafResult(ev, target, idx)
	}
	return eventLeafResolveResult{}
}

// continuePendingChildEvent 处理 pending 状态下的用户后续输入。
func (s *VoiceService) continuePendingChildEvent(ctx context.Context, deviceNo, normalizedTranscript string, events []entity.Event, pending pendingChildEventState) (reply string, exit bool, finishTalk bool, err error) {
	idx := buildEventTreeIndex(events)
	children := idx.childrenByParent[pending.ParentEventId]
	res := s.resolveEventLeaf(ctx, normalizedTranscript, events, children, idx)
	if !res.OK {
		reaskParent := strings.TrimSpace(pending.ParentName)
		if reaskParent == "" {
			reaskParent = "该事件"
		}
		return formatChildDisambiguationReply(reaskParent, children), false, false, nil
	}
	if res.NeedChildDisambiguation {
		s.setPendingChildEvent(deviceNo, pendingChildEventState{
			ParentEventId: res.Event.Id,
			ActionTarget:  pending.ActionTarget,
			ParentName:    res.Event.Name,
		})
		return res.DisambiguationReply, false, false, nil
	}
	s.clearPendingChildEvent(deviceNo)
	return s.applyEventActionTarget(ctx, deviceNo, pending.ActionTarget, res.Event, res.TargetName, normalizedTranscript, 0)
}

// applyEventActionTarget 按动作目标将叶子事件写入 history。
func (s *VoiceService) applyEventActionTarget(ctx context.Context, deviceNo, actionTarget string, event entity.Event, targetName, normalizedTranscript string, quantity int) (reply string, exit bool, finishTalk bool, err error) {
	nowTime := time.Now().Unix()
	targetName = historyRowEventName(event, targetName)
	switch strings.TrimSpace(strings.ToLower(actionTarget)) {
	case ActionTargetTypeStart.String(), "start":
		_, err = DeviceHistory().AddHistory(ctx, entity.History{
			DeviceNo:  deviceNo,
			EventId:   event.Id,
			EventName: targetName,
			EventUnit: historyRowEventUnit(event),
			StartTime: nowTime,
			Remark:    normalizedTranscript,
		})
		if err != nil {
			return "记录失败,请重试", false, true, err
		}
		return fmt.Sprintf("好的，已记录%s开始", targetName), false, true, nil
	case ActionTargetTypeEnd.String(), "end":
		lastEvent, _ := DeviceHistory().GetLatestHistory(ctx, deviceNo)
		if lastEvent.EventId == event.Id {
			_, err = DeviceHistory().EndLatestHistoryIfMatch(ctx, deviceNo, event.Id, nowTime, normalizedTranscript)
			if err != nil {
				return "更新结束时间失败,请重试", false, true, err
			}
			return fmt.Sprintf("好的，已记录%s结束", targetName), false, true, nil
		}
		_, err = DeviceHistory().AddHistory(ctx, entity.History{
			DeviceNo:  deviceNo,
			EventId:   event.Id,
			EventName: targetName,
			StartTime: nowTime,
			EndTime:   nowTime,
			Remark:    normalizedTranscript,
			EventUnit: historyRowEventUnit(event),
		})
		if err != nil {
			return "记录事件失败,请重试", false, true, err
		}
		if lastEvent.EndTime == 0 && lastEvent.EventId > 0 {
			_, _ = DeviceHistory().EndLatestHistoryIfMatch(ctx, deviceNo, lastEvent.EventId, nowTime, "")
			return fmt.Sprintf("好的，已记录%s结束，%s自动结束", targetName, lastEvent.EventName), false, true, nil
		}
		return fmt.Sprintf("好的，已记录%s结束", targetName), false, true, nil
	case ActionTargetTypeOne.String(), "one":
		eventNumber := int64(1)
		if quantity > 0 {
			eventNumber = int64(quantity)
		} else if q, ok := extractNumberFromText(normalizedTranscript); ok && q > 0 {
			eventNumber = int64(q)
		}
		_, err = DeviceHistory().AddHistory(ctx, entity.History{
			DeviceNo:    deviceNo,
			EventId:     event.Id,
			EventName:   targetName,
			EventNumber: eventNumber,
			EventUnit:   historyRowEventUnit(event),
			StartTime:   nowTime,
			EndTime:     nowTime,
			Remark:      normalizedTranscript,
		})
		if err != nil {
			return "记录事件失败,请重试", false, true, err
		}
		if eventNumber > 1 {
			return fmt.Sprintf("好的，已记录 %s %d", targetName, eventNumber), false, true, nil
		}
		return fmt.Sprintf("好的，已记录 %s", targetName), false, true, nil
	default:
		return "我没有理解你的意思", false, false, nil
	}
}

// resolveEventForAction 解析动作链路中的事件；非叶子时写入 pending 并返回追问。
func (s *VoiceService) resolveEventForAction(ctx context.Context, deviceNo, normalizedTranscript string, events []entity.Event, actionTarget string, intent *deepSeekUnifiedIntent) (event entity.Event, targetName string, pendingReply string, ok bool, err error) {
	idx := buildEventTreeIndex(events)
	res := s.resolveEventLeaf(ctx, normalizedTranscript, events, nil, idx)
	if !res.OK && intent != nil {
		target := strings.TrimSpace(intent.ExtraEvent)
		needle := strings.TrimSpace(intent.EventName)
		sorted := append([]entity.Event(nil), events...)
		idx.sortForMatch(sorted)
		for _, ev := range sorted {
			if needle != "" && (strings.EqualFold(ev.Name, needle) || strings.Contains(strings.ToLower(ev.Name), strings.ToLower(needle))) {
				if target == "" {
					target = ev.Name
				}
				res = s.wrapLeafResult(ev, target, idx)
				break
			}
		}
	}
	if !res.OK {
		var dsErr error
		event, targetName, dsErr = s.callDeepSeekEntityExtract(ctx, deviceNo, normalizedTranscript)
		if dsErr != nil {
			return entity.Event{}, "", "", false, dsErr
		}
		idx = buildEventTreeIndex(events)
		res = s.wrapLeafResult(event, targetName, idx)
	}
	if !res.OK && intent != nil {
		needle := strings.TrimSpace(intent.EventName)
		if needle != "" {
			target := strings.TrimSpace(intent.ExtraEvent)
			if target == "" {
				target = needle
			}
			inserted, insErr := DeviceAdmin().InsertOrGetEventByNeedle(ctx, needle, intent.EventType, intent.EventUnit)
			if insErr == nil && inserted.Id > 0 {
				idx = buildEventTreeIndex(events)
				res = s.wrapLeafResult(inserted, target, idx)
			}
		}
	}
	if !res.OK {
		return entity.Event{}, "", "", false, nil
	}
	if res.NeedChildDisambiguation {
		s.setPendingChildEvent(deviceNo, pendingChildEventState{
			ParentEventId: res.Event.Id,
			ActionTarget:  actionTarget,
			ParentName:    res.Event.Name,
		})
		return entity.Event{}, "", res.DisambiguationReply, true, nil
	}
	return res.Event, res.TargetName, "", true, nil
}
