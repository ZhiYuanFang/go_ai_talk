package service

import (
	"hello/internal/model/entity"
)

// 领域对象类型别名，优先复用已生成 entity；无对应结构时使用 model 视图对象。
type (
	DeviceAdminItem   = entity.User
	DeviceEventItem   = entity.Event
	DeviceHistoryItem = entity.History
	DeviceSuggestItem = entity.Suggest
)
