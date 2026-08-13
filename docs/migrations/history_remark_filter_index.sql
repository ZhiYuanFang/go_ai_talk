-- =============================================================================
-- history 表：备注模糊查询支撑（手工复制到 MySQL 执行）
-- 库：生产 ai_voice_history / 测试 ai_voice_history_test
-- 表：history
--
-- 执行前请备份。若某条索引已存在会报 Duplicate key name，跳过该条即可。
-- 应用进程不会自动执行本文件。
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1) remark 保持可空
--    现网 remark 设计为可空：NULL 与空串都表示「这条记录没有备注」。
--    模糊查询必须排除这些行，不得把空备注当成命中。
--    本脚本禁止把 remark 改成 NOT NULL，也不改默认值。
-- -----------------------------------------------------------------------------
-- 不要执行类似：
--   ALTER TABLE history MODIFY COLUMN remark VARCHAR(...) NOT NULL;
-- 可选自检（应看到 Null = YES）：
--   SHOW COLUMNS FROM history LIKE 'remark';

-- -----------------------------------------------------------------------------
-- 2) 为何不单独给 remark 加普通 B-Tree
--    查询形态是 LIKE '%AD%'（关键词在中间）。
--    这种前导通配无法使用 remark 单列 B-Tree 前缀，加了也加速不了中缀模糊。
--    正确做法：先用设备号 + 时间（+ 事件 ID）把行数收窄，再扫备注。
--    探针路径还会强制 limit <= 20。
-- -----------------------------------------------------------------------------

-- 先看现有索引，避免重复加：
--   SHOW INDEX FROM history;

-- 探针 / 时间窗：按设备 + 开始时间收窄
ALTER TABLE history
  ADD INDEX idx_history_device_start (device_no, start_time);

-- 正式点查：按设备 + 事件 + 开始时间收窄，再 AND 备注模糊
ALTER TABLE history
  ADD INDEX idx_history_device_event_start (device_no, event_id, start_time);

-- -----------------------------------------------------------------------------
-- 3) 可选：备注 FULLTEXT + ngram（中文短词、如 AD）
--    空串 / NULL 不会进入 FULLTEXT，符合「备注可空」。
--    需要 MySQL 已安装 ngram 插件（SHOW VARIABLES LIKE 'ngram_token_size';）。
--    未启用 ngram 时不要执行下面这一条；只跑上面两条复合索引即可。
--    本期 Go filter 可用收窄后的 LIKE；FULLTEXT 留给后续 MATCH AGAINST。
-- -----------------------------------------------------------------------------
-- ALTER TABLE history
--   ADD FULLTEXT INDEX ft_history_remark (remark) WITH PARSER ngram;

-- 确认索引：
--   SHOW INDEX FROM history;
