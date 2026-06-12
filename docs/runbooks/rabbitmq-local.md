## Local RabbitMQ Baseline

用于 `Task 5.1`：搭建本地消息队列基础设施并完成 topic/exchange 基线配置。

### 文件位置

- compose: `manifest/docker/docker-compose.rabbitmq.yml`
- init script: `hack/rabbitmq-init.ps1`

### 启动与初始化

1. 启动 RabbitMQ 并初始化交换机/队列绑定（**首次部署、新数据卷、或确认缺失 `voice.events` 时执行一次即可**；业务上「加历史事件」不必重做本步骤）：
  - **Windows / PowerShell**：`powershell -ExecutionPolicy Bypass -File "hack/rabbitmq-init.ps1"`
  - **Linux / macOS**（仓库根目录，需已安装 `docker` 与 `curl`）：
    - `chmod +x hack/rabbitmq-init.sh && ./hack/rabbitmq-init.sh`
    - 若 Rabbit 不在本机回环：先 `export RABBIT_API_BASE=http://<主机>:15672/api` 再执行；仅拓扑、容器已起：`SKIP_UP=1 ./hack/rabbitmq-init.sh`
2. 打开管理台：
  - [http://127.0.0.1:15672](http://127.0.0.1:15672)
  - 用户名/密码：`guest` / `guest`

### 基线拓扑

- Exchange:
  - `voice.events`（type=`topic`, durable）
- Queues（UCG 审核与推荐分）:
  - `ucg.post.created.q`（`ucg.post.created`）
  - `ucg.comment.created.q`（`ucg.comment.created`）
  - `ucg.profile.patch.submitted.q`（`ucg.profile.patch.submitted`）
  - `ucg.chat.msg.created.q`（`ucg.chat.msg.created`）
  - `ucg.recommend.score.q`（`ucg.post.published` / `unpublished` / `liked` / `unliked` / `comment.published` / `comment.removed`）

> **已移除**：`voice.task.*`、`history.events.q`、`notify.events.q` 等无 consumer 的 orphan 队列（原 worker fan-out 专用）。

### UCG 双协议说明（Publisher HTTP / Consumer AMQP）

- **Publish**（发帖/评论/资料/私信触发审核；推荐分增量事件）：`ucg-service` 经 **HTTP Management API** `MQ_HTTP_API_BASE`（默认 `:15672/api`）向 `voice.events` 发布。
- **Consume**（Green 审核 + 推荐分）：`ucg-service` 经 **AMQP** `:5672` push 订阅 UCG 队列，`autoAck=false`，处理成功后 manual Ack；可重试错误 Nack(requeue)。**Audit 与 recommend 共用单 AMQP connection**，每队列独立 channel。
- **环境变量**（ucg-service）：`RABBITMQ_AMQP_URL`（或 `RABBITMQ_HOST` + `RABBITMQ_AMQP_PORT` + `MQ_USER`/`MQ_PASSWORD`）、`UCG_AUDIT_MQ_PREFETCH`（默认 5）、`UCG_AUDIT_MQ_CONSUMER_ENABLED`、`UCG_RECOMMEND_MQ_CONSUMER_ENABLED`（默认 true）、`UCG_RECOMMEND_MQ_PREFETCH`（默认 5）。
- **审核 Publish outbox**（`ucg-audit-publish-outbox`）：业务事务内写入 `ucg_audit_publish_outbox`，`StartAuditPublishRelayWorker` 仅轮询 outbox 表重试 HTTP Publish；配置 `ucg.auditPublish.relayIntervalMs`（默认 1000）、`maxAttempts`（默认 20）。**禁止** pending 业务表 reconciler（无 `UCG_AUDIT_RECONCILE_*`）。

### UCG 审核 MQ 验收（`ucg-mq-green-audit` / `ucg-amqp-audit-consumer` / `ucg-audit-publish-outbox`）

- **Publish 可靠投递**：帖/评/资料/私信 submit 同事务入队 outbox；RabbitMQ 不可用时 API 仍 200，relay worker 恢复后补 Publish。
- **存量 pending**（无 outbox 行）：**非 runtime** 一次性 `INSERT INTO ucg_audit_publish_outbox`（读当前 `audit_version`）或手工 Publish；禁止定时扫 `ucg_post.status=1` 等补发。
- **再提审防脏写**：用户再提审使 `audit_version` 递增后，队列中旧版本消息 CAS `RowsAffected=0` 应 ACK 跳过，不得覆盖新 pending 状态。
- **部署验收清单**（手动）：
  1. 发帖 / 评论 / 改头像 / 发私信后，RabbitMQ 对应队列可见载荷含 `auditVersion`。
  2. consumer 为 AMQP push：消息到达后应 **无 2s 轮询延迟**；日志含 `[ucg-mq] shared AMQP started`。
  3. relay worker 日志含 `[ucg-audit-outbox] relay worker started`；**无** `[ucg-audit-reconcile]`。
  3. 重复消费或过期版本消息：日志记录 skip，状态不被翻转，消息 Ack 不重投。
  4. Green/DB 临时失败：消息 Nack requeue，恢复后可再次消费。
  5. 私信模式 A：收件人先见 pending，驳回后仅发送方可见且在线收件人收到 `msg_hidden`。

### UCG 推荐分 MQ 增量（`ucg-recommend-mq-incremental`）

- **禁止全表刷新**：不得再运行 `RefreshRecommendScores` / 全表 `All()` ticker；热区仅分页 reconciler 兜底。
- **一致性**：MQ consumer 低延迟近似；热区 reconciler 负责时间衰减与 throttle 漏算的最终收敛；冷区仅 MQ、无 reconciler。
- **DDL**：部署前执行 `hack/sql/ucg_recommend_mq.sql`（`ucg_recommend_hot_scan_state`）。
- **Throttle**：Redis key `ucg:recommend:throttle:{postId}`，`SET NX EX 500ms`（`ucg.recommend.likeThrottleMs`）；只限频、不保证 like/unlike 方向；跳过仍 Ack。
- **`unpublished`**：`DELETE FROM ucg_post_recommend WHERE post_id=?`；0 行正常；**永远 Ack**（DB 异常也 Ack）。
- **热区 reconciler**：轮首（`last_post_id=0`）固定 `round_hot_cutoff`，分页内禁止用 `NOW()` 重算 cutoff；无互动帖也 `RecomputeRecommendScore`。
- **验收**（手动）：
  1. 帖过审后 Feed 可见推荐分；删帖/驳回后 recommend 行删除。
  2. like 风暴：500ms 内同帖最多 1 次重算；热区 reconciler 最终收敛计数。
  3. 日志含 `[ucg-mq] shared AMQP started` 与 `[ucg-recommend-hot] started`。
  4. 无新 gateway-app App API。

### 停止与清理

- 停止：
  - `docker compose -f manifest/docker/docker-compose.rabbitmq.yml down`
- 删除数据卷（重置）：
  - `docker compose -f manifest/docker/docker-compose.rabbitmq.yml down -v`

### 验收清单

- RabbitMQ 管理台可访问
- `voice.events` exchange 创建成功
- 5 个基线队列创建成功并完成绑定（4 个 UCG 审核队列 + 1 个推荐分队列多 binding）
- 能在管理台观察到基础路由键流转

