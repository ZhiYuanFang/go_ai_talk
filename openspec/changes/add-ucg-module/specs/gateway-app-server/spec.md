## MODIFIED Requirements

### Requirement: App 网关进程独立运行

系统 SHALL 提供名为 gateway-app-server 的独立 HTTP 服务进程，具备与现有 gateway 相当的静态资源与领域反向代理能力，并额外承载 App 鉴权、令牌、版本检查、历史 WebSocket，以及 **UCG HTTP 反向代理**（`/ucg/app/api/*` → ucg-service）与 **UCG 聊天 WebSocket 升级代理**（`/ucg/app/ws/chat` → ucg-service `/ws/chat`）。App 对外 UCG 流量 MUST 仅经本进程暴露，与现有 App API 同域。

#### Scenario: 进程启动与配置隔离

- **WHEN** 使用 gateway-app-server 专用配置文件启动进程
- **THEN** 服务 SHALL 仅加载该进程所需的数据库分组（含 ai_voice_app）与下游 URL 配置（含 `UCG_SERVICE_BASE_URL`、`UCG_WS_PROXY_URL`），且 SHALL NOT 将 voiceChat 等业务配置错误合并到错误进程的权威配置源中（遵循仓库既有配置边界约定）

#### Scenario: UCG HTTP 代理可用

- **WHEN** 配置 `UCG_SERVICE_BASE_URL` 且 ucg-service 健康
- **THEN** 对 `/ucg/app/api/*` 的请求 SHALL 经 Bearer 鉴权与头注入后转发至 ucg-service

#### Scenario: UCG 聊天 WS 升级代理可用

- **WHEN** 配置 `UCG_WS_ROUTE_MODE=proxy` 且 `UCG_WS_PROXY_URL` 指向可达的 ucg-service `/ws/chat`
- **AND** 客户端对 `/ucg/app/ws/chat` 发起 WebSocket Upgrade
- **THEN** gateway-app SHALL 将握手与后续双向帧透传至 ucg-service，行为与 `ws_route_proxy.go` voice WS 透传一致

### Requirement: 鉴权白名单

系统 SHALL 对 `POST /device/app/api/user/login`（经 device-service 暴露；与网关聚合 `POST /device/app/api/login` 区分）、gateway-app 的登录与刷新接口、版本检查（若产品要求公开）、WebSocket 握手路径（含 **`/ucg/app/ws/chat`**），以及 **产品配置的 UCG 匿名只读路径（如 `GET /ucg/app/api/feed/recommend`）** 等无需 Bearer 的路径配置中间件白名单，使其不触发 Bearer 解析失败。

#### Scenario: 无令牌访问登录

- **WHEN** 客户端无 Authorization 头调用白名单内的登录接口
- **THEN** 请求 SHALL 进入对应处理器且 SHALL NOT 被 Bearer 中间件以「未授权」拦截

#### Scenario: 匿名访问推荐 Feed

- **WHEN** 客户端无 Bearer 调用白名单内的 UCG 推荐 Feed
- **THEN** 请求 SHALL 被代理至 ucg-service 且 SHALL NOT 被 Bearer 中间件拦截

#### Scenario: UCG 聊天 WS Upgrade 无 HTTP Bearer

- **WHEN** 客户端对 `/ucg/app/ws/chat` 发起 WebSocket Upgrade 且未携带 `Authorization`
- **THEN** gateway-app MUST 允许进入 WS 代理链，不得仅因缺少 Bearer 拒绝 Upgrade；JWT 认证 SHALL 在连接后首帧由 ucg-service 处理
