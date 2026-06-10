package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AiQuotaDefaultDao is the data access object for table ai_quota_default.
type AiQuotaDefaultDao struct {
	table   string
	group   string
	columns AiQuotaDefaultColumns
}

// AiQuotaDefaultColumns defines column names for table ai_quota_default.
type AiQuotaDefaultColumns struct {
	Id                  string
	PolishMonthlyLimit  string
	VoiceAiMonthlyLimit string
	UpdatedAt           string
}

var aiQuotaDefaultColumns = AiQuotaDefaultColumns{
	Id:                  "id",
	PolishMonthlyLimit:  "polish_monthly_limit",
	VoiceAiMonthlyLimit: "voice_ai_monthly_limit",
	UpdatedAt:           "updated_at",
}

func NewAiQuotaDefaultDao() *AiQuotaDefaultDao {
	return &AiQuotaDefaultDao{
		group:   "default",
		table:   "ai_quota_default",
		columns: aiQuotaDefaultColumns,
	}
}

func (dao *AiQuotaDefaultDao) DB() gdb.DB              { return g.DB(dao.group) }
func (dao *AiQuotaDefaultDao) Table() string           { return dao.table }
func (dao *AiQuotaDefaultDao) Columns() AiQuotaDefaultColumns { return dao.columns }
func (dao *AiQuotaDefaultDao) Group() string           { return dao.group }
func (dao *AiQuotaDefaultDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}
