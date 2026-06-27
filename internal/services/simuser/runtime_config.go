package simuser

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// RuntimeConfigDB 持久化于 sim_config.runtime_json 的运行时项（秒级周期）。
type RuntimeConfigDB struct {
	TaskRegister  bool `json:"taskRegister"`
	TaskComment   bool `json:"taskComment"`
	TaskPostImage bool `json:"taskPostImage"`
	TaskPostVideo bool `json:"taskPostVideo"`
	TaskChat      bool `json:"taskChat"`
	TaskFollow    bool `json:"taskFollow"`

	IntervalRegisterSec             int64 `json:"intervalRegisterSec"`
	IntervalCommentSec              int64 `json:"intervalCommentSec"`
	IntervalPostImageSec            int64 `json:"intervalPostImageSec"`
	IntervalPostVideoSec            int64 `json:"intervalPostVideoSec"`
	IntervalChatSec                 int64 `json:"intervalChatSec"`
	IntervalFollowSec               int64 `json:"intervalFollowSec"`
	IntervalPostVideoPollSec        int64 `json:"intervalPostVideoPollSec"`
	IntervalPostVideoPollMaxWaitSec int64 `json:"intervalPostVideoPollMaxWaitSec"`
	StartupStaggerSec               int64 `json:"startupStaggerSec"`
	RateLimitRps                    float64 `json:"rateLimitRps"`
	RateLimitBurst                  int     `json:"rateLimitBurst"`

	// 迁移期：读旧 runtime_json 中的 P1 active 间隔，不再写入新配置。
	IntervalVideoPollActiveSec int64 `json:"intervalVideoPollActiveSec,omitempty"`
}

// DefaultRuntimeConfigDB 与 env 默认一致。
func DefaultRuntimeConfigDB() RuntimeConfigDB {
	env := defaultRuntimeFromEnv(context.Background())
	return runtimeConfigFromFlags(env)
}

func defaultRuntimeFromEnv(ctx context.Context) RuntimeFlags {
	return RuntimeFlags{
		TaskRegister:          envBool("SIM_TASK_REGISTER_ENABLED", true),
		TaskComment:           envBool("SIM_TASK_COMMENT_ENABLED", true),
		TaskPostImage:         envBool("SIM_TASK_POST_IMAGE_ENABLED", true),
		TaskPostVideo:         envBool("SIM_TASK_POST_VIDEO_ENABLED", true),
		TaskChat:              envBool("SIM_TASK_CHAT_ENABLED", true),
		TaskFollow:            envBool("SIM_TASK_FOLLOW_ENABLED", true),
		IntervalRegister:      envDuration("SIM_INTERVAL_REGISTER", 24*time.Hour),
		IntervalComment:       envDuration("SIM_INTERVAL_COMMENT", 6*time.Hour),
		IntervalPostImage:     envDuration("SIM_INTERVAL_POST_IMAGE", 3*time.Hour+30*time.Minute),
		IntervalPostVideo:     envDuration("SIM_INTERVAL_POST_VIDEO", 6*time.Hour+30*time.Minute),
		IntervalChat:          envDuration("SIM_INTERVAL_CHAT", time.Hour),
		IntervalFollow:        envDuration("SIM_INTERVAL_FOLLOW", 7*time.Hour),
		PostVideoPollInterval: envDuration("SIM_POST_VIDEO_POLL_INTERVAL", 2*time.Minute),
		PostVideoPollMaxWait:  envDuration("SIM_POST_VIDEO_POLL_MAX_WAIT", 30*time.Minute),
		StartupStaggerMax:     envDuration("SIM_STARTUP_STAGGER_MAX", 30*time.Minute),
		DefaultPassword:       strings.TrimSpace(g.Cfg().MustGet(ctx, "simUser.defaultPassword").String()),
	}
}

func runtimeConfigFromFlags(f RuntimeFlags) RuntimeConfigDB {
	return RuntimeConfigDB{
		TaskRegister: f.TaskRegister, TaskComment: f.TaskComment,
		TaskPostImage: f.TaskPostImage, TaskPostVideo: f.TaskPostVideo,
		TaskChat: f.TaskChat, TaskFollow: f.TaskFollow,
		IntervalRegisterSec:             int64(f.IntervalRegister / time.Second),
		IntervalCommentSec:              int64(f.IntervalComment / time.Second),
		IntervalPostImageSec:            int64(f.IntervalPostImage / time.Second),
		IntervalPostVideoSec:            int64(f.IntervalPostVideo / time.Second),
		IntervalChatSec:                 int64(f.IntervalChat / time.Second),
		IntervalFollowSec:               int64(f.IntervalFollow / time.Second),
		IntervalPostVideoPollSec:        int64(f.PostVideoPollInterval / time.Second),
		IntervalPostVideoPollMaxWaitSec: int64(f.PostVideoPollMaxWait / time.Second),
		StartupStaggerSec:               int64(f.StartupStaggerMax / time.Second),
		RateLimitRps:                    envFloat("SIM_UCG_RATE_LIMIT_RPS", 2.0),
		RateLimitBurst:                  envInt("SIM_UCG_RATE_LIMIT_BURST", 4),
	}
}

func (c RuntimeConfigDB) toRuntimeFlags(ctx context.Context) RuntimeFlags {
	f := RuntimeFlags{
		TaskRegister:          c.TaskRegister,
		TaskComment:           c.TaskComment,
		TaskPostImage:         c.TaskPostImage,
		TaskPostVideo:         c.TaskPostVideo,
		TaskChat:              c.TaskChat,
		TaskFollow:            c.TaskFollow,
		IntervalRegister:      secDuration(c.IntervalRegisterSec, 24*time.Hour),
		IntervalComment:       secDuration(c.IntervalCommentSec, 6*time.Hour),
		IntervalPostImage:     secDuration(c.IntervalPostImageSec, 3*time.Hour+30*time.Minute),
		IntervalPostVideo:     secDuration(c.IntervalPostVideoSec, 6*time.Hour+30*time.Minute),
		IntervalChat:          secDuration(c.IntervalChatSec, time.Hour),
		IntervalFollow:        secDuration(c.IntervalFollowSec, 7*time.Hour),
		PostVideoPollInterval: secDuration(c.IntervalPostVideoPollSec, 2*time.Minute),
		PostVideoPollMaxWait:  secDuration(c.IntervalPostVideoPollMaxWaitSec, 30*time.Minute),
		StartupStaggerMax:     secDuration(c.StartupStaggerSec, 30*time.Minute),
		DefaultPassword:       strings.TrimSpace(g.Cfg().MustGet(ctx, "simUser.defaultPassword").String()),
	}
	if c.RateLimitRps > 0 {
		setActiveRateLimit(RateLimitSettings{RPS: c.RateLimitRps, Burst: c.RateLimitBurst})
	}
	return f
}

func secDuration(sec int64, def time.Duration) time.Duration {
	if sec <= 0 {
		return def
	}
	return time.Duration(sec) * time.Second
}

// LoadRuntimeFromDB 读取 sim_config.runtime_json；空或缺失时回退 env 默认（不写 DB）。
func LoadRuntimeFromDB(ctx context.Context) (RuntimeConfigDB, error) {
	raw, err := loadRuntimeJSONRaw(ctx)
	if err != nil {
		return RuntimeConfigDB{}, err
	}
	if strings.TrimSpace(raw) == "" || raw == "{}" {
		return DefaultRuntimeConfigDB(), nil
	}
	var c RuntimeConfigDB
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return DefaultRuntimeConfigDB(), nil
	}
	def := DefaultRuntimeConfigDB()
	if c.IntervalRegisterSec <= 0 {
		c.IntervalRegisterSec = def.IntervalRegisterSec
	}
	if c.IntervalCommentSec <= 0 {
		c.IntervalCommentSec = def.IntervalCommentSec
	}
	if c.IntervalPostImageSec <= 0 {
		c.IntervalPostImageSec = def.IntervalPostImageSec
	}
	if c.IntervalPostVideoSec <= 0 {
		c.IntervalPostVideoSec = def.IntervalPostVideoSec
	}
	if c.IntervalChatSec <= 0 {
		c.IntervalChatSec = def.IntervalChatSec
	}
	if c.IntervalFollowSec <= 0 {
		c.IntervalFollowSec = def.IntervalFollowSec
	}
	if c.IntervalPostVideoPollSec <= 0 {
		if c.IntervalVideoPollActiveSec > 0 {
			c.IntervalPostVideoPollSec = c.IntervalVideoPollActiveSec
		} else {
			c.IntervalPostVideoPollSec = def.IntervalPostVideoPollSec
		}
	}
	if c.IntervalPostVideoPollMaxWaitSec <= 0 {
		c.IntervalPostVideoPollMaxWaitSec = def.IntervalPostVideoPollMaxWaitSec
	}
	if c.StartupStaggerSec <= 0 {
		c.StartupStaggerSec = def.StartupStaggerSec
	}
	if c.RateLimitRps <= 0 {
		c.RateLimitRps = def.RateLimitRps
	}
	if c.RateLimitBurst <= 0 {
		c.RateLimitBurst = def.RateLimitBurst
	}
	return c, nil
}

func loadRuntimeJSONRaw(ctx context.Context) (string, error) {
	var row struct {
		RuntimeJSON string `json:"runtime_json"`
	}
	err := g.DB().Model("sim_config").Ctx(ctx).Fields("runtime_json").Where("id", 1).Scan(&row)
	if err != nil {
		return "", err
	}
	return row.RuntimeJSON, nil
}

// ParseDurationInput 解析 Admin 提交的周期字符串（如 6h、3h30m）。
func ParseDurationInput(s string, def time.Duration) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return def, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return 0, err
	}
	return d, nil
}
