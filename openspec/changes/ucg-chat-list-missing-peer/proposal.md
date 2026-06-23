## Why

`GET /ucg/app/api/conversations` 在加载某条会话时，若对方成员行缺失（如账号注销后清理、历史脏数据），`loadConversationDTO` → `peerWxID` 返回「会话成员不存在」并导致**整页列表接口失败**。用户仍应看到与该会话的历史记录入口；对方展示信息缺失时不应阻断列表。

## What Changes

- 会话列表读路径：对方成员行不存在或对方 wx 已注销时，**MUST NOT** 使列表接口返回错误。
- 列表项 **MUST** 仍返回 `id`、`unreadCount`、`lastPreview`、`pinned`、`updatedAt` 等会话自身字段。
- **`peerWxId` MUST 保留历史 id**：当 `ucg_conversation_member` 中仍存在对方成员行时，响应 `peerWxId` SHALL 为该行的 `wx_id`；仅当无法解析对方成员时 `peerWxId` MAY 为 `0`。
- 对方不可展示时，`peerNickname`、`peerAvatarKey`、`peerAvatarUrl`、`peerAvatarThumbnailUrl` SHALL 为空（省略或空串）。
- 发消息、创建会话等写路径 **MUST NOT** 放宽：`peerWxID` 严格语义保持不变（无对方则仍报错）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `ucg-chat-ws`：会话列表在对方成员缺失或账号已注销时的容错与 `peerWxId` 保留规则。

## Impact

- `internal/services/ucg/chat_service.go`（`ListConversations` / `loadConversationDTO` / 可选 peer 查询 helper）
- API 字段名不变；列表错误率下降，客户端可对空 peer 展示占位 UI
- 无 DB 迁移、无 Redis 变更
