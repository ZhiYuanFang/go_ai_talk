## ADDED Requirements

### Requirement: Gateway SHALL 支持 voice WebSocket 边缘透传
`gateway-service` MUST 在 `/voice/chat/ws` 提供可配置透传能力，将 WebSocket 连接转发到 `voice-service` 目标地址。

#### Scenario: WS 透传启用且目标可达
- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且 `VOICE_WS_PROXY_URL` 为可达地址
- **THEN** gateway MUST 将 `/voice/chat/ws` 的握手与后续双向消息透传至目标服务

#### Scenario: WS 透传启用但目标不可达
- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且目标服务不可达或握手失败
- **THEN** gateway MUST 返回明确的握手/代理失败错误，且 MUST NOT 回退本地业务执行

### Requirement: Gateway MUST 提供 WS 透传配置约束
gateway MUST 通过环境变量控制 WS 路由行为，并对非法配置执行可预测回退。

#### Scenario: 路由模式非法
- **WHEN** `VOICE_WS_ROUTE_MODE` 非 `proxy`
- **THEN** gateway MUST 将 WS 路由模式视为 `local`

#### Scenario: 代理地址为空或非法
- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且 `VOICE_WS_PROXY_URL` 为空或非法
- **THEN** gateway MUST 视为未启用可用代理目标并返回可诊断错误，不得出现静默成功
