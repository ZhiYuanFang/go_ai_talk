package types

// AdminActionItem 管理端动作预设行（含目标类型中文说明）。
type AdminActionItem struct {
	Id              int64  `json:"id"`
	Name            string `json:"name"`
	TargetType      string `json:"targetType"`
	TargetTypeLabel string `json:"targetTypeLabel"`
}
