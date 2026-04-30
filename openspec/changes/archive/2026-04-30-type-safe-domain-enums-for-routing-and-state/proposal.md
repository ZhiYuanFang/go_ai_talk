## Why

当前项目在 `routing_key`、状态机状态、`target_type`、`mode`、`event type` 等关键领域值上大量使用裸字符串匹配，存在拼写漂移、重构脆弱、跨模块语义不一致等风险。随着缓存投影和异步链路复杂度提升，需要通过类型化枚举（代码层）建立统一约束，同时保持 wire 协议字符串兼容。

## What Changes

- 为核心领域常量引入类型化枚举（`type X string + const`）与集中注册表，替代散落裸字符串。
- 为 `routing_key` 增加合法值校验与白名单机制，拒绝未注册路由键进入 outbox/MQ 链路。
- 为状态机状态、`target_type`、`mode`、`event type` 增加统一转换与校验函数，统一日志与错误语义。
- 在关键入口（producer/consumer/controller/service）收敛到枚举值匹配，减少 ad-hoc 字符串判断。
- 保持数据库字段和消息协议层仍为字符串，不引入破坏性协议变更；仅提升代码层类型安全。

## Capabilities

### New Capabilities
- `typed-domain-enums`: 定义核心领域值的类型化枚举、注册与校验能力。
- `routing-key-governance`: 定义路由键白名单、校验与拒绝策略，覆盖 outbox 与发布链路。
- `enum-adapter-compatibility`: 定义枚举与 wire 字符串之间的双向适配与兼容规则。

### Modified Capabilities
- （无）

## Impact

- Affected code:
  - `internal/services/async/**`
  - `internal/platform/eventkit/**`
  - `internal/services/voice/**`
  - `internal/services/history/**`
  - `internal/services/device/**`
  - `internal/services/contracts/**`
- Affected systems:
  - outbox + MQ 事件发布/消费链路（routing_key 校验）
  - 语音意图和模式路由（target_type/mode 枚举化）
  - 缓存投影事件类型治理（event type）
- Runtime/ops:
  - 新增非法路由键告警与拒绝日志
  - 增加枚举兼容转换错误观测
