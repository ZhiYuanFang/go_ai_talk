## Why

Phase 1（`ucg-video-normalize-phase1`）已为 Web `POST /ucg/app/api/media/upload` 建立 **v1 宽验真 + 直传 OSS** 闸门，不合规视频一律 4xx；无 wasm 的 Web/Flutter Web 用户无法上传 HEVC、无音轨、非 mp4 等常见源文件。Phase 1 已将 `NormalizeVideo` 与 internal `upload-video` 交付供 sim 使用，**Phase 2** 复用同一转码能力，在 Web 代理上传路径对 **v1 不合规但可解码** 的视频自动转码为 v2 canonical 后再入 OSS，扩大 Web 可用格式面且无需客户端 ffmpeg。

## What Changes

- **Web 视频代理上传**（`POST /ucg/app/api/media/upload`，`mediaKind=2`）：先验 `v1`；**通过**则保持 Phase 1 行为（原始字节直传 OSS，`contentHash` 为原始字节，`transformVersion` 提示 `v1`）；**未通过但 ffprobe 可解码**（含视频轨）则调用 `NormalizeVideo` → PUT v2 canonical mp4 → 响应 `contentHash` 为转码后字节、`transformVersion` 提示 `v2`；**不可解码**则 4xx 且不创建 OSS 对象。
- **v1 已合规仅缺 faststart**：**不** remux，仍直传并 register `v1`（与 Phase 1 v1 规则一致）。
- 扩展 `UploadMediaObject`（`internal/services/ucg/oss_upload.go`）与 controller 响应：成功时可选返回 `transformVersion`（`v1` | `v2`）供 Web/Flutter Web register。
- **复用** Phase 1 已有 `NormalizeVideo`、`UploadVideoTranscodedObject`、转码并发/超时配置；**不**新增 internal API；**不**改 sim-user-service。
- **不在本变更**：Flutter Web 客户端读取 `transformVersion` 并 register（可 sibling repo follow-up）；历史 OSS backfill；presign 原生路径变更。

## Capabilities

### New Capabilities

（无 — 复用 Phase 1 已引入的 `ucg-video-validate`、`ucg-video-transcode` 能力，仅修改 Web 上传挂载行为。）

### Modified Capabilities

- `ucg-oss-presign`：Web 视频代理上传由「仅 v1 直传或拒绝」扩展为「v1 直传 **或** 服务端转码 v2 兜底」；响应增 `transformVersion` 提示字段。
- `ucg-video-validate`：Web 上传闸门语义变更 — v1 失败时区分可解码（转码）与不可解码（4xx）；新增可解码探测 requirement。

## Impact

- **进程**：仅 `ucg-service`（`UploadMediaObject`、`ucg_media_upload` controller）；部署须重建 ucg-service 镜像（ffmpeg 已在 Phase 1 Dockerfile，无新增系统依赖）。
- **代码**：`internal/services/ucg/oss_upload.go`、可选 `video_validate.go`（可解码探测 helper）、`internal/controller/ucg_media_upload.go`；`docs/runbooks/release-deploy-and-run.md` 补充 Phase 2 行为说明。
- **API**：`POST /ucg/app/api/media/upload` 响应 **增字段** `transformVersion`（非 BREAKING）；失败语义对原 4xx 场景部分变为成功转码（行为增强，非破坏性结构变更）。
- **客户端**：Web/Flutter Web 应使用响应 `transformVersion` 与 `contentHash` register；未升级客户端若一律 register `v1` 将在转码路径失败 register（产品需同步发版或文档说明）。
- **DB / Redis**：无表结构变更；无新增 Redis 读缓存。
- **OpenSpec 基线**：增量对照 Phase 1 change 中 `ucg-oss-presign`、`ucg-video-validate` Requirement；归档前须与 v2.0.5+ 合并。
- **App API 使用统计**：无新 App 路径；upload 响应增字段，**不计入**新 usage 登记（与 Phase 1 结论一致）。
