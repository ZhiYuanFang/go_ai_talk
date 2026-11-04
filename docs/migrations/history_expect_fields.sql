-- =============================================================================
-- history 表：预计字段 expect_*（手工复制到 MySQL 执行）
-- 库：生产 ai_voice_history / 测试 ai_voice_history_test
-- 表：history
-- 应用进程：history-service（HISTORY_DB_LINK → GoFrame database.default）
--
-- 执行前请备份。若某列已存在会报 Duplicate column name，跳过该条即可。
-- 应用进程不会自动执行本文件。
-- =============================================================================

-- 预计开始时间（unix 秒，可空；老行 / 未传保持 NULL）
ALTER TABLE history
  ADD COLUMN expect_start_time BIGINT NULL COMMENT '预计开始时间 unix 秒' AFTER video_key;

-- 预计结束时间（unix 秒，可空）
ALTER TABLE history
  ADD COLUMN expect_end_time BIGINT NULL COMMENT '预计结束时间 unix 秒' AFTER expect_start_time;

-- 预计数量（浮点，可空；禁止 INT 截断）
ALTER TABLE history
  ADD COLUMN expect_amount DOUBLE NULL COMMENT '预计数量（浮点）' AFTER expect_end_time;

-- 确认：
--   SHOW COLUMNS FROM history LIKE 'expect_%';
