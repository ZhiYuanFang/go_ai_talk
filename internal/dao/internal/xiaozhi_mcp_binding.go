// ==========================================================================
// 小智 MCP 绑定表 xiaozhi_mcp_binding 的 DAO（手写，属 device 库）。
// 业务：wx 账号下挂 N 个小智音箱 token；mcp_token 全局唯一；喂养落点为 device_no。
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// XiaozhiMcpBindingDao is the data access object for table xiaozhi_mcp_binding.
type XiaozhiMcpBindingDao struct {
	table   string
	group   string
	columns XiaozhiMcpBindingColumns
}

// XiaozhiMcpBindingColumns 列名。
type XiaozhiMcpBindingColumns struct {
	Id         string
	WxId       string
	DeviceNo   string
	McpToken   string
	SpeakerMac string
	Alias      string
	Status     string
	CreatedAt  string
	UpdatedAt  string
}

var xiaozhiMcpBindingColumns = XiaozhiMcpBindingColumns{
	Id:         "id",
	WxId:       "wx_id",
	DeviceNo:   "device_no",
	McpToken:   "mcp_token",
	SpeakerMac: "speaker_mac",
	Alias:      "alias",
	Status:     "status",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
}

// NewXiaozhiMcpBindingDao creates and returns a new DAO object.
func NewXiaozhiMcpBindingDao() *XiaozhiMcpBindingDao {
	return &XiaozhiMcpBindingDao{
		group:   "default",
		table:   "xiaozhi_mcp_binding",
		columns: xiaozhiMcpBindingColumns,
	}
}

func (dao *XiaozhiMcpBindingDao) DB() gdb.DB {
	return g.DB(dao.group)
}

func (dao *XiaozhiMcpBindingDao) Table() string {
	return dao.table
}

func (dao *XiaozhiMcpBindingDao) Columns() XiaozhiMcpBindingColumns {
	return dao.columns
}

func (dao *XiaozhiMcpBindingDao) Group() string {
	return dao.group
}

func (dao *XiaozhiMcpBindingDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

func (dao *XiaozhiMcpBindingDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
