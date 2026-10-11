## ADDED Requirements

### Requirement: 规则缺口页默认 TIME/QTY 展示语义

Hub「规则缺口」页在运维未指定维度筛选时，MUST NOT 向列表接口传递 `dimension` 查询参数（从而依赖 Python 默认仅返回 `time` 与 `quantity`）。页面文案 MUST 标明默认展示 TIME/QTY，MUST NOT 再将「空维度」描述为「全部维度」或引导以 type/event/shape 为默认工单。

#### Scenario: 默认进页不传 dimension

- **WHEN** 运维打开规则缺口页且未填写/未选择具体维度
- **THEN** 列表请求省略 `dimension` 参数
- **AND** 页面提示默认仅为 TIME/QTY

#### Scenario: 显式筛选单一维度

- **WHEN** 运维选择或填写 `time` 或 `quantity`（或排障填写 legacy 维度）
- **THEN** 列表请求携带对应 `dimension`
- **AND** 结果按上游过滤返回

### Requirement: 存量 JSONL 清理不经 Go Hub 热路径

本能力 MUST NOT 要求 Go 进程读写 Python `gaps.jsonl`，MUST NOT 新增 Hub「清理存量 JSONL」热按钮或 Go purge API。存量 legacy 维度行的清理若执行，SHALL 由运维在 Python 数据卷按 runbook 可选完成（清空或只保留 time/quantity）。

#### Scenario: Hub 无 JSONL 清理动作

- **WHEN** 运维使用规则缺口管理页
- **THEN** 页面提供列表/筛选/标已处理等既有能力
- **AND** MUST NOT 出现触发 Go 删除/改写 Python JSONL 文件的操作入口
