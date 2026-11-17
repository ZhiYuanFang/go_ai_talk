## Context

版本 APK 现由 `gatewayAppVersionAdminUpload` 流式写入 `apkStorageDir`（默认 `/apk/ai_talk/`），`download_url` 存 path-only `/device/app/apk/<file>.apk`；用户经 gateway-app 源站下载偏慢。UCG 已有 OSS/CDN（bucket、内网 endpoint、`cdnBaseUrl`≈`https://resorce.cuplay.top`）及 `clients/ucg` → `/ucg/internal/api/media/upload` 模式，但现网 logo 上限 2MB、媒体 25MB，且常整文件入内存，不能直接复用。基线：`openspec/specs/v3.1.9` 中 path-only / version-admin / version-check / official-site；对齐 `event-logo-oss-cdn` 边界。

## Goals / Non-Goals

**Goals:**

- 新上传 APK 经方案 A 入 OSS（`apk/`），API 返回 CDN 绝对 URL；App 直接使用。
- 双读：旧 path 行 + 本地 GET 不变；未发新版前官网可用。
- 服务边界：gateway **禁止** import `services/ucg`；经 `clients/ucg` + 内部密钥。
- 删除：OSS key 尽力删对象；path 尽力删本地；`er_code.png` 保留本地。

**Non-Goals:**

- 批量迁移历史磁盘 APK；浏览器预签名直传（方案 B）；gateway 自持 OSS SDK（方案 C）。
- 改邀请群二维码存储；拆 Compose `/apk/ai_talk` 挂载（本阶段保留）。
- 新 App 对外路由；改 usage 统计策略；新增 Redis 读缓存。

## Decisions

1. **上传链路（方案 A）**  
   Admin multipart → gateway-app（校验 `.apk`、大小≤`apkMaxBytes`≈200MB）→ `clients/ucg` HTTP → ucg **新**内部接口（建议 `POST /ucg/internal/api/media/upload-apk` 或 `kind=apk` 且独立大小/流式路径）→ 内网 `PutObject` → 返回 `objectKey` + `cdnUrl` → gateway 写库。  
   **不选 B**：管理页改动大；上传已内网不慢。  
   **不选 C**：避免 OSS SDK/配置扩散到 gateway。

2. **库字段语义**  
   新行 `download_url` = **objectKey**（`apk/...`，无 scheme/域名），对齐 logo。  
   **不**把 CDN URL 原样入库再走现有 `assetpath.Normalize`（会剥 host，官网拼错）。

3. **HTTP 边界映射**  
   集中 helper（建议 `internal/shared` 或 `gatewayapp` 内，类比 `featurelogo.CdnURL`）：  
   - `apk/` 或已是 `http(s)` → CDN/原样绝对 URL；  
   - `/device/app/apk/` → check：**旧契约**可仍返回 path，或 site 拼 `publicBaseUrl`；为统一「客户端可直用」，**新行** check/site/admin **一律 CDN 绝对**；**旧 path 行**：check 可继续 path（App 拼 gateway）或 site 拼绝对——与现网一致即可。  
   明确：**最新行为 OSS 行时**，`version/check` / site / 上传响应 / 管理列表该项的 `downloadUrl` MUST 为 CDN 绝对 URL。

4. **ucg 实现约束**  
   - 前缀常量 `apk/`；`Content-Type: application/vnd.android.package-archive`（或等价）。  
   - **流式** `PutObject`（避免 200MB `ReadAll`）；上限与 gateway `apkMaxBytes` 对齐（可配置）。  
   - **不**走 ownership / media_blob / Green（非社交媒资）。  
   - clients 超时显著长于 logo 的 60s（按内网 200MB 预留）。

5. **删除**  
   - path 前缀 → 现有 `tryRemoveApkForDownloadPath`。  
   - `apk/` objectKey → gateway 调 ucg 内部删除（新增轻量 delete-by-key 或复用无 ownership 的管理删除）；失败仅 WARN，删行仍成功。

6. **本地 handler / 挂载**  
   `GET /device/app/apk/*`、auth exempt、Compose bind mount **保留**（旧包 + `er_code.png`）。

7. **配置**  
   复用 `ucg.oss.*` / `UCG_OSS_*`；gateway 依赖 `UCG_SERVICE_*` + `DEVICE_GATEWAY_INTERNAL_SECRET`（与 feature logo 客户端一致）。无需新 DB；仍用 `database.app` / `APP_DB_LINK`。

8. **接口版本**  
   既有 BindHandler / `api/v1` 字段名不变，仅 `downloadUrl` **取值语义**对新行变化；不新开 v2（字段集合未增删；属契约语义 BREAKING，在 README/规格标明即可）。

## Risks / Trade-offs

- [App 仍无脑拼 gateway 基址] → 产品已确认可直用 CDN；发版说明 BREAKING。  
- [gateway→ucg 双跳传 200MB] → 内网 + 流式；拉长超时；失败回滚不写库（ucg 已 Put 则可记日志，无强事务）。  
- [旧最新行仍慢] → 接受；发一版新包即切 CDN。  
- [误用 Normalize 剥 CDN] → 版本读路径禁止对 OSS 行调用「剥 host」的 path Normalize；专用 ResolveDownloadURL。  
- [删 OSS 失败留孤儿对象] → 与现「尽力删盘」同级；运维可扫 `apk/`。

## Migration Plan

1. 先部署 **ucg-service**（upload-apk + 可选 delete），再部署 **gateway-app-server**。  
2. 验证：未上传时官网仍下旧包；上传新 APK 后 check/site 为 CDN 且可下载；删 OSS 行尽力删对象；二维码图仍本地可访问。  
3. 回滚：回退 gateway 上传逻辑即可恢复本地落盘；已上传 OSS 行仍可用 CDN；或手工改最新行为旧 path。  
4. **不做**磁盘→OSS 批迁。

## Open Questions

- （已定）方案 A；CDN 绝对 URL App 直用；只保证新上传走 OSS。  
- 内部路径用独立 `/upload-apk` 还是扩展 `kind=apk`：实现时优先 **独立路由**，避免与 2MB logo ParseMultipart 上限纠缠。  
- prod/test 是否共用同一 OSS bucket 前缀：默认同 bucket + `apk/`；若 test 需隔离可后续加环境前缀（本版可不做）。
