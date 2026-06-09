-- UCG 私信持久化：ai_voice_ucg 库执行（test/prod 均需）
-- 字符集 utf8mb4 以支持 emoji

CREATE TABLE IF NOT EXISTS `ucg_chat_message` (
  `id` BIGINT UNSIGNED NOT NULL COMMENT '会话内消息序号，与 Redis seq 一致',
  `conversation_id` BIGINT UNSIGNED NOT NULL,
  `client_msg_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '客户端幂等 ID，空表示无',
  `sender_wx_id` BIGINT UNSIGNED NOT NULL,
  `content` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `image_key` VARCHAR(512) NOT NULL DEFAULT '',
  `video_key` VARCHAR(512) NOT NULL DEFAULT '',
  `media_cdn_url` VARCHAR(1024) NOT NULL DEFAULT '',
  `created_at` BIGINT NOT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'delivered',
  PRIMARY KEY (`conversation_id`, `id`),
  KEY `idx_conv_created` (`conversation_id`, `created_at`),
  KEY `idx_client_msg` (`conversation_id`, `client_msg_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='UCG 私信正文（MySQL 持久权威）';

CREATE TABLE IF NOT EXISTS `ucg_chat_message_outbox` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `conversation_id` BIGINT UNSIGNED NOT NULL,
  `payload` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'ChatMessage JSON',
  `status` VARCHAR(16) NOT NULL DEFAULT 'pending' COMMENT 'pending|done|failed',
  `attempts` INT UNSIGNED NOT NULL DEFAULT 0,
  `last_error` VARCHAR(512) NOT NULL DEFAULT '',
  `created_at` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status_id` (`status`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='UCG 私信异步落库队列';
