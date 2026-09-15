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
// Side Effects: 异步厂商推送；无 token 则打 no_device 日志后跳过。
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
	// 保证 dispatch 日志能带上 bizType（调用方 data 可能为 nil）。
	payloadData := data
	if payloadData == nil {
		payloadData = map[string]string{}
	} else {
		copied := make(map[string]string, len(data)+1)
		for k, v := range data {
			copied[k] = v
		}
		payloadData = copied
	}
	if strings.TrimSpace(payloadData["bizType"]) == "" && bizType != "" {
		payloadData["bizType"] = bizType
	}
	asyncPush(recipientWxID, func(bg context.Context) {
		dispatchPush(bg, recipientWxID, PushPayload{
			Alert:  alert,
			Badge:  badge,
			Silent: silent,
			Data:   payloadData,
		})
	})
}
