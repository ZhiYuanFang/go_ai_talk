-- 参考 DDL：请在目标库执行前核对与现网迁移策略一致。
-- ai_voice_device.wx
CREATE TABLE IF NOT EXISTS `wx` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `union_id` varchar(128) NOT NULL COMMENT '微信开放平台 unionid，多端统一身份',
  `device_no` varchar(64) DEFAULT NULL,
  `platform` varchar(64) DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_wx_union_id` (`union_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ai_voice_app.version
CREATE TABLE IF NOT EXISTS `version` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `latest_version` varchar(64) NOT NULL,
  `release_date` bigint NOT NULL DEFAULT 0 COMMENT '当前版本上线时间，Unix 时间戳（秒）',
  `release_notes` text,
  `download_url` varchar(512) NOT NULL DEFAULT '',
  `force_update` tinyint NOT NULL DEFAULT 0,
  `min_version` varchar(64) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 已有表时增加列（按需执行）
-- ALTER TABLE `version` ADD COLUMN `release_date` bigint NOT NULL DEFAULT 0 COMMENT '上线时间 Unix 秒' AFTER `latest_version`;
-- 若曾误建为 varchar，可改为 bigint：
-- ALTER TABLE `version` MODIFY COLUMN `release_date` bigint NOT NULL DEFAULT 0 COMMENT '上线时间 Unix 秒';
