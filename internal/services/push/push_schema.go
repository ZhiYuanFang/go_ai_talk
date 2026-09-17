package push

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	// pushDeviceUniqueIndexName 同账号同机同通道唯一（注册 upsert 键）。
	pushDeviceUniqueIndexName = "uk_wx_device_channel"
	// pushDeviceTokenUniqueIndexName token 全局唯一：一部手机一个 token，换号后来顶上。
	pushDeviceTokenUniqueIndexName = "uk_token"
)

// EnsureSchema 启动时建表并保证唯一索引（库 ai_voice_push）。
//
// 业务逻辑：
//  1. 建表（新库直接带双唯一键）；
//  2. 幂等补 uk_wx_device_channel（历史迁移）；
//  3. 幂等补 uk_token：先按 token 去重（保留 updated_at 最大，并列取更大 id），再加唯一索引。
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
  UNIQUE KEY uk_token (token),
  KEY idx_wx_id (wx_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		return err
	}
	if err := ensurePushDeviceIndex(ctx, pushDeviceUniqueIndexName, ensureUkWxDeviceChannel); err != nil {
		return err
	}
	if err := ensurePushDeviceIndex(ctx, pushDeviceTokenUniqueIndexName, ensureUkToken); err != nil {
		return err
	}
	return nil
}

// ensurePushDeviceIndex 若索引不存在则执行 migrate（幂等探测 information_schema）。
func ensurePushDeviceIndex(ctx context.Context, indexName string, migrate func(context.Context) error) error {
	n, err := g.DB().GetValue(ctx, `
SELECT COUNT(*) FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'push_device' AND INDEX_NAME = ?`, indexName)
	if err != nil {
		return err
	}
	if n.Int() > 0 {
		return nil
	}
	if err = migrate(ctx); err != nil {
		return err
	}
	glog.Infof(ctx, "[push-schema] push_device.%s 已添加", indexName)
	return nil
}

// ensureUkWxDeviceChannel 清理同 (wx,device_key,channel) 重复行后加唯一索引。
func ensureUkWxDeviceChannel(ctx context.Context) error {
	if _, err := g.DB().Exec(ctx, `
DELETE t1 FROM push_device t1
INNER JOIN push_device t2
  ON t1.wx_id = t2.wx_id AND t1.device_key = t2.device_key AND t1.channel = t2.channel AND t1.id < t2.id`); err != nil {
		return err
	}
	_, err := g.DB().Exec(ctx, `
ALTER TABLE push_device ADD UNIQUE KEY uk_wx_device_channel (wx_id, device_key, channel)`)
	return err
}

// ensureUkToken 按 token 去重后加全局唯一索引。
// 保留规则：同 token 保留 updated_at 最大的一行；并列则保留 id 更大者（后来/更新者赢）。
func ensureUkToken(ctx context.Context) error {
	if _, err := g.DB().Exec(ctx, `
DELETE t1 FROM push_device t1
INNER JOIN push_device t2
  ON t1.token = t2.token
 AND (
   t1.updated_at < t2.updated_at
   OR (t1.updated_at = t2.updated_at AND t1.id < t2.id)
 )`); err != nil {
		return err
	}
	_, err := g.DB().Exec(ctx, `
ALTER TABLE push_device ADD UNIQUE KEY uk_token (token)`)
	return err
}
