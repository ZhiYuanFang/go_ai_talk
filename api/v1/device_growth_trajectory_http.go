package v1

import "github.com/gogf/gf/v2/frame/g"

// DeviceGrowthTrajectoryLatestReq GET 宝宝成长轨迹最新结果。
type DeviceGrowthTrajectoryLatestReq struct {
	g.Meta   `path:"/device/api/growth-trajectory/latest" method:"get" tags:"device" summary:"成长轨迹最新结果"`
	DeviceNo string `json:"deviceNo" p:"deviceNo" v:"required" dc:"设备号（宝宝维度）"`
}

// DeviceGrowthTrajectoryLatestRes 最新结果 data。
type DeviceGrowthTrajectoryLatestRes struct {
	ResultMarkdown *string `json:"resultMarkdown" dc:"最新 Markdown；无结果时为 null"`
	UpdatedAt      int64   `json:"updatedAt" dc:"更新时间 unix 秒；无结果时为 0"`
	SessionId      string  `json:"sessionId" dc:"最近会话 ID"`
	UsedToday      int     `json:"usedToday" dc:"今日已用次数（账号维上海日）"`
	DailyLimit     int     `json:"dailyLimit" dc:"今日上限"`
}

// DeviceGrowthTrajectoryAnswerDTO turn 回答体。
type DeviceGrowthTrajectoryAnswerDTO struct {
	QuestionId string `json:"questionId" dc:"问题 ID"`
	Value      string `json:"value" dc:"用户选择或自由文本"`
}

// DeviceGrowthTrajectoryTurnReq POST 成长轨迹一轮（SSE）。
// 流式推送 thinking/question/result/error/done；结束 data: [DONE]。
// 开通/日限等预检失败时返回普通 JSON envelope（不开流）。
type DeviceGrowthTrajectoryTurnReq struct {
	g.Meta    `path:"/device/api/growth-trajectory/turn" method:"post" tags:"device" summary:"成长轨迹一轮（SSE）"`
	DeviceNo  string                           `json:"deviceNo" v:"required" dc:"设备号"`
	Action    string                           `json:"action" v:"required" dc:"start|answer|restart"`
	SessionId string                           `json:"sessionId" dc:"start 可选；answer 必填"`
	Answer    *DeviceGrowthTrajectoryAnswerDTO `json:"answer" dc:"action=answer 时必填"`
}

// DeviceGrowthTrajectoryTurnRes SSE 无固定 JSON；占位以满足 GoFrame 绑定。
type DeviceGrowthTrajectoryTurnRes struct{}
