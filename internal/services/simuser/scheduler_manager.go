package simuser

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/glog"
)

const adminReloadStaggerMax = 30 * time.Second

// SchedulerManager 管理 sim 周期任务 goroutine 的生命周期（Admin reload 时 Stop→Start）。
type SchedulerManager struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	wg     sync.WaitGroup
	root   context.Context
}

var globalScheduler = &SchedulerManager{}

// InitScheduler 进程启动时调用一次。
func InitScheduler(parent context.Context) {
	flags := LoadRuntimeFlags(parent)
	globalScheduler.Start(parent, flags, false)
}

// ReloadSchedulerFromAdmin Admin 保存调度类配置后热重启 scheduler。
func ReloadSchedulerFromAdmin(parent context.Context) {
	flags := LoadRuntimeFlags(parent)
	globalScheduler.Reload(parent, flags)
}

func (m *SchedulerManager) Start(parent context.Context, flags RuntimeFlags, skipLongStagger bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		m.cancel()
		m.wg.Wait()
	}
	schedCtx, cancel := context.WithCancel(parent)
	m.root = parent
	m.cancel = cancel
	startSchedulerGoroutines(schedCtx, &m.wg, flags, skipLongStagger)
}

func (m *SchedulerManager) Reload(parent context.Context, flags RuntimeFlags) {
	m.Start(parent, flags, true)
}

func (m *SchedulerManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		m.cancel()
		m.wg.Wait()
		m.cancel = nil
	}
}

func startSchedulerGoroutines(ctx context.Context, wg *sync.WaitGroup, flags RuntimeFlags, skipLongStagger bool) {
	if !flags.Enabled {
		glog.Infof(ctx, "[simuser] SIM_USER_SERVICE_ENABLED=false，跳过 scheduler")
		return
	}
	cfg, err := GetConfig(ctx)
	if err == nil && !cfg.Enabled {
		glog.Infof(ctx, "[simuser] sim_config.enabled=false，跳过 scheduler")
		return
	}
	password := flags.DefaultPassword
	if password == "" {
		password = "123456"
	}
	stagger := flags.StartupStaggerMax
	if skipLongStagger {
		stagger = adminReloadStaggerMax
	}
	if flags.TaskRegister {
		runPeriodicTracked(ctx, wg, "register", flags.IntervalRegister, stagger, skipLongStagger, func(c context.Context) {
			RunRegisterTask(c, password)
		})
	}
	if flags.TaskComment {
		runPeriodicTracked(ctx, wg, "comment", flags.IntervalComment, stagger, skipLongStagger, func(c context.Context) {
			RunCommentTask(c, password)
		})
	}
	if flags.TaskPostImage {
		runPeriodicTracked(ctx, wg, "post_image", flags.IntervalPostImage, stagger, skipLongStagger, func(c context.Context) {
			RunPostImageTask(c, password)
		})
	}
	if flags.TaskPostVideo {
		runPeriodicTracked(ctx, wg, "post_video_submit", flags.IntervalPostVideo, stagger, skipLongStagger, func(c context.Context) {
			RunPostVideoSubmitTask(c, password)
		})
	}
	if flags.TaskChat {
		runPeriodicTracked(ctx, wg, "chat_scan", flags.IntervalChat, stagger, skipLongStagger, func(c context.Context) {
			RunChatScanTask(c, password, flags)
		})
	}
	if flags.TaskFollow {
		runPeriodicTracked(ctx, wg, "follow", flags.IntervalFollow, stagger, skipLongStagger, func(c context.Context) {
			RunFollowTask(c, password)
		})
	}
	if flags.VideoPoll {
		runAdaptivePeriodicTracked(ctx, wg, "video_poll", flags.IntervalVideoPollIdle, flags.IntervalVideoPollActive, stagger, skipLongStagger, func(c context.Context) bool {
			return RunVideoPollTask(c, password)
		})
	}
	glog.Infof(ctx, "[simuser] scheduler started skipLongStagger=%v staggerMax=%s videoPollIdle=%s active=%s ephemeralLoop=%s window=%s",
		skipLongStagger, stagger, flags.IntervalVideoPollIdle, flags.IntervalVideoPollActive, flags.EphemeralChatLoop, flags.EphemeralChatWindow)
}

func runPeriodicTracked(parent context.Context, wg *sync.WaitGroup, name string, interval, staggerMax time.Duration, skipLongStagger bool, fn func(context.Context)) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		runPeriodic(parent, name, interval, staggerMax, skipLongStagger, fn)
	}()
}

func runAdaptivePeriodicTracked(parent context.Context, wg *sync.WaitGroup, name string, idle, active, staggerMax time.Duration, skipLongStagger bool, fn func(context.Context) bool) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		runAdaptivePeriodic(parent, name, idle, active, staggerMax, skipLongStagger, fn)
	}()
}
