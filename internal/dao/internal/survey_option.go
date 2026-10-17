// 满意度问卷选项表 survey_option 的 DAO。
// 业务：一行一个满意点；enabled=0 不对 App 展示，已提交快照不受影响。

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SurveyOptionDao 选项表访问对象。
type SurveyOptionDao struct {
	table   string
	group   string
	columns SurveyOptionColumns
}

// SurveyOptionColumns 列名。
type SurveyOptionColumns struct {
	Id         string
	CampaignId string
	Text       string
	Sort       string
	Enabled    string
}

var surveyOptionColumns = SurveyOptionColumns{
	Id:         "id",
	CampaignId: "campaign_id",
	Text:       "text",
	Sort:       "sort",
	Enabled:    "enabled",
}

// NewSurveyOptionDao 构造选项 DAO。
func NewSurveyOptionDao() *SurveyOptionDao {
	return &SurveyOptionDao{group: "default", table: "survey_option", columns: surveyOptionColumns}
}

func (dao *SurveyOptionDao) DB() gdb.DB                   { return g.DB(dao.group) }
func (dao *SurveyOptionDao) Table() string                { return dao.table }
func (dao *SurveyOptionDao) Columns() SurveyOptionColumns { return dao.columns }
func (dao *SurveyOptionDao) Group() string                { return dao.group }
func (dao *SurveyOptionDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}
