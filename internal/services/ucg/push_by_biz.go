package ucg

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// 公共推送业务类型；新增类型时在 PushByBizType 分支登记角标策略。
const (
	// PushBizPredictImminent 预测事项临近提醒；不使用 UCG 社区未读角标。
	PushBizPredictImminent = "predict_imminent"
)

// PushByBizType 按业务类型向用户已注册 token 发送可见推送（复用 ucg_push_device 注册表）。
// 客户端须先调用 POST /ucg/app/api/push/register 注册 APNs/HMS/MiPush token，否则静默无设备可发。
// Args: recipientWxID 接收方；bizType 业务类型；alertBody 可见文案；data 可选深链字段。
// Side Effects: 异步下发厂商推送；无效 token 可能删除。
func PushByBizType(ctx context.Context, recipientWxID int64, bizType, alertBody string, data map[string]string) {
	if recipientWxID <= 0 || strings.TrimSpace(alertBody) == "" {
		return
	}
	bizType = strings.TrimSpace(strings.ToLower(bizType))
	switch bizType {
	case PushBizPredictImminent:
		// 预测临近：可见提醒，badge=0，禁止套用社区未读之和。
		asyncPush(recipientWxID, func(bg context.Context) {
			dispatchPush(bg, recipientWxID, PushPayload{
				Alert:  alertBody,
				Badge:  0,
				Silent: false,
				Data:   data,
			})
		})
	default:
		g.Log().Warningf(ctx, "[ucg-push] unknown bizType=%s wxId=%d", bizType, recipientWxID)
	}
}
