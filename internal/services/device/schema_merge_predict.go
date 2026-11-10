package device

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// EnsureEventMergePredictColumn 启动时保证 event.merge_predict 列存在（幂等）。
// 业务：合并预测标志默认 0；仅「有直接子的根」可置 1，与 is_appointment / event_type 正交。
func EnsureEventMergePredictColumn(ctx context.Context) error {
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'event' AND COLUMN_NAME = 'merge_predict'`)
	if err != nil {
		return err
	}
	if n.Int() > 0 {
		return nil
	}
	_, err = g.DB().Exec(ctx, `
ALTER TABLE event ADD COLUMN merge_predict TINYINT NOT NULL DEFAULT 0 COMMENT '1=合并预测（仅有子根）' AFTER is_appointment`)
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[device-schema] event.merge_predict 列已添加")
	return nil
}
