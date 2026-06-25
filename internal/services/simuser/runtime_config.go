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
	VideoPoll     bool `json:"videoPoll"`

	IntervalRegisterSec         int64 `json:"intervalRegisterSec"`
	IntervalCommentSec          int64 `json:"intervalCommentSec"`
	IntervalPostImageSec        int64 `json:"intervalPostImageSec"`
	IntervalPostVideoSec        int64 `json:"intervalPostVideoSec"`
	IntervalChatSec             int64 `json:"intervalChatSec"`
	IntervalFollowSec           int64 `json:"intervalFollowSec"`
	IntervalVideoPollIdleSec    int64 `json:"intervalVideoPollIdleSec"`
	IntervalVideoPollActiveSec  int64 `json:"intervalVideoPollActiveSec"`
	StartupStaggerSec           int64 `json:"startupStaggerSec"`
	EphemeralChatLoopSec        int64 `json:"ephemeralChatLoopSec"`
	EphemeralChatWindowSec      int64 `json:"ephemeralChatWindowSec"`
	RateLimitRps                float64 `json:"rateLimitRps"`
	RateLimitBurst              int     `json:"rateLimitBurst"`
}

// DefaultRuntimeConfigDB 与变更前 env 默认一致。
func DefaultRuntimeConfigDB() RuntimeConfigDB {
	env := defaultRuntimeFromEnv(context.Background())
	return runtimeConfigFromFlags(env)
}

func defaultRuntimeFromEnv(ctx context.Context) RuntimeFlags {
	return RuntimeFlags{
		TaskRegister:            envBool("SIM_TASK_REGISTER_ENABLED", true),
		TaskComment:             envBool("SIM_TASK_COMMENT_ENABLED", true),
		TaskPostImage:           envBool("SIM_TASK_POST_IMAGE_ENABLED", true),
		TaskPostVideo:           envBool("SIM_TASK_POST_VIDEO_ENABLED", true),
		TaskChat:                envBool("SIM_TASK_CHAT_ENABLED", true),
		TaskFollow:              envBool("SIM_TASK_FOLLOW_ENABLED", true),
		VideoPoll:               envBool("SIM_VIDEO_POLL_ENABLED", true),
		IntervalRegister:        envDuration("SIM_INTERVAL_REGISTER", 24*time.Hour),
		IntervalComment:         envDuration("SIM_INTERVAL_COMMENT", 6*time.Hour),
		IntervalPostImage:       envDuration("SIM_INTERVAL_POST_IMAGE", 3*time.Hour+30*time.Minute),
		IntervalPostVideo:       envDuration("SIM_INTERVAL_POST_VIDEO", 6*time.Hour+30*time.Minute),
		IntervalChat:            envDuration("SIM_INTERVAL_CHAT", time.Hour),
		IntervalFollow:          envDuration("SIM_INTERVAL_FOLLOW", 7*time.Hour),
		IntervalVideoPollIdle:   envDuration("SIM_INTERVAL_VIDEO_POLL_IDLE", 10*time.Minute),
		IntervalVideoPollActive: envDuration("SIM_INTERVAL_VIDEO_POLL_ACTIVE", 2*time.Minute),
		EphemeralChatLoop:       envDuration("SIM_EPHEMERAL_CHAT_LOOP", 5*time.Minute),
		EphemeralChatWindow:     envDuration("SIM_EPHEMERAL_CHAT_WINDOW", 15*time.Minute),
		StartupStaggerMax:       envDuration("SIM_STARTUP_STAGGER_MAX", 30*time.Minute),
		DefaultPassword:         strings.TrimSpace(g.Cfg().MustGet(ctx, "simUser.defaultPassword").String()),
	}
}

func runtimeConfigFromFlags(f RuntimeFlags) RuntimeConfigDB {
	return RuntimeConfigDB{
		TaskRegister: f.TaskRegister, TaskComment: f.TaskComment,
		TaskPostImage: f.TaskPostImage, TaskPostVideo: f.TaskPostVideo,
		TaskChat: f.TaskChat, TaskFollow: f.TaskFollow, VideoPoll: f.VideoPoll,
		IntervalRegisterSec:        int64(f.IntervalRegister / time.Second),
		IntervalCommentSec:         int64(f.IntervalComment / time.Second),
		IntervalPostImageSec:       int64(f.IntervalPostImage / time.Second),
		IntervalPostVideoSec:       int64(f.IntervalPostVideo / time.Second),
		IntervalChatSec:            int64(f.IntervalChat / time.Second),
		IntervalFollowSec:          int64(f.IntervalFollow / time.Second),
		IntervalVideoPollIdleSec:   int64(f.IntervalVideoPollIdle / time.Second),
		IntervalVideoPollActiveSec: int64(f.IntervalVideoPollActive / time.Second),
		StartupStaggerSec:          int64(f.StartupStaggerMax / time.Second),
		EphemeralChatLoopSec:       int64(f.EphemeralChatLoop / time.Second),
		EphemeralChatWindowSec:     int64(f.EphemeralChatWindow / time.Second),
		RateLimitRps:               envFloat("SIM_UCG_RATE_LIMIT_RPS", 2.0),
		RateLimitBurst:             envInt("SIM_UCG_RATE_LIMIT_BURST", 4),
	}
}

func (c RuntimeConfigDB) toRuntimeFlags(ctx context.Context) RuntimeFlags {
	f := RuntimeFlags{
		TaskRegister: c.TaskRegister, TaskComment: c.TaskComment,
		TaskPostImage: c.TaskPostImage, TaskPostVideo: c.TaskPostVideo,
		TaskChat: c.TaskChat, TaskFollow: c.TaskFollow, VideoPoll: c.VideoPoll,
		IntervalRegister:        secDuration(c.IntervalRegisterSec, 24*time.Hour),
		IntervalComment:         secDuration(c.IntervalCommentSec, 6*time.Hour),
		IntervalPostImage:       secDuration(c.IntervalPostImageSec, 3*time.Hour+30*time.Minute),
		IntervalPostVideo:       secDuration(c.IntervalPostVideoSec, 6*time.Hour+30*time.Minute),
		IntervalChat:            secDuration(c.IntervalChatSec, time.Hour),
		IntervalFollow:          secDuration(c.IntervalFollowSec, 7*time.Hour),
		IntervalVideoPollIdle:   secDuration(c.IntervalVideoPollIdleSec, 10*time.Minute),
		IntervalVideoPollActive: secDuration(c.IntervalVideoPollActiveSec, 2*time.Minute),
		StartupStaggerMax:       secDuration(c.StartupStaggerSec, 30*time.Minute),
		EphemeralChatLoop:       secDuration(c.EphemeralChatLoopSec, 5*time.Minute),
		EphemeralChatWindow:     secDuration(c.EphemeralChatWindowSec, 15*time.Minute),
		DefaultPassword:         strings.TrimSpace(g.Cfg().MustGet(ctx, "simUser.defaultPassword").String()),
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
	// 合并零值字段为代码默认，避免部分 JSON 缺键
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
	if c.IntervalVideoPollIdleSec <= 0 {
		c.IntervalVideoPollIdleSec = def.IntervalVideoPollIdleSec
	}
	if c.IntervalVideoPollActiveSec <= 0 {
		c.IntervalVideoPollActiveSec = def.IntervalVideoPollActiveSec
	}
	if c.StartupStaggerSec <= 0 {
		c.StartupStaggerSec = def.StartupStaggerSec
	}
	if c.EphemeralChatLoopSec <= 0 {
		c.EphemeralChatLoopSec = def.EphemeralChatLoopSec
	}
	if c.EphemeralChatWindowSec <= 0 {
		c.EphemeralChatWindowSec = def.EphemeralChatWindowSec
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
