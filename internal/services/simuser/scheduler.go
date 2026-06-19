package simuser

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

// StartScheduler 启动全部周期任务（总开关关闭时不启动）。
func StartScheduler(ctx context.Context, flags RuntimeFlags) {
	if !flags.Enabled {
		glog.Infof(ctx, "[simuser] SIM_USER_SERVICE_ENABLED=false，跳过 scheduler")
		return
	}
	password := flags.DefaultPassword
	if password == "" {
		password = "123456"
	}
	stagger := flags.StartupStaggerMax
	if flags.TaskRegister {
		go runPeriodic(ctx, "register", flags.IntervalRegister, stagger, func(c context.Context) {
			RunRegisterTask(c, password)
		})
	}
	if flags.TaskComment {
		go runPeriodic(ctx, "comment", flags.IntervalComment, stagger, func(c context.Context) {
			RunCommentTask(c, password)
		})
	}
	if flags.TaskPostImage {
		go runPeriodic(ctx, "post_image", flags.IntervalPostImage, stagger, func(c context.Context) {
			RunPostImageTask(c, password)
		})
	}
	if flags.TaskPostVideo {
		go runPeriodic(ctx, "post_video_submit", flags.IntervalPostVideo, stagger, func(c context.Context) {
			RunPostVideoSubmitTask(c, password)
		})
	}
	if flags.TaskChat {
		go runPeriodic(ctx, "chat_scan", flags.IntervalChat, stagger, func(c context.Context) {
			RunChatScanTask(c, password, flags)
		})
	}
	if flags.TaskFollow {
		go runPeriodic(ctx, "follow", flags.IntervalFollow, stagger, func(c context.Context) {
			RunFollowTask(c, password)
		})
	}
	if flags.VideoPoll {
		go runAdaptivePeriodic(ctx, "video_poll", flags.IntervalVideoPollIdle, flags.IntervalVideoPollActive, stagger, func(c context.Context) bool {
			return RunVideoPollTask(c, password)
		})
	}
	glog.Infof(ctx, "[simuser] scheduler started staggerMax=%s videoPollIdle=%s active=%s ephemeralLoop=%s window=%s",
		stagger, flags.IntervalVideoPollIdle, flags.IntervalVideoPollActive, flags.EphemeralChatLoop, flags.EphemeralChatWindow)
}

// randomStartupDelay 首次 tick 前随机等待，避免多任务启动齐射。
func randomStartupDelay(parent context.Context, max time.Duration) {
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

func runPeriodic(parent context.Context, name string, interval, staggerMax time.Duration, fn func(context.Context)) {
	randomStartupDelay(parent, staggerMax)
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
func runAdaptivePeriodic(parent context.Context, name string, idle, active, staggerMax time.Duration, fn func(context.Context) bool) {
	randomStartupDelay(parent, staggerMax)
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
