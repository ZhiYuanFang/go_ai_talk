## ADDED Requirements

### Requirement: gateway-app-server SHALL 注册胖宝 clinic WebSocket 为 App 对外入口

`gateway-app-server` MUST 将 `GET /voice/clinic/ws` 注册为 App 客户端的**唯一对外 WebSocket 入口**（与 `/voice/chat/ws`、`/voice/asr/ws` 同 `apiBaseUrl` 主机、同 `installVoiceWSProxyMiddleware` 透传链）。实现 MUST 将 `/voice/clinic/ws` 加入 `internal/controller/ws_route_proxy.go` 的 `voiceWSProxyPaths`，由 `RegisterGatewayAppHTTP` 已挂载的 `installVoiceWSProxyMiddleware` 将握手与双向消息透传至 `voice-service`（`VOICE_WS_PROXY_URL` 同一目标基址）。App 客户端 MUST NOT 被要求或配置为直连 `voice-service` 内网地址。

#### Scenario: App 经 gateway-app 连接 clinic WS

- **WHEN** Flutter 使用 `wss://{apiBaseUrl host}/voice/clinic/ws` 发起 WebSocket Upgrade
- **AND** `VOICE_WS_ROUTE_MODE=proxy` 且 `VOICE_WS_PROXY_URL` 可达
- **THEN** gateway-app-server MUST 透传握手与后续双向帧至 voice-service `/voice/clinic/ws`
- **AND** gateway-app MUST NOT 在本地执行 clinic 业务逻辑

#### Scenario: WS 透传启用但目标不可达

- **WHEN** `VOICE_WS_ROUTE_MODE=proxy` 且目标服务不可达或握手失败
- **AND** 客户端连接 gateway-app `/voice/clinic/ws`
- **THEN** gateway-app MUST 返回明确的握手/代理失败错误，且 MUST NOT 在 gateway 本地执行 clinic 业务逻辑

#### Scenario: 路由模式非 proxy

- **WHEN** `VOICE_WS_ROUTE_MODE` 非 `proxy`
- **AND** 客户端连接 gateway-app `/voice/clinic/ws`
- **THEN** gateway-app MUST 返回可诊断的配置错误，且 MUST NOT 静默成功

### Requirement: gateway-service MUST 同步 clinic WebSocket 透传路径

`gateway-service` MUST 将 `/voice/clinic/ws` 加入同一 `voiceWSProxyPaths` 列表，行为与 `/voice/asr/ws` 一致，以便管理/通用网关与 App 网关路径对齐；**App 主入口仍为 gateway-app-server**。

#### Scenario: gateway-service 透传 clinic WS

- **WHEN** 客户端连接 gateway-service `/voice/clinic/ws` 且 proxy 模式配置正确
- **THEN** gateway-service MUST 透传至 voice-service，行为与 chat/ASR WS 一致

### Requirement: gateway-app-server SHALL 将 clinic WS 纳入 Bearer 白名单

`gateway-app-server` MUST 将 `GET /voice/clinic/ws`（WebSocket Upgrade）列入 `gateway_app_auth_exempt.go` 的 `gatewayAppAuthExemptExactGET`，与 `/voice/asr/ws` 策略一致：Upgrade 不要求 HTTP 层 Bearer；若客户端仍携带 Bearer，`HookBeforeServe` MAY 注入 `X-Internal-Wx-Id`，但 clinic 身份校验 MUST 由 voice-service 首帧 `auth` 完成。

#### Scenario: 无 Bearer 的 Upgrade 请求

- **WHEN** App 客户端对 gateway-app `/voice/clinic/ws` 发起 WebSocket Upgrade 且未携带 `Authorization`
- **THEN** gateway-app SHALL 允许请求进入 WS 透传链（由 voice-service 首帧 `auth` 校验 wxId；**非** deviceNo 反查）

#### Scenario: 可选 Bearer 仍注入内部头

- **WHEN** App 对 `/voice/clinic/ws` Upgrade 且携带有效 App access Bearer
- **THEN** gateway-app MAY 经 `InjectAccessHeadersFromBearer` 注入 `X-Internal-Wx-Id`
- **AND** voice-service clinic handler MUST NOT 仅以该头作为鉴权依据（首帧 JWT 为准）
