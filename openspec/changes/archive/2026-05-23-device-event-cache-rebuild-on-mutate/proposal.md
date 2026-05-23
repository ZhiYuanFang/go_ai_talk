## Why

`device-service` 在事件 **新增 / 更新 / 删除** 及若干写库路径后，用 `ListEvents()` + `setEventOptions()` 试图刷新 Redis 事件字典缓存。但 `ListEvents()` **优先读 Redis**，常返回变更前的旧列表，再写回缓存会**固化旧数据**，导致管理端以外的读路径（语音理解、内部 `event/options` 等）长期看到过期 `logo`/`color`。需在写库成功后改为**从 DB 重建**缓存快照。

## What Changes

- 在 `internal/services/device/admin.go`（及同类写事件路径）将「刷新事件缓存」统一为调用已有 `RebuildEventCache(ctx)`（直查 `event` 表后 `setEventOptions`），**禁止**在写库后用 `ListEvents()` 作为刷新数据源。
- 抽取内部辅助函数（如 `refreshEventOptionsCacheAfterMutate`）避免多处重复，并在失败时打 Warning 日志。
- 覆盖范围：`AddEvent`、`UpdateEvent`、`DeleteEvent`、`InsertOrGetEventByNeedle`、`ApplyDeepSeekEventExtractPersistence` 等写事件后当前错误刷新的调用点。
- **不修改** `ListEvents` 对外读路径语义（仍先读缓存）；**不修改** history-service 独立 `historyCache`（本变更以 device Redis 为准；history 仍通过委派 device 内部接口 eventual 更新，可选后续变更）。
- 与 `admin-event-inline-color-confirm`（管理页色调确定按钮）正交，可独立或并行 apply。

## Capabilities

### New Capabilities

- `device-event-cache-rebuild-on-mutate`：device 域事件表变更后 Redis 事件选项缓存必须从数据库重建。

### Modified Capabilities

（无。）

## Impact

- **代码**：`internal/services/device/admin.go`（主）、必要时核对 `cache_rebuild.go` 无需改签名。
- **部署**：需重建 **device-service**；Redis 中旧 key 会被新快照覆盖，无需手工删 key。
- **测试**：按仓库约定不新增 `*_test.go`；手工验证改色/改 logo 后 `event/options` 与语音侧字典一致。
