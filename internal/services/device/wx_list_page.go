package device

import (
	"context"
	"strconv"
	"strings"

	"hello/internal/dao"
	"hello/internal/services/contracts"
)

const wxListDefaultPageSize = 20
const wxListMaxPageSize = 100

// ListWxPage 管理端 wx 账号分页；q 模糊匹配 id/deviceNo/unionid/account。
// 业务逻辑：排除模拟号；对本页 deviceNo 批量联查 user.baby_name，填充 BabyName（无则空串）。
func (s *service) ListWxPage(ctx context.Context, page, pageSize int, q string) (contracts.WxPageResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = wxListDefaultPageSize
	}
	if pageSize > wxListMaxPageSize {
		pageSize = wxListMaxPageSize
	}
	cols := dao.Wx.Columns()
	m := dao.Wx.Ctx(ctx).Where(cols.IsSimulated, 0)
	q = strings.TrimSpace(q)
	if q != "" {
		like := "%" + q + "%"
		if id, err := strconv.ParseInt(q, 10, 64); err == nil {
			m = m.Where(
				cols.Id+" = ? OR "+cols.DeviceNo+" LIKE ? OR "+cols.Unionid+" LIKE ? OR "+cols.Account+" LIKE ?",
				id, like, like, like,
			)
		} else {
			m = m.Where(
				cols.DeviceNo+" LIKE ? OR "+cols.Unionid+" LIKE ? OR "+cols.Account+" LIKE ?",
				like, like, like,
			)
		}
	}
	total, err := m.Count()
	if err != nil {
		return contracts.WxPageResult{}, err
	}
	if total == 0 {
		return contracts.WxPageResult{List: []contracts.AdminWxListItem{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}
	rows := make([]contracts.AdminWxListItem, 0)
	err = m.Fields(cols.Id, cols.DeviceNo, cols.Unionid, cols.Platform, cols.Account, cols.CreatedAt).
		OrderDesc(cols.Id).
		Page(page, pageSize).
		Scan(&rows)
	if err != nil {
		return contracts.WxPageResult{}, err
	}
	// 批量补齐宝宝名，避免调用方按 deviceNo N+1 查画像。
	fillWxListBabyNames(ctx, rows)
	return contracts.WxPageResult{List: rows, Total: total, Page: page, PageSize: pageSize}, nil
}

// fillWxListBabyNames 按本页 deviceNo 批量读取 user.baby_name 并写回行。
// Side Effects: 仅内存填充 rows；查库失败时保留空串，不中断列表。
func fillWxListBabyNames(ctx context.Context, rows []contracts.AdminWxListItem) {
	if len(rows) == 0 {
		return
	}
	deviceNos := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		dn := strings.TrimSpace(r.DeviceNo)
		if dn == "" {
			continue
		}
		if _, ok := seen[dn]; ok {
			continue
		}
		seen[dn] = struct{}{}
		deviceNos = append(deviceNos, dn)
	}
	if len(deviceNos) == 0 {
		return
	}
	uCols := dao.User.Columns()
	type babyRow struct {
		DeviceNo string `json:"deviceNo"`
		BabyName string `json:"babyName"`
	}
	var babies []babyRow
	_ = dao.User.Ctx(ctx).
		Fields(uCols.DeviceNo, uCols.BabyName).
		WhereIn(uCols.DeviceNo, deviceNos).
		Scan(&babies)
	byDevice := make(map[string]string, len(babies))
	for _, b := range babies {
		byDevice[strings.TrimSpace(b.DeviceNo)] = strings.TrimSpace(b.BabyName)
	}
	for i := range rows {
		rows[i].BabyName = byDevice[strings.TrimSpace(rows[i].DeviceNo)]
	}
}
