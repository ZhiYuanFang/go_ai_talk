-- 参考：时刻类列统一为 Unix 秒（BIGINT）时的常见写法。执行前须与现网 SHOW CREATE TABLE 核对表名/列名。
-- 本文件为迁移参考模板，不在 CI 中自动执行。

-- 示例：新增秒列 + 回填后切换（双写期应用可读两列）
-- ALTER TABLE `history` ADD COLUMN `start_unix_sec` bigint DEFAULT NULL COMMENT '开始时刻 Unix 秒 UTC' AFTER `start_time`;
-- UPDATE `history` SET `start_unix_sec` = UNIX_TIMESTAMP(STR_TO_DATE(`start_time`, '%Y-%m-%d %H:%i:%s')) WHERE `start_time` REGEXP '^[0-9]{4}-'; -- 按实际旧格式调整
-- 若旧值已是「纯数字秒」字符串：UPDATE history SET start_unix_sec = CAST(start_time AS UNSIGNED) WHERE start_time REGEXP '^[0-9]+$';

-- 示例：version.release_date 已为秒级 bigint 时无需改动，仅作对齐注释：
-- ALTER TABLE `version` MODIFY COLUMN `release_date` bigint NOT NULL DEFAULT 0 COMMENT '上线时间 Unix 秒 UTC';
