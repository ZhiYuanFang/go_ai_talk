## MODIFIED Requirements

### Requirement: Outbound chat messages SHALL be processed via WebSocket handler

`ucg-service` WebSocket handler MUST 在收到 `type=message` 后：校验成员关系；向发送方发送 `message_ack`；**立即**写入 Redis 消息（`audit_status=pending`，`audit_version` 与 `ucg_chat_message` 权威列一致）、增加收件人未读、向收件方发送 `message_delivered`；Publish `ucg.chat.msg.created`（载荷含 `auditVersion`）。**MUST NOT** 在 handler 内同步调用 Green 阻塞投递。

#### Scenario: 发送方先收到 ack

- **WHEN** 客户端发送合法 message 帧
- **THEN** 发送方 MUST 先收到 `message_ack` 再收到异步审核结果

#### Scenario: 收件方先于审核收到消息

- **WHEN** 消息已写入 Redis
- **THEN** 收件方 MUST 收到 `message_delivered`，且 body MUST 含 pending 审态标识与 `audit_version`

## ADDED Requirements

### Requirement: WebSocket SHALL emit post-audit rejection events

Green 事后驳回时，系统 MUST：

- 向发送方 MUST 推送 `audit_failed`（含 `clientMsgId`、`reason`）
- 向在线收件方 MUST 推送 `msg_hidden`（含 `conversationId`、`messageId`）

仅当 CAS（`audit_status='pending' AND audit_version=?`）成功将消息置 rejected 后 MUST 推送上述事件；CAS 0 行（过期/重复消息）MUST NOT 推送。

#### Scenario: 在线收件人隐藏违规消息

- **WHEN** 消息 CAS 为 rejected 且收件方 WS 在线
- **THEN** 系统 MUST 向收件方推送 `msg_hidden`

#### Scenario: 重复 reject 消息不重复推送

- **WHEN** 同一 `auditVersion` 的 reject 事件重复消费且 CAS 影响 0 行
- **THEN** 系统 MUST NOT 再次推送 `audit_failed` 或 `msg_hidden`
