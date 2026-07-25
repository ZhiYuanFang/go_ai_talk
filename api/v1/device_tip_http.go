package v1

import "github.com/gogf/gf/v2/frame/g"

// DeviceTipGenerateReq 小贴士流式生成请求。
// 以 SSE 方式返回 thinking/answer/done 流式事件。
// 月龄与当前时间由 Python tip 图派生，请求体不再携带 babyAgeMonths / currentTime。
type DeviceTipGenerateReq struct {
	g.Meta    `path:"/device/tip/generate" method:"post" tags:"device" summary:"小贴士流式生成（SSE）"`
	DeviceNo  string `json:"deviceNo" v:"required" dc:"设备号"`
	EventId   int64  `json:"eventId" dc:"触发事件 ID；0 表示无特定事件"`
	EventName string `json:"eventName" dc:"触发事件名称"`
}

// DeviceTipGenerateRes 小贴士流式生成响应（SSE，无固定 JSON 结构）。
// 实际以 SSE 事件帧逐帧返回：event: thinking / event: answer / event: done / data: [DONE]
type DeviceTipGenerateRes struct{}
