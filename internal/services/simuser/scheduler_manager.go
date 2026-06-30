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
	glog.Infof(parent, "[simuser] scheduler reload begin trigger=admin_config_save")
	flags := LoadRuntimeFlags(parent)
	globalScheduler.Reload(parent, flags)
}

func (m *SchedulerManager) Start(parent context.Context, flags RuntimeFlags, skipLongStagger bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		glog.Infof(parent, "[simuser] scheduler stop previous goroutines")
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
	trigger := "process_start"
	if skipLongStagger {
		trigger = "admin_reload"
	}
	if !flags.Enabled {
		glog.Infof(ctx, "[simuser] scheduler skip trigger=%s reason=SIM_USER_SERVICE_ENABLED=false", trigger)
		return
	}
	cfg, err := GetConfig(ctx)
	if err != nil {
		glog.Warningf(ctx, "[simuser] scheduler get sim_config failed err=%v trigger=%s 仍尝试启动 goroutine", err, trigger)
	} else if !cfg.Enabled {
		glog.Infof(ctx, "[simuser] scheduler skip trigger=%s reason=sim_config.enabled=false", trigger)
		return
	}
	DiscardPendingVideoJobsOnStartup(ctx)
	password := flags.DefaultPassword
	if password == "" {
		password = "123456"
	}
	stagger := flags.StartupStaggerMax
	if skipLongStagger {
		stagger = adminReloadStaggerMax
	}
	type taskSpec struct {
		name     string
		enabled  bool
		interval time.Duration
		run      func(context.Context)
	}
	specs := []taskSpec{
		{"register", flags.TaskRegister, flags.IntervalRegister, func(c context.Context) { RunRegisterTask(c, password) }},
		{"comment", flags.TaskComment, flags.IntervalComment, func(c context.Context) { RunCommentTask(c, password) }},
		{"post_image", flags.TaskPostImage, flags.IntervalPostImage, func(c context.Context) { RunPostImageTask(c, password) }},
		{"post_video_submit", flags.TaskPostVideo, flags.IntervalPostVideo, func(c context.Context) { RunPostVideoSubmitTask(c, password, flags) }},
		{"chat_scan", flags.TaskChat, flags.IntervalChat, func(c context.Context) { RunChatScanTask(c, password, flags) }},
		{"follow", flags.TaskFollow, flags.IntervalFollow, func(c context.Context) { RunFollowTask(c, password) }},
	}
	started := 0
	for _, spec := range specs {
		if !spec.enabled {
			glog.Infof(ctx, "[simuser] scheduler skip task=%s reason=task_switch_disabled", spec.name)
			continue
		}
		glog.Infof(ctx, "[simuser] scheduler start task=%s interval=%s staggerMax=%s trigger=%s",
			spec.name, spec.interval, stagger, trigger)
		runPeriodicTracked(ctx, wg, spec.name, spec.interval, stagger, skipLongStagger, spec.run)
		started++
	}
	if started == 0 {
		glog.Warningf(ctx, "[simuser] scheduler started with 0 goroutines trigger=%s reason=all_task_switches_disabled", trigger)
	}
	glog.Infof(ctx, "[simuser] scheduler started trigger=%s goroutines=%d skipLongStagger=%v staggerMax=%s postVideoPoll=%s maxWait=%s",
		trigger, started, skipLongStagger, stagger, flags.PostVideoPollInterval, flags.PostVideoPollMaxWait)
}

func runPeriodicTracked(parent context.Context, wg *sync.WaitGroup, name string, interval, staggerMax time.Duration, skipLongStagger bool, fn func(context.Context)) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		runPeriodic(parent, name, interval, staggerMax, skipLongStagger, fn)
	}()
}
