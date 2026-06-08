-- UCG moments compose redesign: media upload ownership + AI config singleton.
-- Database: ai_voice_ucg (prod) / ai_voice_ucg_test (test)
-- Run in maintenance window before deploying ucg-service with new handlers.

CREATE TABLE IF NOT EXISTS `ucg_media_upload` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `wx_id` bigint unsigned NOT NULL COMMENT 'uploader wx id',
  `object_key` varchar(512) NOT NULL COMMENT 'OSS object key',
  `media_kind` tinyint NOT NULL DEFAULT 1 COMMENT '1=image 2=video',
  `created_at` bigint NOT NULL DEFAULT 0 COMMENT 'unix seconds',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_object_key` (`object_key`),
  KEY `idx_wx_id` (`wx_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='UCG media upload ownership log';

CREATE TABLE IF NOT EXISTS `ucg_ai_config` (
  `id` int NOT NULL COMMENT 'singleton row id=1',
  `vision_model` varchar(64) NOT NULL DEFAULT 'qwen3-vl-plus',
  `max_images_per_request` int NOT NULL DEFAULT 9,
  `updated_at` bigint NOT NULL DEFAULT 0 COMMENT 'unix seconds',
  `updated_by` varchar(64) NOT NULL DEFAULT '' COMMENT 'admin operator label',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='UCG AI runtime config singleton';

INSERT INTO `ucg_ai_config` (`id`, `vision_model`, `max_images_per_request`, `updated_at`, `updated_by`)
VALUES (1, 'qwen3-vl-plus', 9, UNIX_TIMESTAMP(), 'seed')
ON DUPLICATE KEY UPDATE `id` = `id`;

-- 已有环境若 seed 为 deepseek-chat / VL2，执行一次 Qwen vision 模型迁移（新部署可忽略）。
UPDATE `ucg_ai_config`
SET `vision_model` = 'qwen3-vl-plus',
    `updated_at` = UNIX_TIMESTAMP(),
    `updated_by` = 'qwen-vl-migration'
WHERE `id` = 1 AND `vision_model` IN (
  'deepseek-chat',
  'deepseek-vl2',
  'deepseek-ai/deepseek-vl2',
  'deepseek-ai/deepseek-vl2-small',
  'deepseek-ai/deepseek-vl2-tiny'
);
