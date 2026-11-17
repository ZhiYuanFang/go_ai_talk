## ADDED Requirements

### Requirement: App SHALL allow authors to soft-delete their own ideas

`POST /device/app/api/predict-pro-ideas/delete` MUST accept `{ "id": <ideaId> }` with a valid JWT. The server MUST soft-delete the idea only when it is non-deleted (`deleted_at IS NULL OR deleted_at = 0`) and the row’s `wx_id` equals the authenticated caller. Soft-delete MUST set `deleted_at` to a positive unix timestamp (same semantics as Admin soft-delete). After success, subsequent App `list` MUST NOT include that idea. Requests for another user’s idea, a missing id, or an already deleted idea MUST fail with a business error and MUST NOT clear `deleted_at` on unrelated rows. The endpoint MUST NOT be auth-exempt and MUST NOT require Pro entitlement. App **必须** 仅允许作者软删本人未删想法；**不得** 匿名；**不得** 以 Pro 权益门禁。

#### Scenario: Author deletes own idea

- **WHEN** 已登录用户对本人未删想法调用 delete
- **THEN** 系统 MUST 将该行 `deleted_at` 写为 >0
- **AND** 随后 list MUST 不再返回该 id

#### Scenario: Non-author rejected

- **WHEN** 已登录用户对他人未删想法调用 delete
- **THEN** 系统 MUST 返回业务错误且该行 MUST 保持未删

#### Scenario: Missing or already deleted

- **WHEN** id 不存在或 `deleted_at > 0`
- **THEN** 系统 MUST 返回业务错误且 MUST NOT 冒充成功
