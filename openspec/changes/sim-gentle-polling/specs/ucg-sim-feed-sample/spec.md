## ADDED Requirements

### Requirement: ucg-service SHALL expose internal published post sample API for sim workloads

系统 MUST 提供 internal 接口 `POST /ucg/internal/api/posts/sample`，供 `sim-user-service` 等受信内部调用方抽取已发布帖子样本。鉴权 MUST 与现有 ucg internal API 一致（`X-Device-Gateway-Internal-Secret`）。响应 MUST 仅包含评论任务所需最小字段，MUST NOT 触发 recommend Feed 的 author/media 富化或 device HTTP 调用。

#### Scenario: Authenticated sample request

- **WHEN** 请求携带有效 internal 密钥且 body `limit` 为 20
- **THEN** 响应 MUST 返回最多 20 条 `status=published` 帖子，每条含 `postId`、`content`、`mediaType`，图文帖 MAY 含 `coverObjectKey`（首图 objectKey）

#### Scenario: Invalid or missing secret

- **WHEN** internal 密钥缺失或错误
- **THEN** MUST 返回 403 且 MUST NOT 查询业务表

#### Scenario: Empty plaza

- **WHEN** 无 `status=published` 帖子
- **THEN** MUST 返回空 `list` 与 HTTP 200（code=0）

### Requirement: sample API MUST use bounded single-query read pattern

抽样读路径 MUST 在 ucg 库内完成，使用有界 SQL（`LIMIT` ≤ 50），MUST NOT 调用 `postsFromResult`、`GetPublicProfile` 或 `Device().BatchWx`。`limit` 默认 20，最大 50；超出 MUST 截断为 50。

#### Scenario: Limit clamp

- **WHEN** 请求 `limit` 为 100
- **THEN** 实际查询 MUST 最多返回 50 条

#### Scenario: No cross-domain DAO

- **WHEN** 代码评审 ucg internal sample 实现
- **THEN** MUST NOT import device 域 DAO 或直连 device 库表
