## MODIFIED Requirements

### Requirement: App 网关进程独立运行

系统 SHALL 提供名为 gateway-app-server 的独立 HTTP 服务进程，具备与现有 gateway 相当的静态资源与领域反向代理能力，并额外承载 App 鉴权、令牌、版本检查、历史 WebSocket，以及 **UCG HTTP 反向代理**（`/ucg/app/api/*` → ucg-service）、**UCG 聊天 WebSocket 升级代理**（`/ucg/app/ws/chat` → ucg-service `/ws/chat`），以及 **voice HTTP 反向代理**（`/voice/app/api/*`、`/voice/admin/api/*` → voice-service）。App 对外 UCG 与 voice App/Admin 流量 MUST 经本进程暴露，与现有 App API 同域。

#### Scenario: 进程启动与配置隔离

- **WHEN** 使用 gateway-app-server 专用配置文件启动进程
- **THEN** 服务 SHALL 仅加载该进程所需的数据库分组（含 ai_voice_app）与下游 URL 配置（含 `UCG_SERVICE_BASE_URL`、`UCG_WS_PROXY_URL`、**`VOICE_API_PROXY_URL`**），且 SHALL NOT 将 voiceChat 等业务配置错误合并到错误进程的权威配置源中（遵循仓库既有配置边界约定）

#### Scenario: UCG HTTP 代理可用

- **WHEN** 配置 `UCG_SERVICE_BASE_URL` 且 ucg-service 健康
- **THEN** 对 `/ucg/app/api/*` 的请求 SHALL 经 Bearer 鉴权与头注入后转发至 ucg-service

#### Scenario: voice HTTP 代理可用

- **WHEN** 配置 `VOICE_API_PROXY_URL` 且 voice-service 健康
- **THEN** 对 `/voice/app/api/*` 的请求 SHALL 经 Bearer 鉴权与头注入后转发至 voice-service
- **AND** 对 `/voice/admin/api/*` 的请求 SHALL 经 Admin JWT 校验与 `VOICE_ADMIN_PASSWORD` 注入后转发至 voice-service

#### Scenario: UCG 聊天 WS 升级代理可用

- **WHEN** 配置 `UCG_WS_ROUTE_MODE=proxy` 且 `UCG_WS_PROXY_URL` 指向可达的 ucg-service `/ws/chat`
- **AND** 客户端对 `/ucg/app/ws/chat` 发起 WebSocket Upgrade
- **THEN** gateway-app SHALL 将握手与后续双向帧透传至 ucg-service，行为与 `ws_route_proxy.go` voice WS 透传一致
