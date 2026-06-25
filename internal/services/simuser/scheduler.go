package simuser

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

// StartScheduler 兼容旧调用；新代码应使用 InitScheduler。
func StartScheduler(ctx context.Context, flags RuntimeFlags) {
	globalScheduler.Start(ctx, flags, false)
}

// randomStartupDelay 首次 tick 前随机等待，避免多任务启动齐射；Admin 热重启时 max 已缩短。
func randomStartupDelay(parent context.Context, max time.Duration, skipLongStagger bool) {
	_ = skipLongStagger
	if max <= 0 {
		return
	}
	n := time.Now().UnixNano() % int64(max)
	if n <= 0 {
		return
	}
	d := time.Duration(n)
	glog.Debugf(parent, "[simuser] startup stagger sleep=%s", d)
	timer := time.NewTimer(d)
	select {
	case <-parent.Done():
		timer.Stop()
	case <-timer.C:
	}
}

func runPeriodic(parent context.Context, name string, interval, staggerMax time.Duration, skipLongStagger bool, fn func(context.Context)) {
	randomStartupDelay(parent, staggerMax, skipLongStagger)
	for {
		wait := Jittered(interval)
		timer := time.NewTimer(wait)
		select {
		case <-parent.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		c := gctx.New()
		fn(c)
	}
}

// runAdaptivePeriodic 按 fn 返回值选择下一等待间隔（true=active，false=idle）。
func runAdaptivePeriodic(parent context.Context, name string, idle, active, staggerMax time.Duration, skipLongStagger bool, fn func(context.Context) bool) {
	randomStartupDelay(parent, staggerMax, skipLongStagger)
	next := idle
	for {
		wait := Jittered(next)
		timer := time.NewTimer(wait)
		select {
		case <-parent.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		c := gctx.New()
		if fn(c) {
			next = active
		} else {
			next = idle
		}
	}
}
