package devicectrl

// device_xiaozhi_mcp_internal.go：内部小智绑定全量（密钥中间件组）。

import (
	"context"

	v1 "hello/api/v1"
	device "hello/internal/services/device"
)

// DeviceXiaozhiMcpInternalCtrl 小智绑定内部接口（须 InternalSecretMiddleware）。
type DeviceXiaozhiMcpInternalCtrl struct{}

// NewDeviceXiaozhiMcpInternalCtrl 构造内部控制器。
func NewDeviceXiaozhiMcpInternalCtrl() *DeviceXiaozhiMcpInternalCtrl {
	return &DeviceXiaozhiMcpInternalCtrl{}
}

// BindingsList GET /device/internal/api/xiaozhi-mcp/bindings
func (c *DeviceXiaozhiMcpInternalCtrl) BindingsList(ctx context.Context, req *v1.DeviceInternalXiaozhiMcpBindingListReq) (res *v1.DeviceInternalXiaozhiMcpBindingListRes, err error) {
	_ = req
	list, err := device.ListActiveXiaozhiMcpBindingsFull(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]v1.DeviceInternalXiaozhiMcpBindingItem, 0, len(list))
	for _, it := range list {
		out = append(out, v1.DeviceInternalXiaozhiMcpBindingItem{
			Id:       it.Id,
			WxId:     it.WxId,
			DeviceNo: it.DeviceNo,
			McpToken: it.McpToken,
			Alias:    it.Alias,
			Status:   it.Status,
		})
	}
	return &v1.DeviceInternalXiaozhiMcpBindingListRes{List: out}, nil
}
