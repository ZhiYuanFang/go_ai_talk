package push

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// PushPayload launcher badge push contract.
type PushPayload struct {
	Alert  string
	Badge  int
	Silent bool
	Data   map[string]string
}

// PushSender sends to a single device token.
type PushSender interface {
	Channel() string
	Send(ctx context.Context, token string, payload PushPayload) (invalidToken bool, err error)
}

// PushDispatcher fans out to channel-specific senders.
type PushDispatcher struct {
	senders map[string]PushSender
}

var defaultPushDispatcher *PushDispatcher

// PushDispatcherInstance returns the process-wide dispatcher (lazy init).
func PushDispatcherInstance() *PushDispatcher {
	if defaultPushDispatcher == nil {
		defaultPushDispatcher = NewPushDispatcher(
			NewApnsSender(),
			NewHmsSender(),
			NewMipushSender(),
		)
	}
	return defaultPushDispatcher
}

// NewPushDispatcher builds a dispatcher from senders.
func NewPushDispatcher(senders ...PushSender) *PushDispatcher {
	m := make(map[string]PushSender, len(senders))
	for _, s := range senders {
		if s == nil {
			continue
		}
		ch := strings.TrimSpace(strings.ToLower(s.Channel()))
		if ch != "" {
			m[ch] = s
		}
	}
	return &PushDispatcher{senders: m}
}

func pushBizTypeFromPayload(payload PushPayload) string {
	if payload.Data == nil {
		return ""
	}
	return strings.TrimSpace(payload.Data["bizType"])
}

func dispatchPush(ctx context.Context, recipientWxID int64, payload PushPayload) {
	bizType := pushBizTypeFromPayload(payload)
	devices, err := ListPushDevicesForWx(ctx, recipientWxID)
	if err != nil {
		g.Log().Warningf(ctx, "[push] skip reason=list_devices_err wxId=%d bizType=%s err=%v", recipientWxID, bizType, err)
		return
	}
	// prod 可见：对齐 HTTP accepted 与后续 skip/send。
	g.Log().Warningf(ctx, "[push] dispatch wxId=%d bizType=%s deviceCount=%d silent=%v",
		recipientWxID, bizType, len(devices), payload.Silent)
	if len(devices) == 0 {
		// 调用方 HTTP 已 200，但无注册 token：常见于未 POST /app/api/push/register。
		g.Log().Warningf(ctx, "[push] skip reason=no_device wxId=%d bizType=%s silent=%v", recipientWxID, bizType, payload.Silent)
		return
	}
	d := PushDispatcherInstance()
	for _, dev := range devices {
		d.sendOne(ctx, dev, payload)
	}
}

func (d *PushDispatcher) sendOne(ctx context.Context, dev entityPushDevice, payload PushPayload) {
	if d == nil {
		return
	}
	bizType := pushBizTypeFromPayload(payload)
	ch := strings.TrimSpace(strings.ToLower(dev.Channel))
	sender, ok := d.senders[ch]
	if !ok || sender == nil {
		g.Log().Warningf(ctx, "[push] skip reason=no_sender channel=%s wxId=%d bizType=%s", ch, dev.WxID, bizType)
		return
	}
	token := strings.TrimSpace(dev.Token)
	if token == "" {
		g.Log().Warningf(ctx, "[push] skip reason=empty_token channel=%s wxId=%d deviceId=%d bizType=%s", ch, dev.WxID, dev.ID, bizType)
		return
	}
	invalid, err := sender.Send(ctx, token, payload)
	if err != nil {
		g.Log().Warningf(ctx, "[push] send_failed channel=%s wxId=%d deviceId=%d bizType=%s err=%v", ch, dev.WxID, dev.ID, bizType, err)
	} else {
		// 成功路径也打 Warning，避免 GF_LOGGER_LEVEL=prod 下完全静默。
		g.Log().Warningf(ctx, "[push] send_ok channel=%s wxId=%d deviceId=%d bizType=%s", ch, dev.WxID, dev.ID, bizType)
	}
	if invalid && dev.ID > 0 {
		if delErr := DeletePushDeviceByID(ctx, dev.ID); delErr != nil {
			g.Log().Warningf(ctx, "[push] delete invalid token failed id=%d err=%v", dev.ID, delErr)
		} else {
			g.Log().Infof(ctx, "[push] deleted invalid token id=%d channel=%s wxId=%d", dev.ID, ch, dev.WxID)
		}
	}
}
