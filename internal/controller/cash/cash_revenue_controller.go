package cashctrl

// 开通功能管理「收益统计」Admin API。
// 只读写手填流水并只读合计订单；不改支付、权益或 App 接口。

import (
	"context"

	v1 "hello/api/v1"
	"hello/internal/services/cash"
)

// AdminRevenueSummary GET /cash/admin/api/revenue/summary — VIP、按功能名的实付、手填与差额。
//
// Args: start/end 可选，作用于订单 paid_at。
// Returns: 标价合计，不是到账净额。
// Side Effects: 只读。
func (c *CashFeatureController) AdminRevenueSummary(ctx context.Context, req *v1.CashAdminRevenueSummaryReq) (*v1.CashAdminRevenueSummaryRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	sum, err := cash.AdminRevenueSummary(ctx, cash.RevenueWindow{Start: req.Start, End: req.End})
	if err != nil {
		return nil, err
	}
	res := &v1.CashAdminRevenueSummaryRes{
		VipFen:       sum.VipFen,
		Features:     make([]v1.CashAdminRevenueFeatureItem, 0, len(sum.Features)),
		ManualInFen:  sum.ManualInFen,
		ManualOutFen: sum.ManualOutFen,
		NetFen:       sum.NetFen,
	}
	for _, line := range sum.Features {
		res.Features = append(res.Features, v1.CashAdminRevenueFeatureItem{
			Name: line.Name, AmountFen: line.AmountFen,
		})
	}
	return res, nil
}

// AdminRevenueEntries GET /cash/admin/api/revenue/entries — 手填流水，发生时间倒序。
func (c *CashFeatureController) AdminRevenueEntries(ctx context.Context, req *v1.CashAdminRevenueEntriesReq) (*v1.CashAdminRevenueEntriesRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	list, err := cash.AdminListManualLedger(ctx, cash.RevenueWindow{Start: req.Start, End: req.End})
	if err != nil {
		return nil, err
	}
	res := &v1.CashAdminRevenueEntriesRes{List: make([]v1.CashAdminRevenueEntryItem, 0, len(list))}
	for _, it := range list {
		res.List = append(res.List, v1.CashAdminRevenueEntryItem{
			Id: it.Id, Direction: it.Direction, Name: it.Name,
			AmountFen: it.AmountFen, OccurredAt: it.OccurredAt,
		})
	}
	return res, nil
}

// AdminRevenueEntryCreate POST /cash/admin/api/revenue/entries — 新增手填进账或出账。
func (c *CashFeatureController) AdminRevenueEntryCreate(ctx context.Context, req *v1.CashAdminRevenueEntryCreateReq) (*v1.CashAdminRevenueEntryCreateRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	id, err := cash.AdminCreateManualLedger(ctx, cash.ManualLedgerInput{
		Direction: req.Direction, Name: req.Name, AmountFen: req.AmountFen, OccurredAt: req.OccurredAt,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CashAdminRevenueEntryCreateRes{Id: id}, nil
}

// AdminRevenueEntryUpdate POST /cash/admin/api/revenue/entries/update — 按 id 修改手填流水。
func (c *CashFeatureController) AdminRevenueEntryUpdate(ctx context.Context, req *v1.CashAdminRevenueEntryUpdateReq) (*v1.CashAdminRevenueEntryUpdateRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	err := cash.AdminUpdateManualLedger(ctx, req.Id, cash.ManualLedgerInput{
		Direction: req.Direction, Name: req.Name, AmountFen: req.AmountFen, OccurredAt: req.OccurredAt,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CashAdminRevenueEntryUpdateRes{Id: req.Id}, nil
}

// AdminRevenueEntryDelete POST /cash/admin/api/revenue/entries/delete — 按 id 物理删除。
func (c *CashFeatureController) AdminRevenueEntryDelete(ctx context.Context, req *v1.CashAdminRevenueEntryDeleteReq) (*v1.CashAdminRevenueEntryDeleteRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	if err := cash.AdminDeleteManualLedger(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.CashAdminRevenueEntryDeleteRes{}, nil
}
