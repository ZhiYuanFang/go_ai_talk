## Why

当前版本 APK 落在 gateway-app 宿主机磁盘，用户经 `GET /device/app/apk/*` 从源站下载，速度慢、带宽压在网关。管理端上传已走内网，瓶颈在下载侧；需将**新上传**的 APK 写入既有 UCG OSS，对外返回 CDN 绝对 URL，旧包在发新版前仍可本地下载。

## What Changes

- **方案 A**：版本管理 `POST /device/app/api/version/admin/upload` 仍收 multipart；gateway-app 经 `internal/clients/ucg` 调 ucg-service 内部接口，由 ucg **内网 endpoint** 直传 OSS（前缀 `apk/`），不再以本地 `/apk/ai_talk/` 作为新包唯一存储。
- 库表 `download_url`：**新写入**存 OSS **objectKey**（无域名，如 `apk/...`）；**HTTP 边界**（`version/check`、site 官网/二维码、管理列表/上传响应）返回 **CDN 绝对 URL**（`https://` + 既有 `cdnBaseUrl`）。
- **BREAKING**：`GET /device/app/api/version/check` 对新发版行的 `downloadUrl` 从「应用内 path、由 App 拼 gateway 基址」改为 **CDN 绝对 URL，客户端直接使用**（已确认 App 可直用）。
- **双读兼容、不迁旧包**：历史 `/device/app/apk/*.apk` 行继续走本地 `GET /device/app/apk/*`；site 仍用 `publicBaseUrl` 拼绝对地址。未上传新版本前，官网下载行为与现网一致。
- 删除版本：objectKey 行尽力删 OSS 对象；path-only 行仍尽力删本地文件。`er_code.png` 与本地下载 handler **保留**。
- 不批量迁移磁盘历史 APK；Compose `/apk/ai_talk` 挂载短期保留（旧包 + 邀请群二维码）。

## Capabilities

### New Capabilities

- `apk-oss-cdn`：APK 经 ucg 内部上传至 OSS（`apk/` 前缀）、大小/流式约束、CDN 映射、删除与双读语义。

### Modified Capabilities

- `gateway-app-path-only-assets`：新 APK 写库不再强制 path-only；`version/check` 新数据返回 CDN 绝对 URL；本地 `GET /device/app/apk/*` 保留供旧行与二维码。
- `gateway-app-version-admin`：上传成功落 OSS + 写 objectKey；不再要求新包仅落 `/apk/ai_talk/`。
- `gateway-app-version-admin-crud`：上传写库与列表/删除语义对齐 OSS/CDN；删除支持 OSS 清理。
- `gateway-app-version-check`：最新行为 OSS 行时 `downloadUrl` 为 CDN 绝对 URL。
- `gateway-app-official-site`：Android 下载/二维码对新行使用 CDN；旧 path 行仍拼 `publicBaseUrl`。

## Impact

- **进程重部署**：`ucg-service`（新内部上传能力）、`gateway-app-server`（上传改调 ucg、读路径 CDN 映射）。
- **不必重部署**：`device-service`、`voice-service`、`history-service`、`cash-service`、主 `gateway-service`（本变更不改其契约）。
- **配置/环境**：复用既有 UCG OSS（`UCG_OSS_*` / `ucg.oss.*`、CDN `UCG_OSS_CDN_BASE_URL`）；gateway-app 需可达 `UCG_SERVICE_URL`/`UCG_SERVICE_BASE_URL` + `DEVICE_GATEWAY_INTERNAL_SECRET`；APK 大小上限仍对齐约 200MB，ucg 侧需独立于 logo 2MB/媒体 25MB 限制；HTTP 客户端超时须覆盖大文件内网上传。
- **API**：上传/列表/check/site 的 `downloadUrl` 对新行变为 CDN；路径与 Admin JWT 不变；无新 App 对外路由，usage 统计策略不因本变更调整。
- **基线对照**：复用/变更 v3.1.9 中 `gateway-app-path-only-assets`、`gateway-app-version-admin`、`gateway-app-version-admin-crud`、`gateway-app-version-check`、`gateway-app-official-site`；OSS 模式对齐 `event-logo-oss-cdn`（clients/ucg → internal upload，禁止 gateway import `services/ucg`）。
