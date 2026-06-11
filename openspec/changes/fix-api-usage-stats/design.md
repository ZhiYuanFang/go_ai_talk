## Context

- 统计写入：`installGatewayAppAPIUsageStatsMiddleware` 在 `BindMiddleware("/*")` 中 `Next()` 之后记录；GoFrame 将更具体路径的反代中间件排在 `/*` 之前，反代 handler 直接 `ServeHTTP` + `ExitAll()` 不调用 `Next()`，导致 `usage_stats` 中间件永不执行。
- 统计读取：`g.Redis().Do("HGETALL")` 返回 `*gvar.Var`，`redisHashToMap` 未 unwrap，始终返回空 map。
- 已验证 test Redis：`gw:usage:d:20260611:g` 含 5 条网关本机 API，页面仍为 0。

## Goals / Non-Goals

**Goals**

- 管理端读 API 正确返回 Redis 中已有统计
- 经 device/ucg/history/voice 反代的 App HTTP 2xx 成功请求计入统计
- 「按用户」Tab 通过 wx 列表选账号查 usage

**Non-Goals**

- WebSocket 使用统计
- 历史数据回填
- 变更 device/history/voice/ucg 服务内业务逻辑

## Decisions

### 1. 写入：ModifyResponse + 本机 Middleware 双路径

- **反代路径**：在 `buildReverseProxy` 的 `ModifyResponse` 中，对 HTTP 2xx 调用 `usagestats.RecordHTTPRequest`（`resp.Request` 含原始路径与注入头）
- **本机 Handler**：保留 `BindMiddleware("/*")` 在 `Next()` 之后调用 `RecordGHTTPRequest`
- 两处共用 `ShouldSkip*` 与 `RecordAsync`，避免重复计数（反代不经过 middleware 的 Next 回写路径）

### 2. 读取：统一 `redisHashToMap` 支持 `*gvar.Var`

对 `*gvar.Var` 调用 `.MapStrStr()` 或 `.Val()` 再递归解析；`HGET` 单值用 `.String()`。

### 3. 维护型 API denylist

`usagestats/maintenance_skip.go` 集中维护不统计的路径（精确 METHOD+path 与前缀）。`ShouldSkipRecord` / `ShouldSkipHTTPRecord` 统一调用。登录、注册、绑定与各业务 API **不**在此列表。

### 4. 读 API 排序

`sortBy=count`（默认）或 `lastAt`；list/detail/user 三处一致。前端提供排序下拉。

### 5. 新增接口统计确认流程

见 `openspec/project.md`「App API 使用统计约定」：新增 App HTTP 路由时 MUST 询问负责人是否计入统计；不统计则改 `maintenance_skip.go`。

- `GET /device/admin/api/wx/list?page=&pageSize=&q=` — `q` 模糊匹配 id/deviceNo/unionid/account
- 鉴权：`X-Admin-Password`（与现有 device admin 一致）；gateway 反代时注入 password
- gateway-app 统计页用 Admin JWT + `adminFetch` 调该路径（gateway Bearer hook 对 admin API 注入下游 password）

## Risks / Trade-offs

- ResponseWriter 包装需正确处理 `Flush`/`Hijack`（WS 已在 skip 中，反代 WS 不走此 Writer 的 HTTP 统计）
- 移除 middleware 写入后，若有极特殊路径不经 Writer 写响应，可能漏记 — 可接受，反代与本机均走 Writer

## Migration

- 无需 Redis 迁移；部署后立即生效
- 旧日桶数据可读（修复读取后即展示）
