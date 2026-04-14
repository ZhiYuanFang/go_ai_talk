package service

import (
	"hello/internal/model/entity"
)

// 领域对象类型别名，优先复用已生成 entity；无对应结构时使用 model 视图对象。
type (
	DeviceAdminItem = entity.User
	DeviceEventItem = entity.Event

	// AdminActionItem 管理端动作预设行（含目标类型中文说明）。
	AdminActionItem struct {
		Id              int64  `json:"id"`
		Name            string `json:"name"`
		TargetType      string `json:"targetType"`
		TargetTypeLabel string `json:"targetTypeLabel"`
	}
	DeviceHistoryItem = entity.History
	DeviceSuggestItem = entity.Suggest
)
