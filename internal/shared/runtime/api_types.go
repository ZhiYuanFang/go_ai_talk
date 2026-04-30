package runtime

import (
	"hello/internal/model/entity"
	sharedtypes "hello/internal/shared/types"
)

// 运行时共享类型：仅保留无领域语义的数据结构别名。
type (
	DeviceAdminItem   = entity.User
	DeviceEventItem   = entity.Event
	DeviceHistoryItem = entity.History
	DeviceSuggestItem = entity.Suggest
	AdminActionItem   = sharedtypes.AdminActionItem
)
