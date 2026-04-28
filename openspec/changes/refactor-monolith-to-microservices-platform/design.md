## Context

当前系统以单个 GoFrame 服务运行，在同一进程内同时处理 HTTP API、WebSocket 语音会话、设备管理、历史查询与 AI 交互编排。这种方式提升了初期交付速度，但限制了独立扩缩容、故障隔离以及对分布式架构（如集群状态、异步流程、编排运行）的实践学习。

目标状态是接近生产形态的微服务平台：在保持现有用户侧行为的前提下，引入服务拆分、分布式数据/缓存模式、消息驱动处理和容器编排运行能力。

## Goals / Non-Goals

**Goals:**

- 将单体拆分为边界清晰、可独立部署和扩缩容的领域服务。
- 引入按域拥有的数据边界与分布式状态处理模式。
- 为高耗时/高延迟 AI 相关任务引入异步处理能力。
- 建立基于 Kubernetes 的运行能力（健康检查、滚动更新、自动扩缩容）。
- 保证迁移按阶段增量推进，并且每个阶段都可回退。

**Non-Goals:**

- 在拆分过程中重写全部内部算法。
- 从第一天起将所有请求路径都改造成完全事件驱动。
- 在跨服务场景中实现全局强一致。
- 在首轮迁移中就做到最终生产级成本/性能最优。

## Decisions

1. **服务边界按领域行为划分，而非按代码包划分**
  - 服务包括：gateway、device、history、voice-session、ai-worker。
  - 理由：更符合独立扩缩容与故障域隔离，也与当前模块职责对齐。
  - 备选方案：
    - 仅保留“模块化单体”：短期更容易，但无法达成分布式系统学习目标。
    - 按技术层拆分（controller/service/dao 服务化）：会造成高频调用耦合与所有权不清。

### 服务边界冻结（Task 1.1）

以下边界在 M1 阶段冻结，后续实现阶段不得跨边界直接访问他服务私有数据：


| 服务                    | 核心职责（In）                    | 非职责（Out）       | 数据所有权               |
| --------------------- | --------------------------- | -------------- | ------------------- |
| gateway               | 外部统一入口、鉴权、路由聚合、关联 ID 注入     | 业务规则决策、数据持久化   | 无业务主数据，仅网关配置        |
| device-service        | 设备注册、设备管理、设备状态查询/维护         | 语音流式状态机、历史聚合查询 | 设备域主数据              |
| history-service       | 对话/事件历史查询与检索接口              | 设备写操作、实时语音编排   | 历史域主数据              |
| voice-session-service | WebSocket 会话编排、实时交互状态机      | 设备主数据管理、跨域聚合报表 | 会话临时状态（Redis）与必要元数据 |
| ai-worker             | 异步 AI 任务执行（ASR/TTS/LLM 子任务） | 外部统一入口、长连接会话管理 | 任务执行状态与任务结果索引       |


冻结期接口原则：

- 同步调用：仅允许 `gateway -> domain service`，以及 `voice-session-service` 在必要时调用受控内部接口。
- 异步调用：`domain service/voice-session-service -> MQ -> ai-worker`。
- 禁止项：跨服务直接连对方数据库；跨边界共享进程内内存状态。

### API 契约与版本策略（Task 1.2）

外部 API 契约（North-South）：

- 统一由 `gateway` 对外暴露，外部协议先标准化为 `HTTP + OpenAPI`。
- URL 版本策略：`/api/v1/...`，仅在破坏性变更时升级主版本（`v2`）。
- 向后兼容原则：新增字段仅允许“可选”；不得删除/重命名既有字段；错误码保持稳定。
- WebSocket 契约：消息体必须包含 `type` 与 `code`；新增消息类型不得影响旧类型语义。

内部 API 契约（East-West）：

- 服务间同步调用先采用 `HTTP JSON`（降低学习与调试门槛），后续可演进为 gRPC。
- 每个内部接口需定义：超时、重试策略、幂等要求、降级行为。
- 内部契约版本策略：`/internal/v1/...`；兼容变更走小版本（文档更新），破坏性变更走 `v2` 并保留过渡窗口。

契约治理规则：

- 契约优先：先更新 OpenAPI/接口文档，再改实现。
- 兼容校验：每次变更必须通过契约测试（consumer/provider contract test）。
- 废弃流程：标记 deprecated -> 发布公告 -> 观测无流量后移除。

### 异步事件模型与可靠投递策略（Task 1.3）

事件 schema 标准（统一信封）：
- 必填字段：`event_id`、`event_type`、`event_version`、`producer`、`occurred_at`、`trace_id`、`payload`。
- 语义要求：
  - `event_id` 全局唯一，用于幂等去重。
  - `event_version` 用于事件演进，默认 `v1`，破坏性变更时升级主版本。
  - `payload` 内仅放业务数据，避免混入传输控制字段。
- 示例事件类型：`history.query.requested`、`voice.session.committed`、`ai.task.requested`、`ai.task.completed`、`ai.task.failed`。

topic/exchange 命名规范：
- 统一格式：`<domain>.<entity>.<action>.<version>`，例如 `voice.session.committed.v1`。
- 路由规则：
  - 领域内事件优先按 domain 分组，降低跨域误订阅。
  - 广播类事件采用 fanout/topic；命令型消息采用 direct/keyed routing。
- 命名约束：使用小写英文和点分隔，不使用语义不清缩写。

重试语义（至少一次投递）：
- 默认语义：At-least-once，消费者必须幂等。
- 重试策略：指数退避（例如 1s, 5s, 30s, 2m, 10m），设置最大重试次数。
- 不可重试错误（如参数非法）应快速失败并进入死信，不占用重试预算。
- 可重试错误（如网络抖动、下游超时）走重试链路并记录重试原因。

死信（DLQ）策略：
- 每个关键业务队列配置对应 DLQ，命名：`<queue>.dlq`。
- 触发条件：超过最大重试次数、消息反序列化失败、显式拒绝且不可重试。
- 运维要求：
  - DLQ 消息需保留原始 `event_id`、错误原因、失败时间、最近消费者标识。
  - 建立 DLQ 监控告警与人工重放流程（按 `event_id` 防止重复副作用）。

### 关联 ID、错误模型与鉴权透传规范（Task 1.4）

关联 ID 规范（Correlation ID / Trace ID）：
- 外部入口统一使用 `X-Request-Id`；若客户端未携带，由 gateway 生成并注入。
- 内部链路统一透传 `X-Request-Id` 与 `traceparent`，确保 HTTP、WebSocket、异步事件全链路可关联。
- 事件模型中 `trace_id` 必填，并与入口 `X-Request-Id` 建立映射关系。
- 日志强制字段：`request_id`、`service`、`endpoint_or_topic`、`error_code`、`latency_ms`。

统一错误模型（HTTP/WebSocket/Async）：
- HTTP 错误响应结构统一为：
  - `code`：稳定机器可读错误码（如 `AUTH_UNAUTHORIZED`、`RESOURCE_NOT_FOUND`）。
  - `message`：面向调用方的可读描述。
  - `request_id`：用于排障定位。
  - `details`：可选扩展信息（校验字段、下游原因摘要）。
- WebSocket 错误消息统一包含：`type=error`、`code`、`message`、`request_id`。
- 异步任务失败记录统一包含：`event_id`、`error_code`、`retryable`、`request_id`。
- 错误码分层：
  - `AUTH_*`（鉴权与权限）
  - `VALIDATION_*`（输入校验）
  - `BUSINESS_*`（业务规则）
  - `DEPENDENCY_*`（下游依赖）
  - `SYSTEM_*`（系统内部错误）

鉴权头透传规范：
- 外部访问由 gateway 完成鉴权，向内部透传最小必要身份头：
  - `X-User-Id` / `X-Device-No`
  - `X-Auth-Scope`
  - `X-Request-Id`
- 内部服务禁止直接信任外部原始鉴权头，必须信任 gateway 重新签发/注入的内部身份头。
- 对异步消息，生产者需将最小身份上下文写入消息元数据（例如 `actor_type`、`actor_id`、`scope`），用于审计与授权校验。

### 适配器切换层（Task 2.1）

为避免一次性拆分导致的高风险变更，先在服务调用侧引入“本地/远程可切换”适配器层：
- 入口保持不变：控制器仍通过 `service.DeviceHistory()` 获取 `DeviceHistoryContract`。
- 适配器决策：通过环境变量 `HISTORY_SERVICE_MODE` 决定调用本地实现或远程实现（默认本地）。
- 配置约定：
  - `HISTORY_SERVICE_MODE=local`（默认）：调用单体本地实现。
  - `HISTORY_SERVICE_MODE=remote`：切换到远程客户端实现（当前阶段先落接口骨架与错误提示）。
  - `HISTORY_SERVICE_URL`：远程模式下的 history 服务地址配置项。
- 迁移价值：在不改控制器签名的前提下，支持后续灰度切换与回滚。

### 路由与服务装配解耦（Task 2.2）

在 `RegisterHTTP` 层引入按领域分组的装配方式，减少 history 与 device 相关依赖交叉：
- 新增 `HTTPDeps` 作为控制器装配依赖容器，集中管理 `History/Voice/Admin` 契约实例。
- 将路由绑定拆分为 `registerHistoryRoutes`、`registerAdminRoutes`、`registerVoiceTextRoutes`，以领域边界组织绑定逻辑。
- `register` 层不再直接散落地调用多个服务构造函数，降低后续拆分服务时的修改面。
- 为后续按模块切换远程适配器（例如 history 独立服务）保留单一注入入口。

### 契约测试基线（Task 2.3）

在服务抽离前，先为 `history` 控制器建立契约测试，锁定关键外部行为：
- 参数校验契约：缺失 `deviceNo`、非法 `id` 时返回 `CodeInvalidParameter`。
- 响应结构契约：`List` 返回结构保持 `DeviceHistoryListRes{List: ...}` 形态稳定。
- 调用语义契约：控制器将规范化后的输入传递给 `DeviceHistoryContract`，并正确转发业务结果。

该测试基线用于后续 remote adapter 与 gateway 转发改造时的回归保护，确保“改调用路径不改外部行为”。

### 功能开关与回滚策略（Task 2.4）

为支持受控流量切换与快速回滚，history 适配层引入以下开关：
- `HISTORY_SERVICE_MODE`
  - `local`：全量本地调用（默认）
  - `remote`：全量远程调用
  - `canary`：按设备维度灰度远程调用
- `HISTORY_SERVICE_CANARY_PERCENT`：canary 模式下远程流量百分比（0-100）。
- `HISTORY_SERVICE_REMOTE_FAILOVER_LOCAL`：远程失败是否自动回退本地（默认开启）。

切换行为：
- canary 模式下按 `deviceNo` 的稳定哈希分流，保证同一设备流量落点稳定，便于观察与回滚。
- 远程调用失败时，若开启 failover 则自动降级到本地实现，保证业务连续性。
- 回滚操作只需修改环境变量并重启进程，不需要改代码或改路由。

### 独立 history-service 进程骨架（Task 3.1）

为首个服务抽离建立可运行的独立进程与基础运行配置：
- 新增独立启动入口：`cmd/history-service/main.go`，默认监听 `:9801`，支持 `HISTORY_SERVICE_ADDR` 覆盖。
- 新增 `RegisterHistoryServiceHTTP`，仅注册 history 相关控制器路由，作为拆分后的独立服务入口。
- 新增运行配置样例：`manifest/config/config.history-service.yaml`，用于单独部署时的 server/database 基线配置。
- 新增容器构建文件：`manifest/docker/Dockerfile.history-service`，支持独立镜像构建与运行。

### 网关 history 路由代理（Task 3.2）

在网关入口增加 history API 转发能力，支持“本地处理/代理到 history-service”可切换：
- 新增 `/device/history/api/*` 中间件代理，命中时由 gateway 反向代理到 `HISTORY_API_PROXY_URL`。
- 路由模式开关：
  - `HISTORY_API_ROUTE_MODE=local`：全量本地（默认）
  - `HISTORY_API_ROUTE_MODE=proxy`：全量代理
  - `HISTORY_API_ROUTE_MODE=canary`：按稳定哈希灰度代理
- 灰度比例：`HISTORY_API_PROXY_CANARY_PERCENT`（0-100）。
- 设计原则：默认不改现网路径；仅当配置开启时才代理，便于快速回滚到本地控制器。

### history 数据所有权边界（Task 3.3）

在 history-service 进程中，数据库配置与主服务解耦，形成独立数据所有权边界：
- history-service 启动时默认加载独立配置：`manifest/config/config.history-service.yaml`。
- 独立配置中的数据库默认链接指向 history 域库（示例：`ai_voice_history`），不再复用主服务默认库名。
- 支持环境变量 `HISTORY_DB_LINK` 覆盖数据库链接，便于不同环境按需接入独立 history 库。
- 容器镜像中设置 `GF_GCFG_FILE=manifest/config/config.history-service.yaml`，确保服务内使用独立配置。

### 分阶段流量迁移与回滚演练（Task 3.4）

为 `history` 抽离建立可执行的迁移流程与回滚路径：
- 新增脚本：`hack/history-rollout.ps1`，支持 `local -> canary10 -> canary50 -> canary100 -> proxy -> rollback` 阶段切换。
- 脚本统一维护 gateway 代理开关：
  - `HISTORY_API_ROUTE_MODE`
  - `HISTORY_API_PROXY_CANARY_PERCENT`
  - `HISTORY_API_PROXY_URL`
- 新增运行手册：`docs/history-rollout-runbook.md`，定义每阶段通过条件、观察窗口和回滚触发条件。
- 回滚路径标准化：任一阶段异常均执行 `rollback`，重启 gateway 后恢复本地处理链路。

### Redis 集群基础设施（Task 4.1）

为后续会话状态外置与多实例一致性准备分布式缓存底座：
- 新增本地集群编排：`manifest/docker/docker-compose.redis-cluster.yml`（6 节点：3 主 + 3 从）。
- 新增初始化脚本：`hack/redis-cluster-init.ps1`，可一键启动并执行 `redis-cli --cluster create`。
- 新增运行文档：`docs/redis-cluster-local.md`，包含启动、验证、清理与验收清单。
- 目标是先建立可重复执行的集群拓扑，后续任务再接入业务代码与会话状态迁移。

### voice-session 状态外置（Task 4.2）

在 `VoiceService` 中引入“Redis 优先、内存兜底”的会话读取/写入路径：
- 新增环境开关：
  - `VOICE_SESSION_BACKEND=redis` 启用 Redis 会话存储
  - `VOICE_SESSION_REDIS_PREFIX` 配置会话 key 前缀（默认 `voice:session:`）
- 会话读写链路（`buildChatMessages`、`appendChatHistory`、理解流程的“上一轮消息读取”）统一走会话访问方法，避免业务层直接操作内存 map。
- Redis 不可用时自动降级为本地内存行为，保证对话功能连续性。
- 保持默认兼容：未开启 Redis 开关时，运行行为与原单体内存会话逻辑一致。

### Redis 幂等键与限流状态（Task 4.3）

在文本对话入口加入 Redis 侧幂等与限流保护，减少重复请求与突发打爆风险：
- 新增开关：`VOICE_GUARD_REDIS_ENABLED`（开启后启用 Redis guard）。
- 新增限流配置：`VOICE_TEXT_RATE_LIMIT_PER_MINUTE`（默认每设备每分钟 30 次）。
- 新增幂等配置：`VOICE_TEXT_IDEMPOTENCY_TTL_SECONDS`（默认 3 秒去重窗口）。
- 覆盖入口：`TextChat`、`HandleTranscriptChatOnly`、`HandleTranscriptForStreaming`。
- Redis 异常时采用“放行并告警”策略，避免因缓存故障阻断核心对话链路。

### 多实例会话连续性与故障切换验证（Task 4.4）

建立可重复执行的多实例验证基线，确保 Redis 会话外置后的运行可靠性：
- 新增验证脚本：`hack/voice-session-verify.ps1`，用于对两个 gateway 实例发起同设备连续请求并检查响应可用性。
- 新增验证文档：`docs/voice-session-failover-verify.md`，定义前置条件、自动化步骤、Redis 故障演练与验收清单。
- 新增开关行为回归测试（`VOICE_SESSION_BACKEND`、`VOICE_GUARD_REDIS_ENABLED`），保证配置切换行为稳定。
- 通过“自动脚本 + 故障演练清单”形成上线前/变更后的标准验证流程。

### 消息队列基线拓扑（Task 5.1）

建立 RabbitMQ 本地基础设施与统一命名规范，作为异步链路落地前置条件：
- 新增本地编排：`manifest/docker/docker-compose.rabbitmq.yml`（RabbitMQ + 管理台）。
- 新增初始化脚本：`hack/rabbitmq-init.ps1`，自动创建 exchange/queue/binding。
- 新增运行文档：`docs/rabbitmq-local.md`，定义启动、验证、清理与验收步骤。
- 基线命名：
  - exchange: `voice.events`（topic）
  - queues: `voice.task.requested.q`、`voice.task.completed.q`、`voice.task.failed.q`、`notify.events.q`
  - routing keys: `voice.task.requested`、`voice.task.completed`、`voice.task.failed`、`notify.*`

### 长耗时任务生产者投递（Task 5.2）

在语音文本对话入口增加异步任务投递，形成“同步响应 + 异步事件”双路径：
- 新增 `voiceTaskProducer`，通过 RabbitMQ HTTP API 向 `voice.events` 发布 `voice.task.requested` 事件。
- 事件结构采用统一信封（`event_id`、`event_type`、`event_version`、`producer`、`occurred_at`、`trace_id`、`payload`）。
- 文本对话入口 `TextChat` 在主链路执行前尝试投递任务事件。
- 通过环境变量控制：
  - `MQ_PRODUCER_ENABLED`
  - `MQ_HTTP_API_BASE`
  - `MQ_USER`
  - `MQ_PASSWORD`
- 失败策略：投递失败仅记录告警，不阻断主对话链路。

### 幂等消费者与重试死信（Task 5.3）

新增最小消费者骨架，覆盖“拉取 -> 幂等 -> 重试 -> 完成/失败事件”流程：
- 消费队列：`voice.task.requested.q`（RabbitMQ HTTP API 拉取）。
- 幂等策略：使用 Redis `voice:mq:done:<event_id>` 记录已处理事件，避免重复副作用。
- 重试策略：`MQ_CONSUMER_MAX_RETRIES` 控制最多重试次数（默认 3），指数式短退避。
- 结果事件：
  - 成功投递 `voice.task.completed`
  - 最终失败投递 `voice.task.failed`（包含 reason）
- 失败不崩溃：消费失败记录告警，后台轮询继续处理后续任务。

### 异步 worker 池扩展（Task 5.4）

将单消费者轮询扩展为可配置并发 worker 池：
- 新增配置：
  - `MQ_CONSUMER_WORKERS`：worker 并发数（默认 1）
  - `MQ_CONSUMER_POLL_INTERVAL_MS`：轮询间隔（默认 2000ms，最小 100ms）
- 每个 worker 独立执行 `consumeOnce` 循环，统一共享幂等与重试/死信机制。
- 通过 worker id 打点日志，便于观察并发消费与定位异常。
- 保持兼容：未开启 `MQ_CONSUMER_ENABLED` 时不会启动任何消费者协程。

### 领域所有权拆库/schema（Task 6.1）

在主服务内先完成“逻辑分库路由”，为后续物理拆库与读写分离打基础：
- 在 `database` 下新增 `device`、`voice`、`history` 三个数据库组。
- DAO 层按表做领域路由：
  - `user/event/action` -> `device`
  - `qa/suggest` -> `voice`
  - `history` -> `history`
- 兼容策略：当目标组未配置 `link` 时自动回退 `default`，避免一次性强依赖环境完成全量分库。
- 启动期支持环境变量覆盖：
  - `DEVICE_DB_LINK` -> `GF_DATABASE_DEVICE_LINK`
  - `VOICE_DB_LINK` -> `GF_DATABASE_VOICE_LINK`
  - `HISTORY_DB_LINK` -> `GF_DATABASE_HISTORY_LINK`

### 热点读写分离与只读副本（Task 6.2）

在 history 热点查询链路先行落地读写分离：
- 新增 `history_ro` 数据库组，承接历史查询读流量（可指向只读副本）。
- `history` DAO 新增：
  - `ReadCtx`：优先 `history_ro`，不存在时回退 `history`，再回退 `default`。
  - `WriteCtx`：固定走 `history`（未配置时按既有回退到 `default`）。
- 业务侧在 `device_history` 中切换调用：
  - 列表查询使用 `ReadCtx`
  - 新增/更新/删除使用 `WriteCtx`
- 启动期新增覆盖变量：`HISTORY_RO_DB_LINK` -> `GF_DATABASE_HISTORY_RO_LINK`，便于环境按需注入只读副本连接。

### 跨服务最终一致（Task 6.3）

引入最小 outbox 流程，先覆盖 history 域跨服务事件：
- 在 history 写路径（新增/更新/删除）中，将“业务数据写入 + outbox 事件写入”放入同一个数据库事务。
- 新增 outbox 表 `domain_outbox`（见 `manifest/sql/domain_outbox.sql`），记录事件状态、尝试次数、最近错误。
- 新增后台 relay worker（可配置并发与轮询），从 `domain_outbox` 拉取 `pending/failed` 事件并投递到 `voice.events`。
- 投递成功标记 `published`，失败标记 `failed` 并累计 `attempts`，后续继续重试直到阈值。
- 关键配置：
  - `OUTBOX_RELAY_ENABLED`
  - `OUTBOX_RELAY_WORKERS`
  - `OUTBOX_RELAY_POLL_INTERVAL_MS`
  - `OUTBOX_RELAY_MAX_ATTEMPTS`
- 路由键示例：`history.record.created`、`history.record.updated`、`history.record.deleted`（已在 RabbitMQ 初始化脚本增加 `history.#` 绑定队列）。

### 故障一致性与恢复验证（Task 6.4）

围绕 outbox 最终一致链路增加故障演练与恢复手册：
- 新增验证脚本：`hack/outbox-recovery-verify.ps1`
  - 统计 `domain_outbox` 各状态数量
  - 查看 `pending/failed` 明细
  - 支持将 `failed` 批量重置为 `pending` 触发重放
- 新增运行手册：`docs/outbox-consistency-recovery.md`
  - 覆盖 MQ 故障、MQ 恢复自动补投、人工重放三类场景
  - 提供明确验收清单，验证“业务写入可用 + 最终补投收敛”。

### 容器化与健康探针基线（Task 7.1）

为 gateway、history-service、worker 建立统一容器化与探针基线：
- 新增镜像构建：
  - `Dockerfile.gateway-service`（主入口服务）
  - `Dockerfile.worker-service`（异步消费与 outbox relay）
  - `Dockerfile.history-service` 补充镜像内 `HEALTHCHECK`
- 新增本地编排清单：`docker-compose.microservices.yml`，可一键拉起 gateway/history/worker。
- 探针策略：
  - gateway: `GET /api.json`（9701）
  - history-service: `GET /api.json`（9801）
  - worker: `GET /healthz`（9901，worker 进程内置轻量 http 健康端点）
- 运行约束：gateway 默认关闭消费者，worker 负责异步消费与 relay，避免重复消费。

### Kubernetes 部署清单与环境化配置（Task 7.2）

基于 kustomize 生成 deployment/service/ingress 清单：
- `base` 下提供 gateway、history-service、worker 三套 `Deployment + Service`，并新增 `Ingress`（入口到 gateway）。
- `overlays/develop` 提供环境差异：
  - 镜像 tag 覆盖（`images` 字段，等价 values）
  - 副本数与关键环境变量 patch
  - ingress host 覆盖
- 补充使用文档：`docs/kubernetes-deploy-kustomize.md`，包含渲染、部署与参数调整说明。

### 自动扩缩容与滚动发布策略（Task 7.3）

在 kustomize develop 环境中补齐弹性与发布策略：
- Deployment 统一滚动发布配置：
  - `strategy.type=RollingUpdate`
  - `maxSurge=25%`
  - `maxUnavailable=0`
- 为 gateway/history-service/worker 增加 HPA（`autoscaling/v2`）：
  - 基于 CPU 利用率触发扩缩容（70%~75% 阈值）
  - 为不同服务分别设置 `minReplicas/maxReplicas`
- 为三类服务增加 PDB（`policy/v1`，`minAvailable: 1`），降低节点维护/驱逐时中断风险。

### 日志、指标与分布式追踪看板（Task 7.4）

建立最小可观测基线，覆盖 metrics/logs/traces 三类信号：
- Kubernetes 工作负载层：
  - 在 gateway/history-service/worker Deployment 增加 `prometheus.io/*` 抓取注解。
  - 统一注入 OTel 导出环境变量（`OTEL_SERVICE_NAME`、`OTEL_EXPORTER_OTLP_ENDPOINT`、`OTEL_EXPORTER_OTLP_INSECURE`）。
- 本地看板栈：
  - 新增 `docker-compose.observability.yml`，拉起 Prometheus、Loki、Tempo、Grafana。
  - 新增 Prometheus/Tempo 基础配置（`observability/prometheus.yml`、`observability/tempo.yaml`）。
- 文档化：
  - 新增 `docs/observability-dashboards.md`，包含启动方式、入口地址与验收清单。

### SLO、告警与故障响应/回滚手册（Task 7.5）

为生产化运维补齐目标、规则与应急流程：
- 新增告警规则清单：`overlays/develop/prometheus-rules.yaml`
  - 可用性类告警：gateway/history-service/worker down
  - 质量类告警：gateway 5xx 错误率
  - 资源类告警：worker 高 CPU
- 新增运行手册：`docs/slo-alerts-incident-runbook.md`
  - 定义首版 SLO 目标（可用性、错误率、任务成功率、补投延迟）
  - 规定告警分级与响应时限（critical/warning）
  - 给出标准 incident 处置步骤与回滚演练 checklist
  - 明确一致性恢复路径（含 outbox failed 重放）。

1. **采用混合通信模型（sync + async）**
  - 同步：gateway 到领域服务，承接用户侧请求-响应链路。
  - 异步：队列 + worker，承接长耗时 AI 处理与副作用流程。
  - 理由：既保留低延迟链路，又能解耦高耗时工作负载。
  - 备选方案：
    - 全同步：心智简单但易形成超时级联、韧性不足。
    - 全异步：会过早增加用户交互链路复杂度。
2. **先落实按域数据所有权，再考虑高级分片**
  - 初始数据模型：按服务所有权拆分 schema/database。
  - 演进方式：在有明确热点证据后，再考虑只读副本与分片。
  - 理由：先解决所有权和耦合，再处理规模扩展机制。
  - 备选方案：
    - 立即跨服务分片：首轮迁移复杂度过高，且会掩盖核心架构学习目标。
3. **将共享临时状态迁移到 Redis 集群**
  - 语音会话协同、短期缓存、幂等键与限流状态统一存入 Redis。
  - 理由：支持 WebSocket/会话服务水平扩展。
  - 备选方案：
    - 进程内内存：不支持多实例一致性。
    - 仅数据库存储状态：延迟更高且会引入不必要负载。
4. **基于 Kubernetes 部署，并配套基础可观测性**
  - 使用 Deployments、Services、Ingress、HPA、readiness/liveness 探针。
  - 通过 trace-id 传递实现日志/指标/链路关联。
  - 理由：运维成熟度是本次学习目标的一部分。
  - 备选方案：
    - 仅 Docker Compose：适合本地迭代，但不足以覆盖编排学习目标。

## Risks / Trade-offs

- 运维复杂度上升（网络、重试、版本治理） -> 缓解：分阶段迁移 + 契约测试 + 回滚开关。
- 跨服务数据一致性转为最终一致 -> 缓解：幂等处理、outbox/事件模式、补偿流程。
- WebSocket/语音链路在跨服务后可能增大延迟 -> 缓解：关键低延迟状态放 Redis，减少同步 hop 深度。
- 跨服务排障复杂度增加 -> 缓解：统一结构化日志、分布式追踪、关联 ID 规范。
- 团队学习曲线可能拖慢交付 -> 缓解：一次只迁移一个服务切片，并为每个里程碑补齐运行手册。

## Migration Plan

1. **基线与契约阶段**
  - 冻结服务边界、API 契约和事件 schema。
  - 在单体内部增加与 gateway 兼容的接口层，支持受控切换。
2. **首个服务抽离（history service）**
  - 将历史查询 API 迁移为独立服务，由 gateway 路由转发。
  - 保留可切回单体实现的回滚开关。
3. **Redis 集群接入**
  - 将会话/缓存/幂等状态从进程内内存迁移出去。
  - 验证 voice-session 多实例一致性行为。
4. **队列 + worker 接入**
  - 引入异步处理、重试与死信流程。
  - 将高耗时任务转移到 worker，同时保持业务功能一致。
5. **数据拓扑演进**
  - 按服务所有权拆库；按需增加只读副本。
  - 在必要链路引入最终一致工作流。
6. **Kubernetes 生产化**
  - 全服务容器化并部署，启用健康检查和自动扩缩容。
  - 完成 SLO 看板、告警与故障回滚演练。

回滚原则：每个阶段都保留清晰的回退路径，能返回到上一稳定路由/部署形态。

## Open Questions

- 基于本项目学习目标，应优先选 RabbitMQ 还是 Kafka？
- 内部服务调用应先统一为 HTTP+OpenAPI 还是 gRPC？
- 服务拆分后，语音实时链路可接受的延迟/错误率 SLO 应设为多少？
- 首个目标部署环境应选本地 K8s、托管 K8s，还是两者并行？

