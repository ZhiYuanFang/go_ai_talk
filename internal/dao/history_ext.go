package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
)

func (dao historyDao) Group() string {
	return resolveDBGroup(context.Background(), "history")
}

func (dao historyDao) Ctx(ctx context.Context) *gdb.Model {
	return domainModel(ctx, "history", dao.Table())
}

func (dao historyDao) ReadCtx(ctx context.Context) *gdb.Model {
	return domainReadModel(ctx, dao.Table(), "history_ro", "history", "default")
}

func (dao historyDao) WriteCtx(ctx context.Context) *gdb.Model {
	return domainModel(ctx, "history", dao.Table())
}
