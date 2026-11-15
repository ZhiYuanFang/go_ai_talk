## ADDED Requirements

### Requirement: App predict-pro-ideas APIs SHALL require JWT but MUST NOT require Pro entitlement

All `/device/app/api/predict-pro-ideas/*` endpoints MUST require a valid access_token (JWT). The server SHALL derive `wx_id` from the token. Unauthenticated requests MUST return 401. Submit MUST require a bound `deviceNo`. List/like/unlike MAY succeed for any authenticated user without a bound device when not otherwise required. The server MUST NOT call cash or read `predict_capability_pro` / `feature_user_entitlement` as a gate. App 想法接口 **必须** JWT；**任意登录用户**可提交/点赞；提交另 **必须** 已绑定设备；**不得**以 Pro 开通态拦截。

#### Scenario: Unauthenticated list

- **WHEN** 请求 `GET /device/app/api/predict-pro-ideas/list` 无有效 JWT
- **THEN** 系统 MUST 返回 401

#### Scenario: Non-Pro user may submit

- **WHEN** 已登录已绑定用户未持有预测 Pro 权益并提交合法 text
- **THEN** 系统 SHALL 仍允许插入（若未超日限额），MUST NOT 因缺少 Pro 而拒绝

### Requirement: App SHALL list non-deleted ideas newest-first with public fields

`GET /device/app/api/predict-pro-ideas/list` SHALL return non-deleted rows ordered by `created_at` descending. Each item MUST expose `id`, `text` (`display_text`), `babyName`, `likeCount`, `likedByMe`, and `isMine`. The response MUST NOT include `wxId` or `deviceNo`. `isMine` MUST be true iff the row’s `wx_id` equals the authenticated caller. `likedByMe` MUST reflect whether a like row exists for that caller. 公开列表 **必须** 新在前且含宝宝名/热度/已赞/本人标记；**不得** 下发 wxId。

#### Scenario: Empty list

- **WHEN** 无未删除想法
- **THEN** 系统 SHALL 返回空 list 且业务成功

#### Scenario: isMine for author

- **WHEN** 调用者是某想法的提交者
- **THEN** 该项 `isMine` MUST 为 true

### Requirement: App SHALL submit an idea with validated text and baby name snapshot

`POST /device/app/api/predict-pro-ideas/submit` MUST accept `{ "text": "<content>" }`. After trim, text MUST be non-empty and MUST NOT exceed 200 characters. On success the server SHALL insert a row with `content` and `display_text` equal to the trimmed text, `wx_id` from JWT, `device_no` from the bound session, and `baby_name` from the bound device user. Missing bind MUST fail without insert. 提交 **必须** 校验并快照宝宝名。

#### Scenario: Successful submit

- **WHEN** 已登录已绑定用户提交合法 text
- **THEN** 系统 SHALL 插入一行并返回新建 `id`（或含 id 的记录）

#### Scenario: Empty text rejected

- **WHEN** text trim 后为空
- **THEN** 系统 MUST 拒绝且 MUST NOT 插入

### Requirement: App submit SHALL rate-limit per wx_id by table daily COUNT

The server MUST enforce a per-`wx_id` submit quota of at most **10** ideas per **Asia/Shanghai calendar day**. The quota MUST be computed by `COUNT(*)` on `predict_pro_idea` for that `wx_id` whose `created_at` falls in the current Shanghai day (soft-deleted rows MUST still count toward the quota). The server MUST NOT introduce Redis (or other cache) solely for this quota. Excess submits MUST be rejected without inserting. 同一账号提交 **必须** 按表按日 COUNT 频控。

#### Scenario: Over quota

- **WHEN** 同 wx_id 在当日 COUNT 已达 10 后再提交
- **THEN** 系统 MUST 返回业务错误且 MUST NOT 插入

#### Scenario: Soft-deleted still consumes quota

- **WHEN** 用户当日已提交并被 Admin 软删后再次提交且当日 COUNT（含软删）已达上限
- **THEN** 系统 MUST 仍拒绝新插入

### Requirement: App SHALL like and unlike an idea idempotently

`POST /device/app/api/predict-pro-ideas/like` and `POST /device/app/api/predict-pro-ideas/unlike` MUST accept `{ "id": <ideaId> }` for a non-deleted idea. Like MUST create the unique like row if absent; unlike MUST remove it if present. Both MUST return current `likeCount` and `likedByMe`. Liking a missing/deleted idea MUST fail with a business error. 点赞/取消 **必须** 幂等并返回热度。

#### Scenario: Like then unlike

- **WHEN** 用户对未赞想法 like 再 unlike
- **THEN** 最终 `likedByMe` MUST 为 false 且 `likeCount` MUST 反映取消后的数量
