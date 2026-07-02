## MODIFIED Requirements

### Requirement: ucg internal SHALL batch query public profiles by wxIds

`ucg-service` MUST 提供 `POST /ucg/internal/api/profiles/batch`，鉴权 MUST 与现有 ucg internal API 一致（`X-Device-Gateway-Internal-Secret`）。

请求体 MUST 含 `wxIds`（int64 数组，允许空）。响应 MUST 含 `list` 数组；对请求体中**每个有效 wxId（>0）** MUST 在 `list` 中返回**恰好一条**记录。每项 MUST 含：

- `wxId`（uint64 或 int64）
- `nickname`（string，展示用昵称）
- `avatarUrl`（string，可选）
- `avatarThumbnailUrl`（string，可选）

当 wxId 存在 `ucg_profile` 行时，`nickname` MUST 与 `GetPublicProfilesByWxIDs` 语义一致（含默认昵称刷新逻辑）。当 wxId **不存在** `ucg_profile` 行时，实现 MUST 经 device 契约取得 `babyName` 并 MUST 使用与 App 一致的推导规则：`babyName` 为空时 `nickname` 为「家长」，否则为 `{babyName}的家长`；**MUST NOT** 为此路径写入 `ucg_profile`。MUST NOT 返回 unionid 等敏感字段。

#### Scenario: Batch returns profile nickname for existing profile

- **WHEN** 受信调用方 POST `{ "wxIds": [1001] }` 且 1001 有 ucg_profile
- **THEN** 响应 `list` MUST 含 wxId=1001 的 nickname

#### Scenario: Batch synthesizes nickname without profile row

- **WHEN** 受信调用方 POST `{ "wxIds": [2002] }` 且 2002 无 ucg_profile 但 device 返回 babyName=`"小宝"`
- **THEN** 响应 `list` MUST 含 wxId=2002 且 nickname=`"小宝的家长"`

#### Scenario: Batch synthesizes default parent nickname

- **WHEN** 受信调用方 POST `{ "wxIds": [2003] }` 且 2003 无 ucg_profile 且 babyName 为空
- **THEN** 响应 `list` MUST 含 wxId=2003 且 nickname=`"家长"`

#### Scenario: Reject without secret

- **WHEN** 无内部密钥调用 profiles batch
- **THEN** HTTP MUST 为 403

#### Scenario: Empty wxIds

- **WHEN** POST `{ "wxIds": [] }`
- **THEN** 响应 `list` MUST 为空数组且 MUST NOT 500
