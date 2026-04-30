package eventkit

import "strings"

// RouteKey 表示事件路由键（wire 仍为字符串）。
type RouteKey string

const (
	// RoutingPrefixHistoryRecord 历史记录事件前缀。
	RoutingPrefixHistoryRecord = "history.record."
	// RoutingPrefixDevice 设备域事件前缀。
	RoutingPrefixDevice = "device."
	// RoutingPrefixVoiceTask 语音任务事件前缀。
	RoutingPrefixVoiceTask = "voice.task."

	RoutingHistoryRecordCreated RouteKey = RouteKey(RoutingPrefixHistoryRecord + "created")
	RoutingHistoryRecordUpdated RouteKey = RouteKey(RoutingPrefixHistoryRecord + "updated")
	RoutingHistoryRecordDeleted RouteKey = RouteKey(RoutingPrefixHistoryRecord + "deleted")

	RoutingDeviceEventChanged       RouteKey = RouteKey(RoutingPrefixDevice + "event.changed")
	RoutingDeviceActionChanged      RouteKey = RouteKey(RoutingPrefixDevice + "action.changed")
	RoutingDeviceUserProfileUpdated RouteKey = RouteKey(RoutingPrefixDevice + "user.profile.updated")

	RoutingVoiceTaskRequested RouteKey = RouteKey(RoutingPrefixVoiceTask + "requested")
	RoutingVoiceTaskCompleted RouteKey = RouteKey(RoutingPrefixVoiceTask + "completed")
	RoutingVoiceTaskFailed    RouteKey = RouteKey(RoutingPrefixVoiceTask + "failed")
)

var registeredRoutingKeys = map[RouteKey]struct{}{
	RoutingHistoryRecordCreated:     {},
	RoutingHistoryRecordUpdated:     {},
	RoutingHistoryRecordDeleted:     {},
	RoutingDeviceEventChanged:       {},
	RoutingDeviceActionChanged:      {},
	RoutingDeviceUserProfileUpdated: {},
	RoutingVoiceTaskRequested:       {},
	RoutingVoiceTaskCompleted:       {},
	RoutingVoiceTaskFailed:          {},
}

func (k RouteKey) String() string {
	return string(k)
}

func (k RouteKey) IsValid() bool {
	_, ok := registeredRoutingKeys[k]
	return ok
}

func ParseRoutingKey(raw string) (RouteKey, bool) {
	k := RouteKey(strings.TrimSpace(raw))
	return k, k.IsValid()
}

// HasPrefix 判断路由键是否属于指定前缀分组。
func (k RouteKey) HasPrefix(prefix string) bool {
	return strings.HasPrefix(k.String(), strings.TrimSpace(prefix))
}
