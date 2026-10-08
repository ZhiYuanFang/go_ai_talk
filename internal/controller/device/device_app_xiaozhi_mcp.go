package devicectrl

// device_app_xiaozhi_mcp.go：App 小智 MCP 绑定 CRUD。
//
// 鉴权：gateway-app Bearer → X-Internal-Wx-Id。
// 写路径成功后通知 xiaozhi-mcp-service；通知失败仅告警，不回滚 DB。

import (
	"context"

	v1 "hello/api/v1"
	xiaozhimcpclient "hello/internal/clients/xiaozhimcp"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// DeviceAppXiaozhiMcpCtrl App 小智音箱绑定 API。
type DeviceAppXiaozhiMcpCtrl struct{}

// NewDeviceAppXiaozhiMcpCtrl 构造控制器。
func NewDeviceAppXiaozhiMcpCtrl() *DeviceAppXiaozhiMcpCtrl {
	return &DeviceAppXiaozhiMcpCtrl{}
}

// List GET /device/app/api/xiaozhi-mcp/bindings
func (c *DeviceAppXiaozhiMcpCtrl) List(ctx context.Context, req *v1.DeviceAppXiaozhiMcpBindingListReq) (res *v1.DeviceAppXiaozhiMcpBindingListRes, err error) {
	_ = req
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	list, err := device.ListXiaozhiMcpBindingsForWx(ctx, wxID)
	if err != nil {
		return nil, err
	}
	out := make([]v1.DeviceAppXiaozhiMcpBindingItem, 0, len(list))
	for _, it := range list {
		out = append(out, v1.DeviceAppXiaozhiMcpBindingItem{
			Id:         it.Id,
			Alias:      it.Alias,
			TokenMask:  it.TokenMask,
			SpeakerMac: it.SpeakerMac,
			DeviceNo:   it.DeviceNo,
			Status:     it.Status,
			Connected:  it.Connected,
			CreatedAt:  it.CreatedAt,
			UpdatedAt:  it.UpdatedAt,
		})
	}
	return &v1.DeviceAppXiaozhiMcpBindingListRes{List: out}, nil
}

// Add POST /device/app/api/xiaozhi-mcp/bindings
func (c *DeviceAppXiaozhiMcpCtrl) Add(ctx context.Context, req *v1.DeviceAppXiaozhiMcpBindingAddReq) (res *v1.DeviceAppXiaozhiMcpBindingAddRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	result, err := device.AddXiaozhiMcpBinding(ctx, wxID, req.McpToken, req.Alias, req.SpeakerMac)
	if err != nil {
		return nil, err
	}
	full := result.Binding
	// token 变更：先停旧桥再启新桥；仅 alias 变更可跳过通知。
	if result.TokenChanged {
		if result.PrevMcpToken != "" && result.PrevMcpToken != full.McpToken {
			if err := xiaozhimcpclient.BindingRemove(ctx, full.Id, result.PrevMcpToken); err != nil {
				glog.Warningf(ctx, "[xiaozhi-mcp-binding] remove old notify failed id=%d err=%v", full.Id, err)
			}
		}
		if err := xiaozhimcpclient.BindingUpsert(ctx, full.Id, full.McpToken, full.DeviceNo); err != nil {
			glog.Warningf(ctx, "[xiaozhi-mcp-binding] upsert notify failed id=%d err=%v", full.Id, err)
		}
	}
	return &v1.DeviceAppXiaozhiMcpBindingAddRes{
		Id:         full.Id,
		Alias:      full.Alias,
		TokenMask:  device.MaskXiaozhiMcpToken(full.McpToken),
		SpeakerMac: full.SpeakerMac,
		DeviceNo:   full.DeviceNo,
	}, nil
}

// AliasPut PUT /device/app/api/xiaozhi-mcp/bindings/{id}/alias
func (c *DeviceAppXiaozhiMcpCtrl) AliasPut(ctx context.Context, req *v1.DeviceAppXiaozhiMcpBindingAliasPutReq) (res *v1.DeviceAppXiaozhiMcpBindingAliasPutRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	if err := device.UpdateXiaozhiMcpBindingAlias(ctx, wxID, req.Id, req.Alias); err != nil {
		return nil, err
	}
	return &v1.DeviceAppXiaozhiMcpBindingAliasPutRes{}, nil
}

// Delete DELETE /device/app/api/xiaozhi-mcp/bindings/{id}
func (c *DeviceAppXiaozhiMcpCtrl) Delete(ctx context.Context, req *v1.DeviceAppXiaozhiMcpBindingDeleteReq) (res *v1.DeviceAppXiaozhiMcpBindingDeleteRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID, err := wxIDFromAppUserHeader(r)
	if err != nil {
		return nil, err
	}
	full, err := device.DeleteXiaozhiMcpBinding(ctx, wxID, req.Id)
	if err != nil {
		return nil, err
	}
	if err := xiaozhimcpclient.BindingRemove(ctx, full.Id, full.McpToken); err != nil {
		glog.Warningf(ctx, "[xiaozhi-mcp-binding] remove notify failed id=%d err=%v", full.Id, err)
	}
	return &v1.DeviceAppXiaozhiMcpBindingDeleteRes{}, nil
}
