package ucg

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

func dispatchPush(ctx context.Context, recipientWxID int64, payload PushPayload) {
	devices, err := ListPushDevicesForWx(ctx, recipientWxID)
	if err != nil {
		g.Log().Warningf(ctx, "[ucg-push] list devices failed wxId=%d err=%v", recipientWxID, err)
		return
	}
	if len(devices) == 0 {
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
	ch := strings.TrimSpace(strings.ToLower(dev.Channel))
	sender, ok := d.senders[ch]
	if !ok || sender == nil {
		g.Log().Debugf(ctx, "[ucg-push] no sender for channel=%s wxId=%d", ch, dev.WxID)
		return
	}
	token := strings.TrimSpace(dev.Token)
	if token == "" {
		return
	}
	invalid, err := sender.Send(ctx, token, payload)
	if err != nil {
		g.Log().Warningf(ctx, "[ucg-push] send failed channel=%s wxId=%d deviceId=%d err=%v", ch, dev.WxID, dev.ID, err)
	}
	if invalid && dev.ID > 0 {
		if delErr := DeletePushDeviceByID(ctx, dev.ID); delErr != nil {
			g.Log().Warningf(ctx, "[ucg-push] delete invalid token failed id=%d err=%v", dev.ID, delErr)
		} else {
			g.Log().Infof(ctx, "[ucg-push] deleted invalid token id=%d channel=%s wxId=%d", dev.ID, ch, dev.WxID)
		}
	}
}
