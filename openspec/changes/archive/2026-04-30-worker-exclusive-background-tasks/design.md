## Context

当前代码中 `gateway-service` 启动路径（`internal/cmd/cmd.go`）与 `worker-service` 启动路径（`cmd/worker-service/main.go`）都会调用 `service.StartBackgroundWorkers(ctx)`。这会让网关进程也启动 MQ 消费与 outbox relay，导致网关接触业务异步任务，与“gateway 仅承担流量与策略层”的目标冲突。

该变更需要把后台任务职责明确收敛到 worker，保持 gateway 无状态、无业务 worker 副作用，并确保现有异步链路（voice task consumer + outbox relay）可继续稳定运行。

## Goals / Non-Goals

**Goals:**
- 仅 `worker-service` 启动业务后台任务（voice task consumer / outbox relay）。
- `gateway-service` 启动流程不再触发 `StartBackgroundWorkers`。
- 用明确约束防止后续在 gateway 回归启动业务 worker。
- 保持 worker 任务运行能力、健康探针与依赖检查语义不变。

**Non-Goals:**
- 不重构后台任务内部业务逻辑（重试、幂等、发布语义维持现状）。
- 不调整 MQ/outbox 数据模型与路由键设计。
- 不在本次变更中引入新的任务调度框架。

## Decisions

1. **按进程角色区分后台任务启动权**
   - 决策：仅在 `worker-service` 入口保留 `StartBackgroundWorkers` 调用，gateway 入口移除调用。
   - 理由：最直接满足“网关不接触业务”的边界目标。
   - 备选：保留调用并依赖 `MQ_CONSUMER_ENABLED/OUTBOX_RELAY_ENABLED` 全部关闭。
   - 不采纳原因：入口职责依然混杂，后续配置误改会导致回归。

2. **后台任务入口保留统一函数**
   - 决策：继续使用 `service.StartBackgroundWorkers` 作为 worker 内部统一入口，不拆成多个外部入口函数。
   - 理由：减少改动面，避免破坏已验证的启动顺序与 once 语义。
   - 备选：将 consumer/relay 分别暴露公开入口并在 main 手工编排。
   - 不采纳原因：增加外部耦合点，短期收益不足。

3. **配置与文档双重约束**
   - 决策：在部署清单与文档中明确 gateway 不应承担业务后台任务，worker 为唯一执行者。
   - 理由：避免“代码改了但运维认知未改”导致灰度阶段误配。
   - 备选：仅改代码，不更新文档。
   - 不采纳原因：长期会产生角色漂移与故障排查歧义。

## Risks / Trade-offs

- **[worker 单点承担全部异步任务]** → Mitigation：保留多副本部署能力，结合队列消费幂等策略与健康探针扩容。
- **[迁移窗口期任务处理短暂停顿]** → Mitigation：采用滚动发布，先拉起新 worker 再下线旧 gateway 任务路径。
- **[历史环境仍依赖 gateway 启动任务]** → Mitigation：发布前在各环境执行一次“worker 在跑、gateway 不跑”的显式验证清单。

## Migration Plan

1. 从 gateway 启动路径移除 `StartBackgroundWorkers` 调用。
2. 校验 worker 启动路径保持后台任务启动与健康探针正常。
3. 更新 compose/kustomize 与运行文档，固化服务角色边界。
4. 执行验证：确认任务仅由 worker 消费，gateway 不产生 worker 日志。

回滚策略：
- 快速回滚到上一版本镜像（恢复 gateway 启动背景任务能力）；
- 或临时切换到兼容版本，待 worker 稳定后再次收敛。

## Open Questions

- 是否需要在运行时增加“非 worker 进程禁止启动后台任务”的保护日志或 panic 防护？
- 是否需要新增监控维度区分“任务由哪个服务实例消费”，用于长期边界审计？
