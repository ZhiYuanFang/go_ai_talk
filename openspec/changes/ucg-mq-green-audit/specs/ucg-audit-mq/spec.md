## ADDED Requirements

### Requirement: UCG 审核事件 SHALL 使用已注册的 RabbitMQ routing keys

`ucg-service` 发布 UCG 审核事件时 MUST 使用下列 routing key 之一，且 MUST 经 `eventkit` 注册校验通过：

- `ucg.post.created`
- `ucg.comment.created`
- `ucg.profile.patch.submitted`
- `ucg.chat.msg.created`

#### Scenario: 发帖后发布事件

- **WHEN** 用户 submit 创建帖子且数据库事务提交成功
- **THEN** 系统 MUST Publish `ucg.post.created`，载荷至少含 `postId` 与 `auditVersion`（等于 `ucg_post.audit_version` 当前值）

#### Scenario: 未注册 routing key 拒绝发布

- **WHEN** 代码尝试 Publish 未在 eventkit 注册的 ucg 路由键
- **THEN** Publish MUST 失败且 MUST NOT 静默丢弃

### Requirement: 所有 UCG 审核 MQ 载荷 MUST 携带 auditVersion

四类审核事件的 JSON 载荷 MUST 含非空 `auditVersion`（INT），且 MUST 等于 Publish 时刻对应权威表列的当前值：

- 帖子：`ucg_post.audit_version`
- 评论：`ucg_post_comment.audit_version`
- 资料：`ucg_profile_audit_job.audit_version`（载荷 MUST 含 `jobId`）
- 私信：`ucg_chat_message.audit_version`（载荷 MUST 含 `messageId` 与 `conversationId`）

reconciler 补发 MUST 从数据库读取当前 `audit_version`，MUST NOT 硬编码或从 Redis 推断版本。

#### Scenario: 资料审核载荷含 job 版本

- **WHEN** 用户提交资料变更且 job 行 `audit_version=2`
- **THEN** Publish `ucg.profile.patch.submitted` MUST 含 `jobId` 与 `auditVersion=2`

#### Scenario: 评论审核载荷含版本

- **WHEN** 用户发表评论且评论行 `audit_version=1`
- **THEN** Publish `ucg.comment.created` MUST 含 `commentId` 与 `auditVersion=1`

### Requirement: RabbitMQ 拓扑 SHALL 为 UCG 审核队列绑定 topic exchange

仓库 `hack/rabbitmq-init` 脚本 MUST 为上述四个 routing key 创建 durable 队列并完成与 `voice.events` topic exchange 的 binding。runbook MUST 文档化队列名与初始化步骤。

#### Scenario: 本地/测试环境 init 后队列存在

- **WHEN** 运维执行 rabbitmq init 脚本
- **THEN** 管理台 SHALL 可见 `ucg.post.created.q` 等队列且 binding 正确

### Requirement: UCG 审核消费者 SHALL 驻留 ucg-service 并按 routing key 分发

`ucg-service` MUST 启动审核 MQ consumer（HTTP Pull 或项目等价实现），从 UCG 审核队列拉取消息并调用 Green 审核逻辑。consumer MUST NOT 部署在 worker-service 内直连 ucg 数据库。消费者 MUST 从载荷读取 `auditVersion` 用于 CAS。

#### Scenario: 消费帖子审核消息

- **WHEN** 队列收到合法 `ucg.post.created` 载荷（含 `postId`、`auditVersion`）
- **THEN** consumer MUST 执行 Green 审核并在成功后 CAS 更新帖子状态（条件含 `status` 与 `audit_version`）

### Requirement: 审核消费者 MUST 通过 audit_version CAS 更新且过期消息 SHALL 优雅跳过

所有 MQ 审核消费者在写审态时 MUST 使用条件更新：`WHERE id=? AND status=? AND audit_version=?`（私信为 `audit_status` + `audit_version`）。CAS 成功时 MUST NOT 递增 `audit_version`。`RowsAffected=0` MUST 视为重复投递或过期版本（如用户已再提审 bump 版本），MUST ACK 且 MUST NOT 无限重试，MUST NOT 覆盖新轮次状态。

#### Scenario: 重复消费同一 post 事件

- **WHEN** 同一 `postId` 与 `auditVersion` 的事件被投递两次且首次已 CAS 成功（status 已非 pending）
- **THEN** 第二次 CAS MUST 影响 0 行且 MUST NOT 将已 published 帖改回 pending

#### Scenario: 再提审后旧版本消息过期

- **WHEN** 用户再提审使 `ucg_post.audit_version` 从 2 递增为 3，队列中仍存在 `auditVersion=2` 的消息
- **THEN** consumer CAS `status=1 AND audit_version=2` MUST 影响 0 行且 MUST ACK，当前帖 MUST 保持 version=3 的 pending 状态

#### Scenario: 资料 CAS 仅针对 job 表版本

- **WHEN** consumer 处理 `ucg.profile.patch.submitted` 且载荷 `auditVersion=1`
- **THEN** UPDATE MUST 针对 `ucg_profile_audit_job` 且 MUST 使用 `status=1 AND audit_version=1`；MUST NOT 以 Redis 或 `ucg_profile` 列作为 CAS 版本

### Requirement: MQ Publish 失败 SHALL 可补发且 MUST NOT 阻塞用户主路径（聊天除外可选）

帖/评/资料 submit API 在事务成功但 Publish 失败时 MUST 记录 warning 日志；系统 MUST 提供 reconciler 对长时间 `pending` 且未审条目补发 MQ（补发载荷 MUST 含库内当前 `auditVersion`）。聊天 WS 在 Redis 投递成功后 Publish 失败时 SHOULD 同样可 reconciler 补发。

#### Scenario: 发帖 Publish 失败仍返回成功

- **WHEN** 帖子已写入 `status=1` 但 RabbitMQ 暂不可用
- **THEN** HTTP 创建接口 MAY 仍返回 200 与帖子 DTO；reconciler MUST 可在 MQ 恢复后补审（带当前 `audit_version`）
