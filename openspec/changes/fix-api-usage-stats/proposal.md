## Why

App API 使用统计页显示「已加载 0 条」，但 test Redis 中已有 `gw:usage:*` 数据。初版修复处理了 `*gvar.Var` unwrap 与反代写入，但 **test 环境复现**：写入正常（Redis 含 `gw:usage:d:{day}:g`、`w:{wxId}`），运维页与 `GET /device/admin/api/usage/list` 仍返回空 `list`。**补刀根因**：GoFrame Redis 对 `HGETALL` 的 `[]interface{}` 在 `resultToVar` 中转为 flat **`[]string`**，`redisHashToMap` 仅识别 `[]interface{}`，unwrap 后仍返回空 map。另：领域反代 `ExitAll` 曾导致反代路径未写入（初版已通过 `ModifyResponse` 修复）。

运维无法通过统计页评估真实 API 使用情况；「按用户」Tab 在读路径修复前同样为空。

## What Changes

- **补刀**修复 `usagestats` Redis 读取：`HGETALL` 优先经 `(*gvar.Var).MapStrStr()` 解析（兼容 GoFrame 转 flat `[]string`），管理端 list/detail/user 能展示已有 Redis 数据
- 修复统计写入：在 `HookBeforeServe` 包装 `ResponseWriter`，反代 `ExitAll` 后仍能在 2xx 时记录
- 新增 `GET /device/admin/api/wx/list`（device-service，经 gateway 反代），分页返回 wx 账号列表
- 更新 `api-usage-stats.html`：「按用户」Tab 展示 wx 列表并点选查询；改进空态说明（不含 WebSocket、反代修复前可能偏少）
- **维护型 API denylist**：不统计 token/refresh、version/check、site/home、version/admin/*、**GET 评论列表**（`GET /ucg/app/api/posts/{id}/comments`，负责人已确认）；POST 评论与登录/注册/绑定、各业务 API 仍统计
- **读 API 排序**：`sortBy=count|lastAt`，默认调用次数降序；前端排序控件

## Capabilities

### New Capabilities

（无独立新 capability；wx 列表作为 device-admin 读 API 扩展）

### Modified Capabilities

- `gateway-app-api-usage-stats`：补充反代路径写入与 Redis 读取语义；扩展管理页「按用户」交互（wx 列表）
- `device-admin`：新增 wx 账号分页列表 Admin API

## Impact

- `internal/services/gatewayapp/usagestats/store.go` — Redis 读取
- `internal/controller/gateway_app_usage_stats_middleware.go` — 写入机制（Hook + ResponseWriter）
- `api/v1/device_admin_http.go`、`internal/controller/device_admin.go`、`internal/services/device/` — wx list
- `internal/controller/device_route_proxy.go` — 确保 wx list 路径反代（已有 `/device/admin/api/*` 前缀）
- `resource/public/api-usage-stats.html` — 前端 wx 列表 Tab
- 需 **gateway-app** 与 **device-service** 镜像重建部署
