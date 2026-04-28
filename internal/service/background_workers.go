package service

import (
	"context"
	"sync"
)

var backgroundWorkersOnce sync.Once

func StartBackgroundWorkers(ctx context.Context) {
	backgroundWorkersOnce.Do(func() {
		startVoiceTaskConsumer(ctx)
		startDomainOutboxRelay(ctx)
	})
}
