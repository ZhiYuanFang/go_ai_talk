package device

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// EnsureWxIsSimulatedColumnDropped 启动时幂等 DROP wx.is_simulated。
// 业务逻辑：sim-user 退役后不再读写该列；若列仍存在，先尽量删除 is_simulated=1 行，再 DROP。
// Side Effects: 可能 DELETE 模拟号行并 ALTER TABLE DROP COLUMN。
func EnsureWxIsSimulatedColumnDropped(ctx context.Context) error {
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'wx' AND COLUMN_NAME = 'is_simulated'`)
	if err != nil {
		return err
	}
	if n.Int() == 0 {
		return nil
	}
	// 列仍在：先清模拟号行，避免 DROP 后残留无标记的假账号。
	if _, delErr := g.DB().Exec(ctx, `DELETE FROM wx WHERE is_simulated = 1`); delErr != nil {
		glog.Warningf(ctx, "[device-schema] 清理 is_simulated=1 行失败（继续 DROP）: %v", delErr)
	}
	if _, err = g.DB().Exec(ctx, `ALTER TABLE wx DROP COLUMN is_simulated`); err != nil {
		return err
	}
	glog.Infof(ctx, "[device-schema] wx.is_simulated 列已删除")
	return nil
}
