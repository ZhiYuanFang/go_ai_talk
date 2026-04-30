# routing-key-governance-workflow Specification

## Purpose
TBD - created by archiving change routing-key-prefix-governance-and-dispatch. Update Purpose after archive.
## Requirements
### Requirement: 新增路由键必须遵循注册流程
系统 MUST 定义并执行新增路由键的标准流程：前缀确认、路由键注册、分发映射、观测校验、文档更新。

#### Scenario: 开发者新增路由键
- **WHEN** 开发者新增一个路由键用于新事件
- **THEN** 必须同时完成注册、分发映射和文档更新，缺任一项视为未完成迁移

### Requirement: 迁移验收必须禁止新增核心裸字符串匹配
系统 SHALL 在迁移验收清单中明确要求：核心分发模块不得新增针对 `routing_key` 的裸字符串匹配分支。

#### Scenario: 代码评审检查分发逻辑
- **WHEN** 评审者检查 outbox 与投影分发相关改动
- **THEN** 若发现新增裸字符串匹配而未使用统一前缀/枚举入口，必须拒绝合并

