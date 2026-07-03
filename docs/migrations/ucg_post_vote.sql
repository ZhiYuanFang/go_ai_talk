-- 辩论帖投票表（ai_voice_ucg）

CREATE TABLE IF NOT EXISTS ucg_post_vote (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  post_id BIGINT UNSIGNED NOT NULL,
  voter_wx_id BIGINT UNSIGNED NOT NULL,
  side ENUM('left','right') NOT NULL,
  created_at BIGINT NOT NULL DEFAULT 0,
  updated_at BIGINT NOT NULL DEFAULT 0,
  UNIQUE KEY uk_post_voter (post_id, voter_wx_id),
  KEY idx_post_side (post_id, side)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
