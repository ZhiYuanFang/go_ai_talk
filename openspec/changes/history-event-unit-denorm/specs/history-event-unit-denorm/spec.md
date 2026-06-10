## ADDED Requirements

### Requirement: History 写入 MUST 反规范化事件单位

当新增或更新 `history` 记录且关联的 `eventId` 在 device 事件主档中存在非空 `unit` 时，history-service MUST 在持久化前将单位写入 `history.event_unit`（写入时刻快照）。若请求体已携带非空 `eventUnit`，MUST 优先使用该值；否则 MUST 经 device-service HTTP 契约解析单位。history-service MUST NOT 通过本进程 `default` 数据库连接直查 `event` 表。

#### Scenario: 微服务分库下新增历史带单位

- **WHEN** device 库中 `event.id=10` 的 `unit` 为 `ml`，客户端经 history HTTP 新增一条 `eventId=10` 的历史且请求体未传 `eventUnit`
- **THEN** history-service MUST 经 device 契约解析到 `ml` 并成功 INSERT，`history.event_unit` SHALL 为 `ml`

#### Scenario: 请求体显式携带单位

- **WHEN** 客户端 POST 新增历史且 `eventUnit` 为 `次`
- **THEN** history-service MUST 持久化 `history.event_unit=次`，且 MUST NOT 被主档单位覆盖

#### Scenario: 事件主档无单位

- **WHEN** 关联 `event.unit` 为空且请求体未传 `eventUnit`
- **THEN** history-service MUST 持久化 `history.event_unit` 为 NULL 或等效空值，且 MUST NOT 报错

#### Scenario: 禁止跨库直查 event 表

- **WHEN** history-service 运行于仅配置 `HISTORY_DB_LINK` 的环境
- **THEN** 补全 `event_unit` 的实现 MUST NOT 调用 `dao.Event` 或等价 history 库内 event 表访问

### Requirement: History 读路径 MUST 暴露 eventUnit

`GET /device/history/api/list`、`GET /device/history/api/piece` 及单条查询响应中的 history 实体 MUST 包含 `eventUnit` 字段（JSON camelCase），值与数据库 `event_unit` 一致。

#### Scenario: 列表返回单位

- **WHEN** 数据库行 `event_unit=ml`
- **THEN** 列表 JSON 中对应记录的 `eventUnit` SHALL 为 `ml`

### Requirement: 历史管理页 MUST 展示计数单位

`resource/public/history.html` 对 `eventType=number` 的记录，展示数量时 MUST 在数字后附加单位：优先使用记录 `eventUnit`；若为空且所选事件 option 含 unit，MAY 回退展示 option 单位。

#### Scenario: 有 eventUnit 的计数记录

- **WHEN** 列表项 `eventNumber=120` 且 `eventUnit=ml`
- **THEN** 页面数量列 SHALL 展示含 `ml` 的可读文本（如 `120ml` 或 `120 ml`，实现 MUST 文档化一种格式）
