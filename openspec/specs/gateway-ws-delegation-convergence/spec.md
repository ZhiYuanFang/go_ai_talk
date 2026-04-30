# gateway-ws-delegation-convergence Specification

## Purpose
TBD - created by archiving change gateway-ws-proxy-and-remove-local-ws. Update Purpose after archive.
## Requirements
### Requirement: Gateway MUST 移除本地 voice WS 领域执行
`gateway-service` MUST 不再在 `/voice/chat/ws` 执行本地语音对话业务逻辑，领域处理必须由 `voice-service` 承担。

#### Scenario: Gateway 收到 voice WS 请求
- **WHEN** 客户端连接 `/voice/chat/ws`
- **THEN** gateway MUST 仅执行边缘层职责（路由、策略、元数据透传），并将领域执行委派给 `voice-service`

### Requirement: Gateway SHALL 保持对外 WS 入口契约稳定
迁移到委派模式时，gateway MUST 保持外部 WS 路径与接入方式稳定，避免要求前端同步改地址。

#### Scenario: 前端继续使用既有 WS 地址
- **WHEN** 前端仍连接 gateway 既有 `/voice/chat/ws` 地址
- **THEN** 系统 MUST 可完成握手与消息收发，且业务执行由下游 `voice-service` 完成

