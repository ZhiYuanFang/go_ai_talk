package voice

import (
	"context"
	"errors"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/services/contracts"
)

const qaListMaxPageSize = 100

// ListQaPage 管理端分页查询问答库（id 倒序，较新记录在前）。
func ListQaPage(ctx context.Context, page, pageSize int) (contracts.QaPageResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > qaListMaxPageSize {
		pageSize = qaListMaxPageSize
	}
	total, err := dao.Qa.Ctx(ctx).Count()
	if err != nil {
		return contracts.QaPageResult{}, err
	}
	if total == 0 {
		return contracts.QaPageResult{List: []entity.Qa{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}
	rows := make([]entity.Qa, 0)
	err = dao.Qa.Ctx(ctx).Fields(
		dao.Qa.Columns().Id,
		dao.Qa.Columns().Question,
		dao.Qa.Columns().Replay,
		dao.Qa.Columns().Attack,
	).OrderDesc(dao.Qa.Columns().Id).Page(page, pageSize).Scan(&rows)
	if err != nil {
		return contracts.QaPageResult{}, err
	}
	return contracts.QaPageResult{List: rows, Total: total, Page: page, PageSize: pageSize}, nil
}

// ListQaAllForAdmin 兼容内部全量拉取（最多 qaListMaxPageSize 条，id 倒序）。
func ListQaAllForAdmin(ctx context.Context) ([]entity.Qa, error) {
	res, err := ListQaPage(ctx, 1, qaListMaxPageSize)
	if err != nil {
		return nil, err
	}
	return res.List, nil
}

// DeleteQa 按主键删除问答库行。
func DeleteQa(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("问答库 id 无效")
	}
	_, err := dao.Qa.Ctx(ctx).Where(dao.Qa.Columns().Id, id).Delete()
	return err
}
