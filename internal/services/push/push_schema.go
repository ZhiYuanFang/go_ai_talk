package push

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const pushDeviceUniqueIndexName = "uk_wx_device_channel"

// EnsureSchema 启动时建表并保证唯一索引（库 ai_voice_push）。
func EnsureSchema(ctx context.Context) error {
	if _, err := g.DB().Exec(ctx, `
CREATE TABLE IF NOT EXISTS push_device (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  wx_id BIGINT NOT NULL,
  channel VARCHAR(16) NOT NULL,
  token VARCHAR(512) NOT NULL,
  device_key VARCHAR(64) NOT NULL,
  updated_at BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_wx_device_channel (wx_id, device_key, channel),
  KEY idx_wx_id (wx_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		return err
	}
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'push_device' AND INDEX_NAME = ?`, pushDeviceUniqueIndexName)
	if err != nil {
		return err
	}
	if n.Int() > 0 {
		return nil
	}
	if _, err = g.DB().Exec(ctx, `
DELETE t1 FROM push_device t1
INNER JOIN push_device t2
  ON t1.wx_id = t2.wx_id AND t1.device_key = t2.device_key AND t1.channel = t2.channel AND t1.id < t2.id`); err != nil {
		return err
	}
	if _, err = g.DB().Exec(ctx, `
ALTER TABLE push_device ADD UNIQUE KEY uk_wx_device_channel (wx_id, device_key, channel)`); err != nil {
		return err
	}
	glog.Infof(ctx, "[push-schema] push_device.%s 已添加", pushDeviceUniqueIndexName)
	return nil
}
