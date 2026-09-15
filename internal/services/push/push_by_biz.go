package push

import (
	"context"
	"strings"
)

// 业务类型常量；调用方经 internal API 传入。
const (
	PushBizPredictImminent = "predict_imminent"
	PushBizUcgAlert        = "ucg_alert"
	PushBizUcgSilentBadge  = "ucg_silent_badge"
)

// PushByBizType 按业务类型异步下发；badge 由调用方决定（默认 0）。
// Side Effects: 异步厂商推送；无 token 则静默跳过。
func PushByBizType(ctx context.Context, recipientWxID int64, bizType, alertBody string, badge int, silent bool, data map[string]string) {
	if recipientWxID <= 0 {
		return
	}
	bizType = strings.TrimSpace(strings.ToLower(bizType))
	if !silent && strings.TrimSpace(alertBody) == "" && bizType != PushBizUcgSilentBadge {
		return
	}
	if badge < 0 {
		badge = 0
	}
	alert := strings.TrimSpace(alertBody)
	asyncPush(recipientWxID, func(bg context.Context) {
		dispatchPush(bg, recipientWxID, PushPayload{
			Alert:  alert,
			Badge:  badge,
			Silent: silent,
			Data:   data,
		})
	})
}
