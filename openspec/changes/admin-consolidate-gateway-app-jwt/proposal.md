## Why

Web 管理页分散在主网关（`:9701`）与 App 网关（`:9702`）双入口，新增模块（UCG 管理、功能使用统计等）易写错跳转导致 404；App 网关 Bearer 中间件与 `X-Admin-Password` 口令鉴权并存，9702 上部分管理 API 实际不可用。运维需记忆多套口令（设备、UCG、版本管理），与 `GATEWAY_APP_VERSION_ADMIN_PASSWORD` 等 env 配置不一致。需要将管理面收敛到 App 网关，并以 Admin JWT 统一入口鉴权，一次登录覆盖全部模块（含版本管理 B1）。

## What Changes

- **BREAKING**：Web 管理页**唯一入口**为 App 网关（`:9702` / 测试 `:19702`）；主网关 `gateway-service` 移除 admin 静态路由，对 `/device/admin*`、`/device/history/*`（壳页）提供 **302** 重定向至 `GATEWAY_APP_PUBLIC_BASE_URL` 对应路径（迁移期）。
- 新增 **`POST /device/admin/api/login`**：账号 + 密码（env 配置）签发 **Admin JWT**（`aud=gateway-admin`，与用户 access JWT 隔离）。
- gateway-app **Bearer 中间件分流**：管理 API 须 Admin JWT；App/UCG 用户 API 须用户 JWT；**禁止** Admin JWT 访问用户 API、禁止用户 JWT 访问管理 API。
- 校验 Admin JWT 后，网关**服务端**向反代请求注入 `X-Admin-Password`（device/ucg 各自 env），**剥离**客户端伪造的该 Header；下游 device-service / ucg-service **契约不变**。
- 管理静态页与 `admin-common.js` / `admin-modules.js`：统一 `Authorization: Bearer`；移除各页独立口令 Header 与 `gatewayAppBase()` 端口跳转。
- **B1**：版本管理（`/device/app/api/version/admin/*`）并入同一 Admin JWT，废弃独立口令登录与 `gw_ver_admin` Cookie 会话；废弃 `GATEWAY_APP_VERSION_ADMIN_PASSWORD`（由 `GATEWAY_APP_ADMIN_PASSWORD` 统一）。
- `GatewayAppUsageAdminCtrl` 改为依赖 Admin JWT 上下文，不再读取浏览器 `X-Admin-Password`。
- env 化：`GATEWAY_APP_ADMIN_USERNAME`、`GATEWAY_APP_ADMIN_PASSWORD`、`DEVICE_ADMIN_PASSWORD`、`UCG_ADMIN_PASSWORD`；更新 `.env.example`、compose、runbook。

## Capabilities

### New Capabilities

- `gateway-admin-jwt`：App 网关统一运维登录（Admin JWT 签发与校验、中间件分流、下游口令注入、9701 重定向、admin 模块登记与共享前端约定）。

### Modified Capabilities

- `gateway-app-version-admin`：鉴权从独立口令 + Cookie 会话改为 Admin JWT；登录 API 废弃或改为 Hub 共用 login。
- `gateway-app-version-admin-crud`：写操作鉴权同上（B1）。
- `gateway-app-api-usage-stats`：管理读 API 鉴权从浏览器 `X-Admin-Password` 改为 gateway-app Admin JWT 校验。

## Impact

- **进程**：`gateway-app-server`（JWT、中间件、login handler、版本/usage handler、静态路由）；`gateway-service`（移除 admin 静态页、302 redirect）。
- **静态资源**：`resource/public/admin*.html`、`qa-records.html`、`feedback-records.html`、`api-usage-stats.html`、`ucg-admin.html`、`gateway-app-version-admin.html`；新增 `admin-common.js`、`admin-modules.js`。
- **控制器**：`gateway_app_middleware.go`、`gateway_app_auth_exempt.go`、`gateway_app_version_admin.go`、`gateway_app_usage_admin.go`、`register.go`、`gateway_app_register.go`；新增 `admin_static_pages.go`、`admin_jwt.go`（或 `gatewayapp/admin_jwt.go`）。
- **配置**：`config.gateway-app-server.yaml`、`.env.*`、`docker-compose.microservices.yml`；文档 `README.MD`、`docs/runbooks/release-deploy-and-run.md`。
- **边界**：不修改 device-service / ucg-service 管理 API 的 `X-Admin-Password` 契约；不引入多账号 RBAC；不新增测试文件（仓库约定）。
