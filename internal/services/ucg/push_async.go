package ucg

import (
	"context"
	"sync"
)

var pushWorkerOnce sync.Once

func asyncPush(recipientWxID int64, fn func(context.Context)) {
	pushWorkerOnce.Do(func() {})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// best-effort; must not crash process
			}
		}()
		ctx := context.Background()
		fn(ctx)
	}()
	_ = recipientWxID
}
