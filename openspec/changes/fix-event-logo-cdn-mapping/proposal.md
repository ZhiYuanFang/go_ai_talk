## Why

管理后台修改任意事件 logo 后，App 调用 `GET /device/history/api/event/options` 会短暂返回 OSS objectKey（如 `event/2026/06/xxx.png`）而非 CDN 绝对 URL，导致**全部**事件 logo 无法加载；过一段时间后（cache miss 或 TTL 过期）才恢复。根因是 device-service 与 history-service 共用 Redis 键 `device:event:options:all`，device 重建缓存写入 DB 原始 objectKey，而 history 对外响应未在边界做 `MapEventsLogoCdn`，与 v2.0.3 基线「对外 API logo 必须为 CDN URL」不符。

## What Changes

- `GET /device/history/api/event/options` 返回前统一 `MapEventsLogoCdn`，无论 Redis 缓存命中与否，`logo` 均为 `https://` CDN URL
- history-service 写 Redis 事件选项缓存时仅持久化 objectKey（避免与 device 混写 CDN URL）
- 明确 Redis 事件选项缓存语义：**仅存 objectKey**；CDN 映射仅在 HTTP 响应边界执行

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `device-event-logo-color`：强化 history event options 在缓存命中场景下仍返回 CDN URL；统一 Redis 缓存 logo 字段为 objectKey

## Impact

- `internal/controller/device_history.go` — EventOptions 响应映射
- `internal/services/history/local.go`、`adapter.go` — 可选：读路径或写缓存 normalize
- `internal/services/history/cache_repo.go` — 写缓存前 strip CDN 为 objectKey（若采用方案 B）
- 需 **history-service** 镜像重建部署；device-service 行为不变（已正确在 internal/admin 边界映射）
- App 端无需改动（恢复接收 CDN URL）
