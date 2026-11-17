## 1. ucg 内部 APK 上传 / 删除

- [x] 1.1 在 ucg 增加 APK 上传能力：前缀 `apk/`、流式 PutObject、大小上限对齐约 200MB、`.apk` 校验、返回 `objectKey`+`cdnUrl`；独立于 logo/媒体上限与 ownership
- [x] 1.2 注册内部路由（优先 `POST /ucg/internal/api/media/upload-apk`），鉴权与现有 internal media 相同（`DEVICE_GATEWAY_INTERNAL_SECRET`）
- [x] 1.3 提供内部删除 `apk/` objectKey 的能力（无 ownership 门禁），供 gateway 删版本时调用
- [x] 1.4 中文注释覆盖上传/删除业务边界与大小/超时约束

## 2. gateway 客户端与下载 URL 映射

- [x] 2.1 在 `internal/clients/ucg` 增加 UploadApk / DeleteApkObject（或等价）HTTP 客户端：长超时、multipart/流式、解析 `objectKey`/`cdnUrl`
- [x] 2.2 实现版本 `download_url` 双读映射 helper（`apk/`→CDN；path-only→保持/拼 publicBaseUrl；已是 https→原样）；禁止对 OSS/CDN 值误用剥 host 的 `Normalize`
- [x] 2.3 确认 gateway-app 进程具备 `UCG_SERVICE_*` 与内部密钥环境注入（compose / `.env.example` / runbook 按需补齐说明）

## 3. 上传 / 列表 / 删除改造

- [x] 3.1 改造 `gatewayAppVersionAdminUpload`：校验后调 ucg 上传；写库 `download_url`=objectKey；响应 `downloadUrl`=CDN；失败不写库
- [x] 3.2 改造管理列表/单条：对外 `downloadUrl` 经映射返回（OSS→CDN，旧 path 保持可点开）
- [x] 3.3 改造删除：path 行尽力删本地；`apk/` 行尽力调 ucg 删 OSS；失败仅 WARN；不误伤 `er_code.png`
- [x] 3.4 上传成功后继续失效最新版本缓存

## 4. version/check 与官网

- [x] 4.1 `buildVersionRes` / `version/check`：OSS 行返回 CDN 绝对 URL；历史 path 行仍返回 path-only
- [x] 4.2 `buildGatewayAppSiteAndroidDownload` / 二维码：OSS 行用 CDN；path 行仍 `publicBaseUrl` 拼接
- [x] 4.3 更新 README 中 path-only BREAKING 说明为「新发版为 CDN 绝对 URL，客户端直用」

## 5. 保留本地下载与验收

- [x] 5.1 确认 `GET /device/app/apk/*`、auth exempt、Compose `/apk/ai_talk` 挂载仍在（旧包 + 二维码）
- [x] 5.2 手工验收：未发新版官网可下旧包；上传新 APK 后 check/site 为 CDN 且可下载；删 OSS 行尽力删对象
- [x] 5.3 列出重部署：`ucg-service`、`gateway-app-server`；不必重部署其它领域服务
