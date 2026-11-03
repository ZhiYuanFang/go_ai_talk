package cachekit

import (
	"fmt"
	"strconv"
	"strings"
)

// PushLiveActivityEnabledKey 特性开关缓存占位（可选）；主开关以 env PUSH_LIVE_ACTIVITY_ENABLED 为准。
// identifier 固定 "global"；本 builder 供运维旁路探测，业务默认直接读 env。
func PushLiveActivityEnabledKey() (string, error) {
	return Key(DomainPush, "liveactivity", "enabled", "global")
}

// PushLiveActivitySlotMetaKey 可选：push 域侧记录某 wx 当前 Imminent 槽元数据；identifier=wxId。
// 与 voice 域 PredictImminentActivitySlotKey 分工：voice 写权威槽 id，本键留给 push 编排扩展。
func PushLiveActivitySlotMetaKey(wxID int64) (string, error) {
	return Key(DomainPush, "liveactivity", "slot", strconv.FormatInt(wxID, 10))
}

// PushLiveActivityTimingMetaKey 可选：push 域 Timing 槽元数据；identifier=deviceNo:historyId。
func PushLiveActivityTimingMetaKey(deviceNo string, historyID int64) (string, error) {
	id := fmt.Sprintf("%s:%d", strings.TrimSpace(deviceNo), historyID)
	return Key(DomainPush, "liveactivity", "timing", id)
}
