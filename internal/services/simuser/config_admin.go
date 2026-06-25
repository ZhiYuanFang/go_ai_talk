package simuser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// FullConfigDTO 含 DB 业务开关、上限与运行时 JSON。
type FullConfigDTO struct {
	Enabled     bool            `json:"enabled"`
	MaxSimUsers int             `json:"maxSimUsers"`
	Runtime     RuntimeConfigDB `json:"runtime"`
	UpdatedAt   int64           `json:"updatedAt"`
	UpdatedBy   string          `json:"updatedBy"`
}

// ConfigEffect 保存后的生效提示项。
type ConfigEffect struct {
	Kind    string `json:"kind"`
	Task    string `json:"task,omitempty"`
	Message string `json:"message,omitempty"`
}

// TaskScheduleItem 各任务下一跑估算。
type TaskScheduleItem struct {
	Name         string `json:"name"`
	Label        string `json:"label"`
	Enabled      bool   `json:"enabled"`
	IntervalSec  int64  `json:"intervalSec"`
	LastRunAt    int64  `json:"lastRunAt,omitempty"`
	NextRunHint  string `json:"nextRunHint"`
}

// ConfigPutResult PUT config 扩展响应。
type ConfigPutResult struct {
	ScheduleReloaded bool               `json:"scheduleReloaded"`
	Effects          []ConfigEffect     `json:"effects"`
	TaskSchedule     []TaskScheduleItem `json:"taskSchedule"`
}

// ConfigAdminPutDTO Admin PUT 完整 body。
type ConfigAdminPutDTO struct {
	Enabled     bool
	MaxSimUsers int
	Runtime     RuntimeConfigDB
}

// GetFullConfig 读取 sim_config 全量（含 runtime_json）。
func GetFullConfig(ctx context.Context) (FullConfigDTO, error) {
	var row struct {
		Enabled     int    `json:"enabled"`
		MaxSimUsers int    `json:"max_sim_users"`
		RuntimeJSON string `json:"runtime_json"`
		UpdatedAt   int64  `json:"updated_at"`
		UpdatedBy   string `json:"updated_by"`
	}
	err := g.DB().Model("sim_config").Ctx(ctx).Where("id", 1).Scan(&row)
	if err != nil {
		return FullConfigDTO{}, err
	}
	rt, _ := LoadRuntimeFromDB(ctx)
	if strings.TrimSpace(row.RuntimeJSON) != "" && row.RuntimeJSON != "{}" {
		_ = json.Unmarshal([]byte(row.RuntimeJSON), &rt)
	}
	if row.MaxSimUsers <= 0 {
		row.MaxSimUsers = 100
	}
	return FullConfigDTO{
		Enabled: row.Enabled == 1, MaxSimUsers: row.MaxSimUsers,
		Runtime: rt, UpdatedAt: row.UpdatedAt, UpdatedBy: row.UpdatedBy,
	}, nil
}

// UpdateConfigAdmin 持久化配置并按 diff 决定是否 reload scheduler。
func UpdateConfigAdmin(ctx context.Context, req ConfigAdminPutDTO, updatedBy string) (ConfigPutResult, error) {
	before, err := GetFullConfig(ctx)
	if err != nil {
		return ConfigPutResult{}, err
	}
	if req.MaxSimUsers <= 0 {
		req.MaxSimUsers = 100
	}
	en := 0
	if req.Enabled {
		en = 1
	}
	raw, err := json.Marshal(req.Runtime)
	if err != nil {
		return ConfigPutResult{}, err
	}
	now := time.Now().Unix()
	operator := strings.TrimSpace(updatedBy)
	if operator == "" {
		operator = "admin"
	}
	_, err = g.DB().Model("sim_config").Ctx(ctx).Data(g.Map{
		"id": 1, "enabled": en, "max_sim_users": req.MaxSimUsers,
		"runtime_json": string(raw),
		"updated_at":   now, "updated_by": operator,
	}).Save()
	if err != nil {
		return ConfigPutResult{}, err
	}

	scheduleChanged := runtimeScheduleDiff(before, req)
	result := ConfigPutResult{
		ScheduleReloaded: scheduleChanged,
		Effects:          buildConfigEffects(before, req, scheduleChanged),
	}

	if scheduleChanged {
		setActiveRateLimit(RateLimitSettings{RPS: req.Runtime.RateLimitRps, Burst: req.Runtime.RateLimitBurst})
		ReloadSchedulerFromAdmin(ctx)
		result.TaskSchedule = buildTaskSchedule(ctx, req.Runtime)
	} else {
		result.TaskSchedule = buildTaskSchedule(ctx, before.Runtime)
		if req.Runtime.RateLimitRps > 0 && (req.Runtime.RateLimitRps != before.Runtime.RateLimitRps || req.Runtime.RateLimitBurst != before.Runtime.RateLimitBurst) {
			setActiveRateLimit(RateLimitSettings{RPS: req.Runtime.RateLimitRps, Burst: req.Runtime.RateLimitBurst})
		}
	}
	return result, nil
}

func runtimeScheduleDiff(before FullConfigDTO, after ConfigAdminPutDTO) bool {
	if before.Enabled != after.Enabled {
		return true
	}
	br, ar := before.Runtime, after.Runtime
	return br.TaskRegister != ar.TaskRegister || br.TaskComment != ar.TaskComment ||
		br.TaskPostImage != ar.TaskPostImage || br.TaskPostVideo != ar.TaskPostVideo ||
		br.TaskChat != ar.TaskChat || br.TaskFollow != ar.TaskFollow || br.VideoPoll != ar.VideoPoll ||
		br.IntervalRegisterSec != ar.IntervalRegisterSec || br.IntervalCommentSec != ar.IntervalCommentSec ||
		br.IntervalPostImageSec != ar.IntervalPostImageSec || br.IntervalPostVideoSec != ar.IntervalPostVideoSec ||
		br.IntervalChatSec != ar.IntervalChatSec || br.IntervalFollowSec != ar.IntervalFollowSec ||
		br.IntervalVideoPollIdleSec != ar.IntervalVideoPollIdleSec || br.IntervalVideoPollActiveSec != ar.IntervalVideoPollActiveSec ||
		br.StartupStaggerSec != ar.StartupStaggerSec ||
		br.EphemeralChatLoopSec != ar.EphemeralChatLoopSec || br.EphemeralChatWindowSec != ar.EphemeralChatWindowSec ||
		br.RateLimitRps != ar.RateLimitRps || br.RateLimitBurst != ar.RateLimitBurst
}

func buildConfigEffects(before FullConfigDTO, after ConfigAdminPutDTO, reloaded bool) []ConfigEffect {
	var effects []ConfigEffect
	if reloaded {
		effects = append(effects, ConfigEffect{
			Kind: "scheduler_reloaded", Message: "调度器已重启，新周期将在短延迟后生效",
		})
	} else {
		effects = append(effects, ConfigEffect{
			Kind: "config_saved", Message: "配置已保存（未触发调度器全量重启）",
		})
	}
	if before.Runtime.TaskChat && !after.Runtime.TaskChat {
		effects = append(effects, ConfigEffect{
			Kind: "task_disabled", Task: "chat_scan", Message: "T5 聊天扫描已关闭，进行中的 E1 会话可能仍会继续至窗口结束",
		})
	}
	if before.Runtime.EphemeralChatLoopSec != after.Runtime.EphemeralChatLoopSec ||
		before.Runtime.EphemeralChatWindowSec != after.Runtime.EphemeralChatWindowSec {
		effects = append(effects, ConfigEffect{
			Kind: "ephemeral_may_continue", Message: "E1 参数已更新，已启动的临时聊天 goroutine 可能仍使用旧窗口",
		})
	}
	return effects
}

var taskScheduleDefs = []struct {
	name    string
	label   string
	enabled func(RuntimeConfigDB) bool
	interval func(RuntimeConfigDB) int64
}{
	{"register", "T1 注册", func(r RuntimeConfigDB) bool { return r.TaskRegister }, func(r RuntimeConfigDB) int64 { return r.IntervalRegisterSec }},
	{"comment", "T2 评论", func(r RuntimeConfigDB) bool { return r.TaskComment }, func(r RuntimeConfigDB) int64 { return r.IntervalCommentSec }},
	{"post_image", "T3 图文", func(r RuntimeConfigDB) bool { return r.TaskPostImage }, func(r RuntimeConfigDB) int64 { return r.IntervalPostImageSec }},
	{"post_video_submit", "T4 视频", func(r RuntimeConfigDB) bool { return r.TaskPostVideo }, func(r RuntimeConfigDB) int64 { return r.IntervalPostVideoSec }},
	{"chat_scan", "T5 聊天", func(r RuntimeConfigDB) bool { return r.TaskChat }, func(r RuntimeConfigDB) int64 { return r.IntervalChatSec }},
	{"follow", "T6 关注", func(r RuntimeConfigDB) bool { return r.TaskFollow }, func(r RuntimeConfigDB) int64 { return r.IntervalFollowSec }},
	{"video_poll", "P1 视频轮询", func(r RuntimeConfigDB) bool { return r.VideoPoll }, func(r RuntimeConfigDB) int64 { return r.IntervalVideoPollIdleSec }},
}

func buildTaskSchedule(ctx context.Context, rt RuntimeConfigDB) []TaskScheduleItem {
	lastRuns := map[string]int64{}
	var rows []struct {
		TaskName  string `json:"task_name"`
		LastRunAt int64  `json:"last_run_at"`
	}
	_ = g.DB().Model("sim_task_run").Ctx(ctx).Scan(&rows)
	for _, row := range rows {
		lastRuns[row.TaskName] = row.LastRunAt
	}
	out := make([]TaskScheduleItem, 0, len(taskScheduleDefs))
	for _, def := range taskScheduleDefs {
		enabled := def.enabled(rt)
		intervalSec := def.interval(rt)
		item := TaskScheduleItem{
			Name: def.name, Label: def.label, Enabled: enabled, IntervalSec: intervalSec,
			LastRunAt: lastRuns[def.name],
		}
		item.NextRunHint = nextRunHint(enabled, intervalSec, lastRuns[def.name])
		out = append(out, item)
	}
	return out
}

func nextRunHint(enabled bool, intervalSec, lastRunAt int64) string {
	if !enabled {
		return "已关闭"
	}
	if intervalSec <= 0 {
		return "周期未配置"
	}
	if lastRunAt <= 0 {
		return fmt.Sprintf("重启后首轮约 %s 内（含短错峰）", durationLabel(time.Duration(intervalSec)*time.Second))
	}
	eta := time.Unix(lastRunAt, 0).Add(time.Duration(intervalSec) * time.Second)
	if eta.Before(time.Now()) {
		return "已过预计时间，等待当前 tick 或 reload 后首轮"
	}
	return fmt.Sprintf("约 %s 后", durationLabel(time.Until(eta)))
}
