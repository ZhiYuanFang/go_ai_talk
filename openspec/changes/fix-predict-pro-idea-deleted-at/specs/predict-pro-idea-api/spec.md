## MODIFIED Requirements

### Requirement: App SHALL list non-deleted ideas newest-first with public fields

`GET /device/app/api/predict-pro-ideas/list` SHALL return non-deleted rows ordered by `created_at` descending. A row MUST be treated as non-deleted when `deleted_at IS NULL OR deleted_at = 0`；`deleted_at > 0` MUST be excluded. Each item MUST expose `id`, `text` (`display_text`), `babyName`, `likeCount`, `likedByMe`, and `isMine`. The response MUST NOT include `wxId` or `deviceNo`. `isMine` MUST be true iff the row’s `wx_id` equals the authenticated caller. `likedByMe` MUST reflect whether a like row exists for that caller. 公开列表 **必须** 新在前且含宝宝名/热度/已赞/本人标记；**不得** 下发 wxId；**不得** 仅因 `deleted_at=0` 而过滤掉活跃行。

#### Scenario: Empty list

- **WHEN** 无未删除想法（不含 `deleted_at=0`/`NULL` 行）
- **THEN** 系统 SHALL 返回空 list 且业务成功

#### Scenario: Rows with deleted_at zero appear

- **WHEN** 库中存在 `deleted_at = 0` 的想法且无其它过滤
- **THEN** `list` MUST 返回这些想法（`code=0` 且 `data.list` 非空）

#### Scenario: isMine for author

- **WHEN** 调用者是某想法的提交者
- **THEN** 该项 `isMine` MUST 为 true

### Requirement: App SHALL like and unlike an idea idempotently

`POST /device/app/api/predict-pro-ideas/like` and `POST /device/app/api/predict-pro-ideas/unlike` MUST accept `{ "id": <ideaId> }` for a non-deleted idea（`deleted_at IS NULL OR deleted_at = 0`）. Like MUST create the unique like row if absent; unlike MUST remove it if present. Both MUST return current `likeCount` and `likedByMe`. Liking a missing/deleted idea（`deleted_at > 0` 或不存在）MUST fail with a business error. 点赞/取消 **必须** 幂等并返回热度；对 `deleted_at=0` 的想法 MUST 允许操作。

#### Scenario: Like then unlike

- **WHEN** 用户对未赞且未删（含 `deleted_at=0`）想法 like 再 unlike
- **THEN** 最终 `likedByMe` MUST 为 false 且 `likeCount` MUST 反映取消后的数量

## ADDED Requirements

### Requirement: Admin idea list and mutations SHALL use the same alive predicate

Admin 分页列表、改 `display_text`、软删前置加载 MUST 与 App 使用同一未删谓词：`deleted_at IS NULL OR deleted_at = 0`。Hub 与 App **不得** 因谓词不一致而一边有数据一边为空。

#### Scenario: Hub shows zero-deleted_at rows

- **WHEN** 库中存在 `deleted_at = 0` 的想法
- **THEN** Admin list MUST 返回这些行且 total 计入它们
