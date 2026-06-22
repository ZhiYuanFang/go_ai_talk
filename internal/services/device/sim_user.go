package device

import (
	"context"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/frame/g"
)

// SimWxListItem 模拟用户 wxId 列表项。
type SimWxListItem struct {
	WxId    int64  `json:"wxId"`
	Account string `json:"account"`
}

// SimUsernameRegister 内部注册模拟用户：username 注册 + is_simulated=1。
func SimUsernameRegister(ctx context.Context, userName, password string) (int64, error) {
	wxID, err := WxUsernameRegister(ctx, userName, password)
	if err != nil {
		return 0, err
	}
	_, err = dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, wxID).Data(g.Map{
		dao.Wx.Columns().IsSimulated: 1,
	}).Update()
	if err != nil {
		return 0, err
	}
	_ = invalidateWxCaches(ctx, wxID, "")
	_ = cachekit.Default().SetAdd(ctx, cachekit.GatewayUsageSimWxSetKey(), cachekit.GatewayUsageSimWxMember(wxID))
	return wxID, nil
}

// CountSimulatedWx 统计 is_simulated=1 的用户数。
func CountSimulatedWx(ctx context.Context) (int, error) {
	return dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().IsSimulated, 1).Count()
}

// IsSimulatedWx 判断 wx 是否为模拟用户。
func IsSimulatedWx(ctx context.Context, wxID int64) (bool, error) {
	if wxID <= 0 {
		return false, nil
	}
	row, err := wxRowByWxID(ctx, wxID)
	if err != nil {
		return false, err
	}
	if row == nil {
		return false, nil
	}
	return row.IsSimulated == 1, nil
}

// ListSimulatedWx 分页列出模拟用户 wxId。
func ListSimulatedWx(ctx context.Context, page, pageSize int) (list []SimWxListItem, total int, err error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	model := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().IsSimulated, 1)
	total, err = model.Count()
	if err != nil {
		return nil, 0, err
	}
	rows, err := model.
		Fields(dao.Wx.Columns().Id, dao.Wx.Columns().Account).
		OrderAsc(dao.Wx.Columns().Id).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		All()
	if err != nil {
		return nil, 0, err
	}
	list = make([]SimWxListItem, 0, len(rows))
	for _, row := range rows {
		var w entity.Wx
		if sErr := row.Struct(&w); sErr != nil {
			return nil, 0, sErr
		}
		list = append(list, SimWxListItem{WxId: w.Id, Account: w.Account})
	}
	return list, total, nil
}
