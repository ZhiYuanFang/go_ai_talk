package simuser

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ConfigDTO 模拟服务全局配置。
type ConfigDTO struct {
	Enabled     bool  `json:"enabled"`
	MaxSimUsers int   `json:"maxSimUsers"`
	UpdatedAt   int64 `json:"updatedAt"`
	UpdatedBy   string `json:"updatedBy"`
}

// PromptDTO Prompt 模板。
type PromptDTO struct {
	TaskType           string `json:"taskType"`
	SystemPrompt       string `json:"systemPrompt"`
	UserPromptTemplate string `json:"userPromptTemplate"`
	UpdatedAt          int64  `json:"updatedAt"`
	UpdatedBy          string `json:"updatedBy"`
}

// StatusDTO 任务运行状态。
type StatusDTO struct {
	Tasks           []TaskRunDTO `json:"tasks"`
	PendingVideoJobs int         `json:"pendingVideoJobs"`
}

type TaskRunDTO struct {
	TaskName     string `json:"taskName"`
	LastRunAt    int64  `json:"lastRunAt"`
	SuccessCount int64  `json:"successCount"`
	FailCount    int64  `json:"failCount"`
	LastError    string `json:"lastError,omitempty"`
}

func GetConfig(ctx context.Context) (ConfigDTO, error) {
	var row struct {
		Enabled     int    `json:"enabled"`
		MaxSimUsers int    `json:"max_sim_users"`
		UpdatedAt   int64  `json:"updated_at"`
		UpdatedBy   string `json:"updated_by"`
	}
	err := g.DB().Model("sim_config").Ctx(ctx).Where("id", 1).Scan(&row)
	if err != nil {
		return ConfigDTO{}, err
	}
	if row.MaxSimUsers <= 0 {
		row.MaxSimUsers = 100
	}
	return ConfigDTO{
		Enabled: row.Enabled == 1, MaxSimUsers: row.MaxSimUsers,
		UpdatedAt: row.UpdatedAt, UpdatedBy: row.UpdatedBy,
	}, nil
}

func UpdateConfig(ctx context.Context, enabled bool, maxSimUsers int, updatedBy string) error {
	if maxSimUsers <= 0 {
		maxSimUsers = 100
	}
	en := 0
	if enabled {
		en = 1
	}
	now := time.Now().Unix()
	_, err := g.DB().Model("sim_config").Ctx(ctx).Data(g.Map{
		"id": 1, "enabled": en, "max_sim_users": maxSimUsers,
		"updated_at": now, "updated_by": strings.TrimSpace(updatedBy),
	}).Save()
	return err
}

func GetPrompt(ctx context.Context, taskType string) (PromptDTO, error) {
	var row struct {
		SystemPrompt       string `json:"system_prompt"`
		UserPromptTemplate string `json:"user_prompt_template"`
		UpdatedAt          int64  `json:"updated_at"`
		UpdatedBy          string `json:"updated_by"`
	}
	err := g.DB().Model("sim_prompt").Ctx(ctx).Where("task_type", taskType).Scan(&row)
	if err != nil {
		return PromptDTO{}, err
	}
	return PromptDTO{
		TaskType: taskType, SystemPrompt: row.SystemPrompt,
		UserPromptTemplate: row.UserPromptTemplate,
		UpdatedAt: row.UpdatedAt, UpdatedBy: row.UpdatedBy,
	}, nil
}

func UpdatePrompt(ctx context.Context, taskType, systemPrompt, userTemplate, updatedBy string) error {
	now := time.Now().Unix()
	_, err := g.DB().Model("sim_prompt").Ctx(ctx).Data(g.Map{
		"task_type": taskType, "system_prompt": systemPrompt,
		"user_prompt_template": userTemplate, "updated_at": now, "updated_by": strings.TrimSpace(updatedBy),
	}).Save()
	return err
}

func ListPromptTypes() []string {
	return []string{"register_nickname", "register_avatar", "comment", "post_image_text", "post_video_text", "chat_reply"}
}

func NextAccountName(ctx context.Context) (string, error) {
	var seq int
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var row struct{ NextSeq int `json:"nextSeq"` }
		if err := tx.Model("sim_account_seq").Where("id", 1).Scan(&row); err != nil {
			return err
		}
		seq = row.NextSeq
		if seq <= 0 {
			seq = 1
		}
		_, err := tx.Model("sim_account_seq").Where("id", 1).Data(g.Map{"next_seq": seq + 1}).Update()
		return err
	})
	if err != nil {
		return "", err
	}
	return formatPtestAccount(seq), nil
}

func formatPtestAccount(n int) string {
	return "ptest" + itoa(n)
}

func itoa(n int) string {
	if n <= 0 {
		return "1"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func RecordTaskRun(ctx context.Context, name string, success bool, errMsg string) {
	now := time.Now().Unix()
	n, _ := g.DB().Model("sim_task_run").Ctx(ctx).Where("task_name", name).Count()
	data := g.Map{"task_name": name, "last_run_at": now, "last_error": truncateErr(errMsg)}
	if n == 0 {
		data["success_count"] = 0
		data["fail_count"] = 0
		if success {
			data["success_count"] = 1
		} else {
			data["fail_count"] = 1
		}
		_, _ = g.DB().Model("sim_task_run").Ctx(ctx).Data(data).Insert()
		return
	}
	col := "success_count"
	if !success {
		col = "fail_count"
	}
	_, _ = g.DB().Model("sim_task_run").Ctx(ctx).Where("task_name", name).Data(g.Map{
		"last_run_at": now, "last_error": truncateErr(errMsg),
	}).Update()
	_, _ = g.DB().Model("sim_task_run").Ctx(ctx).Where("task_name", name).Increment(col, 1)
}

func GetStatus(ctx context.Context) (StatusDTO, error) {
	var runs []TaskRunDTO
	_ = g.DB().Model("sim_task_run").Ctx(ctx).Scan(&runs)
	pending, _ := g.DB().Model("sim_video_job").Ctx(ctx).
		WhereIn("status", []string{"pending", "processing"}).Count()
	return StatusDTO{Tasks: runs, PendingVideoJobs: pending}, nil
}

func InsertVideoJob(ctx context.Context, wxID int64, content, taskID string) (int64, error) {
	now := time.Now().Unix()
	id, err := g.DB().Model("sim_video_job").Ctx(ctx).Data(g.Map{
		"wx_id": wxID, "content": content, "task_id": taskID,
		"status": "pending", "created_at": now, "updated_at": now,
	}).InsertAndGetId()
	return id, err
}

// DiscardPendingVideoJobsOnStartup 进程启动时将未完成的视频 job 标为 skipped（模拟场景可接受丢失）。
func DiscardPendingVideoJobsOnStartup(ctx context.Context) {
	_, _ = g.DB().Model("sim_video_job").Ctx(ctx).
		WhereIn("status", []string{"pending", "processing"}).
		Data(g.Map{"status": "skipped", "updated_at": time.Now().Unix()}).
		Update()
}

func HasPendingVideoJob(ctx context.Context, wxID int64) (bool, error) {
	n, err := g.DB().Model("sim_video_job").Ctx(ctx).
		Where("wx_id", wxID).WhereIn("status", []string{"pending", "processing"}).Count()
	return n > 0, err
}

func ListPendingVideoJobs(ctx context.Context) ([]videoJobRow, error) {
	var rows []videoJobRow
	err := g.DB().Model("sim_video_job").Ctx(ctx).
		WhereIn("status", []string{"pending", "processing"}).
		OrderAsc("id").Scan(&rows)
	return rows, err
}

type videoJobRow struct {
	Id      int64  `json:"id"`
	WxId    int64  `json:"wx_id"`
	Content string `json:"content"`
	TaskId  string `json:"task_id"`
	Status  string `json:"status"`
}

func UpdateVideoJobStatus(ctx context.Context, id int64, status string) error {
	_, err := g.DB().Model("sim_video_job").Ctx(ctx).Where("id", id).Data(g.Map{
		"status": status, "updated_at": time.Now().Unix(),
	}).Update()
	return err
}

func truncateErr(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 500 {
		return s[:500]
	}
	return s
}
