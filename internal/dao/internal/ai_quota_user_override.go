package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AiQuotaUserOverrideDao is the data access object for table ai_quota_user_override.
type AiQuotaUserOverrideDao struct {
	table   string
	group   string
	columns AiQuotaUserOverrideColumns
}

// AiQuotaUserOverrideColumns defines column names for table ai_quota_user_override.
type AiQuotaUserOverrideColumns struct {
	WxId                string
	PolishMonthlyLimit  string
	VoiceAiMonthlyLimit string
	UpdatedAt           string
}

var aiQuotaUserOverrideColumns = AiQuotaUserOverrideColumns{
	WxId:                "wx_id",
	PolishMonthlyLimit:  "polish_monthly_limit",
	VoiceAiMonthlyLimit: "voice_ai_monthly_limit",
	UpdatedAt:           "updated_at",
}

func NewAiQuotaUserOverrideDao() *AiQuotaUserOverrideDao {
	return &AiQuotaUserOverrideDao{
		group:   "default",
		table:   "ai_quota_user_override",
		columns: aiQuotaUserOverrideColumns,
	}
}

func (dao *AiQuotaUserOverrideDao) DB() gdb.DB { return g.DB(dao.group) }
func (dao *AiQuotaUserOverrideDao) Table() string { return dao.table }
func (dao *AiQuotaUserOverrideDao) Columns() AiQuotaUserOverrideColumns { return dao.columns }
func (dao *AiQuotaUserOverrideDao) Group() string { return dao.group }
func (dao *AiQuotaUserOverrideDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}
