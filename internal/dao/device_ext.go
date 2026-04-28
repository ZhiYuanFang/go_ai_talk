package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
)

func (dao eventDao) Group() string {
	return resolveDBGroup(context.Background(), "device")
}

func (dao eventDao) Ctx(ctx context.Context) *gdb.Model {
	return domainModel(ctx, "device", dao.Table())
}

func (dao actionDao) Group() string {
	return resolveDBGroup(context.Background(), "device")
}

func (dao actionDao) Ctx(ctx context.Context) *gdb.Model {
	return domainModel(ctx, "device", dao.Table())
}

func (dao userDao) Group() string {
	return resolveDBGroup(context.Background(), "device")
}

func (dao userDao) Ctx(ctx context.Context) *gdb.Model {
	return domainModel(ctx, "device", dao.Table())
}
