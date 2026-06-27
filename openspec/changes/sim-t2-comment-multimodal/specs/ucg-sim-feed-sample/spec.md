## MODIFIED Requirements

### Requirement: ucg-service SHALL expose internal published post sample API for sim workloads

系统 MUST 提供 internal 接口 `POST /ucg/internal/api/posts/sample`，供 `sim-user-service` 等受信内部调用方抽取已发布帖子样本。鉴权 MUST 与现有 ucg internal API 一致（`X-Device-Gateway-Internal-Secret`）。响应 MUST 仅包含评论任务所需最小字段，MUST NOT 触发 recommend Feed 的 author/media 富化或 device HTTP 调用。

每条样本 MUST 含 `postId`、`content`、`mediaType`。图文帖（`mediaType=1`）且存在首图 `coverObjectKey` 时 MUST 含 `coverCdnUrl`（全分辨率 CDN URL）。视频帖（`mediaType=2`）且存在封面 media key 时 MUST 含 `coverCdnUrl`（视频首帧 snapshot URL，MUST NOT 为 mp4 直链）。纯文字帖（`mediaType=0`）或无有效 media key 时 MAY 省略 `coverCdnUrl`。`coverObjectKey` MAY 继续返回供调试；LLM 输入 MUST 使用 `coverCdnUrl`。

#### Scenario: Image post sample includes cover CDN URL

- **WHEN** sample 返回 `mediaType=1` 且首条 media 有 objectKey
- **THEN** 响应项 MUST 含非空 `coverCdnUrl`，且 MUST 由 ucg OSS CDN 配置拼装

#### Scenario: Video post sample includes snapshot URL

- **WHEN** sample 返回 `mediaType=2` 且首条 media 有 objectKey
- **THEN** `coverCdnUrl` MUST 为首帧 snapshot URL（含 `x-oss-process=video/snapshot` 语义），MUST NOT 为原始视频文件 URL

#### Scenario: Text-only post omits cover CDN URL

- **WHEN** sample 返回 `mediaType=0`
- **THEN** `coverCdnUrl` MUST 为空或省略

#### Scenario: Invalid or missing secret

- **WHEN** internal 密钥缺失或错误
- **THEN** MUST 返回 403 且 MUST NOT 查询业务表

#### Scenario: Empty plaza

- **WHEN** 无 `status=published` 帖子
- **THEN** MUST 返回空 `list` 与 HTTP 200（code=0）
