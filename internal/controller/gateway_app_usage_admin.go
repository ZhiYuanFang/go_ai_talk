package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/services/gatewayapp/apiregistry"
	"hello/internal/services/gatewayapp/usagestats"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// GatewayAppUsageAdminCtrl App API 使用统计管理读接口（gateway-app 本机 + Redis）。
type GatewayAppUsageAdminCtrl struct{}

func NewGatewayAppUsageAdminCtrl() *GatewayAppUsageAdminCtrl {
	return &GatewayAppUsageAdminCtrl{}
}

func (c *GatewayAppUsageAdminCtrl) requireAdmin(ctx context.Context) error {
	return requireGatewayAdminJWT(ctx)
}

func queryDaysFromReq(days int) int {
	if days < 0 {
		return 7
	}
	return days
}

func summaryOf(apiKey string) string {
	return apiregistry.SummaryOf(apiKey)
}

// UsageList GET /device/admin/api/usage/list
func (c *GatewayAppUsageAdminCtrl) UsageList(ctx context.Context, req *v1.DeviceAdminUsageListReq) (res *v1.DeviceAdminUsageListRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	queryDays := queryDaysFromReq(req.Days)
	sortBy := usagestats.ParseSortBy(req.SortBy)
	items, err := usagestats.ListAPIs(ctx, queryDays, sortBy, summaryOf)
	if err != nil {
		return nil, err
	}
	list := make([]v1.DeviceAdminUsageListItem, 0, len(items))
	for _, it := range items {
		list = append(list, v1.DeviceAdminUsageListItem{
			ApiKey:  it.ApiKey,
			Summary: it.Summary,
			Count:   it.Count,
			LastAt:  it.LastAt,
		})
	}
	return &v1.DeviceAdminUsageListRes{List: list, Days: queryDays, SortBy: sortBy}, nil
}

// UsageDetail GET /device/admin/api/usage/detail
func (c *GatewayAppUsageAdminCtrl) UsageDetail(ctx context.Context, req *v1.DeviceAdminUsageDetailReq) (res *v1.DeviceAdminUsageDetailRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	apiKey := strings.TrimSpace(req.ApiKey)
	if apiKey == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "apiKey 不能为空")
	}
	queryDays := queryDaysFromReq(req.Days)
	sortBy := usagestats.ParseSortBy(req.SortBy)
	items, err := usagestats.ListUsersForAPI(ctx, queryDays, apiKey, sortBy)
	if err != nil {
		return nil, err
	}
	list := make([]v1.DeviceAdminUsageDetailItem, 0, len(items))
	for _, it := range items {
		list = append(list, v1.DeviceAdminUsageDetailItem{
			WxId:   it.WxId,
			Count:  it.Count,
			LastAt: it.LastAt,
		})
	}
	return &v1.DeviceAdminUsageDetailRes{
		ApiKey:  apiKey,
		Summary: summaryOf(apiKey),
		List:    list,
		Days:    queryDays,
		SortBy:  sortBy,
	}, nil
}

// UsageUser GET /device/admin/api/usage/user
func (c *GatewayAppUsageAdminCtrl) UsageUser(ctx context.Context, req *v1.DeviceAdminUsageUserReq) (res *v1.DeviceAdminUsageUserRes, err error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.WxId <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 须为正整数")
	}
	queryDays := queryDaysFromReq(req.Days)
	sortBy := usagestats.ParseSortBy(req.SortBy)
	items, err := usagestats.ListAPIsForUser(ctx, queryDays, req.WxId, sortBy, summaryOf)
	if err != nil {
		return nil, err
	}
	list := make([]v1.DeviceAdminUsageUserItem, 0, len(items))
	for _, it := range items {
		list = append(list, v1.DeviceAdminUsageUserItem{
			ApiKey:  it.ApiKey,
			Summary: it.Summary,
			Count:   it.Count,
			LastAt:  it.LastAt,
		})
	}
	return &v1.DeviceAdminUsageUserRes{WxId: req.WxId, List: list, Days: queryDays, SortBy: sortBy}, nil
}
