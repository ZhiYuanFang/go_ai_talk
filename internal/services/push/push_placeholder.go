package push

import "context"

// PlaceholderSender 未上架厂商的空实现：不访问厂商 API。
type PlaceholderSender struct {
	channel string
}

func NewPlaceholderSender(channel string) *PlaceholderSender {
	return &PlaceholderSender{channel: channel}
}

func (s *PlaceholderSender) Channel() string {
	if s == nil {
		return ""
	}
	return s.channel
}

// Send 不会被 dispatcher 调用；占位判断在 sendOne。保留签名以满足 PushSender。
func (s *PlaceholderSender) Send(ctx context.Context, token string, payload PushPayload) (bool, error) {
	return false, nil
}
