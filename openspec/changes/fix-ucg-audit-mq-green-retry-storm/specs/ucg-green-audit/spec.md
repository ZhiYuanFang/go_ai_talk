## MODIFIED Requirements

### Requirement: Post and profile content SHALL use Green async audit Option A

Submitted posts (text/images/video) and profile fields (avatar/nickname/bio) SHALL enter `pending_audit` visibility: ONLY author MAY see content in feeds/profile until Green pass sets `published` or profile active state; on fail content SHALL be `rejected` with reason visible to author in 我的动态 or profile edit feedback as「违规已下架」.

机审与业务落库 MUST 拆为两阶段：（1）**Moderation** — 对每个 `(entity_id, audit_version)` 至多完成一次 Green 判定，并将 `moderation_verdict`（pass/reject）与 reject 时的 `moderation_reason` 持久化于 MySQL；（2）**Apply** — 基于已持久化 verdict 执行 CAS 状态迁移（post `published`/`rejected`，profile job `approved`/`rejected` 及 profile 行更新）。MQ 重投时若 `moderation_verdict` 已非 0，MUST 跳过 Green 仅重试 Apply。Apply 失败 MUST 使用有界重试；超限后 MUST 进入 apply 失败终态并停止 MQ requeue，不得因 apply 失败反复调用 Green。

#### Scenario: 提交后仅作者可见

- **WHEN** 用户发布帖子且 Green 未完成
- **THEN** 其他用户请求 Feed SHALL NOT 包含该帖；作者 我的动态 SHALL 显示审核中

#### Scenario: 审核通过公开

- **WHEN** Green 返回 pass 且 Apply 阶段成功
- **THEN** post status SHALL 变为 2 且 SHALL 出现在推荐/关注 Feed

#### Scenario: 审核失败

- **WHEN** Green 返回 fail
- **THEN** post status SHALL 变为 3，作者 SHALL 见 reject_reason；其他用户 SHALL NOT 见该帖

#### Scenario: 资料 bio 机审通过后 apply 失败不重复 Green

- **WHEN** 用户修改 profile bio，Green 文本审核 PASS，但 `approveProfileJobCAS` 首次因 DB 错误失败，且 MQ 消息被 requeue
- **THEN** 后续消费 MUST NOT 再次调用 Green 审核该 bio，MUST 仅重试 apply；且 Green 调用总次数对该 `(job_id, audit_version)` MUST NOT 因 requeue 线性增长

#### Scenario: 资料 apply 超限后作者可感知失败

- **WHEN** profile job 机审 pass 但 apply 在达到 `UCG_AUDIT_APPLY_MAX_ATTEMPTS` 后仍失败
- **THEN** job MUST 进入 apply 失败终态，MUST NOT 保持无限 pending 重试，且作者查询资料审核状态 MUST 得到明确失败反馈（非永久「审核中」）
