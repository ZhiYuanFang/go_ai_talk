package simuser

import (
	"fmt"
	"time"
)

// RuntimeAPIInput Admin API 传入的运行时字段（周期为可读字符串）。
type RuntimeAPIInput struct {
	TaskRegister  bool
	TaskComment   bool
	TaskPostImage bool
	TaskPostVideo bool
	TaskChat      bool
	TaskFollow    bool
	VideoPoll     bool

	IntervalRegister        string
	IntervalComment         string
	IntervalPostImage       string
	IntervalPostVideo       string
	IntervalChat            string
	IntervalFollow          string
	IntervalVideoPollIdle   string
	IntervalVideoPollActive string
	StartupStaggerMax       string
	EphemeralChatLoop       string
	EphemeralChatWindow     string
	RateLimitRps            float64
	RateLimitBurst          int
}

// BuildRuntimeConfigFromAPI 将 Admin 请求转为 DB 结构；空周期字段保留 current 值。
func BuildRuntimeConfigFromAPI(in RuntimeAPIInput, current RuntimeConfigDB) (RuntimeConfigDB, error) {
	def := DefaultRuntimeConfigDB()
	out := current
	if out.IntervalRegisterSec <= 0 {
		out = def
	}
	out.TaskRegister = in.TaskRegister
	out.TaskComment = in.TaskComment
	out.TaskPostImage = in.TaskPostImage
	out.TaskPostVideo = in.TaskPostVideo
	out.TaskChat = in.TaskChat
	out.TaskFollow = in.TaskFollow
	out.VideoPoll = in.VideoPoll
	if in.RateLimitRps > 0 {
		out.RateLimitRps = in.RateLimitRps
	}
	if in.RateLimitBurst > 0 {
		out.RateLimitBurst = in.RateLimitBurst
	}

	var err error
	if out.IntervalRegisterSec, err = durSec(in.IntervalRegister, out.IntervalRegisterSec, def.IntervalRegisterSec); err != nil {
		return RuntimeConfigDB{}, fmt.Errorf("interval register: %w", err)
	}
	if out.IntervalCommentSec, err = durSec(in.IntervalComment, out.IntervalCommentSec, def.IntervalCommentSec); err != nil {
		return RuntimeConfigDB{}, fmt.Errorf("interval comment: %w", err)
	}
	if out.IntervalPostImageSec, err = durSec(in.IntervalPostImage, out.IntervalPostImageSec, def.IntervalPostImageSec); err != nil {
		return RuntimeConfigDB{}, fmt.Errorf("interval postImage: %w", err)
	}
	if out.IntervalPostVideoSec, err = durSec(in.IntervalPostVideo, out.IntervalPostVideoSec, def.IntervalPostVideoSec); err != nil {
		return RuntimeConfigDB{}, fmt.Errorf("interval postVideo: %w", err)
	}
	if out.IntervalChatSec, err = durSec(in.IntervalChat, out.IntervalChatSec, def.IntervalChatSec); err != nil {
		return RuntimeConfigDB{}, fmt.Errorf("interval chat: %w", err)
	}
	if out.IntervalFollowSec, err = durSec(in.IntervalFollow, out.IntervalFollowSec, def.IntervalFollowSec); err != nil {
		return RuntimeConfigDB{}, fmt.Errorf("interval follow: %w", err)
	}
	if out.IntervalVideoPollIdleSec, err = durSec(in.IntervalVideoPollIdle, out.IntervalVideoPollIdleSec, def.IntervalVideoPollIdleSec); err != nil {
		return RuntimeConfigDB{}, fmt.Errorf("interval videoPollIdle: %w", err)
	}
	if out.IntervalVideoPollActiveSec, err = durSec(in.IntervalVideoPollActive, out.IntervalVideoPollActiveSec, def.IntervalVideoPollActiveSec); err != nil {
		return RuntimeConfigDB{}, fmt.Errorf("interval videoPollActive: %w", err)
	}
	if out.StartupStaggerSec, err = durSec(in.StartupStaggerMax, out.StartupStaggerSec, def.StartupStaggerSec); err != nil {
		return RuntimeConfigDB{}, fmt.Errorf("interval startupStagger: %w", err)
	}
	if out.EphemeralChatLoopSec, err = durSec(in.EphemeralChatLoop, out.EphemeralChatLoopSec, def.EphemeralChatLoopSec); err != nil {
		return RuntimeConfigDB{}, fmt.Errorf("interval ephemeralLoop: %w", err)
	}
	if out.EphemeralChatWindowSec, err = durSec(in.EphemeralChatWindow, out.EphemeralChatWindowSec, def.EphemeralChatWindowSec); err != nil {
		return RuntimeConfigDB{}, fmt.Errorf("interval ephemeralWindow: %w", err)
	}
	return out, nil
}

func durSec(input string, currentSec, defSec int64) (int64, error) {
	if input == "" {
		if currentSec > 0 {
			return currentSec, nil
		}
		return defSec, nil
	}
	d, err := ParseDurationInput(input, time.Duration(defSec)*time.Second)
	if err != nil {
		return 0, err
	}
	return int64(d / time.Second), nil
}

// RuntimeConfigToAPIIntervals 将 DB 结构转为 Admin 可读周期字符串。
func RuntimeConfigToAPIIntervals(rt RuntimeConfigDB) map[string]string {
	return map[string]string{
		"register":            durationLabel(time.Duration(rt.IntervalRegisterSec) * time.Second),
		"comment":             durationLabel(time.Duration(rt.IntervalCommentSec) * time.Second),
		"postImage":           durationLabel(time.Duration(rt.IntervalPostImageSec) * time.Second),
		"postVideo":           durationLabel(time.Duration(rt.IntervalPostVideoSec) * time.Second),
		"chat":                durationLabel(time.Duration(rt.IntervalChatSec) * time.Second),
		"follow":              durationLabel(time.Duration(rt.IntervalFollowSec) * time.Second),
		"videoPollIdle":       durationLabel(time.Duration(rt.IntervalVideoPollIdleSec) * time.Second),
		"videoPollActive":     durationLabel(time.Duration(rt.IntervalVideoPollActiveSec) * time.Second),
		"startupStaggerMax":   durationLabel(time.Duration(rt.StartupStaggerSec) * time.Second),
		"ephemeralChatLoop":   durationLabel(time.Duration(rt.EphemeralChatLoopSec) * time.Second),
		"ephemeralChatWindow": durationLabel(time.Duration(rt.EphemeralChatWindowSec) * time.Second),
	}
}
