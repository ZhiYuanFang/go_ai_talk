package push

import (
	"os"
	"strings"
)

const pushLiveActivityEnabledEnv = "PUSH_LIVE_ACTIVITY_ENABLED"

// LiveActivityEnabled 特性开关：默认关；true/1/on 开启 Live Activity 发送与 iOS Imminent LA 路径。
// 业务：关时 push 内部 liveactivity API 仍可受理但发送侧短路；voice Imminent 回退普通 alert。
func LiveActivityEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(pushLiveActivityEnabledEnv)))
	return v == "1" || v == "true" || v == "on" || v == "yes"
}

// LiveActivityAttributesType 返回 APNs attributes-type 字符串（与 Flutter ActivityAttributes 名对齐）。
// 可由 PUSH_LA_ATTR_IMMINENT / PUSH_LA_ATTR_TIMING 覆盖。
func LiveActivityAttributesType(attributesKind string) string {
	kind := strings.TrimSpace(strings.ToLower(attributesKind))
	switch kind {
	case LAAttributesImminent:
		if v := strings.TrimSpace(os.Getenv("PUSH_LA_ATTR_IMMINENT")); v != "" {
			return v
		}
		return "PredictImminentAttributes"
	case LAAttributesTiming:
		if v := strings.TrimSpace(os.Getenv("PUSH_LA_ATTR_TIMING")); v != "" {
			return v
		}
		return "PredictTimingAttributes"
	default:
		return ""
	}
}
