## ADDED Requirements

### Requirement: ucg internal SHALL batch query public profiles by wxIds

`ucg-service` MUST 提供 `POST /ucg/internal/api/profiles/batch`，鉴权 MUST 与现有 ucg internal API 一致（`X-Device-Gateway-Internal-Secret`）。

请求体 MUST 含 `wxIds`（int64 数组，允许空）。响应 MUST 含 `list` 数组；每项 MUST 含：

- `wxId`（uint64 或 int64）
- `nickname`（string）
- `avatarUrl`（string，可选）
- `avatarThumbnailUrl`（string，可选）

实现 MUST 复用 `GetPublicProfilesByWxIDs` 语义：无 profile 行的 wxId MUST NOT 出现在 `list` 中。MUST NOT 返回 unionid 等敏感字段。

#### Scenario: Batch returns profiles for sim users

- **WHEN** 受信调用方 POST `{ "wxIds": [1001, 1002] }` 且 1001 有 ucg_profile
- **THEN** 响应 `list` MUST 含 wxId=1001 的 nickname；1002 无 profile 时 MUST 不出现在 list

#### Scenario: Reject without secret

- **WHEN** 无内部密钥调用 profiles batch
- **THEN** HTTP MUST 为 403

#### Scenario: Empty wxIds

- **WHEN** POST `{ "wxIds": [] }`
- **THEN** 响应 `list` MUST 为空数组且 MUST NOT 500
