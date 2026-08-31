package v1

import (
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DeviceInternalEventOptionsReq 内部事件字典列表（供 history/voice 等跨服务调用）。
type DeviceInternalEventOptionsReq struct {
	g.Meta `path:"/device/internal/api/event/options" method:"get" tags:"device" summary:"内部-事件字典"`
}

// DeviceInternalEventOptionsRes 事件字典响应。
type DeviceInternalEventOptionsRes struct {
	List []entity.Event `json:"list"`
}

// DeviceInternalEventNonLeafCountReq 内部：一级根事件计数（供 cash catalog 聚合 totalActivatableCount）。
// 路径名保留 non-leaf-count（历史）；语义为 parent_id=0 的一级根数，含无子根。
type DeviceInternalEventNonLeafCountReq struct {
	g.Meta `path:"/device/internal/api/event/non-leaf-count" method:"get" tags:"device" summary:"内部-一级根事件数（历史路径 non-leaf-count）"`
}

// DeviceInternalEventNonLeafCountRes 一级根事件计数响应。
type DeviceInternalEventNonLeafCountRes struct {
	Count int `json:"count" dc:"一级根数量（parent_id=0，含无子根）"`
}
