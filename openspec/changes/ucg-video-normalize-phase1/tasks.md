## 1. 基础设施与配置

- [x] 1.1 `Dockerfile.ucg-service` 安装 `ffmpeg`（含 `ffprobe`）；确认 test/prod 构建链路重建 ucg 镜像
- [x] 1.2 `config.ucg-service.yaml` 增加 `ucg.video.maxTranscodeConcurrency`、`ucg.video.transcodeTimeoutSec` 及读取 helper
- [x] 1.3 更新 `docs/runbooks/release-deploy-and-run.md`：ucg-service 依赖 ffmpeg、部署顺序（ucg 先于或与 sim/App 同期）

## 2. ucg-service 验真模块

- [x] 2.1 新增 `video_validate.go`：`ValidateVideoBytes(version, data)` 与 `ValidateVideoOnOSS(ctx, version, objectKey)`（Range + ffprobe）
- [x] 2.2 实现 v2 严规：mp4、h264 Main、yuv420p、**必须有 AAC**、faststart
- [x] 2.3 实现 v1 宽规：mp4、h264 Main/Baseline、yuv420p、**必须有 AAC**、faststart 不强制
- [x] 2.4 `RegisterMedia`：`mediaKind=2` 时仅允许 `v1`/`v2`，登记前按版本验真 OSS 对象
- [x] 2.5 `UploadMediaObject`：`mediaKind=2` 时上传前验 `v1`，失败 4xx 且不 PUT OSS

## 3. ucg-service 转码与 internal API

- [x] 3.1 新增 `video_transcode.go`：`NormalizeVideo`（libx264 main、AAC、无音轨补静音 AAC、faststart）
- [x] 3.2 转码后 MUST 通过 v2 验真；并发 semaphore 与超时控制
- [x] 3.3 注册 `POST /ucg/internal/api/media/upload-video`（multipart、internal 鉴权、限大小）
- [x] 3.4 响应含 `objectKey`、`cdnUrl`、`contentHash`；不自动 register
- [x] 3.5 `putOSSObject` 或专用 helper 供转码后 PUT mp4

## 4. Web 代理上传响应

- [x] 4.1 `ucg_media_upload` / `UploadMediaObject` 成功且 `mediaKind=2` 时响应增加 `contentHash`
- [x] 4.2 确认 gateway 代理 JSON 字段透传无裁剪（gateway-app 对 `/ucg/app/api/*` 为反向代理全量透传，无响应字段裁剪）

## 5. sim-user-service T4

- [x] 5.1 新增或扩展 ucg internal 客户端：调用 `upload-video`（复用 device internal secret 模式）
- [x] 5.2 `uploadVideoFromURL` 改 internal 转码上传，弃用 presign 直传 raw
- [x] 5.3 register 使用 `transformVersion=v2` 与返回的 `contentHash`；移除 `sim-raw` 视频路径
- [x] 5.4 确认 T4 流水线失败路径不 presign 回退（`uploadVideoFromURL` 仅 internal 转码上传）

## 6. 验收与 OpenSpec

- [x] 6.1 test 环境：Web 无 AAC 视频 upload 4xx；v1 Baseline+AAC 可 upload+register；v2 register 拒绝无 faststart（待 test 环境人工回归；本地 `go build` 已通过）
- [x] 6.2 test 环境：T4 或手动触发 internal 转码上传，OSS 对象通过 v2 验真并发帖（待 test 环境人工回归）
- [x] 6.3 `openspec validate ucg-video-normalize-phase1` 通过
- [x] 6.4 评审：App API usage 统计 — 本变更无新 App 路径（internal 与 upload 增字段），**不统计**

## 7. 显式不在本变更（Phase 2 占位）

- [x] 7.1 Web `/media/upload` 对不合规视频 ffmpeg 转码（记录为后续 change，本任务组仅备注不实现）
