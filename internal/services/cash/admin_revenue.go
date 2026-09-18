package cash

// 管理端收益统计。
//
// 业务：开通功能管理对比标价收入与手填成本。
// 设计：cash_manual_ledger 只记管理者手填的进账/出账；VIP 与功能实付打开时对订单求和，不落快照。
// 使用场景：Hub「收益统计」面板的增删改查与合计。

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const (
	// LedgerDirectionIn 手填进账。
	LedgerDirectionIn = "in"
	// LedgerDirectionOut 手填出账（抽成、服务器、模型等）。
	LedgerDirectionOut = "out"
)

// RevenueWindow 合计与列表共用的时间窗。Start/End 为 0 表示该侧不限制。
type RevenueWindow struct {
	Start int64
	End   int64
}

// ManualLedgerEntry 一条手填流水。
type ManualLedgerEntry struct {
	Id         int64  `json:"id"`
	Direction  string `json:"direction"`
	Name       string `json:"name"`
	AmountFen  int    `json:"amountFen"`
	OccurredAt int64  `json:"occurredAt"`
}

// ManualLedgerInput 新增或修改手填流水。OccurredAt 为 0 时新增用当前时间，修改则拒绝。
type ManualLedgerInput struct {
	Direction  string
	Name       string
	AmountFen  int
	OccurredAt int64
}

// RevenueFeatureLine 按功能名拆开的实付。
type RevenueFeatureLine struct {
	Name      string `json:"name"`
	AmountFen int    `json:"amountFen"`
}

// RevenueSummary 只读合计。订单金额是用户支付的标价，不是到账净额。
type RevenueSummary struct {
	VipFen       int                  `json:"vipFen"`
	Features     []RevenueFeatureLine `json:"features"`
	ManualInFen  int                  `json:"manualInFen"`
	ManualOutFen int                  `json:"manualOutFen"`
	NetFen       int                  `json:"netFen"`
}

// AdminCreateManualLedger 新增手填流水。
//
// Args: in.Direction 须为 in 或 out；名称去空白后非空且不超过 128 字；金额须≥1。
// Returns: 新行 id。
// Side Effects: INSERT cash_manual_ledger。不写订单。
func AdminCreateManualLedger(ctx context.Context, in ManualLedgerInput) (int64, error) {
	dir, name, err := normalizeLedgerFields(in.Direction, in.Name, in.AmountFen)
	if err != nil {
		return 0, err
	}
	now := time.Now().Unix()
	occurred := in.OccurredAt
	if occurred == 0 {
		occurred = now
	}
	if occurred < 1 {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "发生时间无效")
	}
	res, err := g.DB().Model("cash_manual_ledger").Ctx(ctx).Data(g.Map{
		"direction":   dir,
		"name":        name,
		"amount_fen":  in.AmountFen,
		"occurred_at": occurred,
		"created_at":  now,
		"updated_at":  now,
	}).Insert()
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return id, err
}

// AdminUpdateManualLedger 按 id 修改方向、名称、金额、发生时间。
//
// Args: id 须≥1；occurredAt 须≥1（修改不默认成现在）。
// Side Effects: UPDATE 一行。不存在则拒绝。
func AdminUpdateManualLedger(ctx context.Context, id int64, in ManualLedgerInput) error {
	if id < 1 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "id 无效")
	}
	dir, name, err := normalizeLedgerFields(in.Direction, in.Name, in.AmountFen)
	if err != nil {
		return err
	}
	if in.OccurredAt < 1 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "发生时间无效")
	}
	now := time.Now().Unix()
	res, err := g.DB().Model("cash_manual_ledger").Ctx(ctx).Where("id", id).Data(g.Map{
		"direction":   dir,
		"name":        name,
		"amount_fen":  in.AmountFen,
		"occurred_at": in.OccurredAt,
		"updated_at":  now,
	}).Update()
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		// 值未变化时 MySQL 影响行数为 0，先确认行还在。
		one, qErr := g.DB().Model("cash_manual_ledger").Ctx(ctx).Where("id", id).One()
		if qErr != nil {
			return qErr
		}
		if one.IsEmpty() {
			return gerror.NewCode(gcode.CodeInvalidParameter, "记录不存在")
		}
	}
	return nil
}

// AdminDeleteManualLedger 按 id 物理删除手填流水。
//
// Side Effects: DELETE 一行。不存在则拒绝。不改订单。
func AdminDeleteManualLedger(ctx context.Context, id int64) error {
	if id < 1 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "id 无效")
	}
	res, err := g.DB().Model("cash_manual_ledger").Ctx(ctx).Where("id", id).Delete()
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "记录不存在")
	}
	return nil
}

// AdminListManualLedger 按发生时间倒序列出手填流水。
//
// Args: win.Start>0 则 occurred_at>=start；win.End>0 则 occurred_at<end。都为 0 则全部。
func AdminListManualLedger(ctx context.Context, win RevenueWindow) ([]ManualLedgerEntry, error) {
	m := g.DB().Model("cash_manual_ledger").Ctx(ctx)
	if win.Start > 0 {
		m = m.Where("occurred_at >= ?", win.Start)
	}
	if win.End > 0 {
		m = m.Where("occurred_at < ?", win.End)
	}
	rows, err := m.OrderDesc("occurred_at").OrderDesc("id").All()
	if err != nil {
		return nil, err
	}
	out := make([]ManualLedgerEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, ManualLedgerEntry{
			Id:         r["id"].Int64(),
			Direction:  r["direction"].String(),
			Name:       r["name"].String(),
			AmountFen:  r["amount_fen"].Int(),
			OccurredAt: r["occurred_at"].Int64(),
		})
	}
	return out, nil
}

// AdminRevenueSummary 现场汇总 VIP 实付、按功能名拆开的功能实付、手填进账/出账与差额。
//
// 业务：只计 status=paid 且 channel 为 alipay 或 apple_iap。手工授与退款不计入。
// 差额 = VIP + 各功能付费 + 手填进账 − 手填出账。
// Args: 时间窗作用于订单 paid_at 与手填 occurred_at。
func AdminRevenueSummary(ctx context.Context, win RevenueWindow) (RevenueSummary, error) {
	var out RevenueSummary
	out.Features = []RevenueFeatureLine{}
	vip, err := sumPaidOrders(ctx, "vip_order", win)
	if err != nil {
		return out, err
	}
	out.VipFen = vip
	features, featureSum, err := sumFeaturePaidByName(ctx, win)
	if err != nil {
		return out, err
	}
	out.Features = features
	inFen, outFen, err := sumManualByDirection(ctx, win)
	if err != nil {
		return out, err
	}
	out.ManualInFen = inFen
	out.ManualOutFen = outFen
	out.NetFen = vip + featureSum + inFen - outFen
	return out, nil
}

func normalizeLedgerFields(direction, name string, amountFen int) (string, string, error) {
	dir := strings.ToLower(strings.TrimSpace(direction))
	if dir != LedgerDirectionIn && dir != LedgerDirectionOut {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "方向须为 in 或 out")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "名称不能为空")
	}
	if utf8.RuneCountInString(name) > 128 {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "名称最长 128 字")
	}
	if amountFen < 1 {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "金额须≥1 分")
	}
	return dir, name, nil
}

// sumPaidOrders 对一张订单表求实付金额（分）。
func sumPaidOrders(ctx context.Context, table string, win RevenueWindow) (int, error) {
	// 表名仅来自本文件常量，不拼接外部输入。
	sql := `SELECT COALESCE(SUM(amount_fen),0) FROM ` + table + `
WHERE status=? AND channel IN (?,?)
AND (?=0 OR paid_at>=?) AND (?=0 OR paid_at<?)`
	v, err := g.DB().GetValue(ctx, sql, paidSumArgs(win)...)
	if err != nil {
		return 0, err
	}
	return v.Int(), nil
}

// sumFeaturePaidByName 功能实付按功能名分组。对不上商品的订单用商品编码单独成组。
func sumFeaturePaidByName(ctx context.Context, win RevenueWindow) ([]RevenueFeatureLine, int, error) {
	sql := `SELECT
  IF(IFNULL(fp.feature_id,'')='', fo.product_code, IF(IFNULL(fd.title,'')='', fp.feature_id, fd.title)) AS display_name,
  COALESCE(SUM(fo.amount_fen),0) AS amount_fen
FROM feature_order fo
LEFT JOIN feature_product fp ON fp.product_code = fo.product_code
LEFT JOIN feature_def fd ON fd.feature_id = fp.feature_id
WHERE fo.status=? AND fo.channel IN (?,?)
AND (?=0 OR fo.paid_at>=?) AND (?=0 OR fo.paid_at<?)
GROUP BY
  IF(IFNULL(fp.feature_id,'')='', fo.product_code, fp.feature_id),
  IF(IFNULL(fp.feature_id,'')='', fo.product_code, IF(IFNULL(fd.title,'')='', fp.feature_id, fd.title))
ORDER BY amount_fen DESC`
	rows, err := g.DB().GetAll(ctx, sql, paidSumArgs(win)...)
	if err != nil {
		return nil, 0, err
	}
	lines := make([]RevenueFeatureLine, 0, len(rows))
	sum := 0
	for _, r := range rows {
		amt := r["amount_fen"].Int()
		sum += amt
		lines = append(lines, RevenueFeatureLine{
			Name:      r["display_name"].String(),
			AmountFen: amt,
		})
	}
	return lines, sum, nil
}

func sumManualByDirection(ctx context.Context, win RevenueWindow) (inFen, outFen int, err error) {
	sql := `SELECT direction, COALESCE(SUM(amount_fen),0) AS amount_fen
FROM cash_manual_ledger
WHERE (?=0 OR occurred_at>=?) AND (?=0 OR occurred_at<?)
GROUP BY direction`
	rows, err := g.DB().GetAll(ctx, sql, win.Start, win.Start, win.End, win.End)
	if err != nil {
		return 0, 0, err
	}
	for _, r := range rows {
		switch r["direction"].String() {
		case LedgerDirectionIn:
			inFen = r["amount_fen"].Int()
		case LedgerDirectionOut:
			outFen = r["amount_fen"].Int()
		}
	}
	return inFen, outFen, nil
}

func paidSumArgs(win RevenueWindow) []interface{} {
	return []interface{}{
		OrderPaid, ChannelAlipay, ChannelAppleIAP,
		win.Start, win.Start, win.End, win.End,
	}
}
