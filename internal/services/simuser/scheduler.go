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
func randomStartupDelay(parent context.Context, taskName string, max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	n := time.Now().UnixNano() % int64(max)
	if n <= 0 {
		return 0
	}
	d := time.Duration(n)
	glog.Infof(parent, "[simuser] task=%s stagger sleep=%s", taskName, d)
	timer := time.NewTimer(d)
	select {
	case <-parent.Done():
		timer.Stop()
		return 0
	case <-timer.C:
		glog.Infof(parent, "[simuser] task=%s stagger done slept=%s", taskName, d)
		return d
	}
}

func runPeriodic(parent context.Context, name string, interval, staggerMax time.Duration, skipLongStagger bool, fn func(context.Context)) {
	trigger := "process_start"
	if skipLongStagger {
		trigger = "admin_reload"
	}
	glog.Infof(parent, "[simuser] task=%s goroutine enter trigger=%s interval=%s staggerMax=%s",
		name, trigger, interval, staggerMax)
	staggerSlept := randomStartupDelay(parent, name, staggerMax)
	firstTick := true
	for {
		wait := Jittered(interval)
		if firstTick {
			glog.Infof(parent, "[simuser] task=%s first interval wait=%s eta=%s staggerSlept=%s",
				name, wait, time.Now().Add(wait).Format(time.RFC3339), staggerSlept)
			firstTick = false
		} else {
			glog.Infof(parent, "[simuser] task=%s interval wait=%s eta=%s", name, wait, time.Now().Add(wait).Format(time.RFC3339))
		}
		timer := time.NewTimer(wait)
		select {
		case <-parent.Done():
			timer.Stop()
			glog.Infof(parent, "[simuser] task=%s goroutine exit", name)
			return
		case <-timer.C:
		}
		c := gctx.New()
		glog.Infof(c, "[simuser] task=%s tick begin manual=%v", name, isManualRun(c))
		start := time.Now()
		fn(c)
		glog.Infof(c, "[simuser] task=%s tick end duration=%s", name, time.Since(start).Round(time.Millisecond))
	}
}

// runAdaptivePeriodic 按 fn 返回值选择下一等待间隔（true=active，false=idle）。
func runAdaptivePeriodic(parent context.Context, name string, idle, active, staggerMax time.Duration, skipLongStagger bool, fn func(context.Context) bool) {
	randomStartupDelay(parent, name, staggerMax)
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
