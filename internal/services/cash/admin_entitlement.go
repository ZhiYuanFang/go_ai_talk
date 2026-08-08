// 管理端 VIP 权益只读列表。
//
// 业务说明：
// 运维 Hub「VIP 权益」页通过 GET /cash/admin/api/vip/entitlements 拉取本域数据。
// 主表为 vip_entitlement 全量行（含已过期）；激活金额取该 wx_id 最近一次 status=paid 订单的 amount_fen。
// 禁止 N+1：用相关子查询 LEFT JOIN 最新 paid 订单；本期不加 Redis、不跨库。
package cash

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AdminEntitlementRow 管理端权益列表单行（含最近实付）。
type AdminEntitlementRow struct {
	WxId               int64  `json:"wxId"`
	IsVip              bool   `json:"isVip"`
	ExpireAt           int64  `json:"expireAt"`
	RemainingSeconds   int64  `json:"remainingSeconds"`
	LastPaidAmountFen  int    `json:"lastPaidAmountFen"`
	Channel            string `json:"channel"`
	PaidAt             int64  `json:"paidAt"`
}

// AdminEntitlementListResult 分页结果。
type AdminEntitlementListResult struct {
	List     []AdminEntitlementRow
	Page     int
	PageSize int
	Total    int
}

// ListEntitlementsForAdmin 分页列出 VIP 权益（含已过期）及最近 paid 订单金额。
//
// 业务逻辑：
//   - 主数据：vip_entitlement 全量；可选 wxId 精确过滤；默认按 expire_at DESC 排序。
//   - 激活金额：LEFT JOIN 每 wx_id 在 vip_order 中 status=paid、按 paid_at DESC / id DESC 取一条。
//   - isVip / remainingSeconds：相对当前 unix 秒派生；过期时 isVip=false，remainingSeconds 可为负。
//
// Args:
//   - page: 页码，从 1 起；非法则置 1。
//   - pageSize: 每页条数，默认 20，最大 200。
//   - wxID: >0 时仅查该账号（0 或 1 行）。
//
// Returns: 列表与 total；Side Effects: 仅读本库，无写。
func ListEntitlementsForAdmin(ctx context.Context, page, pageSize int, wxID int64) (AdminEntitlementListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	out := AdminEntitlementListResult{Page: page, PageSize: pageSize, List: []AdminEntitlementRow{}}

	countM := g.DB().Model("vip_entitlement").Ctx(ctx)
	if wxID > 0 {
		countM = countM.Where("wx_id", wxID)
	}
	total, err := countM.Count()
	if err != nil {
		return out, err
	}
	out.Total = total
	if total == 0 {
		return out, nil
	}

	offset := (page - 1) * pageSize
	// 相关子查询：每行权益附带最近一次 paid 订单；避免按 list 再查订单（N+1）。
	// 索引评估：现有 idx_wx_created 可辅助；若慢再加 (wx_id, status, paid_at)，本期默认不加 DDL。
	sql := `
SELECT e.wx_id AS wx_id,
       e.expire_at AS expire_at,
       COALESCE(o.amount_fen, 0) AS last_paid_amount_fen,
       COALESCE(o.channel, '') AS channel,
       COALESCE(o.paid_at, 0) AS paid_at
FROM vip_entitlement e
LEFT JOIN vip_order o ON o.id = (
  SELECT id FROM vip_order
  WHERE wx_id = e.wx_id AND status = ?
  ORDER BY paid_at DESC, id DESC
  LIMIT 1
)`
	args := []interface{}{OrderPaid}
	if wxID > 0 {
		sql += ` WHERE e.wx_id = ?`
		args = append(args, wxID)
	}
	sql += ` ORDER BY e.expire_at DESC LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	records, err := g.DB().GetAll(ctx, sql, args...)
	if err != nil {
		return out, err
	}
	now := time.Now().Unix()
	for _, r := range records {
		expireAt := r["expire_at"].Int64()
		remaining := expireAt - now
		out.List = append(out.List, AdminEntitlementRow{
			WxId:              r["wx_id"].Int64(),
			IsVip:             expireAt > now,
			ExpireAt:          expireAt,
			RemainingSeconds:  remaining,
			LastPaidAmountFen: r["last_paid_amount_fen"].Int(),
			Channel:           r["channel"].String(),
			PaidAt:            r["paid_at"].Int64(),
		})
	}
	return out, nil
}
