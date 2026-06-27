## Context

- T5：`sim-unread-sample` 筛选 `sim unread_count > 0` → LLM → `POST /ucg/internal/api/chat/send`。
- `DeliverChatMessage` 仅对 **recipient** 递增 unread；sim 发送后自身 unread 不变。
- App 用户通过 `POST /conversations/{id}/read` 清零；T5 不走 App API。

## Decision

**方案 A**：在 `ucgInternalChatSend` handler 内，`ProcessOutboundChatMessage` 成功后调用 `MarkConversationRead(senderWxId, conversationId, msgId)`。

- 不修改 WS 发消息路径的已读语义。
- `lastMsgId` 使用本次投递返回的 `msg.ID`。
- mark-read 失败则 internal send 返回 500（消息已发出，T5 记失败；可观测）。

## Non-Goals

- 独立 internal mark-read API。
- T5 增加 usernameLogin + App read。
