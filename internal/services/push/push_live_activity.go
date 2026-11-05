package push

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// LiveActivityOp 内部编排请求：按 action 选择 Push-to-Start 或 activity token 发送。
type LiveActivityOp struct {
	WxID           int64
	Action         string // start | update | end
	AttributesKind string // imminent | timing
	ActivityID     string
	Attributes     map[string]interface{}
	ContentState   map[string]interface{}
	AlertTitle     string
	AlertBody      string
	DismissalDate  int64 // end 时 unix 秒；<=0 由 APNs 组装层填过去时刻（立即摘条），禁止依赖「省略=立即」
}

// FanoutLiveActivity 异步向某用户的全部相关 LA token 扇出。
//
// 业务：
//   - 开关关闭时仅打日志并返回（不发 APNs）；
//   - start 用 push_to_start token；update/end 用 activity token（按 activityId）；
//   - 无 token 时跳过（不视为整次失败）。
//
// Side Effects: 异步 HTTP 打 APNs；可能删除无效 LA token 行；
// action=end 时另按 activityId 主动清全部 activity 行（imminent/timing 共用）。
func FanoutLiveActivity(ctx context.Context, op LiveActivityOp) {
	if op.WxID <= 0 {
		return
	}
	action := strings.TrimSpace(strings.ToLower(op.Action))
	kind := strings.TrimSpace(strings.ToLower(op.AttributesKind))
	activityID := strings.TrimSpace(op.ActivityID)
	if action == "" || kind == "" {
		return
	}
	if !LiveActivityEnabled() {
		g.Log().Warningf(ctx, "[push-la] skip reason=feature_off wxId=%d action=%s kind=%s activityIdLen=%d",
			op.WxID, action, kind, len(activityID))
		return
	}
	asyncPush(op.WxID, func(bg context.Context) {
		dispatchLiveActivity(bg, op)
	})
}

// cleanupActivityTokensAfterEnd end 扇出尝试结束后删除该 activityId 的 activity 行。
// 计时闭合/删除与 Imminent 清屏/换槽共用；发送失败仍删，避免脏行无限积压。
// MUST NOT 删 push_to_start（DeleteLiveActivityTokensByActivityID 仅打 activity）。
func cleanupActivityTokensAfterEnd(ctx context.Context, wxID int64, activityID string) {
	activityID = strings.TrimSpace(activityID)
	if wxID <= 0 || activityID == "" {
		return
	}
	if delErr := DeleteLiveActivityTokensByActivityID(ctx, wxID, activityID); delErr != nil {
		g.Log().Warningf(ctx, "[push-la] delete activity tokens after end failed wxId=%d activityIdLen=%d err=%v",
			wxID, len(activityID), delErr)
		return
	}
	g.Log().Warningf(ctx, "[push-la] activity tokens cleaned after end wxId=%d activityIdLen=%d", wxID, len(activityID))
}

func dispatchLiveActivity(ctx context.Context, op LiveActivityOp) {
	action := strings.TrimSpace(strings.ToLower(op.Action))
	kind := strings.TrimSpace(strings.ToLower(op.AttributesKind))
	activityID := strings.TrimSpace(op.ActivityID)

	var tokens []entityLiveActivityToken
	var err error
	switch action {
	case "start":
		tokens, err = ListLiveActivityTokensForWx(ctx, op.WxID, LATokenKindPushToStart, kind)
	case "update", "end":
		if activityID == "" {
			g.Log().Warningf(ctx, "[push-la] skip reason=empty_activity_id wxId=%d action=%s", op.WxID, action)
			return
		}
		tokens, err = ListLiveActivityTokensByActivityID(ctx, op.WxID, activityID)
	default:
		g.Log().Warningf(ctx, "[push-la] skip reason=bad_action wxId=%d action=%s", op.WxID, action)
		return
	}
	if err != nil {
		g.Log().Warningf(ctx, "[push-la] skip reason=list_token_err wxId=%d action=%s err=%v", op.WxID, action, err)
		// end 仍尝试清库：列表失败时也可能有残留行可按 activityId 删。
		if action == "end" {
			cleanupActivityTokensAfterEnd(ctx, op.WxID, activityID)
		}
		return
	}
	if len(tokens) == 0 {
		g.Log().Warningf(ctx, "[push-la] skip reason=no_la_token wxId=%d action=%s kind=%s activityIdLen=%d",
			op.WxID, action, kind, len(activityID))
		// 无 token 可发 APNs，但 end 仍清可能残留的 activity 行（幂等）。
		if action == "end" {
			cleanupActivityTokensAfterEnd(ctx, op.WxID, activityID)
		}
		return
	}

	attrType := LiveActivityAttributesType(kind)
	payload := LiveActivityAPNsPayload{
		Event:          action,
		AttributesType: attrType,
		Attributes:     op.Attributes,
		ContentState:   op.ContentState,
		AlertTitle:     op.AlertTitle,
		AlertBody:      op.AlertBody,
		DismissalDate:  op.DismissalDate,
	}
	if action == "start" {
		if payload.Attributes == nil {
			payload.Attributes = map[string]interface{}{}
		}
		// 将服务端权威 activityId 放入 attributes，便于客户端回传 update token。
		if activityID != "" {
			payload.Attributes["activityId"] = activityID
		}
	}

	g.Log().Warningf(ctx, "[push-la] dispatch wxId=%d action=%s kind=%s tokenCount=%d activityIdLen=%d",
		op.WxID, action, kind, len(tokens), len(activityID))
	for _, tok := range tokens {
		invalid, sendErr := SendLiveActivityAPNs(ctx, tok.Token, payload)
		if sendErr != nil {
			g.Log().Warningf(ctx, "[push-la] send_failed wxId=%d action=%s tokenId=%d err=%v", op.WxID, action, tok.ID, sendErr)
		} else {
			g.Log().Warningf(ctx, "[push-la] send_ok wxId=%d action=%s tokenId=%d", op.WxID, action, tok.ID)
		}
		if invalid && tok.ID > 0 {
			if delErr := DeleteLiveActivityTokenByID(ctx, tok.ID); delErr != nil {
				g.Log().Warningf(ctx, "[push-la] delete invalid la token failed id=%d err=%v", tok.ID, delErr)
			}
		}
	}
	// imminent 清屏/换槽与 timing 闭合/删除共用：end 后按 activityId 清 activity 行，保留 push_to_start。
	if action == "end" {
		cleanupActivityTokensAfterEnd(ctx, op.WxID, activityID)
	}
}
