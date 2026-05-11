// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// DomainOutboxDao is the data access object for table domain_outbox.
type DomainOutboxDao struct {
	table   string              // table is the underlying table name of the DAO.
	group   string              // group is the database configuration group name of current DAO.
	columns DomainOutboxColumns // columns contains all the column names of Table for convenient usage.
}

// DomainOutboxColumns defines and stores column names for table domain_outbox.
type DomainOutboxColumns struct {
	Id          string //
	EventId     string //
	EventType   string //
	RoutingKey  string //
	Payload     string //
	Status      string //
	Attempts    string //
	LastError   string //
	PublishedAt string //
	CreatedAt   string //
	UpdatedAt   string //
}

// domainOutboxColumns holds the columns for table domain_outbox.
var domainOutboxColumns = DomainOutboxColumns{
	Id:          "id",
	EventId:     "event_id",
	EventType:   "event_type",
	RoutingKey:  "routing_key",
	Payload:     "payload",
	Status:      "status",
	Attempts:    "attempts",
	LastError:   "last_error",
	PublishedAt: "published_at",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}

// NewDomainOutboxDao creates and returns a new DAO object for table data access.
func NewDomainOutboxDao() *DomainOutboxDao {
	return &DomainOutboxDao{
		group:   "default",
		table:   "domain_outbox",
		columns: domainOutboxColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *DomainOutboxDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *DomainOutboxDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *DomainOutboxDao) Columns() DomainOutboxColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *DomainOutboxDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *DomainOutboxDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *DomainOutboxDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
