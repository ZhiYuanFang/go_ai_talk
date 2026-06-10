## Context

当前 Web 管理页在 **主网关**（`gateway-service`，`:9701`）与 **App 网关**（`gateway-app-server`，`:9702`）双入口重复注册静态路由。新增模块（UCG 管理、功能使用统计）使用相对路径链接，从 9701 打开时 API 落在错误网关（主网关无 UCG 反代、无 usage 本机 Handler），导致 404。

App 网关全局 **Bearer** 中间件（`HookBeforeServe`）对非白名单路径要求用户 JWT，而管理 API 使用 **`X-Admin-Password`** 且未列入豁免——9702 上管理 API 实际不可用。版本管理使用第三套鉴权（独立口令 + `gw_ver_admin` Cookie）。口令来源分散：device 硬编码、UCG yaml 空、版本 env `GATEWAY_APP_VERSION_ADMIN_PASSWORD`。

本变更采用 **方案 B + Admin JWT 一次到位 + B1**：管理面唯一入口为 App 网关；Hub 账号密码登录签发 Admin JWT；版本管理并入同一 JWT，废弃独立 Cookie 会话。

## Goals / Non-Goals

**Goals:**

- App 网关为 Web 管理**唯一**入口；主网关 admin 路径 302 至 `GATEWAY_APP_PUBLIC_BASE_URL`。
- `POST /device/admin/api/login` 签发 Admin JWT（`aud=gateway-admin`），与用户 access JWT 严格隔离。
- 管理 API（`/device/admin/api/*`、`/ucg/admin/api/*`、`/device/app/api/version/admin/*`）须 Admin JWT；网关校验后**服务端**注入下游 `X-Admin-Password`，剥离客户端伪造。
- 前端 `admin-common.js` + `admin-modules.js` 统一 Bearer；一次登录覆盖设备/UCG/usage/版本（B1）。
- env：`GATEWAY_APP_ADMIN_USERNAME`、`GATEWAY_APP_ADMIN_PASSWORD`、`DEVICE_ADMIN_PASSWORD`、`UCG_ADMIN_PASSWORD`。
- Go 侧 `RegisterAdminStaticPages` 单点注册（仅 gateway-app）。

**Non-Goals:**

- 多账号 RBAC、审计用户表、OAuth。
- 修改 device-service / ucg-service 管理 API 契约（仍验 Header 口令，由网关注入）。
- 主网关全量对齐 usage 采集或 admin 静态托管。
- 新增 `*_test.go`（仓库约定）。
- HttpOnly Cookie 存 Admin JWT（首期 sessionStorage + Bearer；后续可迭代）。

## Decisions

### 1. Admin JWT 与用户 JWT 分离签发

- **决定**：新增 `SignAdminAccess` / `ParseAdminClaims`，claim 含 `aud: "gateway-admin"`、`sub: "gateway-admin"`（或固定字面量），**禁止**使用 wxId 作为 admin `sub`。
- **理由**：复用 `SignAccess(wxID, deviceNo)` 会导致 admin token 误访问 `/device/app/api/*` 或用户 token 误访问管理 API。
- **备选**：扩展现有 claims 加 `role` 字段 — 可行但易与历史 token 混淆；独立 `aud` 校验更清晰。

### 2. HookBeforeServe 三路分流

- **决定**：
  1. 静态 admin 页、`POST /device/admin/api/login` → 匿名白名单；
  2. 管理 API 前缀 → 解析 Admin JWT，失败 401；成功则注入口令 / 内部标记，**不**注入 wxId；
  3. 其余 → 现有用户 JWT 逻辑；若路径为管理 API 且仅带用户 JWT → 403。
- **理由**：不扩大 `/device/admin/api/` Bearer 白名单（避免双轨鉴权）；与现有 `InjectAccessHeadersFromBearer` 并列。
- **备选**：白名单 + 继续 `X-Admin-Password` — 改动小但不统一 token，已否决。

### 3. 下游口令由网关注入

- **决定**：Admin JWT 校验通过后，`/device/admin/api/*` 与 `/device/app/api/version/admin/*` 注入 `DEVICE_ADMIN_PASSWORD`；`/ucg/admin/api/*` 注入 `UCG_ADMIN_PASSWORD`。请求进入 Hook 时 **Del** 客户端 `X-Admin-Password`（加入 `StripSpoofed` 或等价逻辑）。
- **理由**：浏览器不再持有 device/ucg 口令；下游零改。
- **备选**：前端继续传口令 — 安全与 UX 差。

### 4. B1：版本管理并入 Admin JWT

- **决定**：删除 `POST /device/app/api/version/admin/login`、Redis `gw:app:veradmin:sess:*`、`gw_ver_admin` Cookie；`requireVersionAdminSession` 改为 `requireAdminJWT`（或读 `X-Internal-Admin-Verified: 1` 网关注入头）。废弃 `GATEWAY_APP_VERSION_ADMIN_PASSWORD`。
- **理由**：与「一次登录全站」目标一致。
- **备选**：保留版本独立口令 — 运维仍两套登录，已否决。

### 5. 主网关 302 迁移

- **决定**：`register.go` 移除 admin/history 静态 `ServeFile`；绑定 catch-all handler：`/device/admin` 及子路径、`/device/history/*`（仅 HTML 壳，非 `/device/history/api/*`）→ 302 到 `{GATEWAY_APP_PUBLIC_BASE_URL}{path}`；env 未配置时 fallback 同 host `:9702`。
- **理由**：bookmark 可平滑迁移；API 反代仍在 9701 不受影响。
- **备选**：410 Gone — 运维体验差。

### 6. admin-modules.js 为唯一模块登记

- **决定**：`resource/public/admin-modules.js` 声明 `id`、`title`、`pagePath`、`apiPrefixes`；`admin.html` 导航由 registry 渲染；**无 `gateway` 字段**（恒为 app 网关）。
- **理由**：方案 B 下不需要端口配对；新增模块强制登记。
- **备选**：Go 端 `/device/admin/api/modules` — 可二期；首期 JS registry 足够。

### 7. 配置与 env 默认值

- **决定**：`GATEWAY_APP_ADMIN_USERNAME` 默认 `admin`；`GATEWAY_APP_ADMIN_PASSWORD` 必填（未配置则 login 503）；`DEVICE_ADMIN_PASSWORD` / `UCG_ADMIN_PASSWORD` 默认可与 admin 密码同值（compose 文档说明）；device-service 硬编码口令迁移为读取 `DEVICE_ADMIN_PASSWORD` env（同变更或紧接 follow-up，design 纳入本变更 tasks）。
- **理由**：与 `GATEWAY_APP_VERSION_ADMIN_PASSWORD` 等既有 env 模式一致。

## Risks / Trade-offs

- **[Risk] Admin JWT 在 sessionStorage 被 XSS 窃取** → 短 TTL（默认 8h）、运维入口 IP 限制（Nginx/runbook）、不在公网暴露 admin 路径。
- **[Risk] Admin JWT 与用户 JWT 混淆** → 硬编码 `aud` 双向拒绝；单测/验收场景覆盖交叉调用。
- **[Risk] 9701 bookmark 失效** → 302 + runbook 公告；保留至少一个版本周期。
- **[Risk] 网关持有全部下游口令** → 进程 env 注入，不入 git；与现有部署模式一致。
- **[Trade-off] 主网关规格「管理页常用入口 :9701」失效** → OpenSpec delta 与 runbook 更新。

## Migration Plan

1. **部署顺序**：先部署 gateway-app（JWT + 9702 admin 完整可用）→ 再部署 gateway（302 + 移除静态页）。
2. **env 准备**：在 `.env.prod` / `.env.test` 增加 `GATEWAY_APP_ADMIN_*`、`DEVICE_ADMIN_PASSWORD`、`UCG_ADMIN_PASSWORD`；可暂时设与现 `a521521521` 相同；移除或留空 `GATEWAY_APP_VERSION_ADMIN_PASSWORD`。
3. **验收**：9702 Hub 登录 → 设备/UCG/usage/版本全流程；9701 `/device/admin` 302；Admin JWT 调用户 API 403。
4. **回滚**：回滚 gateway-app 与 gateway 镜像；恢复双入口静态页（旧 HTML 仍可用口令 Header，需旧 gateway-app 版本）。

## Open Questions

- （已决）版本管理 B1 并入 — 本变更 scope 内实施。
- device-service `fixedDeviceAdminPassword` 是否本变更一并 env 化：**建议纳入 tasks**，避免网关注入口令与 device 校验不一致。
