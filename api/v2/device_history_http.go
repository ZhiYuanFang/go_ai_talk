package v2

import (
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DeviceHistoryListReq 设备历史列表 v2 请求，扩展时间范围和 limit 参数。
// 不传新参数时行为与 v1 完全一致。
type DeviceHistoryListReq struct {
	g.Meta    `path:"/device/history/api/v2/list" method:"get" tags:"device" summary:"设备历史列表v2（支持时间范围）"`
	DeviceNo  string `json:"deviceNo"  p:"deviceNo"  dc:"设备号"`
	Page      int    `json:"page"      p:"page"      dc:"页码，从 1 开始"`
	PageSize  int    `json:"pageSize"  p:"pageSize"  dc:"每页条数"`
	StartTime int64  `json:"startTime" p:"startTime" dc:"开始时间，Unix 秒；0 表示不限制"`
	EndTime   int64  `json:"endTime"   p:"endTime"   dc:"结束时间，Unix 秒；0 表示不限制"`
	Limit     int    `json:"limit"     p:"limit"     dc:"返回条数上限；>0 时替代 pageSize，page 忽略"`
}

// DeviceHistoryListRes 设备历史列表 v2 响应，结构与 v1 完全一致。
type DeviceHistoryListRes struct {
	List     []entity.History `json:"list"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}
