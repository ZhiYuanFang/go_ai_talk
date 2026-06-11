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
	m := dao.Wx.Ctx(ctx)
	q = strings.TrimSpace(q)
	if q != "" {
		like := "%" + q + "%"
		cols := dao.Wx.Columns()
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
	cols := dao.Wx.Columns()
	rows := make([]contracts.AdminWxListItem, 0)
	err = m.Fields(cols.Id, cols.DeviceNo, cols.Unionid, cols.Platform, cols.Account).
		OrderDesc(cols.Id).
		Page(page, pageSize).
		Scan(&rows)
	if err != nil {
		return contracts.WxPageResult{}, err
	}
	return contracts.WxPageResult{List: rows, Total: total, Page: page, PageSize: pageSize}, nil
}
