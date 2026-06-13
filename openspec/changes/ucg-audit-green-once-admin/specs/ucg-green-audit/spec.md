## MODIFIED Requirements

### Requirement: Post and profile content SHALL use Green async audit Option A

Submitted posts (text/images/video) and profile fields (avatar/nickname/bio) SHALL enter `pending_audit` visibility: ONLY author MAY see content in feeds/profile until Green pass sets `published` or profile active state; on fail content SHALL be `rejected` with reason visible to author in 我的动态 or profile edit feedback as「违规已下架」.

机审与业务落库 MUST 拆为两阶段：（1）**Moderation** — 对每个 `(entity_id, audit_version)` **至多发起一次 Green 判定**；Green 结论（pass/reject）SHALL 写入 `moderation_verdict`，Green API 错误或 persist verdict 失败 SHALL 进入机审失败终态（profile：`ProfileJobStatusModerationFailed`；post：等价 `moderation_failed`），且 MQ handler MUST Ack、MUST NOT 再次调用 Green；（2）**Apply** — 基于已持久化 verdict 执行 CAS 状态迁移。MQ 重投时若 `moderation_verdict` 已非 0 或 status 已为机审失败终态，MUST 跳过 Green。Apply 失败 MUST 有界重试；超限后 MUST 进入 apply 失败终态并停止 MQ requeue，不得因 apply 失败反复调用 Green。

评论（`ucg_post_comment`）MUST 采用与 post 相同的两阶段字段与语义（`moderation_verdict`、apply 计数等），不得再使用「单函数内 Green + 一步 CAS」且 Green err 无限 requeue 的模式。

资料机审失败终态（`ProfileJobStatusModerationFailed`）对 App 作者 MUST NOT 展示专用审核中文案或「审核中」态；作者查询 profile MUST 仍呈现已发布 `ucg_profile` 内容（未 apply 的 patch 不可见），直至人工或后续流程处理。

#### Scenario: 提交后仅作者可见

- **WHEN** 用户发布帖子且 Green 未完成
- **THEN** 其他用户请求 Feed SHALL NOT 包含该帖；作者 我的动态 SHALL 显示审核中

#### Scenario: 审核通过公开

- **WHEN** Green 返回 pass 且 Apply 阶段成功
- **THEN** post status SHALL 变为 2 且 SHALL 出现在推荐/关注 Feed

#### Scenario: 审核失败

- **WHEN** Green 返回 fail 且 Apply 阶段成功
- **THEN** post status SHALL 变为 3，作者 SHALL 见 reject_reason；其他用户 SHALL NOT 见该帖

#### Scenario: 资料 bio 机审通过后 apply 失败不重复 Green

- **WHEN** 用户修改 profile bio，Green 文本审核 PASS，但 `approveProfileJobCAS` 首次因 DB 错误失败，且 MQ 消息被 requeue
- **THEN** 后续消费 MUST NOT 再次调用 Green 审核该 bio，MUST 仅重试 apply；且 Green 调用总次数对该 `(job_id, audit_version)` MUST NOT 因 requeue 线性增长

#### Scenario: 资料 apply 超限后作者可感知失败

- **WHEN** profile job 机审 pass 但 apply 在达到 `UCG_AUDIT_APPLY_MAX_ATTEMPTS` 后仍失败
- **THEN** job MUST 进入 apply 失败终态，MUST NOT 保持无限 pending 重试，且作者查询资料审核状态 MUST 得到明确失败反馈（非永久「审核中」）

#### Scenario: Green API 错误不重复调用

- **WHEN** 资料或帖子 Phase1 已调用 Green API 且返回网络/额度/5xx 类错误，或 persist `moderation_verdict` 失败
- **THEN** 系统 MUST 进入机审失败终态并 Ack MQ 消息，后续同 `(entity_id, audit_version)` 的消费 MUST NOT 再次调用 Green API

#### Scenario: 资料机审失败作者无专用展示

- **WHEN** profile audit job 处于 `ProfileJobStatusModerationFailed`
- **THEN** 作者 GET profile MUST NOT 返回 `auditPending=true` 或 moderation_failed 专用 reject 文案，且 MUST 展示已发布 profile 字段

#### Scenario: 评论两阶段不重复 Green

- **WHEN** 评论 Phase1 已写入 `moderation_verdict=pass` 但 Phase2 publish CAS 失败并重投 MQ
- **THEN** 后续消费 MUST NOT 再次调用 Green，MUST 仅重试 apply

### Requirement: Chat messages SHALL use Green audit Option C before delivery

Chat messages SHALL be visible as pending to sender only until Green pass; on pass message MUST be delivered to recipient via WS; on fail sender MUST receive failure notification and recipient MUST NOT receive message.

私信审核 MUST 采用与 post 相同的两阶段 Moderation/Apply 模型（MySQL `ucg_chat_message` 持久化 `moderation_verdict` 等字段）。对每个 `(conversation_id, message_id, audit_version)` Green API MUST 至多发起一次；Green API 错误或 persist verdict 失败 MUST 进入机审失败终态并 Ack，MUST NOT 因 requeue 重复调 Green。

#### Scenario: 发送后收件人不可见

- **WHEN** 用户发送聊天消息且 Green 未完成
- **THEN** 收件人 WS SHALL NOT 收到该消息

#### Scenario: 审核通过后投递

- **WHEN** Green pass 且 Apply 成功
- **THEN** 收件人 SHALL 通过 WS 收到 `message_delivered` 事件

#### Scenario: 审核失败

- **WHEN** Green fail 且 Apply 成功
- **THEN** 发送方 SHALL 收到 `audit_failed` 含 reason；消息 SHALL NOT 进入收件人会话

#### Scenario: 私信 Green API 错误不重复调用

- **WHEN** 私信 Phase1 已发起 Green 且 API 返回错误
- **THEN** 系统 MUST 进入机审失败终态并 Ack，后续消费 MUST NOT 再次调用 Green
