package voice

import (
	"context"
	async "hello/internal/services/async"
)

type voiceTaskProducer struct {
	inner *async.VoiceTaskProducer
}

func newVoiceTaskProducer() *voiceTaskProducer {
	return &voiceTaskProducer{inner: async.NewVoiceTaskProducer()}
}

func (p *voiceTaskProducer) publishTaskRequested(ctx context.Context, deviceNo, transcript, source string) error {
	return p.inner.PublishTaskRequested(ctx, deviceNo, transcript, source)
}
