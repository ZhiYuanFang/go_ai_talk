## Context

事件 logo 现由 `device.SaveEventLogo` 写入 `/ai_talk_images/`，DB 存 path-only；客户端（admin、Flutter、官网）拼接 `www.cuplay.top` 访问。UCG 域已有 OSS 上传与 `BuildCdnURL`（`internal/services/ucg`），bucket `pang-bao`，CDN `https://resorce.cuplay.top`。

## Goals / Non-Goals

**Goals:**

- 新上传与历史迁移后，DB 存 `event/` 前缀 objectKey。
- 对外 API 返回 CDN 绝对 URL；CDN 为 logo 唯一访问域名。
- device-service 经 HTTP 调 ucg-service 上传，不直连 OSS（服务边界）。
- 一次性迁移 prod/test 两库；下线 `/ai_talk_images` 全链路。
- 同变更更新 flutter_ai_talk 与 runbook。

**Non-Goals:**

- 不迁移 APK 等其它 `/apk/ai_talk` 静态资源。
- 不做 `/ai_talk_images/` 双读兼容期。
- 不新增 `*_test.go` 测试文件。

## Decisions

### 1. objectKey 前缀 `event/`

- 新上传：`event/{yyyy}/{mm}/{random32}.{ext}`（与 ucg `buildObjectKey` 一致，前缀可配置为 `event/`）。
- 迁移脚本：`event/{id}/logo.{ext}`（按 event id 幂等）。

### 2. ucg internal upload API

- 新增 `POST /ucg/internal/api/media/upload`（multipart，字段 `file`；可选 `objectKeyPrefix=event`）。
- 鉴权：`X-Device-Gateway-Internal-Secret`（与 device internal 一致）。
- 返回 `{ objectKey, cdnUrl }`；不写 ucg_media_upload ownership（管理端资源）。

### 3. device-service 上传路径

- `saveEventLogoFromRequest` 改为 HTTP 调 ucg internal upload，将返回的 objectKey 写入 DB。
- 删除 `SaveEventLogo` 本地落盘与 `EventLogoAbsPath`。

### 4. API 响应层拼 CDN

- 新增 `device.EventLogoCdnURL(objectKey)`（读 ucg OSS cdnBaseUrl 或 device 配置镜像）。
- `ListEvents` 缓存仍存 objectKey；HTTP 层（admin list、internal/history options）序列化前映射 `logo` → CDN URL。
- 避免缓存存 CDN URL（CDN 域名变更时仅改配置）。

### 5. 域名策略

- 迁移读源：`https://www.cuplay.top`（prod）、test 网关基址（test 库）。
- 迁移后：仅 `https://resorce.cuplay.top/event/...` 可访问 logo。

### 6. Flutter

- `resolveEventLogoUrl`：https 原样返回；移除对网关 path 的依赖（无双读）。
- 原生仍后台下载 CDN URL 到本地；Web 用 `Image.network`。

## Risks / Trade-offs

- **[Risk] 迁移窗口内顺序错误** → runbook 强制：脚本 100% 成功 → Redis DEL → 部署去静态 → Flutter。
- **[Risk] CDN 不可达** → 发版前 curl 抽查；OSS 公开读与 UGC 一致。
- **[Risk] device 调 ucg 需 UCG_SERVICE_URL** → compose 已有；local 文档说明。

## Migration Plan

1. 备份两库 `event` 表。
2. `go run hack/migrate-event-logos-to-oss`（dry-run 后正式）。
3. SQL 校验无 `/ai_talk_images/` 残留。
4. Redis `DEL device:event:options:all`（prod/test 各环境）。
5. 部署 go_ai_talk + 移除 compose volume。
6. 发布 flutter_ai_talk。
7. 冒烟：admin、event/options、App 首次安装。

**Rollback**：仅 DB 备份可恢复；OSS 对象可保留。

## Open Questions

- 无（objectKey 前缀、无双读、CDN 唯一域名已确认）。
