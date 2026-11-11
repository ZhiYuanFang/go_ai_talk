package push

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// LiveActivityOp 内部编排请求。
// Action: start | update | end | upsert（有 activity token→update，否则 start）。
type LiveActivityOp struct {
	WxID          int64
	Action        string
	Attributes    map[string]interface{}
	ContentState  map[string]interface{}
	AlertTitle    string
	AlertBody     string
	DismissalDate int64 // end 时 unix 秒；<=0 由 APNs 组装层填过去时刻
}

// FanoutLiveActivity 异步向某用户的相关 LA token 扇出。
//
// 业务：
//   - 开关关闭时仅打日志并返回；
//   - start / upsert(无 activity) 用 push_to_start；update / upsert(有 activity) / end 用 activity token；
//   - start 前须该 wx 仍有至少一条 push_device（任意 channel，近似代理「消息通知」未关）；无则 skip；
//   - update 遇 invalidToken：仅删该 activity 行，不同拍 restart；后续靠 upsert/clear 再走 start；
//   - 无 token 时跳过（不视为整次失败）；
//   - action=end 后清该 wx 全部 activity 行，保留 push_to_start；
//   - 划掉后脏 activity 由客户端鉴权 clear（App 打开无 LA 时）清理，下次 upsert 走 start。
func FanoutLiveActivity(ctx context.Context, op LiveActivityOp) {
	if op.WxID <= 0 {
		return
	}
	action := strings.TrimSpace(strings.ToLower(op.Action))
	if action == "" {
		return
	}
	if !LiveActivityEnabled() {
		g.Log().Warningf(ctx, "[push-la] skip reason=feature_off wxId=%d action=%s", op.WxID, action)
		return
	}
	asyncPush(op.WxID, func(bg context.Context) {
		dispatchLiveActivity(bg, op)
	})
}

// cleanupActivityTokensAfterEnd end 后删除该 wx 全部 activity 行。
func cleanupActivityTokensAfterEnd(ctx context.Context, wxID int64) {
	if wxID <= 0 {
		return
	}
	if delErr := DeleteLiveActivityTokensByWxActivity(ctx, wxID); delErr != nil {
		g.Log().Warningf(ctx, "[push-la] delete activity tokens after end failed wxId=%d err=%v", wxID, delErr)
		return
	}
	g.Log().Warningf(ctx, "[push-la] activity tokens cleaned after end wxId=%d", wxID)
}

func dispatchLiveActivity(ctx context.Context, op LiveActivityOp) {
	action := strings.TrimSpace(strings.ToLower(op.Action))

	var tokens []entityLiveActivityToken
	var err error
	sendEvent := action

	switch action {
	case "start":
		tokens, err = ListLiveActivityTokensForWx(ctx, op.WxID, LATokenKindPushToStart)
	case "update", "end":
		tokens, err = ListLiveActivityTokensForWx(ctx, op.WxID, LATokenKindActivity)
	case "upsert":
		// 有 activity update token → update；否则 push_to_start → start。
		actTokens, listErr := ListLiveActivityTokensForWx(ctx, op.WxID, LATokenKindActivity)
		if listErr != nil {
			g.Log().Warningf(ctx, "[push-la] skip reason=list_token_err wxId=%d action=upsert err=%v", op.WxID, listErr)
			return
		}
		if len(actTokens) > 0 {
			tokens = actTokens
			sendEvent = "update"
		} else {
			tokens, err = ListLiveActivityTokensForWx(ctx, op.WxID, LATokenKindPushToStart)
			sendEvent = "start"
		}
	default:
		g.Log().Warningf(ctx, "[push-la] skip reason=bad_action wxId=%d action=%s", op.WxID, action)
		return
	}
	if err != nil {
		g.Log().Warningf(ctx, "[push-la] skip reason=list_token_err wxId=%d action=%s err=%v", op.WxID, action, err)
		if action == "end" {
			cleanupActivityTokensAfterEnd(ctx, op.WxID)
		}
		return
	}
	if len(tokens) == 0 {
		g.Log().Warningf(ctx, "[push-la] skip reason=no_la_token wxId=%d action=%s sendEvent=%s",
			op.WxID, action, sendEvent)
		if action == "end" {
			cleanupActivityTokensAfterEnd(ctx, op.WxID)
		}
		return
	}

	// 仅 start：无 push_device 则视为消息通知已关（unregister），跳过弹出；update/end 不闸。
	if sendEvent == "start" {
		devices, listDevErr := ListPushDevicesForWx(ctx, op.WxID)
		if listDevErr != nil {
			g.Log().Warningf(ctx, "[push-la] skip reason=list_push_device_err wxId=%d action=%s err=%v",
				op.WxID, action, listDevErr)
			return
		}
		if len(devices) == 0 {
			g.Log().Warningf(ctx, "[push-la] skip reason=no_push_device wxId=%d action=%s sendEvent=start",
				op.WxID, action)
			return
		}
	}

	attrs := op.Attributes
	if sendEvent == "start" {
		attrs = ensureLiveActivityStartAttributes(op.WxID, attrs)
	}

	g.Log().Warningf(ctx, "[push-la] dispatch wxId=%d action=%s sendEvent=%s tokenCount=%d",
		op.WxID, action, sendEvent, len(tokens))
	for _, tok := range tokens {
		// ContentState 仅业务字段；不再注入 ack / 短超时回执。
		payload := LiveActivityAPNsPayload{
			Event:          sendEvent,
			AttributesType: LiveActivityAttributesType(),
			Attributes:     attrs,
			ContentState:   op.ContentState,
			AlertTitle:     op.AlertTitle,
			AlertBody:      op.AlertBody,
			DismissalDate:  op.DismissalDate,
		}

		invalid, sendErr := SendLiveActivityAPNs(ctx, tok.Token, payload)
		if sendErr != nil {
			g.Log().Warningf(ctx, "[push-la] send_failed wxId=%d action=%s tokenId=%d err=%v", op.WxID, sendEvent, tok.ID, sendErr)
		} else {
			g.Log().Warningf(ctx, "[push-la] send_ok wxId=%d action=%s tokenId=%d", op.WxID, sendEvent, tok.ID)
		}
		// 仅 invalidToken 删行；非 invalid 的失败保留 token。不同拍 restart。
		if !invalid || tok.ID == 0 {
			continue
		}
		if delErr := DeleteLiveActivityTokenByID(ctx, tok.ID); delErr != nil {
			g.Log().Warningf(ctx, "[push-la] delete invalid la token failed id=%d err=%v", tok.ID, delErr)
		}
	}
	if action == "end" {
		cleanupActivityTokensAfterEnd(ctx, op.WxID)
	}
}

// ensureLiveActivityStartAttributes 保证 start 载荷带稳定 activityId（与首启分支一致）。
func ensureLiveActivityStartAttributes(wxID int64, attrs map[string]interface{}) map[string]interface{} {
	if attrs == nil {
		attrs = map[string]interface{}{}
	}
	if _, ok := attrs["activityId"]; !ok {
		attrs["activityId"] = fmt.Sprintf("baby:%d", wxID)
	}
	return attrs
}
