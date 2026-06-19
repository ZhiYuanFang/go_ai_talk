package simuser

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// RuntimeFlags 环境开关与任务周期（周期写死，管理页不可改）。
type RuntimeFlags struct {
	Enabled           bool
	TaskRegister      bool
	TaskComment       bool
	TaskPostImage     bool
	TaskPostVideo     bool
	TaskChat          bool
	TaskFollow        bool
	VideoPoll         bool
	IntervalRegister  time.Duration
	IntervalComment   time.Duration
	IntervalPostImage time.Duration
	IntervalPostVideo time.Duration
	IntervalChat      time.Duration
	IntervalFollow    time.Duration
	IntervalVideoPoll time.Duration
	DefaultPassword   string
}

// LoadRuntimeFlags 读取 sim-user-service 运行时开关。
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
		IntervalRegister:  24 * time.Hour,
		IntervalComment:   6 * time.Hour,
		IntervalPostImage: 3*time.Hour + 30*time.Minute,
		IntervalPostVideo: 6*time.Hour + 30*time.Minute,
		IntervalChat:      time.Hour,
		IntervalFollow:    7 * time.Hour,
		IntervalVideoPoll: time.Minute,
		DefaultPassword:   strings.TrimSpace(g.Cfg().MustGet(ctx, "simUser.defaultPassword").String()),
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
	// 简易 jitter：取 base 的 90%–110%（用纳秒低位近似）。
	n := time.Now().UnixNano() % int64(j*2)
	return base - time.Duration(j) + time.Duration(n)
}
