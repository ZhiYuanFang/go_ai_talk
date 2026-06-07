-- ucg_notification: comment/@ inbox notifications (Option A — no auto DM)
CREATE TABLE IF NOT EXISTS `ucg_notification` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `recipient_wx_id` bigint unsigned NOT NULL COMMENT '接收者 wxId',
  `type` varchar(32) NOT NULL COMMENT 'comment_on_post | mention_in_comment',
  `post_id` bigint unsigned NOT NULL,
  `comment_id` bigint unsigned NOT NULL,
  `actor_wx_id` bigint unsigned NOT NULL COMMENT '评论者',
  `preview` varchar(200) NOT NULL DEFAULT '' COMMENT '评论摘要',
  `read_at` bigint DEFAULT NULL COMMENT 'NULL=未读',
  `created_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_recipient_created` (`recipient_wx_id`, `created_at`),
  KEY `idx_recipient_read` (`recipient_wx_id`, `read_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
