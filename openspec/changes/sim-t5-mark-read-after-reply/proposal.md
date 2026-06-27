## Why

T5 经 internal `chat/send` 回复真人未读后，sim 侧 `ucg_conversation_member.unread_count` 未清零，下一轮 `sim-unread-sample` 仍命中同一会话，造成重复回复（真人未新发消息时）。

## What Changes

- **ucg internal `chat/send`**：消息投递成功后 MUST 对 `senderWxId` 调用 `MarkConversationRead`（方案 A，handler 内闭环）。
- `ProcessOutboundChatMessage` 返回投递消息 ID，供 mark-read 写入 `last_read_msg_id`。
- sim-user T5 编排不变（仍单次 internal send）；**不新增** App HTTP、Redis、`*_test.go`。

## Capabilities

### Modified Capabilities

- `ucg-sim-chat-internal`：internal send 成功后 mark sender 已读。
- `sim-user-service`：T5 回复后 sim 侧 unread 须清零，下 tick 不得重复 sample 同 conv（无新真人消息时）。

## Impact

- **代码**：`internal/controller/ucg_internal_chat.go`、`internal/services/ucg/chat_service.go`、`internal/controller/ucg_chat_ws.go`。
- **部署**：ucg-service 先行；sim-user 可不变更镜像（逻辑在 ucg）。
