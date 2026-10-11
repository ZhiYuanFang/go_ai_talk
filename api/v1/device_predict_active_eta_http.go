package v1

import "github.com/gogf/gf/v2/frame/g"

// PredictActiveEtaItem 客户端上报的单条进行中计时预计结束（与 Flutter / Python 字段对齐）。
// 服务端清洗：须 rootEventId、leafEventId、startAt、expectEndAt 均 >0 且 expectEndAt>=startAt；title 可空；已过期仍保留。
type PredictActiveEtaItem struct {
	RootEventId int64  `json:"rootEventId" dc:"根事件 ID"`
	LeafEventId int64  `json:"leafEventId" dc:"叶子事件 ID（同表去重键）"`
	HistoryId   string `json:"historyId,omitempty" dc:"历史记录 ID；可选"`
	Title       string `json:"title,omitempty" dc:"展示名；可空"`
	StartAt     int64  `json:"startAt" dc:"开始 unix 秒"`
	ExpectEndAt int64  `json:"expectEndAt" dc:"预计结束 unix 秒"`
}

// DevicePredictActiveEtaPendingPutReq 全量替换宝宝进行中 ETA 列表（最后写入赢；空 items 清空）。
// 客户端：孪生仓 Flutter 在进行中计时变化后调用；须登录且 deviceNo 与会话绑机一致。
// 本接口只写 Redis，不发延时 MQ、不推送（与 imminent pending 分离）。
type DevicePredictActiveEtaPendingPutReq struct {
	g.Meta   `path:"/device/api/predict/active-eta/pending" method:"put" tags:"device" summary:"全量同步进行中计时 ETA（Redis）"`
	DeviceNo string                 `json:"deviceNo" v:"required" dc:"宝宝设备号"`
	Items    []PredictActiveEtaItem `json:"items" dc:"进行中 ETA 全量列表；空数组清空"`
}

// DevicePredictActiveEtaPendingPutRes 同步结果。
type DevicePredictActiveEtaPendingPutRes struct {
	Count int `json:"count" dc:"清洗后写入条数"`
}

// VoiceInternalPredictActiveEtaGetReq 内部按 deviceNo 读取进行中 ETA（供 Python）。
// 须经 voice internal 密钥中间件；缺键返回空 items。
type VoiceInternalPredictActiveEtaGetReq struct {
	g.Meta   `path:"/voice/internal/api/predict/active-eta" method:"get" tags:"voice" summary:"内部-进行中计时 ETA 列表"`
	DeviceNo string `json:"deviceNo" in:"query" v:"required" dc:"宝宝设备号"`
}

// VoiceInternalPredictActiveEtaGetRes 内部读响应。
type VoiceInternalPredictActiveEtaGetRes struct {
	Items []PredictActiveEtaItem `json:"items" dc:"当前 ETA 列表；无键时为空数组"`
}
