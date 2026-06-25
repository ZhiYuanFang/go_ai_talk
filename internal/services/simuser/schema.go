package simuser

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// EnsureSchema 创建 sim 服务自有表（幂等）。
func EnsureSchema(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS sim_config (
			id TINYINT PRIMARY KEY,
			enabled TINYINT NOT NULL DEFAULT 1,
			max_sim_users INT NOT NULL DEFAULT 100,
			runtime_json TEXT NOT NULL,
			updated_at BIGINT NOT NULL DEFAULT 0,
			updated_by VARCHAR(64) NOT NULL DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS sim_llm_lane_config (
			lane VARCHAR(32) PRIMARY KEY,
			provider VARCHAR(32) NOT NULL DEFAULT '',
			model VARCHAR(128) NOT NULL DEFAULT '',
			max_in_flight INT NOT NULL DEFAULT 1,
			max_waiters INT NOT NULL DEFAULT 0,
			timeout_sec INT NOT NULL DEFAULT 0,
			updated_at BIGINT NOT NULL DEFAULT 0,
			updated_by VARCHAR(64) NOT NULL DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS sim_prompt (
			task_type VARCHAR(32) PRIMARY KEY,
			system_prompt TEXT,
			user_prompt_template TEXT NOT NULL,
			updated_at BIGINT NOT NULL DEFAULT 0,
			updated_by VARCHAR(64) NOT NULL DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS sim_account_seq (
			id TINYINT PRIMARY KEY,
			next_seq INT NOT NULL DEFAULT 1
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS sim_video_job (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			wx_id BIGINT NOT NULL,
			content TEXT NOT NULL,
			task_id VARCHAR(128) NOT NULL,
			status VARCHAR(16) NOT NULL DEFAULT 'pending',
			created_at BIGINT NOT NULL,
			updated_at BIGINT NOT NULL,
			KEY idx_status (status),
			KEY idx_wx (wx_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS sim_task_run (
			task_name VARCHAR(32) PRIMARY KEY,
			last_run_at BIGINT NOT NULL DEFAULT 0,
			success_count BIGINT NOT NULL DEFAULT 0,
			fail_count BIGINT NOT NULL DEFAULT 0,
			last_error VARCHAR(512) NOT NULL DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, sql := range stmts {
		if _, err := g.DB().Exec(ctx, sql); err != nil {
			return err
		}
	}
	if err := ensureSimConfigRuntimeColumn(ctx); err != nil {
		return err
	}
	if err := seedDefaults(ctx); err != nil {
		return err
	}
	if err := EnsureSimLLMLaneDefaultRows(ctx); err != nil {
		return err
	}
	glog.Infof(ctx, "[simuser] schema ensured")
	return nil
}

// ensureSimConfigRuntimeColumn 迁移旧库：补 runtime_json 列。
func ensureSimConfigRuntimeColumn(ctx context.Context) error {
	_, err := g.DB().Exec(ctx, `ALTER TABLE sim_config ADD COLUMN runtime_json TEXT NOT NULL DEFAULT '' AFTER max_sim_users`)
	if err != nil && !isDuplicateColumnErr(err) {
		return err
	}
	return nil
}

func isDuplicateColumnErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Duplicate column") || strings.Contains(msg, "duplicate column name")
}

func seedDefaults(ctx context.Context) error {
	n, err := g.DB().Model("sim_config").Ctx(ctx).Where("id", 1).Count()
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	runtimeSeed, _ := json.Marshal(DefaultRuntimeConfigDB())
	if n == 0 {
		_, err = g.DB().Model("sim_config").Ctx(ctx).Data(g.Map{
			"id": 1, "enabled": 1, "max_sim_users": 100,
			"runtime_json": string(runtimeSeed),
			"updated_at": now, "updated_by": "seed",
		}).Insert()
		if err != nil {
			return err
		}
	} else {
		var row struct {
			RuntimeJSON string `json:"runtime_json"`
		}
		_ = g.DB().Model("sim_config").Ctx(ctx).Fields("runtime_json").Where("id", 1).Scan(&row)
		if row.RuntimeJSON == "" || row.RuntimeJSON == "{}" {
			_, _ = g.DB().Model("sim_config").Ctx(ctx).Where("id", 1).Data(g.Map{
				"runtime_json": string(runtimeSeed),
				"updated_at":   now,
			}).Update()
		}
	}
	n, err = g.DB().Model("sim_account_seq").Ctx(ctx).Where("id", 1).Count()
	if err != nil {
		return err
	}
	if n == 0 {
		_, err = g.DB().Model("sim_account_seq").Ctx(ctx).Data(g.Map{"id": 1, "next_seq": 1}).Insert()
		if err != nil {
			return err
		}
	}
	for taskType, tmpl := range defaultPrompts() {
		c, cErr := g.DB().Model("sim_prompt").Ctx(ctx).Where("task_type", taskType).Count()
		if cErr != nil {
			return cErr
		}
		if c > 0 {
			continue
		}
		_, err = g.DB().Model("sim_prompt").Ctx(ctx).Data(g.Map{
			"task_type":            taskType,
			"system_prompt":        tmpl.system,
			"user_prompt_template": tmpl.user,
			"updated_at":           now,
			"updated_by":           "seed",
		}).Insert()
		if err != nil {
			return err
		}
	}
	return nil
}

type promptSeed struct {
	system string
	user   string
}

func defaultPrompts() map[string]promptSeed {
	return map[string]promptSeed{
		"register_nickname": {user: "我是一位宝妈，帮我想一个昵称，只输出昵称本身，不要解释。"},
		"register_avatar":   {user: "我是一位宝妈，帮我生成一个头像"},
		"comment":           {user: "作为宝妈，请结合帖子内容和图片，写一条简短真实的评论。\n帖子：{{post_content}}"},
		"post_image_text":   {user: "从母婴专家的角度，写一条适合发朋友圈的短文，100字以内。\n主题：{{topic}}"},
		"post_video_text":   {user: "从母婴专家的角度，写一条适合拍短视频的口播文案，80字以内。\n主题：{{topic}}"},
		"chat_reply":        {user: "作为宝妈，根据聊天记录用自然口语回复对方。\n历史：\n{{chat_history}}"},
	}
}
