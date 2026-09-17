// 推送设备 Admin 读模型：分页列表 push_device，供运维 Hub 查看与手动删除。
package push

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const (
	adminDeviceListDefaultPageSize = 20
	adminDeviceListMaxPageSize     = 200
)

// AdminDeviceListFilter Admin 列表筛选条件。
type AdminDeviceListFilter struct {
	Page       int
	PageSize   int
	WxID       int64
	Channel    string
	TokenPrefix string
}

// AdminDeviceListResult 分页结果。
type AdminDeviceListResult struct {
	List     []entityPushDevice
	Page     int
	PageSize int
	Total    int
}

// ListPushDevicesForAdmin 分页列出 push_device（可选 wxId / channel / token 前缀）。
func ListPushDevicesForAdmin(ctx context.Context, f AdminDeviceListFilter) (*AdminDeviceListResult, error) {
	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize < 1 {
		pageSize = adminDeviceListDefaultPageSize
	}
	if pageSize > adminDeviceListMaxPageSize {
		pageSize = adminDeviceListMaxPageSize
	}

	model := g.DB().Model(pushDeviceTable).Ctx(ctx)
	if f.WxID > 0 {
		model = model.Where("wx_id", f.WxID)
	}
	channel := strings.TrimSpace(strings.ToLower(f.Channel))
	if channel != "" {
		if err := validatePushChannel(channel); err != nil {
			return nil, err
		}
		model = model.Where("channel", channel)
	}
	tokenPrefix := strings.TrimSpace(f.TokenPrefix)
	if tokenPrefix != "" {
		// 前缀模糊：运维排查用；禁止过短以免全表拖垮（至少 4 字符）。
		if len(tokenPrefix) < 4 {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, "token 前缀至少 4 个字符")
		}
		model = model.WhereLike("token", tokenPrefix+"%")
	}

	total, err := model.Clone().Count()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "统计推送设备失败")
	}

	rows, err := model.OrderDesc("id").Page(page, pageSize).All()
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeDbOperationError, err, "查询推送设备失败")
	}
	list := make([]entityPushDevice, 0, len(rows))
	for _, row := range rows {
		list = append(list, entityPushDevice{
			ID:        row["id"].Uint64(),
			WxID:      row["wx_id"].Int64(),
			Channel:   row["channel"].String(),
			Token:     row["token"].String(),
			DeviceKey: row["device_key"].String(),
			UpdatedAt: row["updated_at"].Int64(),
		})
	}
	return &AdminDeviceListResult{
		List:     list,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}
