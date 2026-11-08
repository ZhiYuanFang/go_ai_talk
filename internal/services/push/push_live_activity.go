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
//   - update 遇 invalidToken：删 activity 行并同拍用同 device_key 的 push_to_start 发 start；
//   - update send_ok 后注入 ack 并短超时等回执；超时未 ack 同样删行 + push_to_start（start/end 不做）；
//   - 无 token 时跳过（不视为整次失败）；
//   - action=end 后清该 wx 全部 activity 行，保留 push_to_start（end invalid 不 restart）。
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

	attrs := op.Attributes
	if sendEvent == "start" {
		attrs = ensureLiveActivityStartAttributes(op.WxID, attrs)
	}

	g.Log().Warningf(ctx, "[push-la] dispatch wxId=%d action=%s sendEvent=%s tokenCount=%d",
		op.WxID, action, sendEvent, len(tokens))
	for _, tok := range tokens {
		// DB 补刀：上一轮 update 已超时未 ack（进程丢 AfterFunc）→ 先失效重启，禁止继续 update。
		if sendEvent == "update" && failoverStaleActivityIfAckTimedOut(ctx, op, tok) {
			continue
		}

		cs := op.ContentState
		ack := ""
		if sendEvent == "update" {
			ack = newLiveActivityAck()
			cs = cloneContentStateWithAck(op.ContentState, ack)
		}
		payload := LiveActivityAPNsPayload{
			Event:          sendEvent,
			AttributesType: LiveActivityAttributesType(),
			Attributes:     attrs,
			ContentState:   cs,
			AlertTitle:     op.AlertTitle,
			AlertBody:      op.AlertBody,
			DismissalDate:  op.DismissalDate,
		}

		invalid, sendErr := SendLiveActivityAPNs(ctx, tok.Token, payload)
		if sendErr != nil {
			g.Log().Warningf(ctx, "[push-la] send_failed wxId=%d action=%s tokenId=%d err=%v", op.WxID, sendEvent, tok.ID, sendErr)
		} else {
			g.Log().Warningf(ctx, "[push-la] send_ok wxId=%d action=%s tokenId=%d", op.WxID, sendEvent, tok.ID)
			// 仅 update：持久化 sent ack 并单次 AfterFunc 等回执（start/end 不做）。
			if sendEvent == "update" && ack != "" && tok.ID > 0 {
				if markErr := MarkLiveActivityAckSent(ctx, tok.ID, ack); markErr != nil {
					g.Log().Warningf(ctx, "[push-la] mark_ack_sent_failed tokenId=%d err=%v", tok.ID, markErr)
				} else {
					scheduleLiveActivityAckTimeout(op, tok, ack)
				}
			}
		}
		// 方案 B：仅 invalidToken 删行；非 invalid 的失败保留 token、不 restart。
		if !invalid || tok.ID == 0 {
			continue
		}
		if delErr := DeleteLiveActivityTokenByID(ctx, tok.ID); delErr != nil {
			g.Log().Warningf(ctx, "[push-la] delete invalid la token failed id=%d err=%v", tok.ID, delErr)
			continue
		}
		// update（含 upsert→update）失效：同拍用同机 push_to_start 重启；end MUST NOT restart。
		if sendEvent == "update" {
			restartLiveActivityViaPushToStart(ctx, op, tok)
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

// restartLiveActivityViaPushToStart update 失效后同拍用同 device_key 的 push_to_start 发 start。
//
// 业务：
//  1. 进门：APNs invalidToken 或 update ack 超时（调用方已删 activity 行）；
//  2. 无匹配 push_to_start 则 WARN 跳过；复用本次 ContentState / Attributes / alert；Attributes 补 activityId；
//  3. start 失败仅 WARN，同 dispatch 不重试；若 start 亦 invalid 则删该 push_to_start 行；
//  4. start 本身 MUST NOT 再注册 ack 超时。
//
// Side Effects: 可能 APNs start；可能 DELETE push_to_start 行。
func restartLiveActivityViaPushToStart(ctx context.Context, op LiveActivityOp, deadActivity entityLiveActivityToken) {
	deviceKey := strings.TrimSpace(deadActivity.DeviceKey)
	g.Log().Warningf(ctx, "[push-la] la_update_invalid_restart wxId=%d deviceKey=%s activityTokenId=%d",
		op.WxID, deviceKey, deadActivity.ID)
	if deviceKey == "" {
		g.Log().Warningf(ctx, "[push-la] la_update_invalid_restart skip reason=empty_device_key wxId=%d activityTokenId=%d",
			op.WxID, deadActivity.ID)
		return
	}
	pts, ok, findErr := FindLiveActivityTokenByWxDeviceKind(ctx, op.WxID, deviceKey, LATokenKindPushToStart)
	if findErr != nil {
		g.Log().Warningf(ctx, "[push-la] la_update_invalid_restart skip reason=find_pts_err wxId=%d deviceKey=%s err=%v",
			op.WxID, deviceKey, findErr)
		return
	}
	if !ok || strings.TrimSpace(pts.Token) == "" {
		g.Log().Warningf(ctx, "[push-la] la_update_invalid_restart skip reason=no_push_to_start wxId=%d deviceKey=%s",
			op.WxID, deviceKey)
		return
	}

	startPayload := LiveActivityAPNsPayload{
		Event:          "start",
		AttributesType: LiveActivityAttributesType(),
		Attributes:     ensureLiveActivityStartAttributes(op.WxID, op.Attributes),
		ContentState:   op.ContentState,
		AlertTitle:     op.AlertTitle,
		AlertBody:      op.AlertBody,
	}
	invalid, sendErr := SendLiveActivityAPNs(ctx, pts.Token, startPayload)
	if sendErr != nil {
		g.Log().Warningf(ctx, "[push-la] la_update_invalid_restart start_failed wxId=%d deviceKey=%s ptsTokenId=%d err=%v",
			op.WxID, deviceKey, pts.ID, sendErr)
	} else {
		g.Log().Warningf(ctx, "[push-la] la_update_invalid_restart start_ok wxId=%d deviceKey=%s ptsTokenId=%d",
			op.WxID, deviceKey, pts.ID)
	}
	// start 自身 invalid：清脏 push_to_start，禁止同循环再试。
	if invalid && pts.ID > 0 {
		if delErr := DeleteLiveActivityTokenByID(ctx, pts.ID); delErr != nil {
			g.Log().Warningf(ctx, "[push-la] la_update_invalid_restart delete_pts_failed id=%d err=%v", pts.ID, delErr)
		}
	}
}
