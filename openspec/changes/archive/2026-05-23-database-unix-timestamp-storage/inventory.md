# 时刻类字段盘点（仓库内模型，截至本变更）

基于 `internal/model/entity` 与手工维护实体；**实际 MySQL 类型以现网 DDL 为准**，实施前须 `SHOW CREATE TABLE` 核对。

## ai_voice_device（device-service / default 组）


| 表/实体    | 字段                          | 当前 Go 类型 | 语义            | 目标            |
| ------- | --------------------------- | -------- | ------------- | ------------- |
| user    | active_time, last_talk_time | string   | 注释含「UTF8」历史表述 | BIGINT Unix 秒 |
| user    | birthday                    | string   | 业务生日，见附录      | 见附录           |
| history | start_time, end_time        | string   | 与区间查询可比       | BIGINT Unix 秒 |
| suggest | time                        | string   | 建议产生时刻        | BIGINT Unix 秒 |
| wx      | —                           | —        | 无时刻列          | —             |


## ai_voice_app（gateway-app version 等）


| 表/实体                 | 字段           | 当前 Go 类型 | 语义                                                    | 目标  |
| -------------------- | ------------ | -------- | ----------------------------------------------------- | --- |
| version / AppVersion | release_date | int64    | 已为 Unix **秒**（与 `hack/ddl_wx_and_app_version.sql` 一致） | 保持  |


## 其他 gen 表（event / action / qa）


| 表/实体              | 字段               | 说明                                        |
| ----------------- | ---------------- | ----------------------------------------- |
| event, action, qa | 无独立「时刻」列于 entity | 若库内有 `created_at` 等未映射到 entity，须 DBA 补充盘点 |


## voice / worker

- 以各服务实际连接库为准单独开表；本仓库若后续增加 entity，须回填本清单。

