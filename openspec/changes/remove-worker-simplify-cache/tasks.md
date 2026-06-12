## 1. 规格与全局约束

- [x] 1.1 更新 `AGENTS.md`：删除 worker/domain_outbox 强制入队表述；新增「循环后台任务须 OpenSpec 批准」约定
- [x] 1.2 更新 `openspec/project.md`：新增「背景循环任务约定」章节（与 `background-loop-task-governance` 一致）
- [x] 1.3 运行 `openspec validate remove-worker-simplify-cache --strict` 并修复校验问题

## 2. 删除 worker-service 进程与部署

- [x] 2.1 删除 `cmd/worker-service/main.go` 及 `manifest/config/config.worker-service.yaml`
- [x] 2.2 从 `manifest/deploy/kustomize/base/kustomization.yaml` 移除 worker 资源；删除或归档 `worker-deployment.yaml`、`worker-service.yaml`
- [x] 2.3 从 `manifest/docker/docker-compose.microservices.yml` 与 `.env.example` 移除 worker 服务及 `WORKER_OUTBOX_DB_LINK`
- [x] 2.4 从 history/device/voice/gateway/gateway-app deployment 移除 `WORKER_SERVICE_URL`、`OUTBOX_RELAY_ENABLED` 等 worker 相关 env
- [x] 2.5 更新 `docs/runbooks/release-deploy-and-run.md`、`dao-sync-by-domain.md`：无 worker、无 domain_outbox 入队

## 3. 删除 outbox / workeroutbox / async relay 代码

- [x] 3.1 删除 `internal/services/workeroutbox/**` 包
- [x] 3.2 删除 `internal/services/async/domain_outbox.go`；自 `cmd/worker-service` 已删后清理 `async` 包内仅 worker 使用的符号
- [x] 3.3 删除 `history/local.go` 中 `EnqueueDomainOutbox` 调用与 `isOutboxRelayEnabled`
- [x] 3.4 删除 `device/admin.go` 中 `enqueueDeviceProjectionEvent`、`enqueueDomainOutboxToHistoryRelay` 及 outbox env 分支
- [x] 3.5 删除 `history/projection_reconcile_worker.go` 及 worker 调用的 HTTP reconcile 端点（若存在且仅 worker 使用）
- [x] 3.6 删除 `history/cache_repo.go` 中 `ApplyProjection`（若无其他引用）或保留为 dead code 清理
- [x] 3.7 清理 `internal/services/contracts/http_targets.go` 中 Worker 相关 target（若无引用）

## 4. 删除 voice.task.* 脚手架

- [x] 4.1 删除 `internal/services/async/voice_task_consumer.go`、`voice_task_producer.go` 及相关类型
- [x] 4.2 删除 `voice/voice_chat.go` 中 `publishTaskRequested` 调用
- [x] 4.3 从 `eventkit/routing_keys.go` 移除或标记 deprecated `voice.task.*`（若无 UCG 外引用则删除）
- [x] 4.4 更新 `hack/rabbitmq-init.ps1` / runbook：移除 `voice.task.*` queue/binding

## 5. 加强 history 同步 patch（无 worker 兜底）

- [x] 5.1 `patchHistoryOnAdd`：列表 miss 时仍 `setLatestHistory`；`setHistoryList` 失败时 best-effort DEL list key
- [x] 5.2 `patchHistoryOnUpdate` / `patchHistoryOnDelete`：patch 失败时对称 DEL list key（必要时 latest）
- [x] 5.3 确认 voice `switchAdapter.AddHistory` remote 路径本地 patch 仍保留
- [x] 5.4 保留 `RebuildHistoryCacheByDevice` 供运维手动调用；文档说明无自动 repair ticker

## 6. device 同步缓存确认

- [x] 6.1 确认所有 event/action/user 写路径仍调用 `refreshEventOptionsCacheAfterMutate` / `setUserProfile`；删除 outbox 后无回归
- [x] 6.2 删除 device `ApplyProjection` 若仅 worker HTTP 回调使用（保留 rebuild 函数）

## 7. MQ 与配置清理

- [x] 7.1 从 rabbitmq-init 移除仅 worker fan-out 的 binding（`history.record.*`、`device.*` 等 orphan）
- [x] 7.2 清理 manifest 配置注释中的 `OUTBOX_RELAY`、`WORKER_SERVICE_URL` 说明
- [x] 7.3 确认 ucg-service MQ consumer 与 ticker **未改动**

## 8. 构建与验证

- [x] 8.1 `go build ./...` 通过
- [x] 8.2 全局 grep 无残留 `workeroutbox`、`worker-service`、`OUTBOX_RELAY_ENABLED`（除 changelog/spec 外）
- [x] 8.3 手动验证：AddHistory → GetLatestHistory / ListHistory 正确；device ListEvents 在 admin 变更后可见
