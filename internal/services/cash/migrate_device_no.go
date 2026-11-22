package cash

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// MigrateDeviceNo 将对机权益与次数从 old device_no 迁到 new（宝宝号改名）。
// 业务逻辑：若新号上已存在同 feature 权益或 allowed_count 行则失败（避免静默覆盖）；
// feature_order 等审计快照不跟迁。
func MigrateDeviceNo(ctx context.Context, oldDeviceNo, newDeviceNo string) error {
	old := strings.TrimSpace(oldDeviceNo)
	neu := strings.TrimSpace(newDeviceNo)
	if old == "" || neu == "" {
		return fmt.Errorf("deviceNo 不能为空")
	}
	if old == neu {
		return fmt.Errorf("新旧 deviceNo 不能相同")
	}
	db := g.DB()
	return db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 冲突探测：新号已有 entitlement
		n, err := tx.Model("feature_entitlement").Ctx(ctx).Where("device_no", neu).Count()
		if err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("新号已存在对机权益行，拒绝迁移以免覆盖")
		}
		n2, err := tx.Model("feature_allowed_count").Ctx(ctx).Where("device_no", neu).Count()
		if err != nil {
			return err
		}
		if n2 > 0 {
			return fmt.Errorf("新号已存在预测次数行，拒绝迁移以免覆盖")
		}
		if _, err := tx.Model("feature_entitlement").Ctx(ctx).
			Data("device_no", neu).Where("device_no", old).Update(); err != nil {
			return fmt.Errorf("migrate feature_entitlement: %w", err)
		}
		if _, err := tx.Model("feature_allowed_count").Ctx(ctx).
			Data("device_no", neu).Where("device_no", old).Update(); err != nil {
			return fmt.Errorf("migrate feature_allowed_count: %w", err)
		}
		// 遗留 invite device grant（若仍有行）
		if _, err := tx.Model("feature_invite_device_grant").Ctx(ctx).
			Data("device_no", neu).Where("device_no", old).Update(); err != nil {
			glog.Warningf(ctx, "[cash] migrate feature_invite_device_grant: %v", err)
		}
		glog.Infof(ctx, "[cash] MigrateDeviceNo ok oldLen=%d newLen=%d", len(old), len(neu))
		return nil
	})
}
