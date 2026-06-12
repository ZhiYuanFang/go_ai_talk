## ADDED Requirements

### Requirement: 循环后台任务 MUST 经 OpenSpec 明确批准后方可引入

系统 MUST **默认禁止**在 `internal/services/**` 业务实现中新增循环后台任务，包括但不限于：`time.NewTicker` / `time.Tick` 轮询、`for { select {} }` 常驻 goroutine 扫描 MySQL/Redis/outbox 表、定时 reconciler、pending 业务表全表/分页扫描兜底。

新增上述任务 **MUST** 在 OpenSpec **proposal 与 design** 中写明：任务名称、宿主进程、周期/触发条件、环境开关、失败语义、关闭方式，以及 **为何不能** 在请求内同步完成或使用 AMQP push consumer 替代。

#### Scenario: 未批准的后台 ticker 被拒绝合入

- **WHEN** PR 在业务包内新增 `Start*Reconciler` 或 ticker 扫表且无对应 OpenSpec 变更引用
- **THEN** 评审 MUST 要求补充已批准的变更或删除该后台任务

#### Scenario: 已批准 UCG outbox relay 仍允许

- **WHEN** 变更引用已归档或进行中的 UCG OpenSpec（如 audit publish outbox、chat persist）且在 design 中声明 ticker 语义
- **THEN** 该后台任务 MAY 存在于 `ucg-service` 进程

### Requirement: AMQP push consumer 与 ticker 扫表 MUST 区分治理

经 RabbitMQ **broker push**（`autoAck=false`）的消息 consumer **不视为** ticker 扫表任务，但 **MUST** 在 OpenSpec 变更中声明队列名、routing key 与宿主进程。HTTP Management API Pull 轮询队列 **视为** 循环后台任务，适用批准流程。

#### Scenario: UCG AMQP consumer 合规

- **WHEN** `ucg-service` 启动 AMQP push consumer 消费 `ucg.audit.*` 或 `ucg.recommend.score.q`
- **THEN** 该 consumer MUST 有 OpenSpec 依据且 MUST NOT 部署在已删除的 worker-service 内

## MODIFIED Requirements

### Requirement: 角色边界变更 MUST 伴随文档更新

当后台任务执行角色或宿主进程发生调整时，运行文档与部署说明 MUST 同步更新，以确保运维与开发对角色边界认知一致。

#### Scenario: 移除 worker-service 后文档更新

- **WHEN** 完成 worker-service 删除与缓存同步简化
- **THEN** runbook 与部署清单 MUST 说明 history/device 不再依赖 worker；gateway MUST NOT 启动业务后台任务；UCG 后台任务宿主为 ucg-service

## REMOVED Requirements

### Requirement: 后台任务 MUST 仅由 worker-service 启动

**Reason**: worker-service 已删除；domain outbox relay 与 voice task consumer 移除；后台任务改为「经批准的域内任务」而非单一 worker 进程独占。

**Migration**: 删除 worker 部署；history/device 依赖同步 patch；UCG 任务保留在 ucg-service；见 `background-loop-task-governance`。

### Requirement: 部署配置审查（业务后台任务仅 worker-service）

**Reason**: 原 Scenario 要求确认后台任务仅在 worker-service，与移除 worker 冲突。

**Migration**: 替换为「gateway 无后台任务 + 各域声明宿主进程」审查项（见 MODIFIED 文档更新 Requirement）。
