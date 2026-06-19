## ADDED Requirements

### Requirement: ucg internal chat send SHALL allow simulated users only

`ucg-service` MUST 提供 `POST /ucg/internal/api/chat/send`，要求有效内部网关密钥（与 device ucg internal 一致）。请求体 MUST 含 `senderWxId`、`conversationId`、`clientMsgId`，以及 `content` 或 `imageKey`/`videoKey` 之一（互斥规则与 App WS 一致）。

发送前 MUST 经 device internal 确认 `senderWxId` 对应 `is_simulated=1`。否则 MUST 返回 403。发送方 MUST 为会话成员。成功 MUST 调用 `ProcessOutboundChatMessage`（含 Green 异步审核与真人 peer push）。

#### Scenario: Sim user sends message

- **WHEN** `senderWxId` 为 sim 且为会话成员且 content 非空
- **THEN** 消息 MUST 持久化并进入正常聊天审核流程

#### Scenario: Real user rejected

- **WHEN** `senderWxId` 的 `is_simulated=0`
- **THEN** API MUST 返回 403 且 MUST NOT 发送消息

#### Scenario: Not a member

- **WHEN** sender 非 conversation 成员
- **THEN** API MUST 返回 404 或等价业务错误

### Requirement: sim-user-service MUST NOT use WebSocket for outbound chat

模拟用户出站聊天 MUST 经上述 internal HTTP 契约；sim-user-service MUST NOT 建立 `/ucg/app/ws/chat` 长连接。

#### Scenario: No WS from sim service

- **WHEN** sim 用户回复私聊
- **THEN** 实现 MUST 使用 `POST /ucg/internal/api/chat/send` 而非 WebSocket
