package historyctrl

import (
	"context"
	"strings"

	"hello/internal/platform/httpmeta"
	"hello/internal/services/history"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// HistoryInternalMigrateDeviceNoReq POST 内部：宝宝号原地改名时迁移 history 分区键。
type HistoryInternalMigrateDeviceNoReq struct {
	g.Meta      `path:"/history/internal/api/device-no/migrate" method:"post" tags:"history-internal" summary:"内部：迁移 history 的 device_no"`
	OldDeviceNo string `json:"oldDeviceNo" v:"required"`
	NewDeviceNo string `json:"newDeviceNo" v:"required"`
	// DryRun 为 true 时仅探测库表可读，不 UPDATE。
	DryRun bool `json:"dryRun"`
}

// HistoryInternalMigrateDeviceNoRes 迁移结果。
type HistoryInternalMigrateDeviceNoRes struct {
	Ok bool `json:"ok"`
}

// MigrateDeviceNo POST /history/internal/api/device-no/migrate
func (c *HistoryInternalCtrl) MigrateDeviceNo(ctx context.Context, req *HistoryInternalMigrateDeviceNoReq) (*HistoryInternalMigrateDeviceNoRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil || !httpmeta.ValidateInternalSecret(httpmeta.InternalSecretFromRequest(r)) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "内部接口未授权")
	}
	old := strings.TrimSpace(req.OldDeviceNo)
	neu := strings.TrimSpace(req.NewDeviceNo)
	if req.DryRun {
		if err := history.ProbeMigrateDeviceNo(ctx); err != nil {
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, err.Error())
		}
		return &HistoryInternalMigrateDeviceNoRes{Ok: true}, nil
	}
	if err := history.MigrateDeviceNo(ctx, old, neu); err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidOperation, err.Error())
	}
	return &HistoryInternalMigrateDeviceNoRes{Ok: true}, nil
}
