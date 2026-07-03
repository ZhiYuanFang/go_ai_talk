## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: sample API SHALL support excludeDebate and onlyDebate filters

`POST /ucg/internal/api/posts/sample` 请求 body MAY 含 `excludeDebate`（bool）与 `onlyDebate`（bool）。二者 MUST NOT 同时为 true。判定 MUST 与 UCG `isDebatePost` 一致：`(debate_left_label != '' AND debate_right_label != '') OR type = 'debate'`。

- `excludeDebate=true`：MUST 仅返回非辩论帖；random 的 MIN/MAX 与 probe MUST 应用相同 filter。
- `onlyDebate=true`：MUST 仅返回辩论帖；random 的 MIN/MAX 与 probe MUST 应用相同 filter。
- 均未传或为 false：MUST 保持变更前行为（含辩论帖）。

响应 `list` 每项 MUST 含 `debateLeft`、`debateRight`（字符串；非辩论帖 MUST 为空字符串或省略）。T2 MAY 忽略；T8 MUST 使用。

#### Scenario: T2 exclude debate in SQL

- **WHEN** 请求 `{ "mode": "random", "excludeDebate": true }` 且存在 published 非辩论帖
- **THEN** 返回样本 MUST NOT 为辩论帖

#### Scenario: T8 only debate in SQL

- **WHEN** 请求 `{ "mode": "random", "onlyDebate": true }` 且存在 published 辩论帖
- **THEN** 返回样本 MUST 为辩论帖且 MUST 含非空 `debateLeft` 与 `debateRight`

#### Scenario: Mutually exclusive flags rejected

- **WHEN** 请求同时 `excludeDebate=true` 且 `onlyDebate=true`
- **THEN** MUST 返回 400 且 MUST NOT 查询

#### Scenario: Empty when only debate and none exist

- **WHEN** `onlyDebate=true` 且无 published 辩论帖
- **THEN** MUST 返回空 `list` 与 HTTP 200（code=0）

#### Scenario: Debate fields on debate sample

- **WHEN** sample 返回辩论帖
- **THEN** `debateLeft` 与 `debateRight` MUST 各 ≤5 字且非空
