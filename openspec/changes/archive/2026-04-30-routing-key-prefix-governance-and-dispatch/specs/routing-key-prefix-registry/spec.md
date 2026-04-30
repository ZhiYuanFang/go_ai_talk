## ADDED Requirements

### Requirement: 路由键前缀必须集中注册
系统 MUST 在统一注册入口维护路由键前缀常量，并将前缀作为领域分组的唯一来源，禁止在业务模块重复定义核心前缀字面量。

#### Scenario: 新增前缀时集中维护
- **WHEN** 开发者需要新增一个事件族前缀
- **THEN** 必须在统一注册入口新增前缀常量并在路由键定义处复用该常量

### Requirement: 路由键定义必须采用前缀与后缀组合
系统 SHALL 通过“前缀常量 + 后缀常量/字面量”生成路由键枚举值，保证同一事件族的命名一致性。

#### Scenario: 定义 history 事件路由键
- **WHEN** 定义 `history.record.created`、`history.record.updated`、`history.record.deleted`
- **THEN** 这些路由键必须共享 `history.record.` 前缀常量并仅通过后缀区分
