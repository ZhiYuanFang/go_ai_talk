## MODIFIED Requirements

### Requirement: Conversation list SHALL support unread counts and pin/delete flags

API SHALL return conversations with unread_count, pinned, last message preview; member soft-delete via `deleted_at` on `ucg_conversation_member`. List ordering SHALL use `pinned DESC, updated_at DESC` from `ucg_conversation_member`.

`GET /ucg/app/api/conversations` MUST NOT 因单条会话的对方成员缺失或对方账号已注销而返回整页错误。对每条当前用户未软删的会话，响应 MUST 仍包含 `id`、`unreadCount`、`lastPreview`、`pinned`、`updatedAt`。

当 `ucg_conversation_member` 中仍存在对方成员行时，列表项 `peerWxId` MUST 为该行的历史 `wx_id`。当无法解析对方成员行时，`peerWxId` MAY 为 `0`。

当对方成员行不存在，或 device internal `ValidateWx` 表明该 `peerWxId` 对应 wx 不存在时，列表项 MUST NOT 填充 `peerNickname`、`peerAvatarKey`、`peerAvatarUrl`、`peerAvatarThumbnailUrl`（省略或空串）。发消息、创建会话等写路径 MUST 保持原有严格校验，不在本 Requirement 中放宽。

#### Scenario: 未读计数

- **WHEN** 收件人收到新消息
- **THEN** 其 `unread_count` SHALL 递增直至调用 read API

#### Scenario: 对方成员行缺失时会话仍出现在列表

- **WHEN** 用户请求 `GET /ucg/app/api/conversations` 且某 direct 会话中除本人外无其他 `ucg_conversation_member` 行
- **THEN** 该会话 MUST 仍出现在 `list` 中且接口 MUST 返回成功
- **AND** 该项 `peerWxId` MAY 为 `0`
- **AND** `peerNickname` 与 avatar 相关字段 MUST 为空

#### Scenario: 对方已注销时保留 peerWxId 且展示为空

- **WHEN** 用户请求会话列表且某会话对方 `ucg_conversation_member` 行仍存在 `wx_id=W`，但 `ValidateWx(W)` 返回 `exists=false`
- **THEN** 该项 `peerWxId` MUST 为 `W`
- **AND** `peerNickname` 与 avatar 相关字段 MUST 为空
- **AND** 接口 MUST 返回成功

#### Scenario: 发消息路径仍要求有效对方

- **WHEN** 客户端经 WebSocket 向会话发送消息且对方成员不存在
- **THEN** 系统 MUST 拒绝发送且 MUST NOT 因本变更而静默成功
