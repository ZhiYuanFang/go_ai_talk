package device

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"math"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/database/gdb"
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

// PickRandomSimulatedWx 经 ID 探测从全库 is_simulated=1 用户随机取 1 条。
func PickRandomSimulatedWx(ctx context.Context) (SimWxListItem, bool, error) {
	bounds, err := simWxSimulatedModel(ctx).
		Fields(
			"MIN("+dao.Wx.Columns().Id+") as min_id",
			"MAX("+dao.Wx.Columns().Id+") as max_id",
		).
		One()
	if err != nil {
		return SimWxListItem{}, false, err
	}
	if bounds.IsEmpty() {
		return SimWxListItem{}, false, nil
	}
	minID := bounds["min_id"].Int64()
	maxID := bounds["max_id"].Int64()
	if minID <= 0 || maxID <= 0 {
		return SimWxListItem{}, false, nil
	}

	anchor := minID
	if minID < maxID {
		u, uErr := simWxRandomUnit()
		if uErr != nil {
			return SimWxListItem{}, false, uErr
		}
		span := float64(maxID - minID)
		// 均匀锚点：所有 simulated 用户在 id 轴上等概率，不做新用户偏置。
		anchor = minID + int64(math.Floor(span*u))
		if anchor > maxID {
			anchor = maxID
		}
	}

	row, err := simWxSimulatedModel(ctx).
		Fields(dao.Wx.Columns().Id, dao.Wx.Columns().Account).
		Where("id >= ?", anchor).
		OrderAsc(dao.Wx.Columns().Id).
		Limit(1).
		One()
	if err != nil {
		return SimWxListItem{}, false, err
	}
	if row.IsEmpty() {
		// 锚点落在 id 空洞时回退 minId，保证 eligible 用户存在时必命中一条。
		row, err = simWxSimulatedModel(ctx).
			Fields(dao.Wx.Columns().Id, dao.Wx.Columns().Account).
			Where(dao.Wx.Columns().Id, minID).
			Limit(1).
			One()
		if err != nil {
			return SimWxListItem{}, false, err
		}
	}
	if row.IsEmpty() {
		return SimWxListItem{}, false, nil
	}
	var w entity.Wx
	if sErr := row.Struct(&w); sErr != nil {
		return SimWxListItem{}, false, sErr
	}
	return SimWxListItem{WxId: w.Id, Account: w.Account}, true, nil
}

func simWxSimulatedModel(ctx context.Context) *gdb.Model {
	return dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().IsSimulated, 1)
}

// simWxRandomUnit 返回 [0,1) 均匀随机数，供锚点探测使用。
func simWxRandomUnit() (float64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return float64(binary.BigEndian.Uint64(b[:])>>11) / float64(1<<53), nil
}
