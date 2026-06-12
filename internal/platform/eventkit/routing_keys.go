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
	// RoutingPrefixUcg UCG 审核事件前缀。
	RoutingPrefixUcg = "ucg."

	RoutingHistoryRecordCreated RouteKey = RouteKey(RoutingPrefixHistoryRecord + "created")
	RoutingHistoryRecordUpdated RouteKey = RouteKey(RoutingPrefixHistoryRecord + "updated")
	RoutingHistoryRecordDeleted RouteKey = RouteKey(RoutingPrefixHistoryRecord + "deleted")

	RoutingDeviceEventChanged       RouteKey = RouteKey(RoutingPrefixDevice + "event.changed")
	RoutingDeviceActionChanged      RouteKey = RouteKey(RoutingPrefixDevice + "action.changed")
	RoutingDeviceUserProfileUpdated RouteKey = RouteKey(RoutingPrefixDevice + "user.profile.updated")

	RoutingVoiceTaskRequested RouteKey = RouteKey(RoutingPrefixVoiceTask + "requested")
	RoutingVoiceTaskCompleted RouteKey = RouteKey(RoutingPrefixVoiceTask + "completed")
	RoutingVoiceTaskFailed    RouteKey = RouteKey(RoutingPrefixVoiceTask + "failed")

	RoutingUcgPostCreated          RouteKey = RouteKey(RoutingPrefixUcg + "post.created")
	RoutingUcgCommentCreated       RouteKey = RouteKey(RoutingPrefixUcg + "comment.created")
	RoutingUcgProfilePatchSubmitted RouteKey = RouteKey(RoutingPrefixUcg + "profile.patch.submitted")
	RoutingUcgChatMsgCreated       RouteKey = RouteKey(RoutingPrefixUcg + "chat.msg.created")

	RoutingUcgPostPublished   RouteKey = RouteKey(RoutingPrefixUcg + "post.published")
	RoutingUcgPostUnpublished   RouteKey = RouteKey(RoutingPrefixUcg + "post.unpublished")
	RoutingUcgPostLiked         RouteKey = RouteKey(RoutingPrefixUcg + "post.liked")
	RoutingUcgPostUnliked       RouteKey = RouteKey(RoutingPrefixUcg + "post.unliked")
	RoutingUcgCommentPublished  RouteKey = RouteKey(RoutingPrefixUcg + "comment.published")
	RoutingUcgCommentRemoved    RouteKey = RouteKey(RoutingPrefixUcg + "comment.removed")
)

var registeredRoutingKeys = map[RouteKey]struct{}{
	RoutingHistoryRecordCreated:     {},
	RoutingHistoryRecordUpdated:     {},
	RoutingHistoryRecordDeleted:     {},
	RoutingDeviceEventChanged:       {},
	RoutingDeviceActionChanged:      {},
	RoutingDeviceUserProfileUpdated: {},
	RoutingVoiceTaskRequested:        {},
	RoutingVoiceTaskCompleted:        {},
	RoutingVoiceTaskFailed:           {},
	RoutingUcgPostCreated:            {},
	RoutingUcgCommentCreated:         {},
	RoutingUcgProfilePatchSubmitted:  {},
	RoutingUcgChatMsgCreated:         {},
	RoutingUcgPostPublished:          {},
	RoutingUcgPostUnpublished:        {},
	RoutingUcgPostLiked:              {},
	RoutingUcgPostUnliked:            {},
	RoutingUcgCommentPublished:       {},
	RoutingUcgCommentRemoved:         {},
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
