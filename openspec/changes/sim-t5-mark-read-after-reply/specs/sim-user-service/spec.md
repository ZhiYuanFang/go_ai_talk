## MODIFIED Requirements

### Requirement: T5 chat scan SHALL poll-reply once per eligible unread without E1

（继承 `sim-t5-unread-sample` 编排。）每 tick 在 `chat/send` 成功后，sim 侧该会话 MUST 不再满足 `sim-unread-sample` 的 `unread_count > 0` 条件，直至真人 peer 再次发消息。

#### Scenario: No repeat sample after reply without new peer message

- **WHEN** T5 已成功对某 `(simWxId, conversationId)` 回复且真人未再发消息
- **THEN** 下一次 T5 tick 的 `sim-unread-sample` MUST NOT 再次命中该会话
