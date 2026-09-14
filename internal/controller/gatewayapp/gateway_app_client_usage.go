package gatewayappctrl

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/services/gatewayapp/clientusage"
	"hello/internal/services/gatewayapp/usagestats"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// GatewayAppClientUsageCtrl 客户端功能使用：App 上报 + Admin 读（gateway-app 本机 Redis）。
type GatewayAppClientUsageCtrl struct{}

func NewGatewayAppClientUsageCtrl() *GatewayAppClientUsageCtrl {
	return &GatewayAppClientUsageCtrl{}
}

// Report POST /device/app/api/client-usage/report — 须登录；不计入 App API usage。
func (c *GatewayAppClientUsageCtrl) Report(ctx context.Context, req *v1.GatewayAppClientUsageReportReq) (*v1.GatewayAppClientUsageReportRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	wxID := usagestats.WxIDFromRequest(r)
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "须登录后上报")
	}
	if err := clientusage.Report(ctx, wxID, req.FeatureId, req.Description); err != nil {
		return nil, err
	}
	return &v1.GatewayAppClientUsageReportRes{}, nil
}

func (c *GatewayAppClientUsageCtrl) requireAdmin(ctx context.Context) error {
	return requireGatewayAdminJWT(ctx)
}

// ClientUsageFeatures GET /device/admin/api/client-usage/features
func (c *GatewayAppClientUsageCtrl) ClientUsageFeatures(ctx context.Context, req *v1.DeviceAdminClientUsageFeaturesReq) (*v1.DeviceAdminClientUsageFeaturesRes, error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	days := clientusage.NormalizeDays(req.Days)
	sortBy := usagestats.ParseSortBy(req.SortBy)
	items, err := clientusage.ListFeatures(ctx, days, sortBy)
	if err != nil {
		return nil, err
	}
	list := make([]v1.DeviceAdminClientUsageFeatureItem, 0, len(items))
	for _, it := range items {
		list = append(list, v1.DeviceAdminClientUsageFeatureItem{
			FeatureId: it.FeatureId, Description: it.Description, Count: it.Count, LastAt: it.LastAt,
		})
	}
	return &v1.DeviceAdminClientUsageFeaturesRes{List: list, Days: days, SortBy: sortBy}, nil
}

// ClientUsageFeatureUsers GET /device/admin/api/client-usage/feature-users
func (c *GatewayAppClientUsageCtrl) ClientUsageFeatureUsers(ctx context.Context, req *v1.DeviceAdminClientUsageFeatureUsersReq) (*v1.DeviceAdminClientUsageFeatureUsersRes, error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	fid := strings.TrimSpace(req.FeatureId)
	days := clientusage.NormalizeDays(req.Days)
	sortBy := usagestats.ParseSortBy(req.SortBy)
	items, err := clientusage.ListUsersForFeature(ctx, days, fid, sortBy)
	if err != nil {
		return nil, err
	}
	wxIDs := make([]int64, 0, len(items))
	for _, it := range items {
		wxIDs = append(wxIDs, it.WxId)
	}
	nickMap := usagestats.FetchProfileNicknames(ctx, wxIDs)
	list := make([]v1.DeviceAdminClientUsageFeatureUserItem, 0, len(items))
	for _, it := range items {
		list = append(list, v1.DeviceAdminClientUsageFeatureUserItem{
			WxId: it.WxId, Nickname: nickMap[it.WxId], Count: it.Count, LastAt: it.LastAt,
		})
	}
	return &v1.DeviceAdminClientUsageFeatureUsersRes{
		FeatureId: fid, List: list, Days: days, SortBy: sortBy,
	}, nil
}

// ClientUsageUser GET /device/admin/api/client-usage/user
func (c *GatewayAppClientUsageCtrl) ClientUsageUser(ctx context.Context, req *v1.DeviceAdminClientUsageUserReq) (*v1.DeviceAdminClientUsageUserRes, error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.WxId <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 须为正整数")
	}
	days := clientusage.NormalizeDays(req.Days)
	sortBy := usagestats.ParseSortBy(req.SortBy)
	items, err := clientusage.ListFeaturesForUser(ctx, days, req.WxId, sortBy)
	if err != nil {
		return nil, err
	}
	tl, err := clientusage.ListTimeline(ctx, req.WxId, int64(req.TimelineLimit))
	if err != nil {
		return nil, err
	}
	list := make([]v1.DeviceAdminClientUsageUserFeatureItem, 0, len(items))
	for _, it := range items {
		list = append(list, v1.DeviceAdminClientUsageUserFeatureItem{
			FeatureId: it.FeatureId, Description: it.Description, Count: it.Count, LastAt: it.LastAt,
		})
	}
	timeline := make([]v1.DeviceAdminClientUsageTimelineItem, 0, len(tl))
	for _, it := range tl {
		timeline = append(timeline, v1.DeviceAdminClientUsageTimelineItem{
			FeatureId: it.FeatureId, Description: it.Description, At: it.At,
		})
	}
	return &v1.DeviceAdminClientUsageUserRes{
		WxId: req.WxId, List: list, Timeline: timeline, Days: days, SortBy: sortBy,
	}, nil
}

// ClientUsageWxList GET /device/admin/api/client-usage/wx-list
func (c *GatewayAppClientUsageCtrl) ClientUsageWxList(ctx context.Context, req *v1.DeviceAdminClientUsageWxListReq) (*v1.DeviceAdminClientUsageWxListRes, error) {
	if err := c.requireAdmin(ctx); err != nil {
		return nil, err
	}
	page, pageSize := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	rows, total, outPage, outPageSize, err := usagestats.DeviceWxListPage(ctx, page, pageSize, req.Q)
	if err != nil {
		return nil, err
	}
	wxIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		wxIDs = append(wxIDs, row.Id)
	}
	nickMap := usagestats.FetchProfileNicknames(ctx, wxIDs)
	list := make([]v1.DeviceAdminClientUsageWxListItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, v1.DeviceAdminClientUsageWxListItem{
			Id: row.Id, DeviceNo: row.DeviceNo, Unionid: row.Unionid,
			Platform: row.Platform, Account: row.Account, CreatedAt: row.CreatedAt,
			Nickname: nickMap[row.Id],
		})
	}
	return &v1.DeviceAdminClientUsageWxListRes{
		List: list, Total: total, Page: outPage, PageSize: outPageSize,
	}, nil
}
