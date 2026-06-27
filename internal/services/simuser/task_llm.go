package simuser

import (
	"hello/internal/services/aimodel"
)

// TaskAIModelItem 管理页展示：某调度任务使用的一条 LLM lane 及当前生效模型。
type TaskAIModelItem struct {
	LaneKey  string `json:"laneKey"`
	Usage    string `json:"usage,omitempty"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// taskLLMUsage 调度任务名 → 所用 lane 及用途说明（与 tasks.go 调用保持一致）。
var taskLLMUsage = map[string][]struct {
	lane  aimodel.Lane
	usage string
}{
	"register": {
		{aimodel.LaneSimText, "昵称"},
		{aimodel.LaneSimImageGen, "头像"},
	},
	"comment": {
		{aimodel.LaneSimVision, "评论"},
	},
	"post_image": {
		{aimodel.LaneSimText, "文案"},
		{aimodel.LaneSimImageGen, "配图"},
	},
	"post_video_submit": {
		{aimodel.LaneSimText, "文案"},
		{aimodel.LaneSimVideoGen, "提交与轮询"},
	},
	"chat_scan": {
		{aimodel.LaneSimText, "未读回复"},
	},
	// follow 不使用 LLM
}

// BuildTaskAIModelCatalog 解析各任务当前生效的 provider/model（lane 配置来自 DB/env）。
func BuildTaskAIModelCatalog(lanes map[string]aimodel.LaneProfileDTO) map[string][]TaskAIModelItem {
	out := make(map[string][]TaskAIModelItem, len(taskLLMUsage))
	for taskName, defs := range taskLLMUsage {
		items := make([]TaskAIModelItem, 0, len(defs))
		for _, def := range defs {
			key := string(def.lane)
			prof, ok := lanes[key]
			item := TaskAIModelItem{LaneKey: key, Usage: def.usage}
			if ok {
				item.Provider = prof.Provider
				item.Model = prof.Model
			}
			items = append(items, item)
		}
		out[taskName] = items
	}
	return out
}
