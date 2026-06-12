## Why

`worker-service` 的 domain outbox relay 与 history/device **写路径同步 patch Redis** 高度重叠：用户感知路径（喂养反馈、App Pub/Sub、读 miss 回源 MySQL）不依赖 worker。保留 worker 增加独立进程、专用库 `ai_voice_worker`、`WORKER_SERVICE_URL` 与 ticker 轮询复杂度，且 v2.0.3「后台任务 MUST 仅由 worker-service 启动」与 UCG 已在域内自管 outbox/MQ consumer 的现状矛盾。在业务可接受 history 列表 **cache hit 旧数据最长约 60s**（latest 约 30s）的前提下，应 **彻底移除 worker**，改为 **同步投影 + 读路径自愈**，并建立 **「未经开发者同意禁止新增循环后台任务」** 的全局约束。

## What Changes

- **BREAKING**：删除 `worker-service` 进程、`cmd/worker-service`、`manifest/config/config.worker-service.yaml`、K8s/Compose worker 部署与 `WORKER_SERVICE_URL` / `WORKER_OUTBOX_DB_LINK` / `OUTBOX_RELAY_ENABLED` 全链路。
- **BREAKING**：删除 `domain_outbox` relay、`internal/services/workeroutbox` 包、`internal/services/async/domain_outbox.go`、`voice.task.*` 生产/消费脚手架（含 `voice.task.requested` best-effort 发布，consumer 为 stub）。
- history/device 写路径 **仅保留同步 Redis 投影**（`patchHistoryOnAdd` 等 + device `refresh*CacheAfterMutate`）；**加强**冷缓存与 patch 失败语义（design 详述）。
- 删除 history/device 中所有 `EnqueueDomainOutbox` 调用与 outbox 相关 env 分支。
- 删除 RabbitMQ 上 **仅服务于 worker fan-out** 的 binding/队列（`history.record.*`、`device.*` fan-out、`voice.task.*`），**保留 UCG 审核/推荐 MQ**。
- 新增全局约定：**禁止**在 `internal/services/**` 未经 OpenSpec 批准新增 ticker/扫表类循环后台任务（UCG 已批准项除外）。
- 更新 `AGENTS.md`、`openspec/project.md`、runbook、部署清单；废止 v2.0.3 中 worker 独占后台任务与 worker 强依赖 RabbitMQ 启动等 Requirement。

## Capabilities

### New Capabilities

- `background-loop-task-governance`：循环后台任务（ticker 轮询、扫 outbox/业务表 reconciler 等）的批准流程、默认禁止与评审检查项。
- `history-device-sync-cache-projection`：history/device 读模型在 **无 worker** 下的同步 patch、TTL/miss 回源与可接受的一致性窗口。

### Modified Capabilities

- `async-cache-projection-sync`：由「写后异步 domain_outbox + worker 投影」改为「写后同步 patch + 读 miss 重建」。
- `worker-exclusive-background-runtime`：废止 worker 独占后台任务 Requirement（worker 进程删除）。
- `gateway-no-business-workers`：更新部署审查场景，不再要求「业务后台任务仅 worker-service」。
- `cache-and-messaging-hard-dependencies`：移除 worker-service 启动 MUST 失败于 RabbitMQ 不可达的 Requirement。
- `voice-textchat-resilience`：移除 `voice.task.requested` 发布相关 Requirement（该链路删除）。

## Impact

- **代码**：`cmd/worker-service`、`internal/services/workeroutbox`、`internal/services/async/domain_outbox.go`、`voice_task_consumer.go`、`voice_task_producer.go`；`history/local.go`、`device/admin.go`、`history/projection_reconcile_worker.go`（若仅 worker 用）；`manifest/deploy/**` worker 资源。
- **配置/环境**：移除 `OUTBOX_RELAY_ENABLED`、`WORKER_SERVICE_URL`、`WORKER_OUTBOX_DB_LINK`、`CACHE_PROJECTION_REPAIR_*`、`MQ_CONSUMER_*`（worker 侧）；history/device/voice/gateway deployment env 清理。
- **数据库**：`ai_voice_worker.domain_outbox` 可弃用（文档标注，不强制删表）。
- **MQ**：`hack/rabbitmq-init` 收 worker 专用 queue/binding；**不影响** UCG 队列。
- **Non-Goals（本变更不做）**：UCG 已有 ticker/outbox/MQ consumer；Feed Redis ZSET；新增测试文件。
