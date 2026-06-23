## Context

- 现状：`internal/services/ucg/oss_cdn.go` 中 `BuildImageThumbnailURL` 对 `BuildCdnURL(objectKey)` 追加 `x-oss-process=image/auto-orient,1/resize,m_lfit,w_200/quality,q_90/format,jpg`。
- 调用方：`post.go`、`profile.go`、`chat_store.go`、`notification.go`（写入时快照）等 6 处；字段名已暴露给 App/Admin。
- 上传路径：App presign → 客户端 PUT 原图 → `RegisterMedia`；另有 `UploadMediaObject` / `UploadEventLogoObject` 服务端直传（事件 logo 本变更不纳入缩略图范围）。
- 约束：服务边界、Redis 无新增读缓存；不新增 `*_test.go`；中文注释覆盖关键流程。

## Goals / Non-Goals

**Goals:**

- 图片 `thumbnailUrl` 使用独立 OSS 对象，CDN URL path 与原图不同，避免窜缓存。
- `_thumb` 后缀全局单一定义；thumb key 为 `stem_thumb.ext`，扩展名与原图一致。
- 新上传与历史数据（backfill）使用同一 `EnsureImageThumb` 逻辑。
- 删除媒体时成对清理 thumb 对象。

**Non-Goals:**

- 视频首帧 `BuildVideoSnapshotURL`（仍用 `x-oss-process`）。
- 事件 logo（`event/`）缩略图与管理端展示改造。
- 批量更新 `ucg_notification.post_thumb_url` 历史行。
- 客户端/SDK 改动（继续消费服务端返回的 `thumbnailUrl`）。

## Decisions

### 1. 命名与 helper 位置

- **决定**：`internal/shared/mediacdn` 导出 `ImageThumbSuffix = "_thumb"`、`ThumbObjectKey(key)`、`IsThumbObjectKey(key)`。
- **理由**：ucg 与潜在其他域共用；禁止业务层拼后缀字符串。
- **算法**：对 objectKey 取最后一段扩展名 `ext`，`stem = TrimSuffix(key, "."+ext)`；若 `IsThumbObjectKey` 则原样返回；否则 `stem + ImageThumbSuffix + "." + ext`。

### 2. 缩略图生成方式

- **决定**：`EnsureImageThumb(ctx, objectKey)` 内：`HEAD` thumb 已存在则跳过；`GetObject(原图, oss.Process(...))` 拉取处理后字节 → `PutObject(thumbKey)`；Content-Type 与原图一致。
- **参数**：`auto-orient,1/resize,m_lfit,w_200/quality,q_90`；按扩展名追加 `format,png|webp|gif`（jpg/jpeg 可省略 format 或 `format,jpg`）。
- **备选**：Go 本地 imaging resize — 需新依赖与 EXIF 处理，与现网像素一致性难保证；未采用。
- **说明**：CDN 读路径不再带 `x-oss-process`，但**服务端生成阶段**仍可用 OSS 图片处理能力，二者不冲突。

### 3. 挂钩点

| 路径 | 时机 |
|------|------|
| `RegisterMedia` miss（新 blob，原图已存在） | 事务成功后、`return` 前调用 `EnsureImageThumb` |
| `RegisterMedia` dedup hit | thumb 应已存在；`EnsureImageThumb` 幂等跳过 |
| `putOSSObject` | `PutObject` 原图成功后，若 `mediaKind==1` 则 `EnsureImageThumb` |

`PresignUpload` 不生成 thumb（客户端尚未上传）。

### 4. 读路径

- `BuildImageThumbnailURL(objectKey)` → `BuildCdnURL(ThumbObjectKey(objectKey))`。
- 移除图片侧 `imageThumbProcess` 与 `appendOssProcess` 用于图片的分支；保留 `BuildVideoSnapshotURL`。

### 5. 删除

- `DeleteOwnedMedia` / blob ref 归零删 OSS 时：删除原图 key 后 MUST `DeleteObject(ThumbObjectKey(key))`（忽略 404）。

### 6. 历史数据（方案 A）

- CLI `cmd/ucg-image-thumb-backfill`：从 UCG DB 收集去重图片 objectKey（`ucg_media_blob` media_kind=1、`ucg_post_media`、`ucg_profile.avatar_key`、`ucg_chat_message.image_key` 等），过滤已是 thumb key、非图片扩展名。
- 标志：`--dry-run`、`--limit`、`--concurrency`（默认 4）；输出 ok/skipped/missing/failed 汇总。
- **发布顺序**（推荐）：
  1. 部署写路径（register/直传生成 thumb）
  2. 各环境跑 backfill 并抽样验证
  3. 部署读路径切换（`BuildImageThumbnailURL`）
- 数据量小，不做读时 fallback 到 `x-oss-process`。

### 7. 失败语义

- `EnsureImageThumb` 在 register 路径失败 SHOULD 使 register 失败（避免登记成功但列表裂图）。
- backfill 单条失败记日志，脚本可重跑（幂等）。

## Risks / Trade-offs

- [Register 延迟增加] → 单次 OSS process + PUT；数据量与图片尺寸可控，可接受。
- [backfill 未完成即切读路径] → runbook 强制先 backfill；部署检查清单。
- [OSS 图片处理不可用] → register 失败可观测；与现网依赖同级。
- [通知表历史 URL 仍为 query 形式] → 接受；新通知写入新 URL。
- [GIF 动图] → OSS process 默认首帧；与现网列表缩略图语义一致。

## Migration Plan

1. 合并并部署含写路径的版本（可先保留读路径为 oss-process，或同事务一次发布）。
2. test 环境：`go run ./cmd/ucg-image-thumb-backfill --dry-run` → 正式执行 → 抽样 HEAD/CURL thumb URL。
3. prod 重复步骤 2。
4. 若分阶段发布：backfill 完成后部署读路径切换。
5. 回滚：读路径回退 `x-oss-process` 可快速恢复展示；已写入的 thumb 对象可保留无害。

## Open Questions

- 无（探索阶段已确认：`abc_thumb.jpg`、扩展名一致、方案 A backfill、OSS process 生成）。
