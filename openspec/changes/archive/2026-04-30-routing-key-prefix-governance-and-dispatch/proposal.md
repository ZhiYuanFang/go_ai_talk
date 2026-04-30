## Why

当前路由键治理已具备类型化枚举和白名单校验，但在投影分发链路仍存在“逐项枚举 case”重复问题，新增路由键时需要在多处同步扩展。随着事件数量持续增加，需要引入“前缀分组 + 白名单校验”协同机制，降低维护成本并保持安全边界。

## What Changes

- 在 `eventkit` 增加路由键前缀常量（如 `history.record.`、`device.`、`voice.task.`）并统一命名约定。
- 将路由键常量改为“前缀 + 后缀”组合生成，减少重复裸字符串片段。
- 在 outbox 投影分发中采用“先白名单校验、后前缀分组路由”的双层策略。
- 将 repair/reconcile 等需要按域识别事件的路径统一为前缀入口，避免核心分支持续增长。
- 增加治理文档与迁移验收项，明确新增路由键时必须同步前缀/注册表/分发规则。

## Capabilities

### New Capabilities

- `routing-key-prefix-registry`: 定义路由键前缀常量、前缀命名规范和前缀到业务域的映射规则。
- `validated-prefix-dispatch`: 定义 outbox 链路“白名单校验 + 前缀分发”的统一行为与错误语义。
- `routing-key-governance-workflow`: 定义新增路由键时的注册、分发、文档与验收流程。

### Modified Capabilities

- （无）

## Impact

- Affected code:
  - `internal/platform/eventkit/**`
  - `internal/services/async/domain_outbox.go`
  - `internal/services/history/**`
  - `internal/services/device/**`
- Runtime/system impact:
  - 事件分发仍保持兼容字符串 wire 协议，不改变 MQ/DB 字段格式。
  - 新增前缀治理后，新增路由键时的代码改动点减少，分发分支更稳定。
