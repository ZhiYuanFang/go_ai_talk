# gateway-ws-edge-proxy Specification

## Purpose
TBD - created by archiving change gateway-ws-proxy-and-remove-local-ws. Update Purpose after archive.
## Requirements
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

### Requirement: Gateway SHALL 支持听写 WebSocket 边缘透传

`gateway-service` 与 `gateway-app-server` MUST 在 `/voice/asr/ws` 提供与 `/voice/chat/ws` 相同的可配置 WebSocket 透传能力，将连接转发至 `voice-service`（与 `VOICE_WS_PROXY_URL` 同一目标基址）。

#### Scenario: WS 透传启用且目标可达

- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且 `VOICE_WS_PROXY_URL` 为可达地址
- **AND** 客户端连接 `/voice/asr/ws`
- **THEN** gateway MUST 将该路径的握手与后续双向消息透传至 voice-service，行为与 `/voice/chat/ws` 一致

#### Scenario: WS 透传启用但目标不可达

- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且目标服务不可达或握手失败
- **AND** 客户端连接 `/voice/asr/ws`
- **THEN** gateway MUST 返回明确的握手/代理失败错误，且 MUST NOT 在 gateway 本地执行听写业务逻辑

#### Scenario: 路由模式非 proxy

- **WHEN** `VOICE_WS_ROUTE_MODE` 非 `proxy`
- **AND** 客户端连接 `/voice/asr/ws`
- **THEN** gateway MUST 返回可诊断的配置错误（与 chat WS 一致），且 MUST NOT 静默成功

### Requirement: App 网关 SHALL 将听写 WS 纳入 Bearer 白名单

`gateway-app-server` MUST 将 `GET /voice/asr/ws`（WebSocket Upgrade）列入 Bearer 鉴权豁免路径，与 `/voice/chat/ws` 策略一致。

#### Scenario: 无 Bearer 的 Upgrade 请求

- **WHEN** App 客户端对 `/voice/asr/ws` 发起 WebSocket Upgrade 且未携带 `Authorization`
- **THEN** gateway-app MUST 允许进入透传或 voice-service 处理链，不得仅因缺少 Bearer 拒绝 Upgrade

