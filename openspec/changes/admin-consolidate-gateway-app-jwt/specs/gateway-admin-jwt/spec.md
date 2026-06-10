## ADDED Requirements

### Requirement: gateway-app SHALL provide admin login issuing Admin JWT

`gateway-app-server` MUST 提供 `POST /device/admin/api/login`（`Content-Type: application/json`），请求体含 `username` 与 `password`。校验 MUST 对照环境变量 `GATEWAY_APP_ADMIN_USERNAME`（默认 `admin`）与 `GATEWAY_APP_ADMIN_PASSWORD`。成功时响应 MUST 为 `{ code: 0, data: { accessToken, expiresIn } }`，其中 `accessToken` 为 HS256 JWT，`aud` MUST 为 `gateway-admin`，`iss` MUST 为 `gateway-app-server`。未配置 `GATEWAY_APP_ADMIN_PASSWORD` 时 MUST 返回 503 且 SHALL NOT 签发 token。登录接口 MUST 列入 Bearer 白名单。

#### Scenario: 正确账号密码登录

- **WHEN** 客户端提交与 env 一致的 username/password
- **THEN** 系统 SHALL 返回 `code=0` 及非空 `accessToken`

#### Scenario: 密码错误

- **WHEN** 客户端提交错误 password
- **THEN** 系统 SHALL 返回 401 且 SHALL NOT 签发 token

#### Scenario: 未配置 admin 密码

- **WHEN** `GATEWAY_APP_ADMIN_PASSWORD` 为空且客户端请求 login
- **THEN** 系统 SHALL 返回 503 语义（管理未启用）

### Requirement: Admin JWT and user access JWT SHALL be mutually isolated

Bearer 中间件 MUST 区分 Admin JWT（`aud=gateway-admin`）与用户 access JWT。管理 API 路径（见下条）MUST 要求有效 Admin JWT。App/UCG 用户 API MUST 要求用户 JWT 且 MUST 拒绝仅含 Admin JWT 的请求。用户 JWT MUST NOT 访问管理 API（返回 403）。

#### Scenario: Admin JWT 访问设备管理 API

- **WHEN** 客户端对 `GET /device/admin/api/user/list` 携带有效 Admin JWT
- **THEN** 请求 SHALL 通过 gateway-app Bearer 校验并进入下游或本机 handler

#### Scenario: Admin JWT 访问用户 profile API

- **WHEN** 客户端对受保护 App API 仅携带 Admin JWT
- **THEN** gateway-app SHALL 返回 403

#### Scenario: 用户 JWT 访问管理 API

- **WHEN** 客户端对 `GET /device/admin/api/event/list` 仅携带用户 access JWT
- **THEN** gateway-app SHALL 返回 403

### Requirement: gateway-app SHALL inject downstream admin passwords server-side

校验 Admin JWT 成功后，gateway-app MUST 在转发或本机处理前注入：`/device/admin/api/*` 与 `/device/app/api/version/admin/*` 注入 `X-Admin-Password` 等于 `DEVICE_ADMIN_PASSWORD`；`/ucg/admin/api/*` 注入 `X-Admin-Password` 等于 `UCG_ADMIN_PASSWORD`。Hook 入口 MUST 删除客户端传入的 `X-Admin-Password`，防止伪造。

#### Scenario: 浏览器不传 X-Admin-Password 仍可调用 device 管理 API

- **WHEN** 管理员仅携带 Admin JWT 请求 `GET /device/admin/api/user/list`
- **THEN** gateway-app 反代至 device-service 的请求 SHALL 含服务端注入的有效 `X-Admin-Password` 且 device-service SHALL 返回业务数据

#### Scenario: 客户端伪造 X-Admin-Password 无效

- **WHEN** 客户端携带错误 `X-Admin-Password` 与有效 Admin JWT
- **THEN** 下游 SHALL 仍使用网关注入口令且 SHALL NOT 使用客户端伪造值

### Requirement: Web admin pages SHALL be served only by gateway-app

下列静态路由 MUST 仅由 `gateway-app-server` 注册：`/device/admin`、`/device/admin/qa-records`、`/device/admin/feedback-records`、`/device/admin/api-usage-stats`、`/device/admin/ucg-admin.html`、`/device/app/version-admin.html`、`/device/history/*deviceNo`（history 壳页）。`gateway-service` MUST NOT 再 `ServeFile` 上述路径。

#### Scenario: 9702 可打开设备管理页

- **WHEN** 客户端 `GET /device/admin` 访问 App 网关
- **THEN** 系统 SHALL 返回 `admin.html`

#### Scenario: 9701 不直接提供 admin 静态页

- **WHEN** 客户端 `GET /device/admin` 访问主网关且迁移已完成
- **THEN** 系统 SHALL NOT 返回 admin 静态 HTML 作为 200 正文（SHALL 302 或 runbook  documented 等价行为）

### Requirement: gateway-service SHALL redirect legacy admin URLs to App gateway

主网关 MUST 对 `/device/admin` 及子路径（静态 admin 页）、以及 `/device/history/*` 壳页（不含 `/device/history/api/*`）返回 **302**，`Location` 为 `GATEWAY_APP_PUBLIC_BASE_URL` 与请求路径拼接；env 未配置时 MAY fallback 至同 host 的 App 网关端口（如 `:9702`）。

#### Scenario: 旧 bookmark 跳转

- **WHEN** 用户访问 `https://example.com:9701/device/admin`
- **THEN** 浏览器 SHALL 被重定向至 App 网关等价路径

### Requirement: admin front-end SHALL use shared Bearer client

仓库 MUST 提供 `resource/public/admin-common.js`（登录、`Authorization: Bearer` fetch、logout）与 `resource/public/admin-modules.js`（模块登记）。所有管理静态页 MUST 通过 admin-common 调用 API，MUST NOT 在浏览器侧发送 `X-Admin-Password`。`admin.html` MUST 从 admin-modules 渲染导航链接，MUST NOT 使用 `gatewayAppBase()` 端口配对。

#### Scenario: Hub 登录后子页共用 token

- **WHEN** 管理员在 `/device/admin` 登录成功
- **THEN** 打开 `/device/admin/qa-records` SHALL 使用同一 Admin JWT 加载数据而无需再次输入口令

### Requirement: New admin modules SHALL be registered in admin-modules.js

新增 Web 管理模块时 MUST 在 `admin-modules.js` 增加条目（至少含 `id`、`title`、`pagePath`、`apiPrefixes`），并在 gateway-app 静态路由单点注册表中增加 `pagePath`。PR MUST NOT 新增仅主网关可见的 admin 静态路由。

#### Scenario: 登记与路由一致

- **WHEN** admin-modules 声明 `pagePath=/device/admin/foo`
- **THEN** gateway-app MUST 注册该路径对应的静态文件 handler
