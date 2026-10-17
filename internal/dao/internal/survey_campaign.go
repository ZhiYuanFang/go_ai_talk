// 满意度问卷活动表 survey_campaign 的 DAO。
// 业务：开关、上海时区起止 unix 秒、说明；App 只消费 id 最大的一轮。

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SurveyCampaignDao 活动表访问对象。
type SurveyCampaignDao struct {
	table   string
	group   string
	columns SurveyCampaignColumns
}

// SurveyCampaignColumns 列名。
type SurveyCampaignColumns struct {
	Id        string
	Enabled   string
	StartAt   string
	EndAt     string
	Intro     string
	CreatedAt string
}

var surveyCampaignColumns = SurveyCampaignColumns{
	Id:        "id",
	Enabled:   "enabled",
	StartAt:   "start_at",
	EndAt:     "end_at",
	Intro:     "intro",
	CreatedAt: "created_at",
}

// NewSurveyCampaignDao 构造活动 DAO。
func NewSurveyCampaignDao() *SurveyCampaignDao {
	return &SurveyCampaignDao{group: "default", table: "survey_campaign", columns: surveyCampaignColumns}
}

func (dao *SurveyCampaignDao) DB() gdb.DB                     { return g.DB(dao.group) }
func (dao *SurveyCampaignDao) Table() string                  { return dao.table }
func (dao *SurveyCampaignDao) Columns() SurveyCampaignColumns { return dao.columns }
func (dao *SurveyCampaignDao) Group() string                  { return dao.group }
func (dao *SurveyCampaignDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}
func (dao *SurveyCampaignDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) error {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
