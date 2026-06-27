## Context

Phase 1 已实现：

- `ValidateVideoBytes` / `ValidateVideoOnOSS`：v1 宽规与 v2 严规分叉验真。
- Web `UploadMediaObject`：`mediaKind=2` 时 v1 验真通过才 PUT OSS，否则 4xx。
- `NormalizeVideo` + `UploadVideoTranscodedObject`：internal `POST /ucg/internal/api/media/upload-video` 供 sim T4 转码上传。
- `Dockerfile.ucg-service` 已含 ffmpeg/ffprobe；`ucg.video.maxTranscodeConcurrency`、`transcodeTimeoutSec` 已配置。

Phase 2 将 **同一 NormalizeVideo 管线** 挂载到 App Web 代理上传，使浏览器/Flutter Web（经 gateway 同域 upload）在源文件不满足 v1 时仍可获得 v2 canonical OSS 对象，无需客户端 wasm ffmpeg。

约束：无新增 `*_test.go`；转码仅在 ucg-service；不新增 Redis 读缓存；不修改 sim-user-service。

## Goals / Non-Goals

**Goals:**

- Web 视频 upload：v1 合规 → 直传 + `transformVersion=v1` 提示；v1 不合规但可解码 → `NormalizeVideo` → OSS v2 + `transformVersion=v2` 提示。
- 响应始终含 `objectKey`、`cdnUrl`、`contentHash`（对 **OSS 最终字节** 计算）。
- 不可解码输入 MUST 4xx，且不留下 OSS 对象。
- 复用 `UploadVideoTranscodedObject` / `NormalizeVideo` 与 Phase 1 转码 semaphore、超时。
- runbook 文档化 Phase 2 行为与 register 约定。

**Non-Goals:**

- Flutter Web 客户端改造（sibling repo；本变更仅服务端 + design 备注）。
- v1 合规对象因缺 faststart 而 remux（**决定：仍直传 v1**，与 Phase 1 v1 规则一致）。
- presign 原生路径 upload 时转码（仍依赖客户端 ffmpeg + v2 register）。
- 历史 OSS 对象 backfill。
- 新增 internal API 或 gateway 新路由。

## Decisions

### 1. Web upload 三分支流程

| 分支 | 条件 | 动作 | register 版本 |
|------|------|------|---------------|
| A 直传 | `ValidateVideoBytes(v1)` 通过 | 原始字节 PUT OSS | `v1` |
| B 转码 | v1 失败 **且** `ProbeVideoDecodable` 为 true | `NormalizeVideo` → PUT mp4 | `v2` |
| C 拒绝 | v1 失败 **且** 不可解码 | 4xx，不 PUT | — |

**Rationale**：与 Phase 1 向后兼容 — 已能直传的 v1 文件行为不变；扩大 Web 格式面仅对 B 分支生效。

**Alternatives**：v1 失败一律尝试 ffmpeg — 已采纳为 B 分支，但 C 分支在 ffprobe 完全失败时短路 4xx，避免无意义转码与 5xx 混淆。

### 2. 「可解码」判定

- **决定**：新增 `ProbeVideoDecodable(data []byte) error`（或等价）：写入临时文件 → `runFFprobe` 成功 **且** streams 中存在 `codec_type=video` 即视为可解码；否则不可解码。
- **不** 要求容器为 mp4、h264 或 AAC — 这些由 v1 验真负责；HEVC/mp3 音轨/mov 等 ffprobe 可解析且有视频轨即走 B 分支。
- ffprobe 失败（损坏、非视频）→ C 分支 4xx「视频格式不合规：无法解析视频」。

### 3. v1 合规但缺 faststart

- **决定**：若 v1 验真通过（faststart 不强制），**直传分支 A**，register `v1`；**不** 为 faststart 单独 remux。
- **Rationale**：Phase 1 产品决策已接受 v1 非 canonical；Phase 2 不扩大 remux 范围以免意外改变 contentHash/dedup。

### 4. 实现挂载点

- **决定**：扩展 `UploadMediaObject`（`mediaKind=2`）：
  1. 读满 body（已有）。
  2. `ValidateVideoBytes(VideoTransformV1, data)` — 成功 → 直传（现有逻辑）。
  3. 失败 → `ProbeVideoDecodable(data)` — 失败 → 返回 v1 验真错误或合并 4xx。
  4. 成功 → 调用 `UploadVideoTranscodedObject(ctx, data)` 或内联 `NormalizeVideo` + `putOSSObjectBytes`（**优先复用 `UploadVideoTranscodedObject`** 避免重复）。
- `UploadMediaResult` 增加 `TransformVersion string`（`v1` | `v2`）。
- Controller `ucg_media_upload` 在 `mediaKind=2` 成功响应写入 `transformVersion`。

### 5. contentHash 语义

- **A 直传**：SHA-256 of **原始上传字节**（与 Phase 1 一致）。
- **B 转码**：SHA-256 of **转码后 OSS 字节**（与 internal upload-video 一致）。
- register MUST 使用响应 `transformVersion` 与 `contentHash` 配对；`(hash, version)` dedup 不变。

### 6. 失败语义

| 场景 | HTTP |
|------|------|
| v1 失败 + 不可解码 | 4xx（CodeInvalidParameter） |
| v1 失败 + 可解码但 NormalizeVideo 失败/超时 | 5xx |
| 超过 MaxMediaUploadBytes | 4xx |

转码失败 MUST NOT 留下 OSS 对象（`UploadVideoTranscodedObject` 先转码后 PUT，已满足）。

### 7. 并发与资源

- 复用 Phase 1 `maxTranscodeConcurrency` semaphore；Web B 分支与 sim internal 转码共享同一池。
- 临时文件：`writeTempVideoFile` / 转码输出，defer 清理（与 Phase 1 一致）。

### 8. Flutter Web follow-up

- Flutter Web 经 gateway `POST /ucg/app/api/media/upload` 上传视频时，**应**读取 `transformVersion` 与 `contentHash` 再 `RegisterMedia`（sibling `flutter_ai_talk` 任务，本仓库 design 备注，不阻塞 Phase 2 服务端交付）。

### 9. App API usage

- 无新路径；响应增 `transformVersion` 为 additive JSON 字段 → **不** 修改 usagestats。

## Risks / Trade-offs

- [Web 转码 CPU 延迟] → 共享 semaphore；产品预期 B 分支慢于 A。
- [未升级 Web 客户端仍 register v1] → register 验真失败；文档 + Flutter Web follow-up。
- [HEVC 等大文件转码超时] → 沿用 `transcodeTimeoutSec`；超时 5xx，用户可压缩后重试。
- [同文件 v1 直传与转码 contentHash 不同] → 预期行为；dedup 按 `(hash, version)` 隔离。

## Migration Plan

1. 部署 ucg-service 新镜像（含 Phase 2 代码；ffmpeg 无变更）。
2. test 回归：v1 合规仍直传；HEVC/无音轨可转码成功且 register v2；损坏文件 4xx。
3. Web/Flutter Web 客户端按需发版以消费 `transformVersion`（可与服务端独立）。
4. 回滚：回滚 ucg-service → Web 恢复 Phase 1 纯 v1 闸门（不合规再 4xx）。

## Open Questions

- Flutter Web 发版节奏 — 由 sibling repo 负责人排期；服务端可先上线。
- 是否在响应增加 `transcoded: true` 布尔 — **否**，`transformVersion` 足够表达。
