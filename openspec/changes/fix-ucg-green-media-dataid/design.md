## Context

`internal/services/ucg/green_client.go` 中 `ModerateImageURL` / `ModerateVideoURL` 将 `ServiceParameters.dataId` 设为完整 CDN URL（约 79 字符，且含 `:`、`/` 等非法字符）。阿里云 ImageModeration 要求 `dataId` 可选、最长 64、字符集 `[A-Za-z0-9_.-]`。

受影响 call site 均在 `audit_moderation.go`（帖子媒体、资料头像、私信图/视频），均只传 URL，不改接口即可集中修复。

典型 objectKey：`social/2026/06/{32随机}.jpg`（约 51 字符 path）；CDN URL 为 `https://resorce.cuplay.top/{objectKey}`。

## Goals / Non-Goals

**Goals:**

- 图片/视频 Green 请求的 `dataId` 合规，使 `baselineCheck` / `videoDetection` 能通过参数校验并成功发起审核。
- 在 `green_client` 单点实现，call site 零改。
- 改进 image/video 解析层错误信息（含 business `Code` / `Msg`），便于日志与运维。

**Non-Goals:**

- 扩展 `GreenModerator` 接口传入业务 ID（如 `post-{id}-s{sort}`）。
- 添加 `infoType`、切换 `postImageCheck`、OSS 授权检测（`ossBucketName` 等）。
- 为 post status=5 增加 Admin 重审 UI 或自动重跑 Green。
- 修改 `parseVideoModeration` 对风险 label 的解析（仍为 MVP pass-on-200）。

## Decisions

### 1. 在 `green_client` 内从 URL 推导 `dataId`（方案 B）

**选择**：新增 `greenDataIDFromMediaURL(url string) string`：

1. `url.Parse` 取 `Path`，去首尾 `/`。
2. 将 `/` 替换为 `_`；过滤掉非 `[A-Za-z0-9_.-]` 字符。
3. 若结果为空，**不传** `dataId`（字段 optional）。
4. 若长度 > 64，截断至 64（object path 正常约 51 字符，截断仅作兜底）。

**备选**：完全省略 `dataId`——能修复校验失败，但 Green 控制台无法关联 OSS 对象；规范化 path 更可观测。

**备选**：扩展接口让 call site 传 `post-{id}-m-{mediaId}`——可观测性更好，但改动面大；本变更不采用。

### 2. 图片与视频同步修复

视频当前同样 `dataId=URL`；虽可能未触发失败，为一致性与未来校验收紧，**两路径共用** `greenDataIDFromMediaURL`。

### 3. 构建 ServiceParameters 时条件写入 dataId

仅当 `greenDataIDFromMediaURL` 返回非空字符串时写入 JSON；避免传空字符串触发另一类参数错误。

### 4. 错误信息对齐文本路径

`parseImageModeration` / `parseVideoModeration` 在 `body.Code != 200` 时返回 `fmt.Errorf("green image: code %d msg %s", code, msg)`（与 `parseTextModeration` 风格一致）；HTTP 非 200 时带上 status code。

## Risks / Trade-offs

- **[Risk] 截断 64 字符导致极少数 path 碰撞** → 正常 `social/YYYY/MM/` 前缀下 path 约 51 字符，碰撞概率可忽略；若未来 prefix 变长需改用 hash。
- **[Risk] URL path 与 objectKey 不一致的 CDN 配置** → 当前 `BuildCdnURL` 为 `cdnBase + "/" + objectKey`，path 即 objectKey；若 CDN 加前缀 rewrite 需另开变更。
- **[Risk] 修复后 status=5 存量帖不自动恢复** → runbook 说明：Admin 处理或作者重新提交；green-once 语义不变。
- **[Trade-off] 仍用 CDN URL 拉图而非 OSS 授权** → 若 CDN 防盗链/3s 超时仍会导致失败，需后续变更；本修复只解决 dataId 参数问题。

## Migration Plan

1. 部署 `ucg-service` 新版本（仅代码变更，无 DDL）。
2. 验证：新发带图帖 → Green 控制台出现 `baselineCheck`；日志无 `green image: code not ok`（或 Msg 明确为其它原因）。
3. 存量 status=5：按需 SQL 查 `reject_reason` 含 `green image`；人工通过/驳回或让用户重发。
4. 回滚：还原 `green_client.go` 即可，无数据迁移。

## Open Questions

- 修复 dataId 后若图片仍失败，是否在同一变更追加 `infoType` 或 OSS 授权检测？**当前决策：否，另开变更。**
