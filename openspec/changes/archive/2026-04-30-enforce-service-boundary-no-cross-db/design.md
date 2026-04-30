## Context

当前仓库已拆分出 `voice-service`、`history-service`、`device-service`，网关侧也具备按域代理能力；但 `voice-service` 内仍存在直接访问 `dao.History`、`dao.User` 等他域表的实现路径，导致“服务进程拆分了、数据边界未拆分”。同时 `DeviceHistory` 远程适配器尚未实现，`remote` 模式会返回 `not implemented yet`，设备画像读取也未完全通过跨服务契约完成，使服务边界治理无法真正落地。

该变更需要在不破坏现有功能可用性的前提下，把 voice 对 history 与 device/profile 领域的读取与写入逐步迁移到 HTTP 契约，明确迁移期兜底策略，并最终收敛为“跨服务走接口，不跨库直查”。

## Goals / Non-Goals

**Goals:**
- 为 voice -> history 建立稳定的内部 HTTP 契约（查询历史、查询/保存生日、事件列表、增删改历史等）。
- 为 voice -> device 建立稳定的内部 HTTP 契约（设备信息、用户画像、注册状态等）。
- 实现 `historyRemoteClient`，让 `HISTORY_SERVICE_MODE=remote|canary` 可真实工作。
- 将 `voice-service` 中直接依赖 `dao.History` 的路径改为通过 `DeviceHistoryContract` 调用。
- 保留迁移期可控回退（`HISTORY_SERVICE_REMOTE_FAILOVER_LOCAL`），避免一次切换导致核心流程中断。
- 统一错误语义：远程不可达、超时、业务错误在 voice 侧可观测且可判定。

**Non-Goals:**
- 不在本次变更中重构大模型业务提示词与对话策略。
- 不引入新的消息中间件模式替代当前同步查询（该场景仍以同步接口为主）。
- 不改造 history/device 领域数据库模型与表结构。

## Decisions

1. **以 `DeviceHistoryContract` 为唯一访问入口**
   - 决策：voice 仅通过 `DeviceHistory()` 返回的契约访问 history 数据，不允许在 voice 业务代码中直接使用 `dao.History`。
   - 理由：最小化业务代码改动面，先收敛访问入口再推进具体调用方式迁移。
   - 备选：直接在 voice 代码内散点替换 HTTP 调用。
   - 不采纳原因：会造成调用细节扩散，后续维护与灰度控制困难。

2. **优先补齐 `historyRemoteClient`，保留迁移期 failover**
   - 决策：实现远程客户端完整方法；若开启 `HISTORY_SERVICE_REMOTE_FAILOVER_LOCAL`，远程失败时允许回落 local。
   - 理由：既满足边界治理目标，又给迁移期保留可控稳定性缓冲。
   - 备选：强制 remote 失败即整体失败。
   - 不采纳原因：切换初期可用性风险过高，不利于分阶段落地。

3. **继续采用 `local|remote|canary` 模式切换**
   - 决策：沿用已有模式与分流键策略（deviceNo 稳定分流），避免引入新的配置模型。
   - 理由：与现有网关/服务治理方式一致，运维心智成本低。
   - 备选：移除 canary，仅保留 local/remote 二态。
   - 不采纳原因：失去渐进放量能力，回滚颗粒度过粗。

4. **同步查询场景不改为 MQ**
   - 决策：查询历史记录用于当前对话回复，仍通过同步 HTTP 获取。
   - 理由：该场景是请求-响应闭环，MQ 异步模型不适合承担即时读取职责。
   - 备选：voice 发 MQ 请求 history 再异步回填。
   - 不采纳原因：增加链路复杂度，且无法满足用户实时响应诉求。

5. **设备画像与历史查询统一走契约层**
   - 决策：voice 获取生日、性别、设备档案等数据时统一通过契约层调用，不再直接触达 `dao.User` 或 device/history 表。
   - 理由：避免“历史改了边界，设备画像仍跨库直查”的半收敛状态。
   - 备选：先仅治理 history，设备画像后续再处理。
   - 不采纳原因：会留下明显架构洞，后续再收敛成本更高。

## Risks / Trade-offs

- **[远程调用延迟高于本地查库]** → Mitigation：设置超时、连接复用与必要缓存，优先保障核心查询接口响应。
- **[迁移期双路径并存导致行为不一致]** → Mitigation：统一通过契约层切换并记录命中路径日志，减少散点分叉。
- **[failover 掩盖远程故障]** → Mitigation：保留显式告警与指标，failover 命中需记录 warning 并统计占比。
- **[跨服务契约变更导致兼容问题]** → Mitigation：契约文档版本化，history 先向后兼容再切流量。
- **[设备画像接口与历史接口演进节奏不一致]** → Mitigation：按域定义契约版本，允许独立灰度但统一错误码规范。

## Migration Plan

1. 定义并补齐 history 与 device 内部 HTTP 契约与 handler（查询/写入能力）。
2. 实现 `historyRemoteClient` 与设备画像远程访问客户端全量方法，接入统一错误解析与超时控制。
3. 将 voice 里 `dao.History`、`dao.User` 直接访问路径替换为契约层调用。
4. 在开发环境按 `local -> canary -> remote` 逐步验证 history/device 查询一致性与性能。
5. 生产先开小流量 canary，观察失败率与延迟后再全量 remote。

回滚策略：
- 快速切回 `HISTORY_SERVICE_MODE=local`；
- 保留 remote 客户端代码，但停止流量命中，待问题定位后再重启 canary。

## Open Questions

- history 内部接口是否需要独立认证头（仅内网白名单）以限制误调用？
- voice 查询历史的默认时间窗口与分页参数是否需要标准化为统一契约字段？
- failover 兜底策略在生产是否长期保留，还是仅限迁移窗口期？
