## Context

- App 流量经 **gateway-app-server** 进入，Bearer 中间件在 `HookBeforeServe` 注入 `X-Internal-Wx-Id` / `X-Internal-Device-No`。
- 现有 `installDeviceAPIAccessTouchMiddleware` 在请求前异步更新 device-service `user.last_api_path`，仅保留最后一笔，且不区分 HTTP 状态码。
- 管理页模式：`qa-records.html` 独立页 + `X-Admin-Password` + `/device/admin/api/*` 调用；`/device/admin/api/*` 默认反代至 device-service。
- `api/v1/*.go` 的 `g.Meta` 已含 `path`、`method`、`summary`（中文），可作为 API 注册表与 OpenAPI 同源。
- 探索阶段已确认：默认 7 天、用户下钻主展示 wxId、只统计 2xx、独立 `api-usage-stats.html`、鉴权复用设备管理口令；入口放在 **设备记录**卡片 `card-actions`。

## Goals / Non-Goals

**Goals:**

- 在 gateway-app 边缘采集 App 相关 HTTP API 的 2xx 成功调用，按归一化 `METHOD /path` 聚合。
- 提供管理端三视图：API 频率总览、按 API 看 wxId 分布、按 wxId 看 API 分布。
- 列表附带中文 `summary`；默认时间窗口 7 天。
- 独立管理页 + `admin.html` 设备记录区入口。

**Non-Goals:**

- WebSocket（语音对话、ASR、UCG 聊天 WS）连接/消息统计。
- 主网关 `gateway-service` 全量对齐（可后续扩展；本变更聚焦 app 网关）。
- 非 2xx、业务 body `code!=0` 的细粒度过滤。
- wxId=0 纯设备会话的用户下钻（仍计入 API 全局计数）。
- 操作审计、导出 CSV、Prometheus/Grafana 替代方案。

## Decisions

### 1. 采集位置与时机

- **决定**：在 gateway-app 新增 `installGatewayAppAPIUsageStatsMiddleware`，在 `r.Middleware.Next()` **之后**读取 `r.Response.Status`，仅 `200 <= status < 300` 时异步写 Redis。
- **理由**：满足「只统计 2xx」；与现有 touch 中间件解耦，互不影响。
- **备选**：请求前计数 — 无法过滤状态码，不符合需求。

### 2. 用户维度主键

- **决定**：用户下钻以 **`wxId > 0`** 为条件；Redis 用户键使用 `wx:{wxId}`；展示列主显示 wxId，可选附带 `deviceNo`（当请求头存在时）。
- **理由**：UCG/账号类功能以 wx 为主；探索阶段已确认。
- **备选**：deviceNo 主键 — 与 UCG 账号统计口径不一致。

### 3. 路径归一化

- **决定**：启动时扫描 `api/v1` 构建路由表，将动态段（如 `/posts/123`）归一化为 `/posts/{id}`，与 `g.Meta path` 一致；未命中路由的记原始 path，`summary` 显示「未登记」。
- **理由**：避免 UCG 等动态路径把统计打碎；`summary` 可自动挂载。
- **备选**：仅原始 path — 频率榜单不可读。

### 4. Redis 存储模型

- **决定**：按日分桶 + 最近时间戳：
  - `usage:day:{YYYYMMDD}:api:{apiKey}` — 全局日计数 INCR
  - `usage:day:{YYYYMMDD}:wx:{wxId}:api:{apiKey}` — 用户×API 日计数 INCR（仅 wxId>0）
  - `usage:last:api:{apiKey}` — 全局最近成功调用 Unix 秒
  - `usage:last:wx:{wxId}:api:{apiKey}` — 用户×API 最近成功调用 Unix 秒
- **TTL**：日桶 key 90 天；查询默认聚合最近 7 天，`days` 参数支持 7/30/0（全部在 TTL 内）。
- **理由**：支持滚动窗口且写入 O(1)；与探索方案一致。
- **备选**：MySQL 明细表 — 写入压力大、跨服务边界复杂。

### 5. 跳过路径

- **决定**：不统计 WebSocket upgrade、`/device/internal/*`、`/device/admin/api/*`（管理操作）、静态页/HTML、`/swagger`、`/api.json`。
- **理由**：聚焦 App 功能热度，避免管理行为与静态资源污染。

### 6. 管理读 API 与反代排除

- **决定**：读 API 由 gateway-app 本机 Handler 提供，路径前缀 `/device/admin/api/usage/`；在 `installDeviceProxyMiddleware` 绑定反代时**排除**该前缀（先于 `/*` 注册本机路由或中间件内判断跳过反代）。
- **鉴权**：复用 `device.DeviceAdmin().VerifyPassword`（与 `DeviceAdminCtrl` 一致），Header `X-Admin-Password`。
- **理由**：统计数据在 gateway Redis，不宜经 device-service；URL 风格与 `qa-records` 的 `/device/admin/api/*` 一致。
- **备选**：数据推到 device-service — 跨域统计语义错位。

### 7. 管理页与入口

- **决定**：新增 `api-usage-stats.html`，路由 `/device/admin/api-usage-stats`；双 Tab「按 API」「按用户」；登录流程对齐 `qa-records.html`。
- **入口**：`admin.html` **设备记录**卡片 `card-actions` 增加链接「功能使用统计」→ `/device/admin/api-usage-stats`（登录后可见，与 qa/feedback 的 `hidden` 模式一致）。
- **理由**：用户明确要求与 qa-records 同级独立页，入口在设备记录头部。

### 8. 与现有 touch 共存

- **决定**：保留 `TouchLastAPIAccess` 中间件不变。
- **理由**：运营仍需要每台设备「最近接口」快照；统计模块为开发者工具，职责不同。

## Risks / Trade-offs

- **[Risk] 高 QPS 下 Redis 写入放大** → 异步 goroutine 批量/管道写入；失败静默，不阻塞响应（与 touch 一致）。
- **[Risk] 未登记路径 summary 为空** → UI 显示「未登记」+ 原始 path；后续可补注册表。
- **[Risk] wxId=0 用户看不到自己** → 页脚说明口径；API 全局计数仍准确。
- **[Risk] HTTP 200 但业务失败仍计入** → 接受；读 body 成本高，非本阶段目标。
- **[Trade-off] 仅 app 网关** → 主网关流量不计入；文档注明统计范围。

## Migration Plan

1. 部署 gateway-app-server（中间件 + Redis 键 + 读 API + 静态页）。
2. 更新 `admin.html` 入口；无需 DB 迁移。
3. 回滚：移除中间件与路由即可；Redis 键可自然过期。
4. Redis 为硬依赖：写入失败记录 warning 日志，不影响业务响应。

## Open Questions

（无。探索阶段决策已闭合。）
