## Context

- Redis 键：`cachekit.EventOptionsKey()`，TTL 约 10 分钟（`eventOptionsTTL`）。
- `ListEvents`：缓存命中 → 直接返回；未命中 → DB Scan → `setEventOptions`。
- `RebuildEventCache`：始终 DB Scan（含 `logo`、`color` 字段）→ `normalizeEventRows` → `setEventOptions`。
- `UpdateEvent` 成功后当前代码：

```go
if rows, listErr := s.ListEvents(ctx); listErr == nil {
    _ = deviceCache.setEventOptions(ctx, rows) // rows 可能来自旧缓存
}
```

- 异步 `enqueueDeviceProjectionEvent` → worker outbox → `ApplyProjection` 路径**会**从 DB 重建，但 compose 常 `OUTBOX_RELAY_ENABLED=false`，且不能替代同步正确性。

## Goals / Non-Goals

**Goals:**

- 任意成功写入 `event` 表的路径，同步刷新 Redis 事件列表快照与 DB 一致。
- 管理端 `GET /device/admin/api/event/list` 经 `ListEvents` 在刷新后读到新数据（下一次请求命中新缓存）。

**Non-Goals:**

- 不改 history 进程内 `historyCache` 失效策略（可观测：history 缓存 TTL 内仍可能旧，直到 miss 后委派 device；若需强一致另开变更）。
- 不改 action 缓存（除非发现相同反模式；本变更聚焦 **event**）。
- 不改 outbox/worker 投影逻辑（已正确从 DB 重建）。

## Decisions

### 1. 统一用 `RebuildEventCache`

写库成功后：

```go
if err := RebuildEventCache(ctx); err != nil {
    glog.Warningf(ctx, "[device-admin] 重建事件 Redis 缓存失败 err=%v", err)
}
```

**弃用**：`ListEvents` + `setEventOptions` 作为写后刷新。

### 2. 辅助函数

```go
// refreshEventOptionsCacheAfterMutate 事件表变更后从 DB 重建 Redis 事件字典（勿用 ListEvents，避免读旧缓存）。
func refreshEventOptionsCacheAfterMutate(ctx context.Context) {
    if err := RebuildEventCache(ctx); err != nil {
        glog.Warningf(ctx, "[device-admin] 重建事件 Redis 缓存失败 err=%v", err)
    }
}
```

替换 `admin.go` 中所有写事件后的错误刷新块。

### 3. 版本键（可选）

`ApplyProjection` 会 `setVersion` on `DeviceEventVersionKey`。同步 `RebuildEventCache` 后不强制 bump version（outbox 消费者仍可按 version 去重）；若双路径并发，以最后一次 `setEventOptions` 为准即可。

### 4. 与 admin 页关系

管理页 `event/list` 走 device admin `ListEvents`；修复后行内/弹窗改 `color`/`logo` 保存成功即更新 Redis，与 UI 列表刷新一致。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 每次写事件多一次 DB 全表 Scan | 事件表规模小，可接受；优于错误缓存 |
| history 仍可能短期旧 | 文档说明；必要时后续 history 收到 device 变更事件时删 history 事件 key |

## Migration Plan

1. 部署新 device-service。
2. 可选：对现网错误缓存执行一次 `RebuildEventCache`（重启后首次写事件也会重建）或删 `EventOptions` Redis key。

## Open Questions

- 无。
