package voice

import (
	"context"

	"hello/internal/dao"
	"hello/internal/model/entity"
)

// ListQaAllForAdmin 返回问答库全表行（供管理后台与 device 域 HTTP 委派；权威数据在 voice 库 qa 表）。
func ListQaAllForAdmin(ctx context.Context) ([]entity.Qa, error) {
	rows := make([]entity.Qa, 0)
	err := dao.Qa.Ctx(ctx).Fields(
		dao.Qa.Columns().Id,
		dao.Qa.Columns().Question,
		dao.Qa.Columns().Replay,
		dao.Qa.Columns().Attack,
	).OrderAsc(dao.Qa.Columns().Id).Scan(&rows)
	return rows, err
}
