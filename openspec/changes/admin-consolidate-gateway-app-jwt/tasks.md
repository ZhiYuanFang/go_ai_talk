## 1. Admin JWT 与登录 API

- [x] 1.1 在 `internal/services/gatewayapp/` 新增 `admin_jwt.go`：`SignAdminAccess`、`ParseAdminClaims`（`aud=gateway-admin`）、TTL 与 secret 复用 `GATEWAY_APP_JWT_SECRET`
- [x] 1.2 实现 `AdminCredentials` 读取 env：`GATEWAY_APP_ADMIN_USERNAME`（默认 admin）、`GATEWAY_APP_ADMIN_PASSWORD`
- [x] 1.3 新增 `GatewayAdminLoginCtrl` 或 handler：`POST /device/admin/api/login`，含 IP 节流与常量时间密码比较
- [x] 1.4 在 `gateway_app_register.go` 注册 login 路由；`gateway_app_auth_exempt.go` 将 login 列入白名单

## 2. Bearer 中间件分流与口令注入

- [x] 2.1 重构 `gateway_app_middleware.go` HookBeforeServe：识别管理 API 前缀（`/device/admin/api/`、`/ucg/admin/api/`、`/device/app/api/version/admin/`）
- [x] 2.2 实现 Admin JWT 校验路径：成功则设置内部上下文或 `X-Internal-Admin-Verified`；失败 401
- [x] 2.3 用户 JWT 路径：管理 API 返回 403；Admin JWT 访问 App 用户 API 返回 403
- [x] 2.4 扩展 `StripSpoofedInternalHeaders`（或等价）：删除客户端 `X-Admin-Password`；Admin JWT 通过后按路径注入 `DEVICE_ADMIN_PASSWORD` / `UCG_ADMIN_PASSWORD`
- [x] 2.5 从 `gateway_app_auth_exempt.go` 移除 `/ucg/admin/api/` 任意方法豁免；补充 admin 静态子页 GET 白名单（qa-records、feedback-records、api-usage-stats 等）

## 3. 本机 Handler 与 B1 版本管理

- [x] 3.1 `GatewayAppUsageAdminCtrl` 改为校验 Admin JWT 上下文，移除读 Header 口令
- [x] 3.2 版本管理：`requireVersionAdminSession` 替换为 Admin JWT 校验；删除 `gatewayAppVersionAdminLogin`、Cookie 会话 Redis、`gw_ver_admin`
- [x] 3.3 版本管理 upload/list/update/delete 走 Admin JWT；更新 `gateway_app_auth_exempt.go` 移除 version admin login 与 Cookie 依赖的 GET 豁免调整
- [x] 3.4 删除或废弃 `GATEWAY_APP_VERSION_ADMIN_PASSWORD` / `VersionAdminPassword` 在版本流程中的使用；更新 `config.gateway-app-server.yaml` 注释

## 4. device-service 口令 env 化

- [x] 4.1 `internal/services/device/admin.go`：`VerifyPassword` 改为读 `DEVICE_ADMIN_PASSWORD` env（保留开发默认 fallback 或空则拒绝）
- [x] 4.2 `manifest/docker/docker-compose.microservices.yml` 与 `.env.example` / `.env.test` / `.env.prod` / `.env.local` 注入 `DEVICE_ADMIN_PASSWORD`、`UCG_ADMIN_PASSWORD`、`GATEWAY_APP_ADMIN_*`

## 5. 静态路由与主网关迁移

- [x] 5.1 新增 `admin_static_pages.go`：`RegisterAdminStaticPages(s)` 单点列表；`gateway_app_register.go` 调用
- [x] 5.2 `register.go` 移除 admin/history 壳页 `ServeFile`；实现 `/device/admin*`、`/device/history/*`（非 api）302 至 `GATEWAY_APP_PUBLIC_BASE_URL`
- [x] 5.3 主网关 compose 增加 `GATEWAY_APP_PUBLIC_BASE_URL` env（用于 redirect）

## 6. 前端 admin-common 与页面改造

- [x] 6.1 新增 `resource/public/admin-common.js`（login、Bearer fetch、logout、requireAdmin）
- [x] 6.2 新增 `resource/public/admin-modules.js` 登记全部模块；`admin.html` 导航由 registry 渲染，删除 `gatewayAppBase()`
- [x] 6.3 改造 `qa-records.html`、`feedback-records.html`、`api-usage-stats.html`、`ucg-admin.html` 使用 admin-common
- [x] 6.4 改造 `gateway-app-version-admin.html`：Hub token、删除独立口令 UI
- [x] 6.5 改造 `admin.html` Hub 为 username+password 登录

## 7. ucg-service 口令 env

- [x] 7.1 `UcgAdminPassword` 支持 env `UCG_ADMIN_PASSWORD` 覆盖 yaml `ucg.admin.password`
- [x] 7.2 更新 `config.ucg-service.yaml` 注释与 compose 注入

## 8. 文档与验收

- [x] 8.1 更新 `README.MD` 管理页表：唯一入口 App 网关 :9702
- [x] 8.2 更新 `docs/runbooks/release-deploy-and-run.md`：Admin JWT env、9701 302、验收步骤
- [x] 8.3 手动验收：9702 Hub 登录 → 设备/UCG/usage/版本全流程；9701 `/device/admin` 302；Admin JWT 与用户 JWT 交叉 403
