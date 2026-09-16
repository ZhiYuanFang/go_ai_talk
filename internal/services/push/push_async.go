package push

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
)

var pushWorkerOnce sync.Once

// asyncPush 在独立 goroutine 中执行厂商发送，避免阻塞 HTTP 受理。
// Side Effects: panic 时打 Error，禁止空 recover 导致静默丢推送。
func asyncPush(recipientWxID int64, fn func(context.Context)) {
	pushWorkerOnce.Do(func() {})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "[push] skip reason=async_panic wxId=%d panic=%v", recipientWxID, r)
			}
		}()
		ctx := context.Background()
		fn(ctx)
	}()
}
