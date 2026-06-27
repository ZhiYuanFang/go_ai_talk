## ADDED Requirements

### Requirement: sample API SHALL support excludeMediaTypes filter

`POST /ucg/internal/api/posts/sample` 请求 body MAY 含 `excludeMediaTypes`（整型数组）。当数组非空时，抽样 MUST 仅返回 `media_type` **不在**该集合中的已发布帖。`mode=random` 的 ID 探测（MIN/MAX 与 `id>=R` probe）MUST 应用与列表查询相同的 `excludeMediaTypes` filter。未传或空数组时 MUST 保持变更前行为（含视频帖）。

#### Scenario: Random sample excludes video

- **WHEN** 请求为 `{ "mode": "random", "excludeMediaTypes": [2] }` 且存在 published 非视频帖
- **THEN** 返回的每条样本 `mediaType` MUST NOT 为 `2`

#### Scenario: Random bounds respect exclude filter

- **WHEN** 请求 `mode=random` 且 `excludeMediaTypes` 为 `[2]`
- **THEN** MIN/MAX id 聚合 MUST 仅在 `media_type NOT IN (2)` 的 published 帖上计算

#### Scenario: No eligible posts after exclude

- **WHEN** 所有 published 帖均为 `media_type=2` 且 `excludeMediaTypes` 含 `2`
- **THEN** MUST 返回空 `list` 与 HTTP 200（code=0）

#### Scenario: Backward compatible without exclude

- **WHEN** 请求未含 `excludeMediaTypes`
- **THEN** random/latest 行为 MUST 与引入本字段前一致

## MODIFIED Requirements

### Requirement: sample API SHALL support random mode via ID probe with mild recency bias

当请求 body `mode` 为 `random` 时，系统 MUST 在 ucg 库内通过有界 ID 探测返回 **0 或 1** 条满足 **status 与 excludeMediaTypes（若提供）** 条件的已发布帖，MUST NOT 使用 `ORDER BY RAND()`。探测 MUST 在过滤后的 published 集合上覆盖全库（非仅最新 N 条）。锚点 MUST 在 eligible 帖的 `[minId, maxId]` 上按 `R = minId + floor((maxId - minId) * U^α)` 生成（`U` 均匀随机，默认 `α = 0.65`），随后 `WHERE … AND id >= R ORDER BY id ASC LIMIT 1`。响应字段 MUST 含 `postId`、`content`、`mediaType`、可选 `coverObjectKey`/`coverCdnUrl`。

#### Scenario: Random mode returns one published post

- **WHEN** 有效 internal 密钥、body `{ "mode": "random" }`（或无 exclude），且存在 eligible published 帖
- **THEN** 响应 `list` MUST 含 1 条 published 帖

#### Scenario: Random mode empty plaza

- **WHEN** 无 eligible published 帖（含 exclude 后为空）
- **THEN** MUST 返回空 `list` 与 HTTP 200（code=0）

#### Scenario: Latest mode unchanged without exclude

- **WHEN** `mode=latest`（或缺省）、`limit=20`、无 `excludeMediaTypes`
- **THEN** MUST 仍按 `published_at DESC` 返回最多 20 条
