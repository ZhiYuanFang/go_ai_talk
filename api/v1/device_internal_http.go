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

// DeviceInternalEventNonLeafCountReq 内部：非叶子事件计数（供 cash catalog 聚合）。
type DeviceInternalEventNonLeafCountReq struct {
	g.Meta `path:"/device/internal/api/event/non-leaf-count" method:"get" tags:"device" summary:"内部-非叶子事件数"`
}

// DeviceInternalEventNonLeafCountRes 非叶子事件计数响应。
type DeviceInternalEventNonLeafCountRes struct {
	Count int `json:"count"`
}
