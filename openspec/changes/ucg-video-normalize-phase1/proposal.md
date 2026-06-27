## Why

UCG 视频当前经 presign 或 Web 代理直传 OSS，无格式校验与转码，OSS 上并存 HEVC、无音轨、moov 后置等文件，导致移动端/Web 播放与 OSS 首帧截帧不稳定。原生将本地每条视频 ffmpeg 为 canonical（H.264 Main + 静音 AAC + faststart）并以 `transformVersion=v2` 登记；Web Phase 1 无 wasm 转码，需在服务端对直传做 **v1 宽验真**（与 v2 刻意分叉）；sim T4 智谱下载视频须经服务端转码为 v2 后再入 OSS。本变更交付 **Phase 1**：验真闸门 + sim 转码；Web canonical 转码留 **Phase 2**。

## What Changes

- 新增 `ucg-service` 视频 ffprobe 验真模块：`transformVersion=v1`（Web 宽规）与 `v2`（canonical 严规）规则分叉；**v1 与 v2 均要求存在 AAC 音轨**（无音轨 MUST 拒绝；Phase 1 不补轨，Phase 2 Web 转码时由服务端补静音 AAC）。
- Web `POST /ucg/app/api/media/upload`（`mediaKind=2`）：上传前 ffprobe 验 `v1`；通过则直传 OSS，失败返回 4xx，**不在 Phase 1 做 ffmpeg 转码**。
- 视频 `RegisterMedia`（`mediaKind=2`）：按 `transformVersion` 对 OSS 对象 ffprobe 验真；仅允许 `v1`、`v2`；拒绝 `sim-raw` 及其他版本；失败 MUST 拒绝登记。
- 原生 presign 路径保留；客户端转码与 `v2` register 由 Flutter 侧交付（本仓库 Phase 1 实现服务端验真）。
- 新增 `POST /ucg/internal/api/media/upload-video`（或等价 internal 路径）：multipart 收视频 → ffmpeg 转码为 v2 canonical → PUT OSS → 返回 `objectKey`、`contentHash`、`cdnUrl`；供 sim T4 与后续 Web Phase 2 复用。
- `sim-user-service`：`uploadVideoFromURL` 改调 internal 转码上传，弃用 presign 直传 raw 视频 + `sim-raw` register。
- `Dockerfile.ucg-service` 安装 `ffmpeg`、`ffprobe`；可配置转码并发与超时。
- upload 响应（Web 代理路径）可选增加 `contentHash` 便于 register（见 design）。
- **不在本变更**：Web Phase 2 canonical 转码、历史 OSS 视频 backfill、Flutter 客户端 ffmpeg 实现。

## Capabilities

### New Capabilities

- `ucg-video-validate`：v1/v2 ffprobe 验真规则、register 闸门、允许/拒绝的 `transformVersion` 列表。
- `ucg-video-transcode`：ffmpeg 转码为 v2 canonical、`NormalizeVideo` 语义、internal 转码上传 API、Docker ffmpeg 依赖。

### Modified Capabilities

- `ucg-oss-presign`：视频 register 必须经验真；视频 blob 与 `transformVersion` 语义扩展（v1 非 canonical、v2 canonical）。
- `sim-user-service`：T4 视频发帖 MUST 经 ucg internal 转码上传，禁止 `sim-raw` presign 直传。

## Impact

- **进程**：`ucg-service`（验真、转码、internal API、Web upload 闸门）、`sim-user-service`（T4 上传路径）；部署须重建 **ucg-service** 镜像（ffmpeg）。
- **代码**：`internal/services/ucg/`（新 `video_validate.go`、`video_transcode.go` 等）、`oss_upload.go`、`media_register.go`、`internal/controller/ucg_media_upload.go`、`ucg_internal` 路由；`internal/services/simuser/clients.go`；`manifest/docker/Dockerfile.ucg-service`。
- **API**：新增 internal 转码上传；Web upload 失败语义变更（不合规视频 4xx）；register 拒绝非 v1/v2 与验真失败。
- **客户端**：Flutter 须并行发版（本地 ffmpeg + `v2`）；Web Phase 1 仅能接受已含 AAC 且满足 v1 宽规的文件；无音轨视频在 Phase 2 前 Web 不可用。
- **DB**：无表结构变更；`ucg_media_blob.transform_version` 将出现 `v1`、`v2`，逐步淘汰 `sim-raw` 新写入。
- **OpenSpec 基线**：对照 v2.0.5 `ucg-oss-presign` 与 `sim-user-service` 相关 Requirement 增量。
- **App API 使用统计**：新增 internal 接口不计入 gateway-app usage；若 Web upload 响应增字段不改变路径则无需新登记。
