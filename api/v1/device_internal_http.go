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
