// 满意度问卷提交表 survey_response 的 DAO。
// 业务：唯一键 (campaign_id, wx_id)；option_snapshot 为提交当时的文案 JSON。

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SurveyResponseDao 提交表访问对象。
type SurveyResponseDao struct {
	table   string
	group   string
	columns SurveyResponseColumns
}

// SurveyResponseColumns 列名。
type SurveyResponseColumns struct {
	Id             string
	CampaignId     string
	WxId           string
	DeviceNo       string
	BabyName       string
	OptionSnapshot string
	Comment        string
	SubmittedAt    string
}

var surveyResponseColumns = SurveyResponseColumns{
	Id:             "id",
	CampaignId:     "campaign_id",
	WxId:           "wx_id",
	DeviceNo:       "device_no",
	BabyName:       "baby_name",
	OptionSnapshot: "option_snapshot",
	Comment:        "comment",
	SubmittedAt:    "submitted_at",
}

// NewSurveyResponseDao 构造提交 DAO。
func NewSurveyResponseDao() *SurveyResponseDao {
	return &SurveyResponseDao{group: "default", table: "survey_response", columns: surveyResponseColumns}
}

func (dao *SurveyResponseDao) DB() gdb.DB                     { return g.DB(dao.group) }
func (dao *SurveyResponseDao) Table() string                  { return dao.table }
func (dao *SurveyResponseDao) Columns() SurveyResponseColumns { return dao.columns }
func (dao *SurveyResponseDao) Group() string                  { return dao.group }
func (dao *SurveyResponseDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}
