package v1

import "github.com/gogf/gf/v2/frame/g"

// PredictImminentPendingEventItem 客户端上报的单条待发生预测节点。
type PredictImminentPendingEventItem struct {
	EventId int64  `json:"eventId" v:"required|min:1" dc:"事件 ID（可为一级根）"`
	NextAt  int64  `json:"nextAt" v:"required|min:1" dc:"下次发生 unix 秒"`
	Title   string `json:"title" dc:"父事件展示名；空则服务端到点不发可见推送"`
}

// DevicePredictImminentPendingPutReq 全量替换宝宝待发生预测列表（最后写入赢）。
// 客户端：孪生仓 Flutter 在每次预测更新后调用；须登录且 deviceNo 与会话绑机一致。
// 离线推送依赖用户已 POST /app/api/push/register。
type DevicePredictImminentPendingPutReq struct {
	g.Meta   `path:"/device/api/predict/imminent/pending" method:"put" tags:"device" summary:"全量同步预测临近待办（Redis）"`
	DeviceNo string                            `json:"deviceNo" v:"required" dc:"宝宝设备号"`
	Events   []PredictImminentPendingEventItem `json:"events" dc:"待发生全量列表；空数组清空"`
}

// DevicePredictImminentPendingPutRes 同步结果。
type DevicePredictImminentPendingPutRes struct {
	Count int `json:"count" dc:"写入条数"`
}
