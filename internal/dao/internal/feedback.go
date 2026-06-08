package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// FeedbackDao is the data access object for table feedback.
type FeedbackDao struct {
	table   string
	group   string
	columns FeedbackColumns
}

// FeedbackColumns defines and stores column names for table feedback.
type FeedbackColumns struct {
	Id            string
	WxId          string
	Question      string
	OfficialReply string
	Status        string
	CreatedAt     string
	UpdatedAt     string
	RepliedAt     string
}

var feedbackColumns = FeedbackColumns{
	Id:            "id",
	WxId:          "wx_id",
	Question:      "question",
	OfficialReply: "official_reply",
	Status:        "status",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
	RepliedAt:     "replied_at",
}

// NewFeedbackDao creates and returns a new DAO object for table data access.
func NewFeedbackDao() *FeedbackDao {
	return &FeedbackDao{
		group:   "default",
		table:   "feedback",
		columns: feedbackColumns,
	}
}

func (dao *FeedbackDao) DB() gdb.DB {
	return g.DB(dao.group)
}

func (dao *FeedbackDao) Table() string {
	return dao.table
}

func (dao *FeedbackDao) Columns() FeedbackColumns {
	return dao.columns
}

func (dao *FeedbackDao) Group() string {
	return dao.group
}

func (dao *FeedbackDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

func (dao *FeedbackDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
