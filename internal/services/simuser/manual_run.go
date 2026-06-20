package simuser

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/os/gctx"
)

type manualRunCtxKey struct{}

// RunnableTaskNames 管理页可手动触发的任务名（与 sim_task_run.task_name、scheduler 一致）。
func RunnableTaskNames() []string {
	return []string{
		"register", "comment", "post_image", "post_video_submit",
		"chat_scan", "follow", "video_poll",
	}
}

// WithManualRun 标记为管理页手动触发（跳过 DB enabled / env 任务开关语义）。
func WithManualRun(ctx context.Context) context.Context {
	return context.WithValue(ctx, manualRunCtxKey{}, true)
}

func isManualRun(ctx context.Context) bool {
	v, _ := ctx.Value(manualRunCtxKey{}).(bool)
	return v
}

var (
	manualRunMu sync.Mutex
	manualBusy  = map[string]bool{}
)

// TryStartManualRun 尝试占用任务执行槽；已占用返回 false。
func TryStartManualRun(taskName string) bool {
	taskName = normalizeRunnableTask(taskName)
	if taskName == "" {
		return false
	}
	manualRunMu.Lock()
	defer manualRunMu.Unlock()
	if manualBusy[taskName] {
		return false
	}
	manualBusy[taskName] = true
	return true
}

// EndManualRun 释放手动任务占用。
func EndManualRun(taskName string) {
	taskName = normalizeRunnableTask(taskName)
	if taskName == "" {
		return
	}
	manualRunMu.Lock()
	delete(manualBusy, taskName)
	manualRunMu.Unlock()
}

// IsManualRunBusy 查询任务是否正在手动执行。
func IsManualRunBusy(taskName string) bool {
	taskName = normalizeRunnableTask(taskName)
	manualRunMu.Lock()
	defer manualRunMu.Unlock()
	return manualBusy[taskName]
}

// NormalizeRunnableTaskPublic 校验任务名是否可手动触发（供 controller 使用）。
func NormalizeRunnableTaskPublic(name string) string {
	return normalizeRunnableTask(name)
}

func normalizeRunnableTask(name string) string {
	name = strings.TrimSpace(name)
	for _, n := range RunnableTaskNames() {
		if n == name {
			return name
		}
	}
	return ""
}

// RunTaskByName 执行指定周期任务一次（供 scheduler 与 admin 手动触发共用）。
func RunTaskByName(ctx context.Context, taskName string, flags RuntimeFlags) error {
	taskName = normalizeRunnableTask(taskName)
	if taskName == "" {
		return fmt.Errorf("未知任务: %s", taskName)
	}
	password := flags.DefaultPassword
	if password == "" {
		password = "123456"
	}
	switch taskName {
	case "register":
		RunRegisterTask(ctx, password)
	case "comment":
		RunCommentTask(ctx, password)
	case "post_image":
		RunPostImageTask(ctx, password)
	case "post_video_submit":
		RunPostVideoSubmitTask(ctx, password)
	case "chat_scan":
		RunChatScanTask(ctx, password, flags)
	case "follow":
		RunFollowTask(ctx, password)
	case "video_poll":
		RunVideoPollTask(ctx, password)
	default:
		return fmt.Errorf("未知任务: %s", taskName)
	}
	return nil
}

// StartManualRunAsync 后台执行一次手动任务（HTTP 立即返回）。
func StartManualRunAsync(taskName string, flags RuntimeFlags) bool {
	if !TryStartManualRun(taskName) {
		return false
	}
	go func() {
		defer EndManualRun(taskName)
		ctx := WithManualRun(gctx.New())
		_ = RunTaskByName(ctx, taskName, flags)
	}()
	return true
}
