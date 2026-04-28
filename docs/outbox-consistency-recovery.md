## Outbox 一致性恢复验证（Task 6.4）

用于验证 `Task 6.3` 引入的 outbox 最终一致链路，在故障场景下的数据一致性与恢复能力。

### 目标

- 验证 MQ 不可用时，业务写入仍成功且 outbox 留存可追踪。
- 验证 MQ 恢复后，relay 能自动补投并收敛到 `published`。
- 验证失败事件可人工重置后再次重放。

### 前置条件

- 已执行 `manifest/sql/domain_outbox.sql` 建表。
- 已启动 RabbitMQ，并完成初始化：
  - `powershell -ExecutionPolicy Bypass -File "hack/rabbitmq-init.ps1"`
- 服务已开启：
  - `OUTBOX_RELAY_ENABLED=true`
  - `MQ_HTTP_API_BASE`、`MQ_USER`、`MQ_PASSWORD` 有效

### 场景 A：MQ 故障时一致性

1. 停止 RabbitMQ（模拟消息系统故障）：
   - `docker compose -f manifest/docker/docker-compose.rabbitmq.yml down`
2. 调用历史写接口（新增/修改/删除任一）。
3. 查看 outbox：
   - `powershell -ExecutionPolicy Bypass -File "hack/outbox-recovery-verify.ps1" -ShowPendingOnly`
4. 预期：
   - 业务数据已写入 `history` 表；
   - `domain_outbox` 中有 `pending/failed` 记录，`attempts` 递增。

### 场景 B：MQ 恢复后自动补投

1. 重启 RabbitMQ 并初始化：
   - `powershell -ExecutionPolicy Bypass -File "hack/rabbitmq-init.ps1"`
2. 保持服务运行 10~30 秒（等待 relay 轮询）。
3. 查看 outbox 状态：
   - `powershell -ExecutionPolicy Bypass -File "hack/outbox-recovery-verify.ps1"`
4. 预期：
   - 之前的 `pending/failed` 记录逐步变为 `published`。

### 场景 C：人工重放失败事件

1. 重置失败记录为待投递：
   - `powershell -ExecutionPolicy Bypass -File "hack/outbox-recovery-verify.ps1" -ResetFailedToPending`
2. 等待 relay 再次消费后复查：
   - `powershell -ExecutionPolicy Bypass -File "hack/outbox-recovery-verify.ps1" -ShowPendingOnly`
3. 预期：
   - 可重放的事件被再次处理，并最终变为 `published`。

### 验收清单

- [ ] MQ 故障时业务写入不中断
- [ ] outbox 能记录未投递事件并可观测
- [ ] MQ 恢复后未投递事件可自动补投
- [ ] 人工重放路径可用（failed -> pending -> published）
