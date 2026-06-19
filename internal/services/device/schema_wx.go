package device

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// EnsureWxIsSimulatedColumn 启动时保证 wx.is_simulated 列存在（幂等）。
func EnsureWxIsSimulatedColumn(ctx context.Context) error {
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'wx' AND COLUMN_NAME = 'is_simulated'`)
	if err != nil {
		return err
	}
	if n.Int() > 0 {
		return nil
	}
	_, err = g.DB().Exec(ctx, `
ALTER TABLE wx ADD COLUMN is_simulated TINYINT NOT NULL DEFAULT 0 COMMENT '1=模拟用户' AFTER password`)
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[device-schema] wx.is_simulated 列已添加")
	return nil
}
