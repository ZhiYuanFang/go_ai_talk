## Why

事件 logo 当前落盘于宿主机 `/ai_talk_images/` 并由 `www.cuplay.top` 网关静态提供，客户端需自行拼接域名。该模式不利于 CDN 加速、多环境一致性与运维备份。需将 logo 统一迁移至 UCG OSS（`event/` 前缀），对外仅通过 CDN（`https://resorce.cuplay.top`）访问。

## What Changes

- **BREAKING**：`event.logo` 库存 OSS objectKey（`event/...`），不再存 `/ai_talk_images/...` path-only。
- **BREAKING**：所有返回事件字典的 API 中 `logo` 字段改为 CDN 绝对 URL，客户端不得再拼接网关 origin。
- 管理端上传 logo 经 device-service HTTP 调用 ucg-service 内部 OSS 上传。
- 一次性迁移脚本：从 `https://www.cuplay.top{logo}` 拉取旧文件上传 OSS 并回填两库（`ai_voice_device`、`ai_voice_device_test`）。
- 迁移完成后立即下线 `/ai_talk_images` 静态反代、device-service 静态读与 Docker 宿主机目录挂载。
- 同步更新 `admin.html` 与 `flutter_ai_talk` 图标获取逻辑。
- 新增 runbook：`docs/runbooks/event-logo-oss-migration.md`。

## Capabilities

### New Capabilities

- `event-logo-oss-cdn`: 事件 logo OSS 存储、CDN 响应、ucg 内部上传契约、迁移 runbook 与旧静态链路退役。

### Modified Capabilities

- `device-event-logo-color`: logo 存储与 API 响应语义由 path-only 改为 objectKey + CDN URL；移除 `/ai_talk_images` 静态读要求。
- `device-admin-event-logo-color-ui`: 管理页 logo 预览改为 CDN URL，不再依赖同源 `/ai_talk_images`。

## Impact

- **go_ai_talk**：device-service、ucg-service（internal upload）、gateway-app（移除反代）、admin.html、Docker compose、Redis 事件缓存、OpenSpec 规格。
- **flutter_ai_talk**：`event_catalog_paths`、`event_logo`、README；游客冷启动 event/options + CDN 下载。
- **运维**：维护窗口内执行迁移脚本、失效 Redis、部署与下线静态目录。
