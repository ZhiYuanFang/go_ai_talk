## Why

运维需要在 Hub 维护「小智音箱」能力宣传与绑定教学两条固定视频，供 Flutter 小智页点播；同时 UCG 同步请求内的服务端 ffmpeg 转码会拖垮共用 ucg 进程，且迟早要拆到独立负载，首版应直接删除转码兜底，统一为「合规直传、不合规拒绝」。

## What Changes

- **BREAKING（UCG 视频上传）**：`POST /ucg/app/api/media/upload`（`mediaKind=2`）删除「v1 失败但可解码 → `NormalizeVideo` 转码再 PUT」分支；v1 验真失败 MUST 直接 4xx，MUST NOT 调用服务端转码。
- **BREAKING（internal）**：`POST /ucg/internal/api/media/upload-video` 改为 **v1 验真 + 直传**（不再 `NormalizeVideo`）；响应 `transformVersion` 与 contentHash 对齐直传语义；sim T4 随之改为 register 匹配版本（通常 `v1`）。
- 删除或停用进程内转码能力（`NormalizeVideo` / 转码 B 分支）；**保留** ffprobe 验真与既有体积上限（与 Flutter 同限，不为 Hub 抬额）。
- cash：双固定槽（`promo` / `tutorial`）元数据表；Hub「小智 MCP 开通」页新增区 D（改标题、上传/覆盖、清空）；文件经 `clients/ucg` 服务端 PutObject（OSS），库内仅存 objectKey，HTTP 边界映 CDN。
- App：`GET /cash/app/api/xiaozhi-mcp/guide-videos`（Bearer）返回已配置槽的 CDN；**计入 usage 统计**（不进 `maintenance_skip`）。
- Flutter（兄弟仓，本迭代必做）：小智页拉取清单；双槽皆空则完全不露入口；按标题点播。

## Capabilities

### New Capabilities

- `xiaozhi-mcp-guide-videos`：双槽引导视频的 cash 存储、Admin/App 契约、Hub 区 D、Flutter 消费与空列表隐藏入口。

### Modified Capabilities

- `ucg-video-transcode`：移除「服务端 NormalizeVideo / internal 必转码」要求；internal upload-video 改为验真直传；容器仍须含 ffprobe（验真），不再要求为转码保留 ffmpeg 能力作为上传前置（实现可保留二进制供 probe，但转码路径删除）。
- `ucg-video-validate`：Web 视频代理上传改为仅 v1 直传或拒绝；移除「可解码则转码」与 Probe 进 B 分支的要求。
- `sim-user-service`：T4 视频上传不再依赖「internal 转码出 v2」；改为合规直传后按响应版本 register。

## Impact

- **ucg-service**：`UploadMediaObject` 视频分支、`InternalMediaUploadVideo`、`NormalizeVideo` / `ProbeVideoDecodable` 调用链；runbook 视频三分支说明。
- **sim-user-service**：T4 调 internal upload 后的 `transformVersion`。
- **cash-service**：schema（引导视频表）、Admin/App API、经 `clients/ucg` 上传。
- **gateway-app**：cash 反代已覆盖；静态页改 `cash-xiaozhi-mcp-admin.html`；新 App GET **计入 usage**。
- **Flutter（flutter_ai_talk）**：小智音箱页入口与播放器；本仓交付契约，兄弟仓同迭代实现。
- **客户端体感**：Web/非合规片源上传将从「自动转码成功」变为 4xx，需客户端或运维先压成 v1 合规 mp4。
