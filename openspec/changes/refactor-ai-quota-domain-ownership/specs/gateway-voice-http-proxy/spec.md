## ADDED Requirements

### Requirement: gateway-app-server SHALL proxy voice App HTTP APIs

`gateway-app-server` MUST 将 **`/voice/app/api/*`** 注册为 HTTP 反向代理至 voice-service（环境变量 `VOICE_API_PROXY_URL`、`VOICE_API_ROUTE_MODE`，模式对齐 `device_route_proxy.go` / `ucg_route_proxy.go`）。对受保护路径 MUST 经 Bearer 鉴权并注入 `X-Internal-Wx-Id`（及有值时的 `X-Internal-Device-No`）后转发。gateway-app MUST NOT 在本地实现 voice App 业务逻辑。

#### Scenario: App 查询 voice 域额度经 gateway 反代

- **WHEN** Flutter 携带有效 Bearer 请求 `GET /voice/app/api/ai-quota`
- **THEN** gateway-app MUST 注入内部头并转发至 voice-service 同路径
- **AND** gateway-app MUST NOT 本地聚合 polish 或 clinic 数据

#### Scenario: 反代目标不可达

- **WHEN** `VOICE_API_ROUTE_MODE=proxy` 且 voice-service 不可达
- **THEN** gateway-app MUST 返回可诊断的代理错误

### Requirement: gateway-app-server SHALL proxy voice Admin HTTP APIs with password injection

`gateway-app-server` MUST 将 **`/voice/admin/api/*`** 反代至 voice-service。Admin JWT 校验通过后，`InjectAdminDownstreamPassword` MUST 对 `/voice/admin/api/` 前缀注入 `X-Admin-Password`（值来自 **`VOICE_ADMIN_PASSWORD`** env / `voice.admin.password` 配置）。

#### Scenario: voice admin API 口令注入

- **WHEN** 已登录 Admin Hub 的用户请求 PUT `/voice/admin/api/ai-quota/default`
- **THEN** gateway-app MUST 注入 `X-Admin-Password` 并转发至 voice-service

#### Scenario: 未登录 Admin 拒绝

- **WHEN** 请求 `/voice/admin/api/*` 且无有效 Admin JWT
- **THEN** gateway-app SHALL 返回未授权且 SHALL NOT 转发

### Requirement: gateway-app-server SHALL remove device ai-quota App proxy

`gateway-app-server` MUST **移除** `device_route_proxy.go` 中对 **`/device/app/api/ai-quota`** 的反代登记。该路径 MUST NOT 再可达 device-service ai-quota 读 API。

#### Scenario: 旧 App 读路径不可用

- **WHEN** 客户端请求 `GET /device/app/api/ai-quota`
- **THEN** gateway-app SHALL NOT 反代至 device ai-quota（返回 404 或由网关本机明确拒绝）

### Requirement: gateway-service SHALL sync voice HTTP proxy paths

`gateway-service` MUST 同步注册 `/voice/app/api/*` 与 `/voice/admin/api/*` 反代，行为与 gateway-app 对齐，以便管理/通用网关与 App 网关路径一致。

#### Scenario: gateway-service 反代 voice App API

- **WHEN** 客户端经 gateway-service 请求 `/voice/app/api/ai-quota` 且 proxy 配置正确
- **THEN** gateway-service MUST 转发至 voice-service
