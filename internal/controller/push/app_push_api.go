package pushctrl

import (
	"context"

	v1 "hello/api/v1"
	"hello/internal/platform/httpmeta"
	pushsvc "hello/internal/services/push"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AppPushCtrl App 推送注册/注销（宿主 push-service）。
type AppPushCtrl struct{}

func wxIDFromHeader(ctx context.Context) (int64, error) {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "缺少请求上下文")
	}
	wxID := httpmeta.ParseHeaderWxID(r.GetHeader(httpmeta.HeaderInternalWxId))
	if wxID <= 0 {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	return wxID, nil
}

// RegisterPost POST /app/api/push/register
func (c *AppPushCtrl) RegisterPost(ctx context.Context, req *v1.AppPushRegisterPostReq) (res *v1.AppPushRegisterPostRes, err error) {
	wxID, err := wxIDFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	if err = pushsvc.RegisterPushDevice(ctx, wxID, req.Channel, req.Token, req.DeviceKey); err != nil {
		return nil, err
	}
	return &v1.AppPushRegisterPostRes{}, nil
}

// UnregisterPost POST /app/api/push/unregister
func (c *AppPushCtrl) UnregisterPost(ctx context.Context, req *v1.AppPushUnregisterPostReq) (res *v1.AppPushUnregisterPostRes, err error) {
	wxID, err := wxIDFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	if err = pushsvc.UnregisterPushDevice(ctx, wxID, req.DeviceKey, req.Channel); err != nil {
		return nil, err
	}
	return &v1.AppPushUnregisterPostRes{}, nil
}
