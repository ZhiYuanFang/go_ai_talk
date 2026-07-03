-- ucg_post 辩论帖字段（ai_voice_ucg）
-- 执行前备份；执行后 make dao.sync 并部署 ucg-service

ALTER TABLE ucg_post
  ADD COLUMN type VARCHAR(16) NOT NULL DEFAULT 'moment' COMMENT 'moment|debate' AFTER author_wx_id,
  ADD COLUMN debate_left_label VARCHAR(5) NOT NULL DEFAULT '' COMMENT '辩论左立场标签' AFTER content,
  ADD COLUMN debate_right_label VARCHAR(5) NOT NULL DEFAULT '' COMMENT '辩论右立场标签' AFTER debate_left_label;

CREATE INDEX idx_ucg_post_type_status_published ON ucg_post (type, status, published_at DESC);
