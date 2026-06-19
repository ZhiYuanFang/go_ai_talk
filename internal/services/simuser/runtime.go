package simuser

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// RuntimeFlags 环境开关与任务周期（周期可由 SIM_INTERVAL_* 覆盖；管理页不可改周期）。
type RuntimeFlags struct {
	Enabled              bool
	TaskRegister         bool
	TaskComment          bool
	TaskPostImage        bool
	TaskPostVideo        bool
	TaskChat             bool
	TaskFollow           bool
	VideoPoll            bool
	IntervalRegister     time.Duration
	IntervalComment      time.Duration
	IntervalPostImage    time.Duration
	IntervalPostVideo    time.Duration
	IntervalChat         time.Duration
	IntervalFollow       time.Duration
	IntervalVideoPollIdle   time.Duration
	IntervalVideoPollActive time.Duration
	EphemeralChatLoop    time.Duration
	EphemeralChatWindow  time.Duration
	StartupStaggerMax    time.Duration
	DefaultPassword      string
}

// LoadRuntimeFlags 读取 sim-user-service 运行时开关与周期（未设置 env 时与变更前默认值一致）。
func LoadRuntimeFlags(ctx context.Context) RuntimeFlags {
	return RuntimeFlags{
		Enabled:           envBool("SIM_USER_SERVICE_ENABLED", false),
		TaskRegister:      envBool("SIM_TASK_REGISTER_ENABLED", true),
		TaskComment:       envBool("SIM_TASK_COMMENT_ENABLED", true),
		TaskPostImage:     envBool("SIM_TASK_POST_IMAGE_ENABLED", true),
		TaskPostVideo:     envBool("SIM_TASK_POST_VIDEO_ENABLED", true),
		TaskChat:          envBool("SIM_TASK_CHAT_ENABLED", true),
		TaskFollow:        envBool("SIM_TASK_FOLLOW_ENABLED", true),
		VideoPoll:         envBool("SIM_VIDEO_POLL_ENABLED", true),
		IntervalRegister:  envDuration("SIM_INTERVAL_REGISTER", 24*time.Hour),
		IntervalComment:   envDuration("SIM_INTERVAL_COMMENT", 6*time.Hour),
		IntervalPostImage: envDuration("SIM_INTERVAL_POST_IMAGE", 3*time.Hour+30*time.Minute),
		IntervalPostVideo: envDuration("SIM_INTERVAL_POST_VIDEO", 6*time.Hour+30*time.Minute),
		IntervalChat:      envDuration("SIM_INTERVAL_CHAT", time.Hour),
		IntervalFollow:    envDuration("SIM_INTERVAL_FOLLOW", 7*time.Hour),
		IntervalVideoPollIdle:   envDuration("SIM_INTERVAL_VIDEO_POLL_IDLE", 10*time.Minute),
		IntervalVideoPollActive: envDuration("SIM_INTERVAL_VIDEO_POLL_ACTIVE", 2*time.Minute),
		EphemeralChatLoop:   envDuration("SIM_EPHEMERAL_CHAT_LOOP", 5*time.Minute),
		EphemeralChatWindow: envDuration("SIM_EPHEMERAL_CHAT_WINDOW", 15*time.Minute),
		StartupStaggerMax:   envDuration("SIM_STARTUP_STAGGER_MAX", 30*time.Minute),
		DefaultPassword:     strings.TrimSpace(g.Cfg().MustGet(ctx, "simUser.defaultPassword").String()),
	}
}

func envBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

func envDuration(key string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return def
	}
	return d
}

// Jittered 为周期添加 ±10% 随机偏移。
func Jittered(base time.Duration) time.Duration {
	if base <= 0 {
		return base
	}
	j := float64(base) * 0.1
	n := time.Now().UnixNano() % int64(j*2)
	return base - time.Duration(j) + time.Duration(n)
}
