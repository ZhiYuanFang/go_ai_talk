package ucg

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const pushDeviceUniqueIndexName = "uk_wx_device_channel"

// EnsurePushDeviceUniqueIndex 启动时保证 (wx_id, device_key, channel) 唯一索引存在，支撑 ON DUPLICATE KEY upsert。
func EnsurePushDeviceUniqueIndex(ctx context.Context) error {
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ucg_push_device' AND INDEX_NAME = ?`, pushDeviceUniqueIndexName)
	if err != nil {
		return err
	}
	if n.Int() > 0 {
		return nil
	}
	// 加唯一索引前删重复行，保留 id 最大者，避免 ALTER 失败。
	if _, err = g.DB().Exec(ctx, `
DELETE t1 FROM ucg_push_device t1
INNER JOIN ucg_push_device t2
  ON t1.wx_id = t2.wx_id AND t1.device_key = t2.device_key AND t1.channel = t2.channel AND t1.id < t2.id`); err != nil {
		return err
	}
	_, err = g.DB().Exec(ctx, `
ALTER TABLE ucg_push_device ADD UNIQUE KEY uk_wx_device_channel (wx_id, device_key, channel)`)
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[ucg-schema] ucg_push_device.%s 已添加", pushDeviceUniqueIndexName)
	return nil
}
