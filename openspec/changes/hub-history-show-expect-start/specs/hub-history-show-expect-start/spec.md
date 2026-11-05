## ADDED Requirements

### Requirement: Hub history event table SHALL show expected start time

The Hub device history page (`history.html`) history events table MUST include a read-only column **预计开始** that displays each row’s `expectStartTime` when present (Unix seconds formatted consistently with the page’s start-time display helpers). When `expectStartTime` is missing, null, or non-positive, the cell MUST show a clear empty placeholder (e.g. `-`). This MUST NOT require new backend API fields beyond what list already returns.

Hub 设备历史页历史事件表 MUST 含只读列 **预计开始**，有 `expectStartTime` 时按与开始时间一致的可读格式展示；缺失/空/非正 MUST 显示占位（如 `-`）。MUST NOT 要求新增后端 list 字段（沿用既有回传）。

#### Scenario: Row with expectStartTime

- **WHEN** list 某行含正 `expectStartTime`
- **THEN** 「预计开始」单元格 MUST 显示对应可读时间

#### Scenario: Row without expect

- **WHEN** list 某行无 `expectStartTime` 或值无效
- **THEN** 「预计开始」单元格 MUST 显示占位符（如 `-`）

### Requirement: Usage stats embedded history tab SHALL show the same expected start column

On `api-usage-stats` and `client-usage-stats` user-detail **历史事件** tables, the column set MUST include **预计开始** with the same semantics and formatting intent as the Hub `history.html` history events table, so operators see consistent fields when drilling from usage stats.

使用统计两页右侧历史 Tab 表 MUST 含 **预计开始**，语义与格式意图与 `history.html` 历史事件表一致。

#### Scenario: Side panel shows expect start

- **WHEN** 运维在使用统计按用户选中设备并打开历史 Tab 且某行有 `expectStartTime`
- **THEN** 该行「预计开始」MUST 可见且非空占位
