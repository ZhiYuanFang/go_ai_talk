## Context

UCG 媒体上传今日三条视频路径均无格式闸门：App presign 直传 OSS、`POST /ucg/app/api/media/upload` Web 代理直传、sim T4 经 presign + `transformVersion=sim-raw` register。图片已在 `EnsureImageThumb` 实现上传后缩略图；视频无等价处理。

产品决策：

- **原生**：每条本地视频客户端 ffmpeg → H.264 **Main** + **静音 AAC** + faststart → register `v2`。
- **Web Phase 1**：无 wasm；`v1` 宽验真 + 直传 OSS；canonical 转码为 **Phase 2**。
- **v1 与 v2**：验真规则 **部分对齐、刻意分叉**；v1 通过验真 **不等于** canonical。
- **v1 无音轨**：**必须存在 AAC 音轨**；无音轨 MUST 拒绝（Phase 1 不补轨；Phase 2 Web 转码时由服务端补静音 AAC）。
- **sim T4**：纳入服务端转码，产出 v2 canonical。

约束：无新增 `*_test.go`；ffmpeg 仅 ucg-service 镜像；sim 经 HTTP internal 调 ucg；不新增 Redis 读缓存。

## Goals / Non-Goals

**Goals:**

- 定义并实现 `v1` / `v2` 分叉 ffprobe 验真规则与 `ValidateVideoForVersion` 入口。
- Web 视频代理上传：验 `v1` 通过后直传 OSS；失败 4xx。
- 视频 `RegisterMedia`：仅允许 `transformVersion` 为 `v1` 或 `v2`；按版本对 OSS 对象验真；拒绝 `sim-raw` 等新写入。
- `NormalizeVideo`：ffmpeg 转码为 v2 canonical（含无音轨时补静音 AAC）；internal API 供 sim 调用。
- sim `uploadVideoFromURL` 改 internal 转码上传，register `v2`。
- `Dockerfile.ucg-service` 安装 ffmpeg/ffprobe。

**Non-Goals:**

- Web Phase 2：不合规视频服务端转码（后续 change）。
- Flutter 客户端 ffmpeg 实现（并行发版，本仓库仅服务端验 `v2`）。
- 历史 OSS 视频 backfill / 批量重转码。
- 视频物理缩略图（仍用 `BuildVideoSnapshotURL` + OSS process）。
- 变更 `ucg_notification` 历史 `post_thumb_url`。

## Decisions

### 1. v1 / v2 验真规则（分叉表）

| 检查项 | v2（canonical / 原生 / sim 转码产出） | v1（Web Phase 1 直传） |
|--------|--------------------------------------|-------------------------|
| 容器 | mp4 | mp4 |
| 视频 codec | h264 | h264 |
| profile | **Main** only | **Main 或 Baseline** |
| pix_fmt | yuv420p | yuv420p |
| 音频 | **必须有 AAC 轨**（静音 AAC 合法） | **必须有 AAC 轨**（静音 AAC 合法）；**无音轨拒绝** |
| faststart | **必须**（moov 在 mdat 前） | **不强制** |
| 大小 | ≤ `MaxMediaUploadBytes`（25MB） | 同左 |

**Rationale**：v1 放宽 profile 与 faststart，便于浏览器/导出 mp4 直传；音轨与 v2 对齐（必须 AAC），避免 feed 出现完全无音轨对象。Phase 1 无转码故无法「补轨」，无 AAC 的用户须等 Phase 2 或自行处理源文件。

**Alternatives**：v1 允许无音轨 — 已否决（用户要求必须补轨 AAC，验真层至少要求 AAC 存在）。

### 2. faststart 检测

- **决定**：v2 验真使用 ffprobe 解析 format/streams，并结合对 object 前若干 MB 的 Range 读取判断 moov 位置；实现可封装 `hasFastStart` helper；不确定时 v2 **fail**。
- **备选**：一律 remux — 仅用于 `NormalizeVideo` 产出，不用于 v1 直传验真。

### 3. 验真挂载点

| 路径 | 时机 | 规则 |
|------|------|------|
| `UploadMediaObject`（Web，`mediaKind=2`） | PUT OSS **之前**，对内存/临时文件 ffprobe | `v1` |
| `RegisterMedia`（`mediaKind=2`） | blob 登记前，OSS GetObject Range + ffprobe | 请求中的 `transformVersion` |
| internal 转码上传 | `NormalizeVideo` 输出后 ffprobe | `v2`（转码失败则 5xx） |

原生 presign **不在 PUT 时验**；依赖 register + `v2` 验真（与 App 客户端转码并行上线）。

### 4. `transformVersion` 允许列表

- **允许 register**：`v1`、`v2` only。
- **拒绝**：`sim-raw`、空、及其他未登记版本 → 400。
- **dedup**：`(contentHash, transformVersion)` 不变；v1 与 v2 天然不跨版本命中。

### 5. Internal 转码 API

- **路径**：`POST /ucg/internal/api/media/upload-video`（multipart 字段 `file`）。
- **鉴权**：与现有 `POST /ucg/internal/api/media/upload` 相同（`X-Gateway-Internal-Secret` / device internal secret）。
- **流程**：读 body（限 `MaxMediaUploadBytes`）→ `NormalizeVideo` → `putOSSObject`（mp4）→ 计算 SHA-256 → 响应 `objectKey`、`cdnUrl`、`contentHash`（**不**自动 register；sim 持用户 token 自行 register `v2`，或 internal 可选写 ownership — **决定：不自动 register**，与现有 upload 一致，sim 继续 `appPost register`）。
- **理由**：sim 已有 token 与 register 流程；保持 ownership 与 wxId 绑定在 App API。

### 6. `NormalizeVideo` ffmpeg 参数（v2 canonical）

与 Flutter 对齐：

```text
-c:v libx264 -profile:v main -pix_fmt yuv420p
-c:a aac -b:a 128k
-movflags +faststart
```

无音轨输入：`ffmpeg` 使用 `anullsrc`（或等价）生成静音 AAC 并 `-shortest` 对齐视频时长。

输出扩展名 **mp4**；objectKey 使用 `buildObjectKey(..., "mp4")`。

### 7. sim T4 改造

- `uploadVideoFromURL`：HTTP POST internal `upload-video`（不经用户 token）→ 拿 `objectKey` + `contentHash` → 用户 token `register` `v2`。
- **备选**：sim 无 internal secret — 查 `device` 包已有 ucg internal 调用模式（`ucg_upload_client`）复用。

### 8. Web upload 响应

- **决定**：`POST /ucg/app/api/media/upload` 在 `mediaKind=2` 成功时额外返回 `contentHash`（小写 hex SHA-256 of **OSS 上字节**，Phase 1 与上传字节相同）、`transformVersion: "v1"` 提示字段（可选），便于 Web register。
- v1 API 结构增量字段，非 BREAKING。

### 9. Docker 与运行时

- `Dockerfile.ucg-service`：`apk add ffmpeg`（含 ffprobe）。
- 配置项（`config.ucg-service.yaml`）：`ucg.video.maxTranscodeConcurrency`（默认 2）、`ucg.video.transcodeTimeoutSec`（默认 120）。
- 临时文件：`os.TempDir()` 下子目录，转码完立即删除。

### 10. 失败语义

- 验真失败（v1/v2）：4xx，`message` 含「视频格式不合规」类中文说明。
- ffmpeg 缺失/转码失败：5xx。
- register 验真失败：不删 OSS 对象（与今日一致）；客户端可 `media/delete` 清理孤儿。可选后续增强。

## Risks / Trade-offs

- [Web Phase 1 无音轨视频不可用] → Phase 2 服务端补轨；产品说明 Web 上传须已含 AAC。
- [Web 仅 mp4+h264+AAC 窄门] → Phase 2 转码兜底；mov/HEVC 仍拒直到 Phase 2。
- [register 延迟增加] → OSS Range + ffprobe；视频 register 频率低于图片。
- [ucg CPU] → sim 转码 + register 验真；并发 semaphore 限制。
- [原生旧版非 v2] → register 拒绝，促升级。
- [v1 非 canonical 起播慢] → 接受至 Phase 2；v2 对象 faststart 正常。

## Migration Plan

1. 部署 **ucg-service** 新镜像（含 ffmpeg）至 test。
2. 部署 **sim-user-service**（internal 上传路径）。
3. 原生 App 发版（本地 ffmpeg + `v2`）与 ucg 验真 **同期或略晚**；先发 ucg 时旧 App register 可能失败。
4. Web：无客户端变更即可获 v1 验真；无 AAC 视频将失败直至 Phase 2。
5. 回滚：回滚 ucg 镜像；sim 回滚后可能再次 presign 直传（不推荐长期）。

## Open Questions

- v1 是否 **拒绝 mov** 容器（当前设计：仅 mp4）— 实现按仅 mp4。
- internal `upload-video` 是否需计入 gateway-app usage — **否**（internal 路径）。
- Phase 2 change 名称与是否合并 `NormalizeVideo` 用于 Web `/media/upload` — 后续 proposal。
