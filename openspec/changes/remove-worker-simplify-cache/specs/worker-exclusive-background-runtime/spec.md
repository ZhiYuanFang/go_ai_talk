## REMOVED Requirements

### Requirement: 后台任务 MUST 仅由 worker-service 启动

**Reason**: worker-service 进程删除；无 domain outbox relay 与 voice task consumer。

**Migration**: 见 `background-loop-task-governance` 与 `history-device-sync-cache-projection`。

### Requirement: 后台任务启动语义 MUST 保持幂等

**Reason**: worker 启动入口删除；幂等要求适用于未来经批准的域内后台任务，由各自 OpenSpec 规定。

**Migration**: 新任务在 proposal/design 中声明 `sync.Once` 或等价幂等启动。
