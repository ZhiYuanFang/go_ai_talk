## ADDED Requirements

### Requirement: 客户端与 API 使用统计页时间范围含 1/3 天且默认 1 天

`client-usage-stats.html` 与 `api-usage-stats.html` 的时间范围选择器 SHALL 提供以下选项（值分别为请求参数 `days`）：近 1 天（`1`）、近 3 天（`3`）、近 7 天（`7`）、近 30 天（`30`）。页面首次加载时 MUST 默认选中近 1 天。切换选项后发起的列表/明细请求 MUST 携带对应该选项的 `days` 查询参数。MUST NOT 因此变更 wx 列表排序语义或新增自填天数控件。

#### Scenario: 默认近 1 天

- **WHEN** 运维打开客户端使用统计页或 App API 使用统计页
- **THEN** 时间范围 MUST 默认选中近 1 天，且首屏请求的 `days` MUST 为 `1`

#### Scenario: 可选近 3 天

- **WHEN** 运维将时间范围改为近 3 天并触发刷新/重载
- **THEN** 后续统计请求的 `days` MUST 为 `3`

#### Scenario: 保留原 7/30

- **WHEN** 运维查看时间范围选项
- **THEN** 列表 MUST 仍包含近 7 天与近 30 天
