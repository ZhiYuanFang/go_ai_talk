## Context

- 事件 logo 在 DB 与 Redis 中存 **OSS objectKey**（`event/2026/06/xxx.png`）；对外 HTTP 须通过 `eventlogo.MapEventsLogoCdn` / `CdnURL` 转为 CDN 绝对 URL。
- device-service 与 history-service 共用 Redis 键 `device:event:options:all`（`cachekit.EventOptionsKey`）。
- device 在 `RebuildEventCache` 写入 objectKey；history 在 cache miss 时经 `device/internal/event/options` 回填，可能写入 CDN URL。
- `device_internal.EventOptions` 与 `device_admin` 列表已在控制器层 `MapEventsLogoCdn`；`device_history.EventOptions` **未**映射，缓存命中时直接返回 Redis 内容。
- 用户已验证：改 logo 后 history API 返回 `event/2026/06/xxx.png`，App 全部 logo 不可见。

## Goals / Non-Goals

**Goals**

- `GET /device/history/api/event/options` 在任何情况下返回 CDN 形式 `logo`
- Redis 事件选项缓存 logo 字段语义统一为 objectKey，避免 device/history 混写不同格式
- 管理后台改 logo 后 App 无需等待 TTL 即可正确展示全部事件 logo

**Non-Goals**

- 修改 App 客户端拼接 CDN 逻辑
- 变更 OSS 上传路径策略（仍为 `event/年/月/随机.ext`）
- 拆分 device/history Redis 键空间（本变更用边界映射 + 写缓存 normalize 即可）
- CDN 刷新 / 缓存 busting 策略

## Decisions

### 1. HTTP 边界统一映射（必做）

在 `device_history.EventOptions` 返回前调用 `eventlogo.MapEventsLogoCdn(ctx, items)`。

- **理由**：与 admin/internal 一致；`CdnURL` 对已含 `https://` 的值原样返回，兼容历史混写。
- **替代**：仅改 App 端拼接 objectKey — 违反基线 spec，且无法解决「全部 logo 空白」。

### 2. history 写 Redis 前 normalize 为 objectKey（推荐）

在 `historyCache.setEventOptions` 或 `ListEventOptions` 写缓存路径，将 logo 字段 strip 回 objectKey（若已是 CDN URL 则解析出 key；若已是 objectKey 则不变）。

- **理由**：device 重建缓存后 history 读到的仍是 objectKey；history 不再把 CDN URL 写回共享键，减少混写复发。
- **实现**：复用 `eventlogo.NormalizeObjectKey`；对 CDN URL 可 strip `CdnBaseURL(ctx)+"/"` 前缀。

### 3. device 侧不改写缓存格式

device `RebuildEventCache` 继续存 objectKey，符合「DB 投影」语义；internal/admin 已在读路径映射。

### 4. 不在 history ApplyProjection 订阅 device.event.changed

device 变更已通过共享 Redis 键同步；本 fix 不引入新 MQ 订阅。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| CDN base 配置缺失时 `CdnURL` 返回空串 | 与现有 admin/internal 行为一致；部署检查 `UCG_OSS_CDN_BASE_URL` |
| 历史 Redis 中已缓存 CDN URL | `MapEventsLogoCdn` 识别 http(s) 原样返回；写路径 normalize 逐步统一 |
| 双写 TTL 10min 内旧 CDN URL 仍在 Redis | 读边界映射后对外始终正确 |

## Migration Plan

1. 部署 history-service 新版本
2. 无需 Redis 手动清理；首次请求即返回 CDN URL
3. 回滚：还原 history 控制器改动即可

## Open Questions

（无 — 根因已通过接口响应验证）
