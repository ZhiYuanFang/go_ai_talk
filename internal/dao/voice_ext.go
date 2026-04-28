package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
)

func (dao qaDao) Group() string {
	return resolveDBGroup(context.Background(), "voice")
}

func (dao qaDao) Ctx(ctx context.Context) *gdb.Model {
	return domainModel(ctx, "voice", dao.Table())
}

func (dao suggestDao) Group() string {
	return resolveDBGroup(context.Background(), "voice")
}

func (dao suggestDao) Ctx(ctx context.Context) *gdb.Model {
	return domainModel(ctx, "voice", dao.Table())
}
