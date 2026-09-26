package device

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// EnsureEventIsAppointmentColumn 启动时保证 event.is_appointment 列存在（幂等）。
// 业务：预约标志默认 0；与 event_type 正交。
func EnsureEventIsAppointmentColumn(ctx context.Context) error {
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'event' AND COLUMN_NAME = 'is_appointment'`)
	if err != nil {
		return err
	}
	if n.Int() > 0 {
		return nil
	}
	_, err = g.DB().Exec(ctx, `
ALTER TABLE event ADD COLUMN is_appointment TINYINT NOT NULL DEFAULT 0 COMMENT '1=预约事件' AFTER parent_id`)
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[device-schema] event.is_appointment 列已添加")
	return nil
}

// EnsureAppointmentNextTable 启动时保证 appointment_next 表存在（幂等）。
// 业务：按 (device_no, event_id) 存下次约定；next_at=0 表示无约定且保留行（不清行）。
func EnsureAppointmentNextTable(ctx context.Context) error {
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.TABLES
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'appointment_next'`)
	if err != nil {
		return err
	}
	if n.Int() > 0 {
		return nil
	}
	_, err = g.DB().Exec(ctx, `
CREATE TABLE appointment_next (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键',
  device_no VARCHAR(64) NOT NULL DEFAULT '' COMMENT '宝宝设备号',
  event_id BIGINT NOT NULL DEFAULT 0 COMMENT '事件字典 ID',
  next_at BIGINT NOT NULL DEFAULT 0 COMMENT '下次约定 unix 秒；0=无约定',
  PRIMARY KEY (id),
  UNIQUE KEY uk_device_event (device_no, event_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='预约事件下次约定（客户端记忆）'`)
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[device-schema] appointment_next 表已创建")
	return nil
}
