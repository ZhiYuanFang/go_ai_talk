package dao

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

const defaultDBGroup = "default"

func resolveDBGroup(ctx context.Context, preferred string) string {
	preferred = strings.TrimSpace(preferred)
	if preferred == "" || strings.EqualFold(preferred, defaultDBGroup) {
		return defaultDBGroup
	}
	link, err := g.Cfg().Get(ctx, "database."+preferred+".link")
	if err != nil {
		return defaultDBGroup
	}
	if link.String() == "" {
		return defaultDBGroup
	}
	return preferred
}

func domainModel(ctx context.Context, preferredGroup, table string) *gdb.Model {
	group := resolveDBGroup(ctx, preferredGroup)
	return g.DB(group).Model(table).Safe().Ctx(ctx)
}

func resolveDBGroupChain(ctx context.Context, preferredGroups ...string) string {
	for _, group := range preferredGroups {
		resolved := resolveDBGroup(ctx, group)
		if resolved != defaultDBGroup || strings.EqualFold(strings.TrimSpace(group), defaultDBGroup) {
			return resolved
		}
	}
	return defaultDBGroup
}

func domainReadModel(ctx context.Context, table string, preferredGroups ...string) *gdb.Model {
	group := resolveDBGroupChain(ctx, preferredGroups...)
	return g.DB(group).Model(table).Safe().Ctx(ctx)
}
