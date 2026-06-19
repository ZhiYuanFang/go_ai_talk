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
	if flags.TaskRegister {
		go runPeriodic(ctx, "register", flags.IntervalRegister, func(c context.Context) {
			RunRegisterTask(c, password)
		})
	}
	if flags.TaskComment {
		go runPeriodic(ctx, "comment", flags.IntervalComment, func(c context.Context) {
			RunCommentTask(c, password)
		})
	}
	if flags.TaskPostImage {
		go runPeriodic(ctx, "post_image", flags.IntervalPostImage, func(c context.Context) {
			RunPostImageTask(c, password)
		})
	}
	if flags.TaskPostVideo {
		go runPeriodic(ctx, "post_video_submit", flags.IntervalPostVideo, func(c context.Context) {
			RunPostVideoSubmitTask(c, password)
		})
	}
	if flags.TaskChat {
		go runPeriodic(ctx, "chat_scan", flags.IntervalChat, func(c context.Context) {
			RunChatScanTask(c, password)
		})
	}
	if flags.TaskFollow {
		go runPeriodic(ctx, "follow", flags.IntervalFollow, func(c context.Context) {
			RunFollowTask(c, password)
		})
	}
	if flags.VideoPoll {
		go runPeriodic(ctx, "video_poll", flags.IntervalVideoPoll, func(c context.Context) {
			RunVideoPollTask(c, password)
		})
	}
	glog.Infof(ctx, "[simuser] scheduler started")
}

func runPeriodic(parent context.Context, name string, interval time.Duration, fn func(context.Context)) {
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
