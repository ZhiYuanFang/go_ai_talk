## MODIFIED Requirements

### Requirement: ucg internal chat send SHALL allow simulated users only

`ucg-service` MUST 提供 `POST /ucg/internal/api/chat/send`，要求有效内部网关密钥（与 device ucg internal 一致）。请求体 MUST 含 `senderWxId`、`conversationId`、`clientMsgId`，以及 `content` 或 `imageKey`/`videoKey` 之一（互斥规则与 App WS 一致）。

发送前 MUST 经 device internal 确认 `senderWxId` 对应 `is_simulated=1`。否则 MUST 返回 403。发送方 MUST 为会话成员。成功 MUST 调用 `ProcessOutboundChatMessage`（含 Green 异步审核与真人 peer push）。**消息投递成功后**，MUST 对 `senderWxId` 调用 `MarkConversationRead`，将其在该会话的 `unread_count` 置 0（T5 未读闭环）；`last_read_msg_id` SHOULD 为本次投递消息 id。

#### Scenario: Sim user sends message

- **WHEN** `senderWxId` 为 sim 且为会话成员且 content 非空
- **THEN** 消息 MUST 持久化并进入正常聊天审核流程

#### Scenario: Sim sender marked read after send

- **WHEN** internal send 成功且发送前 sim 侧 `unread_count > 0`
- **THEN** 发送完成后 sim 侧 `unread_count` MUST 为 0

#### Scenario: Real user rejected

- **WHEN** `senderWxId` 的 `is_simulated=0`
- **THEN** API MUST 返回 403 且 MUST NOT 发送消息

#### Scenario: Not a member

- **WHEN** sender 非 conversation 成员
- **THEN** API MUST 返回 404 或等价业务错误
