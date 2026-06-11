## 1. HTTP 响应边界 CDN 映射

- [x] 1.1 在 `device_history.EventOptions` 返回前调用 `eventlogo.MapEventsLogoCdn(ctx, items)`
- [x] 1.2 补充 import 与中文注释：说明 Redis 缓存可能含 objectKey，边界 MUST 映射

## 2. history Redis 写缓存 normalize

- [x] 2.1 在 `history` 写 `setEventOptions` 路径（或封装 helper）将 logo strip 为 objectKey（CDN URL → key；已是 key 则不变）
- [x] 2.2 复用 `eventlogo.NormalizeObjectKey` / `CdnBaseURL` strip 逻辑，避免 history 将 CDN URL 写回共享 Redis 键

## 3. 验证

- [x] 3.1 `go build ./...` 通过
- [ ] 3.2 本地或 test：改 logo 后 `GET /device/history/api/event/options` 返回 `https://` CDN URL（非 `event/...` objectKey）
- [x] 3.3 `openspec validate fix-event-logo-cdn-mapping --strict` 通过

## 4. 部署

- [ ] 4.1 重建并部署 **history-service**（device-service 无需因本变更单独发布，除非同批部署）
