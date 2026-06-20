package simuser

import (
	"context"
	"os"
	"strings"
	"time"

	"hello/internal/platform/dbcfg"
	"hello/internal/services/aimodel"
)

// RuntimeSnapshotDTO 管理页只读运行时快照（不含密码与完整 DSN）。
type RuntimeSnapshotDTO struct {
	ServiceEnabled    bool
	DbEnabled         bool
	DatabaseName      string
	SimUserCount      int
	SimUserCountError string
	MaxSimUsers       int
	TaskSwitches      RuntimeTaskSwitchesDTO
	Intervals         RuntimeIntervalsDTO
	RateLimitRps      float64
	RateLimitBurst    int
	LLMLanes          map[string]LLMLaneSnapshotDTO
}

// LLMLaneSnapshotDTO 单条 aimodel lane 只读快照。
type LLMLaneSnapshotDTO struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// RuntimeTaskSwitchesDTO 各任务 env 开关。
type RuntimeTaskSwitchesDTO struct {
	Register  bool
	Comment   bool
	PostImage bool
	PostVideo bool
	Chat      bool
	Follow    bool
	VideoPoll bool
}

// RuntimeIntervalsDTO 各任务周期字符串（人类可读）。
type RuntimeIntervalsDTO struct {
	Register            string
	Comment             string
	PostImage           string
	PostVideo           string
	Chat                string
	Follow              string
	VideoPollIdle       string
	VideoPollActive     string
	StartupStaggerMax   string
	EphemeralChatLoop   string
	EphemeralChatWindow string
}

// GetRuntimeSnapshot 组装当前进程生效的运行时配置，供 Admin 只读展示。
func GetRuntimeSnapshot(ctx context.Context) (RuntimeSnapshotDTO, error) {
	flags := LoadRuntimeFlags(ctx)
	cfg, err := GetConfig(ctx)
	if err != nil {
		return RuntimeSnapshotDTO{}, err
	}

	dbLink := strings.TrimSpace(os.Getenv("SIM_DB_LINK"))
	if dbLink == "" {
		dbLink = strings.TrimSpace(os.Getenv("GF_DATABASE_DEFAULT_LINK"))
	}
	rl := LoadRateLimitSettings()

	out := RuntimeSnapshotDTO{
		ServiceEnabled: flags.Enabled,
		DbEnabled:      cfg.Enabled,
		DatabaseName:   dbcfg.DatabaseNameFromLink(dbLink),
		MaxSimUsers:    cfg.MaxSimUsers,
		TaskSwitches: RuntimeTaskSwitchesDTO{
			Register:  flags.TaskRegister,
			Comment:   flags.TaskComment,
			PostImage: flags.TaskPostImage,
			PostVideo: flags.TaskPostVideo,
			Chat:      flags.TaskChat,
			Follow:    flags.TaskFollow,
			VideoPoll: flags.VideoPoll,
		},
		Intervals: RuntimeIntervalsDTO{
			Register:            durationLabel(flags.IntervalRegister),
			Comment:             durationLabel(flags.IntervalComment),
			PostImage:           durationLabel(flags.IntervalPostImage),
			PostVideo:           durationLabel(flags.IntervalPostVideo),
			Chat:                durationLabel(flags.IntervalChat),
			Follow:              durationLabel(flags.IntervalFollow),
			VideoPollIdle:       durationLabel(flags.IntervalVideoPollIdle),
			VideoPollActive:     durationLabel(flags.IntervalVideoPollActive),
			StartupStaggerMax:   durationLabel(flags.StartupStaggerMax),
			EphemeralChatLoop:   durationLabel(flags.EphemeralChatLoop),
			EphemeralChatWindow: durationLabel(flags.EphemeralChatWindow),
		},
		RateLimitRps:   rl.RPS,
		RateLimitBurst: rl.Burst,
		LLMLanes:       mapLLMSnapshot(LoadAllLaneProfiles()),
	}

	count, countErr := countSimUsers(ctx)
	if countErr != nil {
		out.SimUserCount = -1
		out.SimUserCountError = countErr.Error()
	} else {
		out.SimUserCount = count
	}
	return out, nil
}

func mapLLMSnapshot(in map[string]aimodel.LaneProfileDTO) map[string]LLMLaneSnapshotDTO {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]LLMLaneSnapshotDTO, len(in))
	for k, v := range in {
		out[k] = LLMLaneSnapshotDTO{Provider: v.Provider, Model: v.Model}
	}
	return out
}

// durationLabel 将 Duration 格式化为管理页可读字符串（如 5m、6h30m）。
func durationLabel(d time.Duration) string {
	if d <= 0 {
		return "0"
	}
	s := d.Round(time.Second).String()
	if strings.HasSuffix(s, "0s") {
		s = strings.TrimSuffix(s, "0s")
	}
	return s
}
