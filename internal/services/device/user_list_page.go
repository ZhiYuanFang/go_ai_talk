package device

import (
	"context"
	"strings"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/services/contracts"
)

const userListDefaultPageSize = 5
const userListMaxPageSize = 100

// ListUsersPage 管理端设备分页列表；q 非空时对 device_no 模糊匹配。
func (s *service) ListUsersPage(ctx context.Context, page, pageSize int, q string) (contracts.UserPageResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = userListDefaultPageSize
	}
	if pageSize > userListMaxPageSize {
		pageSize = userListMaxPageSize
	}
	m := dao.User.Ctx(ctx)
	q = strings.TrimSpace(q)
	if q != "" {
		m = m.WhereLike(dao.User.Columns().DeviceNo, "%"+q+"%")
	}
	total, err := m.Count()
	if err != nil {
		return contracts.UserPageResult{}, err
	}
	if total == 0 {
		return contracts.UserPageResult{List: []entity.User{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}
	rows := make([]entity.User, 0)
	err = m.Fields(
		dao.User.Columns().DeviceNo,
		dao.User.Columns().ActiveTime,
		dao.User.Columns().LastTalkTime,
		dao.User.Columns().LastTalkAsk,
		dao.User.Columns().LastTalkAnswer,
		dao.User.Columns().LastApiPath,
		dao.User.Columns().LastApiAt,
	).OrderDesc(dao.User.Columns().LastApiAt).OrderDesc(dao.User.Columns().Id).
		Page(page, pageSize).Scan(&rows)
	if err != nil {
		return contracts.UserPageResult{}, err
	}
	return contracts.UserPageResult{List: rows, Total: total, Page: page, PageSize: pageSize}, nil
}
