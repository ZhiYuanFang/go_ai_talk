package simuser

import (
	"context"
	"os"
	"strings"
	"time"
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

// LoadRuntimeFlags 读取运行时配置：DB runtime_json 优先，env 兜底；进程总闸仍读 SIM_USER_SERVICE_ENABLED。
func LoadRuntimeFlags(ctx context.Context) RuntimeFlags {
	dbCfg, err := LoadRuntimeFromDB(ctx)
	flags := dbCfg.toRuntimeFlags(ctx)
	if err != nil {
		flags = defaultRuntimeFromEnv(ctx)
	}
	flags.Enabled = envBool("SIM_USER_SERVICE_ENABLED", false)
	return flags
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
