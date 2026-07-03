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
// logCtx 仅用于日志与读配置；goroutine 生命周期 MUST 绑定进程级 root，不得用 HTTP 请求 context。
func ReloadSchedulerFromAdmin(logCtx context.Context) {
	glog.Infof(logCtx, "[simuser] scheduler reload begin trigger=admin_config_save")
	flags := LoadRuntimeFlags(logCtx)
	globalScheduler.Reload(flags)
}

func (m *SchedulerManager) Start(parent context.Context, flags RuntimeFlags, skipLongStagger bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		glog.Infof(parent, "[simuser] scheduler stop previous goroutines")
		m.cancel()
		m.wg.Wait()
	}
	if !skipLongStagger {
		m.root = parent
	}
	schedParent := parent
	if skipLongStagger {
		schedParent = m.root
		if schedParent == nil {
			schedParent = context.Background()
			glog.Warningf(parent, "[simuser] scheduler admin reload: process root missing, fallback Background")
		}
	}
	schedCtx, cancel := context.WithCancel(schedParent)
	m.cancel = cancel
	startSchedulerGoroutines(schedCtx, &m.wg, flags, skipLongStagger)
}

func (m *SchedulerManager) Reload(flags RuntimeFlags) {
	m.Start(m.schedulerParentForReload(), flags, true)
}

// schedulerParentForReload 返回进程级 context，避免 Admin HTTP 请求结束后 cancel 调度 goroutine。
func (m *SchedulerManager) schedulerParentForReload() context.Context {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.root != nil {
		return m.root
	}
	return context.Background()
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
		{"post_debate", flags.TaskPostDebate, flags.IntervalPostDebate, func(c context.Context) { RunPostDebateTask(c, password) }},
		{"debate_comment", flags.TaskDebateComment, flags.IntervalDebateComment, func(c context.Context) { RunDebateCommentTask(c, password) }},
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
