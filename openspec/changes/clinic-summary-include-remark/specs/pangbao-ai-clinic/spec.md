## MODIFIED Requirements

### Requirement: Clinic SHALL 注入近 7 天喂养事件聚合摘要

每次处理 `question` 前，系统 MUST 取得该设备近 7 天喂养 history，并按 event 聚合为摘要（含 count、amount 合计、duration 合计等），注入 LLM system/context。摘要 JSON MUST 为 object，含：

- **`by_event`**：按 event 聚合统计数组（字段含 count、amount 合计、duration 合计等，与现网一致）；
- **`records_with_remark`**：近 7 天内 **`remark` 非空**（trim 后）的记录列表，每条 MUST 含 `event_name`、`start_time`（本地 `YYYY-MM-DD HH:mm:ss`）、`remark`；MAY 含 `amount_value`、`duration_minutes`。该列表 MUST 按 `start_time` **降序**排列，且 MUST NOT 超过 **30** 条。单条 `remark` MUST 截断至不超过 **200** 字符（或等价 rune 长度限制）。

摘要 MUST NOT 为 7 天内**全部** history 行的全量 dump（无备注记录 MUST NOT 出现在 `records_with_remark` 中，仅计入 `by_event` 聚合）。history 数据 MUST 经 HTTP 契约（如 `DeviceHistory`）获取；voice-service MUST NOT 直连 history 库表。

#### Scenario: 摘要在 prompt 中可见

- **WHEN** 设备近 7 天有 3 次「母乳」记录
- **THEN** 注入 LLM 的上下文中 SHALL 包含 `by_event` 聚合统计
- **AND** token 量 SHALL 小于同等条数全量原始 JSON 行列表

#### Scenario: 有备注记录进入 records_with_remark

- **WHEN** 近 7 天内有 2 条「母乳」记录且其中 1 条 remark 为「左侧，宝宝不太配合」
- **THEN** 摘要 JSON 的 `records_with_remark` SHALL 含 1 条
- **AND** 该条 SHALL 含 `event_name`、`start_time` 与完整 remark 文本（≤200 字）
- **AND** `by_event` 中「母乳」的 count SHALL 仍为 2

#### Scenario: 超过 30 条有备注记录时截断

- **WHEN** 近 7 天内有 40 条 remark 非空的 history
- **THEN** `records_with_remark` MUST 仅含 30 条
- **AND** MUST 为 start_time 最近的 30 条

#### Scenario: 无备注时 records_with_remark 为空

- **WHEN** 近 7 天所有 history 的 remark 均为空
- **THEN** `records_with_remark` SHALL 为 `[]`
- **AND** `by_event` SHALL 仍含聚合统计

#### Scenario: history 契约失败

- **WHEN** history HTTP 契约不可用或返回错误
- **THEN** 系统 SHALL 返回可诊断 `error` 帧且 MUST NOT 在无摘要时静默调用 LLM（除非 design 明确降级为空摘要并记录日志——本 spec 要求显式失败或空摘要+警告日志二选一，实现 MUST 在 design 中择一并在日志中可观测）
