package push

import (
	"os"
	"strings"
)

const pushLiveActivityEnabledEnv = "PUSH_LIVE_ACTIVITY_ENABLED"

// 统一 ActivityAttributes 默认类型名（与 Flutter 对齐；可用 PUSH_LA_ATTR_TYPE 覆盖）。
const defaultLiveActivityAttributesType = "BabyLiveActivityAttributes"

// LiveActivityEnabled 特性开关：默认关；true/1/on 开启 Live Activity 发送。
func LiveActivityEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(pushLiveActivityEnabledEnv)))
	return v == "1" || v == "true" || v == "on" || v == "yes"
}

// LiveActivityAttributesType 返回 APNs attributes-type（统一单类型，不再分 imminent/timing）。
func LiveActivityAttributesType() string {
	if v := strings.TrimSpace(os.Getenv("PUSH_LA_ATTR_TYPE")); v != "" {
		return v
	}
	return defaultLiveActivityAttributesType
}
