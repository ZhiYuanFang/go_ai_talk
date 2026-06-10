## Why

开发者需要了解 App 网关各接口的真实使用热度，以便判断哪些功能值得投入维护与优化。现有 `last_api_path` / `last_api_at` 仅记录每台设备**最后一次** HTTP 调用，无法按 API 聚合频率，也无法按 wxId 下钻查看调用分布。

## What Changes

- 在 **gateway-app-server** 边缘增加响应后统计中间件：仅对 HTTP 2xx 成功响应计数；路径归一化后与 `api/v1` 的 `summary` 中文说明关联。
- 使用 **Redis** 按日分桶存储全局 API 计数与用户（wxId）维度计数；默认查询窗口为近 **7 天**（可扩展 30 天 / 全部）。
- 新增管理读 API（gateway 本机处理，不经 device-service 反代）：
  - `GET /device/admin/api/usage/list` — API 频率列表
  - `GET /device/admin/api/usage/detail` — 某 API 的 wxId 调用分布
  - `GET /device/admin/api/usage/user` — 某 wxId 的 API 调用分布
- 鉴权复用 Header **`X-Admin-Password`**（与设备管理口令一致）。
- 新增静态页 **`resource/public/api-usage-stats.html`**，路由 `/device/admin/api-usage-stats`（对齐 `qa-records.html` 模式）。
- 在 **`admin.html` 设备记录**卡片头部 `card-actions` 增加「功能使用统计」入口链接。
- 保留现有 `TouchLastAPIAccess`（设备最近接口快照），与本统计模块职责分离。

## Capabilities

### New Capabilities

- `gateway-app-api-usage-stats`：App 网关 API 使用统计（采集、Redis 模型、管理读 API、独立管理页、设备管理入口）。

### Modified Capabilities

（无。不修改 v2.0.2 既有 `last_api_path` touch 行为。）

## Impact

- **进程**：`gateway-app-server`（中间件、Redis 读写、管理 API、静态页路由）；主网关 `gateway-service` 可选同步安装采集中间件（本变更以 app 网关为主）。
- **静态资源**：`resource/public/api-usage-stats.html`、`resource/public/admin.html`；`gateway_app_register.go` 与 `register.go` 注册静态路由。
- **反代**：`device_route_proxy.go` 须排除 `/device/admin/api/usage/*`，避免读 API 被透传至 device-service。
- **依赖**：Redis（gateway-app 已有连接）；无新数据库表。
- **边界**：不统计 WebSocket、internal/admin 路径、非 2xx；wxId=0 的请求仅计入 API 全局统计，不出现在用户下钻。
