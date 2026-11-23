# Spec Delta

## ADDED Requirements

### Requirement: History realtime consumers MAY receive mutations via gateway SSE

Feeding history mutation notifications that today reach App clients via history WebSocket MUST also be deliverable via gateway history SSE with the same business JSON fields required for list/get and WS push (including stored expect* when non-null). This change MUST NOT alter history-service persistence rules for add/update/delete. 既有经 history WS 到达 App 的变更通知 MUST 也可经网关历史 SSE 送达，业务字段与 list/WS 一致；MUST NOT 改变 history-service 落库规则。

#### Scenario: expect* 出现在 SSE update 载荷

- **WHEN** 某历史行已存 expect* 且发布 update 推送
- **THEN** 经 SSE 扇出的 payload MUST 在非空时包含这些 expect* 字段（与 WS 推送要求一致）

### Requirement: History Redis notify publish shape MUST stay unchanged for this change

为本变更引入网关 SSE 时，history-service MUST NOT 修改向约定 Redis channel（如 `app:history:notify`）发布的消息体契约：仍 MUST 含 `device_no`、`action`（create|update|delete）与 `payload`。SSE 若需不同包装，MUST 仅在网关 SSE 写出侧完成。History Redis notify publish shape MUST stay unchanged for this change.

#### Scenario: 发布体字段不变

- **WHEN** history-service 在落库成功后 PUBLISH 一条历史变更通知
- **THEN** 消息 MUST 仍包含 `device_no`、`action` 与 `payload`
- **AND** 本变更 MUST NOT 以「适配 SSE」为由删除或重命名这些顶层字段
