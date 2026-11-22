# platform-runtime

## Purpose

Go 微服务运行时、Redis/Rabbit/cachekit、配置边界、包隔离、DB-per-service、导入门禁与进程布局的行为规格。 Owns bootstrap and cross-service platform contracts.

<!-- migrated-from: async-cache-projection-sync, cache-and-messaging-hard-dependencies, cachekit-zrevrange-parse, cash-service-runtime, controller-process-layout, dao-extension-layer-simplification, database-unix-timestamp-storage, device-route-canary-management, domain-package-boundary-enforcement, enum-adapter-compatibility, gdb-empty-query-convention, main-config-boundary-pruning, main-config-without-database, microservice-boundary-final-alignment, push-service-runtime, redis-disaster-recovery-runbook, redis-platform-access, redis-read-model-cache, retire-sim-user-surface, routing-key-governance, routing-key-governance-workflow, routing-key-prefix-registry, runtime-dependency-check, runtime-docs-centralization-and-governance, service-boundary-no-cross-db, service-code-full-cutover, service-dedicated-config-loading, service-import-isolation, service-migration-safety-and-rollback, service-outbound-clients, service-runtime-standardization, sim-runtime-config, sim-user-service, single-default-db-per-service, typed-domain-enums, ucg-feed-redis-store, ucg-service-runtime, validated-prefix-dispatch, worker-dedicated-config-loading, worker-exclusive-background-runtime -->

## Requirements

### Requirement: 缓存投影 MUST 由写路径同步 patch 与读 miss 重建承担

`history`/`device` 读模型缓存更新 MUST 在写库成功后的同请求内同步 patch（见 `history-device-sync-cache-projection`），读路径 cache miss 时 MUST 回源 MySQL 并回填。系统 MUST NOT 依赖 worker-service、domain_outbox relay 或异步投影 consumer 作为唯一更新路径。

#### Scenario: 写后同步 patch 替代 outbox relay

- **WHEN** history 记录新增且事务提交成功
- **THEN** 系统 MUST 在同请求内 patch Redis 读模型，MUST NOT 仅 enqueue domain_outbox 等待 worker

#### Scenario: 无 worker 时读路径仍正确

- **WHEN** Redis list key 不存在且客户端请求 history 列表
- **THEN** 系统 MUST 从 MySQL 重建列表并回填 Redis

---

### Requirement: Redis 必须作为唯一缓存后端

系统 SHALL 将 Redis 作为 voice、device、history 相关缓存状态的唯一后端，并且 SHALL NOT 在请求处理路径执行内存缓存兜底逻辑。

#### Scenario: Redis 不可用时服务启动
- **WHEN** 服务进程启动且 Redis 连通性检查失败
- **THEN** 进程启动 SHALL 立即失败，并在启动日志中输出依赖失败原因

#### Scenario: 运行时缓存操作失败
- **WHEN** 请求处理中发生 Redis 缓存读写失败
- **THEN** 系统 SHALL 返回明确的依赖错误，且不得切换到内存兜底

### Requirement: RabbitMQ 必须作为唯一事件通道

系统 SHALL 将 RabbitMQ 作为跨服务事件通道。**UCG 审核/推荐等必需 consumer 流程** MUST 保持 RabbitMQ；history/device 的 `history.record.*` / `device.*` **fan-out 发布** 随 worker 删除而移除，不再作为必需流程。API 类进程 MAY 在 RabbitMQ 不可达时仍启动。**已删除 worker-service**，不再有「worker 启动 MUST 因 RabbitMQ 失败而失败」的进程角色。

#### Scenario: RabbitMQ 不可用时 API 服务启动

- **WHEN** API 类服务进程启动且 RabbitMQ 连通性检查失败
- **THEN** 进程 MAY 继续启动并记录 MQ 降级 Warning（ucg-service 若启用 consumer 则按 ucg 规格处理）

#### Scenario: 运行时必需 UCG 事件发布失败

- **WHEN** ucg-service 路径要求发布审核/推荐事件且 RabbitMQ 发布失败
- **THEN** 系统 MUST 按 UCG outbox/MQ 规格处理（与 worker 删除无关）

---

### Requirement: ZREVRANGE WITHSCORES 解析兼容嵌套与扁平响应

`cachekit` 的 `SortedSetRevRangeWithScores`（及内部 helper）在解析 Redis `ZREVRANGE key start stop WITHSCORES` 结果时，MUST 同时支持：

- **嵌套形态**：`[[member, score], [member, score], ...]`（go-redis 9.x 经 GoFrame `Do` 常见返回）；
- **扁平形态**：`[member, score, member, score, ...]`（历史/其他 adapter 返回）。

解析输出的每个 `ZSetMemberScore.Member` MUST 为 Redis ZSET 的 member 字符串（如 `"1"`），MUST NOT 为整对 `[member,score]` 的 JSON 或数组字符串。`Score` MUST 为对应浮点分值。

#### Scenario: 嵌套响应解析为正确 member

- **WHEN** `Do(ZREVRANGE ... WITHSCORES)` 返回 `[[ "1", "1.718" ], [ "19", "1.195" ]]`
- **THEN** `SortedSetRevRangeWithScores` MUST 返回 `[{Member:"1", Score:1.718}, {Member:"19", Score:1.195}]`（score 允许浮点误差）

#### Scenario: 扁平响应仍兼容

- **WHEN** 响应为 `[ "1", "1.718", "19", "1.195" ]`
- **THEN** 解析结果 MUST 与嵌套形态等价

#### Scenario: 空 ZSET

- **WHEN** Redis 返回空数组或 nil
- **THEN** MUST 返回空 slice 且无 error

---

### Requirement: cash-service 独立进程与配置

系统 MUST 提供独立进程 `cash-service`，默认配置文件为 `manifest/config/config.cash-service.yaml`（或等价），MUST NOT 将现金域专属配置回流到主配置 `manifest/config/config.yaml`。进程启动 MUST 能通过环境变量 **`CASH_DB_LINK`** 覆盖其 GoFrame `database.default` 连接。

#### Scenario: 默认配置指向本域

- **WHEN** 未设置 `GF_GCFG_FILE` 启动 `cmd/cash-service`
- **THEN** 运行时 MUST 使用 cash-service 专用配置文件，且 MUST NOT 依赖共享主配置承载支付/VIP 业务项

#### Scenario: 数据库连接可被 CASH_DB_LINK 覆盖

- **WHEN** 部署环境设置 `CASH_DB_LINK` 为 `ai_voice_cash` 的 DSN
- **THEN** cash-service MUST 使用该连接访问本域库，MUST NOT 直连 `ai_voice_device` / `ai_voice_voice` 等他域库承载 VIP 表

### Requirement: 部署清单包含 cash-service

微服务 Compose、`.env.example` 与发布 runbook MUST 登记 `cash-service`（镜像/服务名、`CASH_DB_LINK`、`CASH_SERVICE_ADDR`、他服所需的 `CASH_SERVICE_URL`）。

#### Scenario: Compose 可拉起 cash-service

- **WHEN** 按微服务 compose 启动含 cash-service 的栈且已配置 `CASH_DB_LINK`
- **THEN** 容器 MUST 监听配置地址（建议 `:9806`）并提供可探测的 HTTP 健康/API 入口（与现网其它 `*-service` 惯例一致）

---

### Requirement: HTTP controller MUST 按宿主进程分子包

`internal/controller` 下 MUST 按进程划分目录（至少包含：cash、voice、device、history、ucg、gatewayapp、simuser、notify）。各进程的 `Register*ServiceHTTP`（或等价注册入口）MUST 仅绑定本进程子包中的控制器类型。误导性文件名（例如实际由 voice 宿主注册却以 `device_` 为前缀的 care-alert/tip/clinic 控制器）MUST 迁入正确进程子包，并可重命名类型；gateway 反代与跨切面 MUST 归入 gatewayapp（或明确的 proxy 子包并由 gateway-app 注册）。

#### Scenario: voice 注册 care-alert

- **WHEN** 构建并注册 voice-service HTTP
- **THEN** care-alert 控制器 MUST 位于 voice 子包且由 voice 注册入口 Bind，MUST NOT 再与 device-service 注册入口混绑

#### Scenario: cash 仅绑 cash 控制器

- **WHEN** 注册 cash-service
- **THEN** Bind 集合 MUST 仅含 cash 子包控制器（VIP/功能等）

### Requirement: 对外 App 路由字符串 MUST 冻结

本变更中，面向 App/网关已暴露的路由 `path` 字符串（含 `g.Meta` 中的 path 与既有反代匹配模式所依赖的对外前缀）MUST NOT 被修改。允许调整 Go 包路径、类型名与文件名。internal 服务间 path 宜保持不变；若必须调整 MUST 同步所有调用方 clients，且仍 MUST NOT 影响 App 路径。

#### Scenario: care-alert App path 不变

- **WHEN** 完成本变更后的代码审查或自动化检查
- **THEN** App 使用的 `/device/api/care-alert/*`（及同类既有对外 path）MUST 与变更前字符串一致

#### Scenario: 重命名控制器不影响 path

- **WHEN** 将 `DeviceCareAlertController` 重命名并移入 `controller/voice`
- **THEN** 其 `g.Meta` path 字段 MUST 仍为原对外路径

---

### Requirement: DAO extension files SHALL follow minimum-necessary rule

DAO `*_ext.go` files MUST be retained only when they provide business-meaningful extensions beyond generated DAO wrappers.

#### Scenario: Ext file has no added behavior
- **WHEN** an ext file only duplicates generated DAO behavior without business logic
- **THEN** the file MUST be merged away or removed

#### Scenario: Ext file provides service-specific query semantics
- **WHEN** an ext file includes domain query composition or behavior not present in generated DAO
- **THEN** the file MAY be retained with explicit comment/documented rationale

---

### Requirement: 时间类字段落库形态

除经架构评审明确豁免的「纯日历日期」字段外，凡表示**时刻**（事件发生时间、创建/更新时间、最后活跃时间等）的数据库列 **MUST** 以 **Unix 时间戳秒** 数值存储，MySQL 类型 **MUST** 为可表达该范围的整数类型（推荐 `BIGINT`）。**MUST NOT** 将本地墙钟格式化的日期时间字符串作为权威落库值。

#### Scenario: 新表创建

- **WHEN** 新建包含「时刻」语义的表或列
- **THEN** 该列类型为整数型时间戳秒且注释标明 UTC 纪元秒，且应用写入路径使用 UTC 纪元秒（如 `time.Time.Unix()`）而非格式化字符串

### Requirement: API 与 JSON 契约

对外 HTTP JSON 中代表「时刻」的字段 **MUST** 使用数字类型（Unix 秒），与数据库存储单位一致；字段文档或 OpenAPI **MUST** 标明单位为秒。若迁移期需兼容旧客户端，**MUST** 在变更说明中定义弃用截止条件，且服务端权威值仍为数字戳。

#### Scenario: 客户端解析

- **WHEN** 客户端接收代表事件发生时刻的字段
- **THEN** 该值为 JSON number（Unix 秒），客户端在展示给用户时自行按目标时区转换，不依赖服务端返回本地日历字符串作为权威

### Requirement: 迁移与数据完整性

对已有「非数字时刻」列的迁移 **MUST** 提供可重复执行的回填策略，并在切换读写前完成行数一致性与抽样校验。**MUST** 定义 NULL/非法旧值的处置规则（拒绝写入、置 0 或置哨兵值须文档化且经评审）。

#### Scenario: 回填后校验

- **WHEN** 执行从旧列到新秒级列的回填脚本
- **THEN** 存在自动化或清单式校验（行数、非 NULL 比例、时间范围合理性）且通过后应用才切换为只读新列

### Requirement: 服务边界

各服务 **MUST** 仅修改本服务拥有库内的表与 DAO；跨服务时间语义通过契约（HTTP/RPC/事件）传递，传递值 **MUST** 为 Unix 秒或与契约显式声明的单位一致。

#### Scenario: history 与 device 分库

- **WHEN** 在 history-service 所属库中迁移时间列
- **THEN** 不修改 device-service 所属库的表结构于同一提交中混写；各自变更独立可发布

---

### Requirement: Gateway MUST 为 device 路由提供独立可配置代理能力

gateway MUST 以独立中间件管理 `/device/admin/api/*` 路由，并支持 `local|proxy|canary` 三态。

#### Scenario: device 路由进入 local 模式
- **WHEN** `DEVICE_API_ROUTE_MODE=local`
- **THEN** gateway MUST 执行本地处理链路，且 MUST NOT 将请求转发到 device-service

#### Scenario: device 路由进入 proxy 模式
- **WHEN** `DEVICE_API_ROUTE_MODE=proxy` 且 `DEVICE_API_PROXY_URL` 可用
- **THEN** gateway MUST 将 `/device/admin/api/*` 请求全量转发到 device-service

#### Scenario: device 路由进入 canary 模式
- **WHEN** `DEVICE_API_ROUTE_MODE=canary` 且配置了 `DEVICE_API_PROXY_CANARY_PERCENT`
- **THEN** gateway MUST 按稳定分流键执行百分比转发，其余请求保持本地处理

### Requirement: device canary 分流 MUST 保持同键稳定

gateway MUST 采用稳定分流键（如 deviceNo 或请求头标识）对 canary 流量做无状态一致性计算。

#### Scenario: 同一分流键连续请求
- **WHEN** 同一设备在 canary 模式下发起多次 `/device/admin/api/*` 请求
- **THEN** 请求 MUST 稳定命中同一流量路径（proxy 或 local）

---

### Requirement: 领域包边界 MUST 与服务边界一致

迁移后代码 MUST 按领域归属放置在 `internal/services/voice`、`internal/services/device`、`internal/services/history` 等目录，且包语义必须与目录一致。

#### Scenario: 包语义审查
- **WHEN** 审查迁移后的领域目录
- **THEN** 目录内代码包语义 MUST 体现对应领域职责，不得继续使用统一 `service` 包承载多域逻辑

### Requirement: 共享目录准入 MUST 可审计

`internal/shared` MUST 仅容纳无领域语义的通用能力；含领域流程或领域模型耦合的实现 MUST 禁止进入共享目录。

#### Scenario: 共享目录准入检查
- **WHEN** 有文件计划迁入 `internal/shared`
- **THEN** 评审 MUST 给出“无领域语义”依据，否则该文件 MUST 回到对应领域目录

### Requirement: 新增代码 MUST 禁止回流到 `internal/service`

迁移完成后，新增实现文件 MUST 不得再放入 `internal/service`。

#### Scenario: 新增文件路径检查
- **WHEN** 提交包含新增实现文件
- **THEN** 若目标路径为 `internal/service`，该提交 MUST 视为不符合边界规范

---

### Requirement: 渐进迁移兼容层

系统 MUST 在迁移期间保留字符串入口的兼容适配层，并通过统一适配函数将旧字符串路径映射到新枚举实现。

#### Scenario: 旧入口继续可用
- **WHEN** 调用方仍传入历史字符串值
- **THEN** 系统 MUST 通过兼容适配层完成转换并保持行为一致

#### Scenario: 兼容层输出弃用提示
- **WHEN** 旧入口被调用
- **THEN** 系统 SHOULD 输出弃用告警日志，提示迁移到枚举入口

### Requirement: 枚举化迁移可验证

系统 SHALL 提供可验证迁移清单，确保关键模块不再新增裸字符串匹配。

#### Scenario: 核心模块迁移完成检查
- **WHEN** 执行迁移验收
- **THEN** 系统 MUST 能确认 outbox、consumer、voice 关键路径已使用枚举匹配

---

### Requirement: 可能无行的单行查询 MUST 使用 One + IsEmpty

在 `internal/services/**` 与 `internal/controller/**` 中，凡对 GoFrame/gdb 模型执行**单行**查询且业务上允许零行（可选配置行、按条件取最新、按业务键探测是否存在等），实现 MUST 使用 `.One()`，并在 `err == nil` 后通过 `IsEmpty()` 区分「无行」与「有行」；有行后再 `Struct`（或等价字段读取）。MUST NOT 依赖单结构体 `.Scan(&T)` 在空集时返回 `nil`。若因历史原因保留 `Scan`，则 MUST 将 `errors.Is(err, sql.ErrNoRows)`（及驱动等价空结果错误）视为「无行」，MUST NOT 将其作为系统失败向上抛出或打成失败级日志。

#### Scenario: 作者无 pending 资料审核 job

- **WHEN** 调用方请求作者视角 profile 合并，且该 `wx_id` 在 `ucg_profile_audit_job` 中不存在 `status=pending` 行
- **THEN** 系统 MUST 将结果视为「无待审 job」（例如 `ok=false`），MUST NOT 因空结果集将 `sql.ErrNoRows` 记为失败 WARN，MUST 仍返回已发布 profile 视图

#### Scenario: 可选覆盖行不存在

- **WHEN** 业务读取用户级可选覆盖表（如 AI quota override）且该用户无覆盖行
- **THEN** 系统 MUST 回落默认配额/空覆盖视图并成功返回，MUST NOT 因空集 `Scan` 错误导致接口失败

### Requirement: 空集 MUST NOT 冒充系统失败或无限 requeue

当「无行」属于预期业务状态时，系统 MUST NOT：将裸 `sql.ErrNoRows`（或空集 Scan 错误）作为 HTTP/RPC 系统错误返回给调用方；或在审核等 MQ consumer 中因该错误对消息 Nack/requeue。按实体 ID 加载且行已不存在时，consumer MUST 视为过期/缺失并 Ack（或等价丢弃），MUST NOT 因空集读失败反复入队。

#### Scenario: 审核消息指向已删除实体

- **WHEN** 审核 MQ 消息携带的 jobId/postId/commentId/messageId 在库中已不存在
- **THEN** consumer MUST 跳过该消息并 Ack，MUST NOT 因单行空查询错误导致消息持续 requeue

#### Scenario: 业务必须存在的行缺失

- **WHEN** 查询语义要求行必须存在（例如按订单号履约）且库中无行
- **THEN** 系统 MUST 返回明确业务错误（如订单不存在），MUST NOT 向调用方暴露裸 `sql.ErrNoRows`

### Requirement: 列表 Scan 与必有行回读不受本约定禁止

对切片目标的 `.Scan(&[]T)`（列表/分页/批量）以及在同一流程中刚成功插入、或已 `Ensure*Row`/`Count>0` 保证存在后的单行回读，实现 MUST 被允许继续使用 `Scan`；本约定 MUST NOT 要求将其改为 `One`。本约定的禁止范围 MUST 仅针对「空集为正常可能」且「错误被当作系统失败」的单结构体路径。

#### Scenario: 管理端分页列表

- **WHEN** 管理接口按分页查询多行并 `Scan` 到切片
- **THEN** 零行结果 MUST 表现为空列表，且本约定 MUST NOT 要求改为 `One`

### Requirement: 全局文档与评审 MUST 记载空集查询约定

`openspec/project.md` MUST 包含「GoFrame 单行空结果查询」强制约定（含推荐写法、禁止写法、排除项与评审检查）。`AGENTS.md` MUST 包含可执行摘要并指向 `project.md` 对应章节。涉及新增/修改单行「可能无行」查询的变更，评审 MUST 检查是否违反该约定。

#### Scenario: 文档可被 AI 与评审发现

- **WHEN** 贡献者或 AI 阅读 `AGENTS.md` / `openspec/project.md`
- **THEN** 均可定位到单行空结果查询的强制约定，且 `AGENTS.md` 摘要与 `project.md` 权威描述一致指向

---

### Requirement: 主配置 MUST 仅包含网关与全局公共配置

`manifest/config/config.yaml` MUST 仅保留 gateway 与全局共享配置项，不应包含仅属于 `voice-service`、`device-service`、`history-service` 的专属业务配置字段。

#### Scenario: 主配置检查
- **WHEN** 审查 `config.yaml` 字段归属
- **THEN** 所有服务专属字段 MUST 已迁移到对应服务专属配置文件

### Requirement: 删除主配置服务专属项 MUST 保持服务可启动

当主配置移除服务专属字段后，各服务 MUST 仍可通过其专属配置文件独立启动并完成依赖加载。

#### Scenario: 删除 voice 专属段后启动 voice-service
- **WHEN** 主配置不再包含 voice 专属业务配置项
- **THEN** `voice-service` MUST 通过自身配置文件正常启动且功能不缺失

---

### Requirement: Main config SHALL not carry database settings

`manifest/config/config.yaml` MUST NOT contain `database.*` fields after worker dedicated configuration is introduced.

#### Scenario: Review main config fields
- **WHEN** auditing `manifest/config/config.yaml`
- **THEN** no database connection configuration MUST exist in the file

#### Scenario: Gateway runtime without DB dependency
- **WHEN** `gateway-service` starts with main config
- **THEN** gateway MUST run without requiring database fields from main config

---

### Requirement: 配置边界 MUST 与服务边界一致

系统 MUST 保证“服务职责边界、配置归属边界、运行入口边界”一致；任何服务不得通过共享主配置承担他域职责或访问路径。

#### Scenario: gateway 运行角色审查
- **WHEN** 审查 gateway 启动入口与配置项
- **THEN** gateway MUST 仅包含流量与策略层配置，不得加载他域业务执行配置

#### Scenario: voice 跨服务访问
- **WHEN** voice 需要获取 history/device 领域数据
- **THEN** voice MUST 通过服务契约访问，不得通过主配置回流到跨库直查实现

### Requirement: 最终形态迁移 MUST 包含可回滚路径

面向最终微服务形态的配置与边界收敛 MUST 提供清晰的分阶段切换与回滚策略，避免一次性切换导致生产不可用。

#### Scenario: canary 切换失败
- **WHEN** 配置切换到 canary/remote 后出现异常
- **THEN** 系统 MUST 支持按服务维度快速回滚到 local/上一版本配置

---

### Requirement: 系统 MUST 提供独立 push-service 进程

系统 MUST 以独立进程 `push-service` 承载厂商推送（APNs/HMS/MiPush）与设备 token 存储。该进程 MUST NOT 复用 `notify-service`。进程 MUST 使用独立配置文件（如 `config.push-service.yaml`）与可配置监听地址（默认建议 `:9808`）。数据库 MUST 使用同 MySQL 实例上的独立 schema `ai_voice_push`，token 表 MUST NOT 再作为 UCG 业务表写入权威。

#### Scenario: 进程可独立启动

- **WHEN** 运维仅启动 push-service 且数据库 `ai_voice_push` 可用
- **THEN** 服务 MUST 能接受注册与 internal 发送请求，且 MUST NOT 依赖 ucg-service 进程内存活才能连接厂商

### Requirement: Token 存储 MUST 在 ai_voice_push

push-service MUST 将设备推送 token 持久化于 `ai_voice_push` 中的设备表（实现表名可定为 `push_device`），并以 `(wx_id, device_key, channel)` 唯一约束支撑 upsert。channel MUST 限于 `apns`、`hms`、`mipush`。

#### Scenario: 同设备同通道覆盖 token

- **WHEN** 同一 `wxId`+`deviceKey`+`channel` 再次注册新 token
- **THEN** 系统 MUST 更新为新 token，MUST NOT 插入重复唯一键行

### Requirement: 部署清单 MUST 包含 push-service

CI ACR 工作流与 `docker-compose.microservices.yml`（及必要 overlay/env 示例）MUST 纳入 `push-service` 镜像构建/推送与本地编排；推送厂商相关环境变量 MUST 注入 push-service 容器，MUST NOT 再作为 ucg-service 推送实现的必需配置。

#### Scenario: ACR 可构建 push-service

- **WHEN** 工作流以全量或指定 `push-service` 触发构建
- **THEN** MUST 使用对应 Dockerfile 产出 push-service 镜像标签

---

### Requirement: 项目 SHALL 提供 Redis 容灾 runbook 文档

仓库 SHALL 包含 `docs/runbooks/redis-disaster-recovery.md`，说明本项目 Redis 在 Docker 环境下的重启、卷保留、AOF 持久化与数据丢失场景下的恢复步骤。文档 SHALL 使用中文说明性文本；命令、路径、环境变量名可保留英文。

#### Scenario: 运维查阅容器重启恢复

- **WHEN** 运维人员 Redis 容器 stop/start 或 `docker compose up --force-recreate`（未使用 `down -v`）
- **THEN** runbook SHALL 提供验证步骤（PING、CLUSTER INFO、DBSIZE）
- **AND** SHALL 说明无需重复 `cluster create`（生产 cluster 卷已有元数据时）

#### Scenario: 运维查阅 volume 丢失影响

- **WHEN** 运维人员执行 `down -v` 或 volume 损坏
- **THEN** runbook SHALL 列出按数据类型的可恢复性（MySQL 权威 vs 仅 Redis vs 可 lazy warm）
- **AND** SHALL 说明 UCG 私信在方案 A 下可从 MySQL 读时回填 Redis

### Requirement: runbook SHALL 说明本项目 Redis 持久化配置

文档 SHALL 描述测试 standalone 与生产 cluster 的 compose 文件路径、`--appendonly yes`、volume 名称（如 `redis-test-data`、`redis-node-*-data`），以及 AOF 文件位于容器 `/data` 目录。

#### Scenario: 查阅 AOF 与 volume 位置

- **WHEN** 运维需要确认持久化是否启用
- **THEN** runbook SHALL 指向 `manifest/docker/docker-compose.redis-standalone.test.yml` 与 `docker-compose.redis-cluster.yml` 中的相关配置

### Requirement: runbook SHALL 提供 volume 备份与还原指引

文档 SHALL 包含 Docker volume 备份示例（如 `tar` 打包 `/data`）与还原步骤，并注明备份前应尽量降低写入或接受 point-in-time 一致性限制。

#### Scenario: 计划内维护前备份

- **WHEN** 运维在维护窗口前备份 Redis 数据
- **THEN** runbook SHALL 提供可复制的备份命令示例（测试 standalone 与生产 cluster 节点）

### Requirement: runbook SHALL 与 release-deploy-and-run 交叉引用

`docs/runbooks/redis-disaster-recovery.md` SHALL 链接至 `docs/runbooks/release-deploy-and-run.md` 中 Redis 日常重启、cluster create 误报等章节，避免重复维护冲突步骤。

#### Scenario: 生产 cluster 报 Node is not empty

- **WHEN** 运维误执行 cluster create
- **THEN** runbook SHALL 说明先查 `cluster_state:ok` 并引用 release runbook 对应排障节

---

### Requirement: 业务代码 SHALL 经 cachekit WithObserver 访问 Redis KV

`internal/services/**` 与 `internal/controller/**` 中的 Redis 键值读写（含 String、Hash、List、Set、Sorted Set、INCR/DECR、EXPIRE/PERSIST、EVAL 等）MUST 经 `cachekit.Cache` 接口执行，且 MUST 使用 `cachekit.WithObserver(..., cachekit.LoggingObserver{})`（或等价的 `cachekit.Default()`）包装。**SHALL NOT** 直接调用 `g.Redis().Do(...)` 或 `g.Redis()`。唯一允许直连 `g.Redis()` 的 Go 源码 MUST 位于 `internal/platform/cachekit/**` 与 `internal/platform/rediscfg/**`。

#### Scenario: 业务包无 g.Redis 直连

- **WHEN** 对 `internal/services` 与 `internal/controller` 执行 `rg 'g\.Redis\(\)' --glob '*.go'`
- **THEN** 匹配数 SHALL 为 0

#### Scenario: cachekit 操作可观测

- **WHEN** 业务经 `cachekit.WithObserver` 执行 Redis GET 且 Redis 可用
- **THEN** 观测器 SHALL 收到带 `operation`、`key`、`duration` 的回调；失败时 SHALL 以 warning 级别记录

#### Scenario: Redis 不可用错误语义一致

- **WHEN** 经 cachekit 执行操作且 Redis 返回连接/协议错误
- **THEN** 调用方 SHALL 收到 wrapping `cachekit.ErrUnavailable` 的错误

### Requirement: Redis Pub/Sub SHALL 经 redismsgkit WithObserver

Redis **PUBLISH** 与 **SUBSCRIBE**（及进程级订阅 goroutine）MUST 经 `internal/platform/redismsgkit` 抽象，且 MUST 使用 `redismsgkit.WithObserver(..., redismsgkit.LoggingObserver{})`。**SHALL NOT** 在业务或 controller 层直接 `g.Redis().Do("PUBLISH", ...)` 或使用 `github.com/redis/go-redis/v9` 创建客户端。

#### Scenario: 历史变更通知发布经 redismsgkit

- **WHEN** history-service 广播 App 历史增删改
- **THEN** MUST 调用 `redismsgkit` Publisher 向频道 `app:history:notify` 发布，且 SHALL NOT 直接 PUBLISH

#### Scenario: gateway-app 订阅经 redismsgkit

- **WHEN** gateway-app-server 启动历史 WS 通知订阅
- **THEN** MUST 经 `redismsgkit` 订阅 `app:history:notify`，且业务包 SHALL NOT 含 `redis.NewClient` / `redis.NewClusterClient`

### Requirement: Redis 键与 Pub/Sub 频道 SHALL 仅经 platform builder 构造

除 `internal/platform/cachekit/**` 与 `internal/platform/redismsgkit/**` 外，**SHALL NOT** 出现 Redis 键或 Pub/Sub 频道名字面量拼接（含 `fmt.Sprintf("domain:...")`）。既有线上键字符串 MUST 通过 platform builder 返回且与本变更前一致（策略 A：不重命名键空间）。

#### Scenario: ucg 聊天键经 keys_ucg builder

- **WHEN** ucg-service 读写会话消息 List
- **THEN** 键 MUST 由 `cachekit` 域 builder（如 `UCGChatMsgListKey(convID)`）生成，且返回值 SHALL 等于 `ucg:chat:conv:{convID}:msgs`

#### Scenario: 跨服务 ai 配额键单点定义

- **WHEN** voice-service 或 ucg-service 读写 AI 月度配额
- **THEN** 键 MUST 由 `cachekit` 的 `AIQuotaUsageKey`（或等价）生成，且 voice/ucg SHALL NOT 各自维护前缀常量

#### Scenario: App 历史通知频道单点定义

- **WHEN** 任意模块引用 App 历史 WS 通知频道名
- **THEN** MUST 使用 `redismsgkit.ChannelAppHistoryNotify`（或后继等价常量），且字符串 SHALL 为 `app:history:notify`

### Requirement: cachekit SHALL 提供全仓 typed 方法且无 Raw Do 后门

`cachekit.Cache` MUST 暴露本仓库业务所需的 typed 方法（含 `HashGetAll` 正确解析 GoFrame adapter 返回的 flat `[]string`），且 observed 包装 MUST 覆盖全部方法。**SHALL NOT** 向业务暴露通用 `Do(cmd, args...)` 逃逸口。

#### Scenario: HashGetAll 解析 flat []string

- **WHEN** 底层 `HGETALL` 经 GoFrame 返回 flat `[]string`（非 map）
- **THEN** `cachekit.HashGetAll` SHALL 仍返回正确的 `map[string]string`，SHALL NOT 因类型解析失败返回空 map

#### Scenario: 新增 Redis 命令须扩展 typed 接口

- **WHEN** 业务需要本变更未列出的 Redis 命令
- **THEN** MUST 先在 `cachekit` 或 `redismsgkit` 增加 typed 方法与观测，SHALL NOT 在业务层临时直连

### Requirement: 仓库级 AI 与代码评审 SHALL 检查 Redis platform 合规

`AGENTS.md` MUST 包含与 `openspec/project.md` 一致的 Redis 访问与键命名强制条款。PR 评审 MUST 包含对业务/controller 层 `g.Redis()` 与 Redis 键字面量的检查。

#### Scenario: AGENTS.md 含 Redis 强制节

- **WHEN** 查阅仓库根 `AGENTS.md`
- **THEN** SHALL 存在独立的 Redis 访问与 Redis 键命名强制说明，并引用 `cachekit` / `redismsgkit`

---

### Requirement: Redis 优先读取历史与元数据

系统对 `history/action/event/user` 的读取 SHALL 默认优先从 Redis 读模型获取；当缓存缺失、反序列化失败或依赖异常时，系统 MUST 回源权威数据源并在成功后回填 Redis。

#### Scenario: 缓存命中直接返回
- **WHEN** 读取 `history/action/event/user` 请求命中 Redis 且数据有效
- **THEN** 系统 MUST 直接返回缓存结果且不访问数据库

#### Scenario: 缓存缺失触发回源回填
- **WHEN** 读取请求未命中 Redis
- **THEN** 系统 MUST 回源数据库或契约服务获取数据并回填到 Redis 后返回

#### Scenario: 缓存损坏自动降级
- **WHEN** Redis 中对应 key 数据格式错误或反序列化失败
- **THEN** 系统 MUST 降级为回源读取并覆盖修复该缓存 key

### Requirement: 统一缓存键空间与版本语义

系统 SHALL 为 `history/action/event/user` 定义统一域内 key 规则与版本键规则，并 MUST 在读取时识别版本语义以支持后续乱序保护和修复。

#### Scenario: 键命名符合域规范
- **WHEN** 任一模块写入或读取缓存 key
- **THEN** key MUST 满足统一格式（domain:module:kind:identifier）且可由领域缓存仓储一致生成

#### Scenario: 版本键可用于一致性判断
- **WHEN** 读取方发现实体数据键与版本键不一致
- **THEN** 系统 MUST 触发回源修复或异步修复流程并避免返回明显过期快照

### Requirement: DeepSeek 历史上下文读取复用读模型

语音链路在构造 DeepSeek prompt 时，历史与画像读取 MUST 复用 Redis 读模型优先路径，不得绕过读模型直接形成新的 DB 热点通道。

#### Scenario: 构造 prompt 时命中 Redis 历史
- **WHEN** 语音链路需要读取最近历史记录
- **THEN** 系统 MUST 优先读取 Redis 中历史读模型并用于 prompt 构造

#### Scenario: Redis 不可用时语音链路可用
- **WHEN** Redis 短时不可用
- **THEN** 系统 MUST 回源获取历史并继续完成 prompt 构造，同时记录降级日志与指标

---

### Requirement: sim-user-service MUST 整进程下线

系统 MUST 移除 `sim-user-service` 进程（cmd、Dockerfile、compose 服务、ACR 镜像别名、`config.sim-user-service.yaml`、`SIM_*` / `SIM_SERVICE_BASE_URL` 环境变量）、`internal/services/simuser` 与 `controller/simuser`、gateway `/sim/admin/api/*` 反代，以及 Hub `sim-admin.html` 与导航入口。

#### Scenario: compose 无 sim 服务

- **WHEN** 运维查看 microservices compose
- **THEN** MUST NOT 再包含可启动的 `sim-user-service` 服务定义（或等价已删除）

### Requirement: device/ucg sim-only API MUST 删除

device-service MUST 删除 `/device/internal/api/sim/*` 与 `sim_user.go`。ucg-service MUST 删除 `POST /ucg/internal/api/posts/sample`、`POST /ucg/internal/api/chat/sim-unread-sample`、`POST /ucg/internal/api/chat/send`（仅允许模拟发送方的内部发送）。**MUST 保留** `POST /ucg/internal/api/media/upload-video`（小智引导片等）。

#### Scenario: sim register 不可用且 upload-video 仍在

- **WHEN** 调用方请求原 sim username register，同时另一调用方请求 `upload-video`（如 kind=xiaozhi-guide）
- **THEN** sim register MUST 不可用；upload-video MUST 仍可处理合法上传

### Requirement: wx.is_simulated 列 MUST 删除

device-service MUST 停止读写 `wx.is_simulated`，Ensure/迁移 MUST 在列存在时执行 `DROP COLUMN is_simulated`（或等价）。部署 MUST 先清理 `is_simulated=1` 行（删除或注销），再 DROP 列。Admin `wx/list` MUST NOT 再依赖该列过滤。`/device/internal/api/ucg/wx/batch` 响应 MUST NOT 再包含 `isSimulated` 字段。

#### Scenario: 启动后列不存在

- **WHEN** device-service 完成 schema Ensure 且库中曾有 `is_simulated` 列
- **THEN** `wx` 表 MUST 不再包含 `is_simulated` 列

### Requirement: gateway 模拟号 usage 跳过 MUST 删除

gateway-app MUST 删除 `IsSimulatedWx` 及 usagestats/clientusage 中对模拟号的跳过逻辑，以及相关 Redis sim wx set 键写入。删除后 usage 统计 MUST 仅按既有其它维护型 skip 规则过滤，MUST NOT 再查询模拟号集合。

#### Scenario: 统计路径无模拟号分支

- **WHEN** 成功 App 请求进入 usage 记录逻辑
- **THEN** 实现 MUST NOT 调用已删除的 `IsSimulatedWx` 判定

### Requirement: 小智 MCP 链路 MUST 不受影响

删除 sim 表面时，系统 MUST 保留小智 MCP 的 cash 开通/试用/引导视频、device bindings、mcpbridge/voice chat 相关路径，以及 ucg `upload-video`。

#### Scenario: 小智引导片上传仍可用

- **WHEN** Hub/cash 经既有客户端上传小智引导视频
- **THEN** `upload-video` MUST 成功写入对象键（或既有错误语义），MUST NOT 因删除 sim 而路由消失

---

### Requirement: 路由键白名单校验

系统 MUST 对 outbox 写入与事件发布链路中的 `routing_key` 执行白名单校验，未注册路由键不得进入发布流程。

#### Scenario: 合法路由键正常发布
- **WHEN** 业务写入或发布已注册的 `routing_key`
- **THEN** 系统 MUST 允许进入 outbox 与发布流程

#### Scenario: 非法路由键被拒绝
- **WHEN** 业务使用未注册的 `routing_key`
- **THEN** 系统 MUST 拒绝该请求并输出结构化错误日志

### Requirement: 路由键集中定义

系统 SHALL 提供集中路由键定义与查询接口，避免在多个模块重复硬编码路由字符串。

#### Scenario: 新增路由键需要注册
- **WHEN** 开发新增事件路由键
- **THEN** 该路由键 MUST 先在集中注册处声明后才能被调用方引用

---

### Requirement: 新增路由键必须遵循注册流程

系统 MUST 定义并执行新增路由键的标准流程：前缀确认、路由键注册、分发映射、观测校验、文档更新。

#### Scenario: 开发者新增路由键
- **WHEN** 开发者新增一个路由键用于新事件
- **THEN** 必须同时完成注册、分发映射和文档更新，缺任一项视为未完成迁移

### Requirement: 迁移验收必须禁止新增核心裸字符串匹配

系统 SHALL 在迁移验收清单中明确要求：核心分发模块不得新增针对 `routing_key` 的裸字符串匹配分支。

#### Scenario: 代码评审检查分发逻辑
- **WHEN** 评审者检查 outbox 与投影分发相关改动
- **THEN** 若发现新增裸字符串匹配而未使用统一前缀/枚举入口，必须拒绝合并

---

### Requirement: 路由键前缀必须集中注册

系统 MUST 在统一注册入口维护路由键前缀常量，并将前缀作为领域分组的唯一来源，禁止在业务模块重复定义核心前缀字面量。

#### Scenario: 新增前缀时集中维护
- **WHEN** 开发者需要新增一个事件族前缀
- **THEN** 必须在统一注册入口新增前缀常量并在路由键定义处复用该常量

### Requirement: 路由键定义必须采用前缀与后缀组合

系统 SHALL 通过“前缀常量 + 后缀常量/字面量”生成路由键枚举值，保证同一事件族的命名一致性。

#### Scenario: 定义 history 事件路由键
- **WHEN** 定义 `history.record.created`、`history.record.updated`、`history.record.deleted`
- **THEN** 这些路由键必须共享 `history.record.` 前缀常量并仅通过后缀区分

---

### Requirement: API 类进程启动 SHALL 不因 RabbitMQ 不可达而失败

gateway-service、gateway-app-server、device-service、history-service、voice-service、ucg-service 在启动前 SHALL 校验 Redis 连通性；RabbitMQ 管理 API 探活 MAY 失败且不 SHALL 阻断进程启动。探活失败时 SHALL 记录 Warning 级别日志。

#### Scenario: RabbitMQ 宕机时 device-service 启动

- **WHEN** device-service 启动且 Redis 可达但 RabbitMQ 管理 API 不可达
- **THEN** 进程 SHALL 成功启动并监听 HTTP

#### Scenario: Redis 不可达时 API 进程启动

- **WHEN** 任一 API 类进程启动且 Redis 不可达
- **THEN** 进程启动 SHALL 失败

### Requirement: worker-service 启动 SHALL 保持 RabbitMQ 强依赖

worker-service 在启动前 SHALL 同时校验 Redis 与 RabbitMQ；任一失败 SHALL 导致进程启动失败。

#### Scenario: RabbitMQ 宕机时 worker-service 启动

- **WHEN** worker-service 启动且 RabbitMQ 管理 API 不可达
- **THEN** 进程启动 SHALL 失败

---

### Requirement: Runtime docs SHALL be centralized and governed

`dao-sync-by-domain.md` and `release-deploy-and-run.md` MUST be maintained in a dedicated runtime-docs directory, and change governance MUST require synchronized updates when runtime behavior changes. Changes that introduce or alter Compose prod/test dual-stack deployment, registry image tagging, or test seed desensitization MUST update `release-deploy-and-run.md` in the same change.

#### Scenario: Docs location is consolidated

- **WHEN** checking runtime operation documents
- **THEN** both target documents MUST exist under one dedicated new folder

#### Scenario: Governance requires synchronized update

- **WHEN** project runtime/deployment/database boundary rules change
- **THEN** project governance (`openspec/project.md`) MUST require updating both runtime docs

#### Scenario: Dual-stack change updates runbook

- **WHEN** a change adds prod/test Compose overlays or test seed procedures
- **THEN** `release-deploy-and-run.md` MUST be updated in that change before merge

### Requirement: Runbook SHALL 文档化 2G 双栈 survival 配置

`docs/runbooks/release-deploy-and-run.md` SHALL 包含 **2G ECS**（或 documented 低内存同机双栈）专节，至少包括：

- 生产 Redis **3 主 0 从** 与测试 Redis **单机** 拓扑及迁移步骤（含 `down -v` 与数据丢失说明）
- prod/test 启动命令叠加 **`docker-compose.resources.{prod,test}.yml`**
- 默认 **mem_limit / cpus** 对照表与 `docker stats` 验收
- MySQL 同机 **`innodb_buffer_pool_size`** 建议（如 256M 级）
- ASR 验收约定：生产微服务保持 Up，**仅 test 域名** 进行语音压测，避免 prod 并发 ASR
- OOM / swap 排错

#### Scenario: 运维按 2G 文档完成测试 Redis 迁移

- **WHEN** 运维阅读 runbook 2G 专节并按步骤从六节点 test cluster 迁到 standalone
- **THEN** 其 SHALL 能完成 standalone Redis 启动、`.env.test` 更新与微服务 recreate，且 **无需** `cluster create`

#### Scenario: 运维按文档叠加资源 limits 启动双栈

- **WHEN** 运维按 runbook 生产/测试启动命令启动双栈
- **THEN** 命令示例 SHALL 包含 `-f docker-compose.resources.prod.yml` 或 test 等价文件

### Requirement: release-deploy-and-run SHALL 文档化 Compose 镜像版本控制

`docs/runbooks/release-deploy-and-run.md` SHALL 包含章节说明：`docker-compose.microservices.yml` 与 prod/test overlay 的关系；镜像仓库 `${REGISTRY}` 与 `${IMAGE_TAG}` 用法；测试默认 `develop`、生产 semver 钉扎；registry `pull` + `--no-build` 部署；禁止生产使用 `:local`/`develop`；按服务镜像 tag 回滚步骤。

#### Scenario: 运维按 runbook 回滚生产镜像

- **WHEN** 生产发版后需回滚至上一 semver
- **THEN** runbook SHALL 提供将 `.env.prod` 中 `IMAGE_TAG` 改回上一版本并 `pull` + `force-recreate` 的步骤

### Requirement: release-deploy-and-run SHALL 文档化生产测试双栈部署

`docs/runbooks/release-deploy-and-run.md` SHALL 包含生产/测试双栈对照表（网络、端口、库名、静态目录、中间件端口）、测试栈启动顺序（test 网络 → test Redis cluster → test Rabbit 初始化 → microservices test）、Nginx 反代 `test.pangbao.cuplay.top:9701/9702`、健康检查 URL（对外形态与生产一致仅域名不同）、脱敏种子刷新与发版前 checklist。

#### Scenario: 运维首次搭建测试栈

- **WHEN** 运维仅阅读 `release-deploy-and-run.md`
- **THEN** 其 SHALL 能按文档顺序完成测试中间件与微服务启动，并验证 `https://test.pangbao.cuplay.top:9702/api.json`（或 documented 等价 HTTPS 探活）

---

### Requirement: 服务边界 MUST 与数据库边界一致

每个微服务 MUST 仅访问其所属数据库/表；跨服务数据访问 MUST 通过服务契约（HTTP/RPC/事件）进行，MUST NOT 通过 DAO 或 SQL 直连他域数据表。

#### Scenario: Voice 访问 history 数据
- **WHEN** voice 需要读取或更新 history 领域数据
- **THEN** voice MUST 通过 history 服务契约调用完成，代码中 MUST NOT 出现对 history DAO 的直接查询或更新

#### Scenario: Voice 访问 device/user 画像数据
- **WHEN** voice 需要读取设备资料、生日、性别或注册状态
- **THEN** voice MUST 通过 device 服务契约调用完成，代码中 MUST NOT 出现对 `user/device` 领域 DAO 的直接查询

#### Scenario: 代码评审发现跨库直查
- **WHEN** 新增代码出现跨服务数据库直连访问
- **THEN** 该变更 MUST 被视为违反架构约束并在合入前整改

### Requirement: 迁移期分流 MUST 支持可控回退

服务边界治理迁移期 MUST 支持 `local|remote|canary` 切换，并对同一分流键保持稳定命中，确保可渐进放量与快速回滚。

#### Scenario: Canary 分流验证
- **WHEN** 开启 canary 模式并设置百分比
- **THEN** 同一设备标识 MUST 稳定命中同一路径，避免在 local 与 remote 之间抖动

#### Scenario: 远程路径故障回退
- **WHEN** remote 路径连续失败且开启 failover 配置
- **THEN** 调用方 MUST 回退到 local 路径并记录可观测日志

### Requirement: 单服务单库部署下禁止进程内跨域 DAO

当进程配置的 `database.default` 仅包含本服务所属库时，业务代码 MUST NOT 通过 import 他域服务包并调用其基于 DAO 的实现来访问他域表；跨服务数据 MUST 经 HTTP/RPC/消息。禁止依赖「同一代码仓库、不同包」造成可连他域表的假象。

#### Scenario: Voice 进程仅配置 voice 库

- **WHEN** `voice-service` 的配置仅连接 `qa`/`suggest` 所在库
- **THEN** 代码路径 MUST NOT 执行对 `user`/`event`/`action` 等表的 DAO；必须通过 device-service 契约

#### Scenario: 评审发现 voice 包引用他域 DAO

- **WHEN** 代码评审发现 `internal/services/voice` 直接或间接触发他域 `dao` 访问
- **THEN** 该变更 MUST 拒绝合入，直至改为 HTTP 客户端或经批准的同进程例外（文档化且仅限非生产）

### Requirement: Device 进程内 outbox 写入 MUST 使用显式 history 库连接

`domain_outbox` 若仅存在于 history 库，device-service MUST 使用独立配置的 `history_relay`（或等价）连接组写入，MUST NOT 误用 `default` 连接组指向 device 库写 outbox。

#### Scenario: 分库部署

- **WHEN** device 与 history 为不同数据库实例
- **THEN** 未正确配置 relay 时 MUST 跳过或失败可观测，MUST NOT 静默写入错误库

### Requirement: 表归属 MUST 与部署库一致

在分库部署下，`history` 表与 `domain_outbox` MUST 仅由可连接 history 库的进程写入；`user`、`event`、`action` MUST 仅由可连接 device 库的进程写入；`qa`、`suggest` MUST 仅由可连接 voice 库的进程写入。禁止因历史单体代码路径而使用错误默认库连接组访问上述表。

#### Scenario: device 进程不写 history 库中的 outbox（除非显式配置）

- **WHEN** `domain_outbox` 表仅存在于 history 服务数据库中
- **THEN** device-service MUST NOT 使用 `user` 表所在连接组对 `domain_outbox` 执行 Insert，除非运维显式配置为同一物理库且经架构评审

#### Scenario: voice 进程不写 event/action 表

- **WHEN** voice-service 需要新增或查询事件字典、动作记录
- **THEN** voice MUST 通过 device 服务契约完成，MUST NOT 使用 `dao.Event` 或 `dao.Action` 直连 device 库表

### Requirement: history 服务 MUST NOT 冒充他域数据权威

对外 HTTP 或内部契约 MUST NOT 将「生日、事件选项、语音建议」等响应伪装为 history 数据库本地查询结果；若经网关聚合， MUST 在实现上分别调用 device/voice 权威服务，且错误语义可追溯至真实下游。

#### Scenario: 拆分后的 API 归属

- **WHEN** 客户端请求事件选项或用户画像
- **THEN** 响应数据 MUST 来源于 device 域存储与接口，而非 history 进程内对 `event`/`user` 表的 DAO 查询

### Requirement: ucg-service MUST NOT cross-read device database for wx data

ucg-service MUST treat `wx` table as device-domain data. All wx validation, batch profile display fields, and baby_name for default nickname MUST be fetched via device-service internal HTTP APIs with `DEVICE_GATEWAY_INTERNAL_SECRET`. ucg-service MUST NOT import device DAO or execute SQL against device database.

#### Scenario: ucg 读取 wx 展示名
- **WHEN** ucg-service 需要渲染帖子作者昵称且 ucg_profile 缺失
- **THEN** 服务 MUST 调用 device internal batch API，且 MUST NOT 查询 device 库 `wx` 表

#### Scenario: 评审发现 ucg 跨库 DAO
- **WHEN** 代码评审发现 ucg-service 直连 `wx` 表
- **THEN** 变更 MUST 拒绝合入

### Requirement: ucg 表 MUST 仅由 ucg-service 写入 ai_voice_ucg

Tables `ucg_*` MUST reside in database `ai_voice_ucg` and MUST only be written by ucg-service default connection. Other services MUST NOT insert/update ucg tables via cross-DB SQL.

#### Scenario: gateway 不写 ucg 表
- **WHEN** gateway-app 代理 UCG HTTP 请求
- **THEN** gateway MUST NOT 直接写入 `ucg_post` 或任何 ucg 表

---

### Requirement: `internal/service` 实现文件 MUST 全量迁移

系统 MUST 将 `internal/service` 中实现文件按领域归属迁移到 `internal/services/*` 或 `internal/shared/*`，迁移完成后不得遗留可编译业务实现文件。

#### Scenario: 全量迁移完成
- **WHEN** 执行迁移收口检查
- **THEN** `internal/service` 中不得再存在业务实现文件，且对应实现已在目标目录可追踪

### Requirement: 迁移后调用路径 MUST 指向新目录

所有服务入口、控制器和内部调用方 MUST 使用迁移后的包路径，不得继续依赖旧 `internal/service` 路径。

#### Scenario: 调用路径校验
- **WHEN** 对迁移范围执行 import 路径审查
- **THEN** 迁移后的调用引用 MUST 全部指向 `internal/services/*` 或 `internal/shared/*`

---

### Requirement: 服务配置 MUST 按服务进程独立加载

`voice-service`、`device-service`、`history-service`、`gateway-service` MUST 具备独立默认配置文件，服务启动时 MUST 优先使用本服务默认配置，并允许通过 `GF_GCFG_FILE` 显式覆盖。

#### Scenario: voice-service 启动未指定 GF_GCFG_FILE
- **WHEN** 启动 `voice-service` 且环境变量未设置 `GF_GCFG_FILE`
- **THEN** 系统 MUST 加载 `voice-service` 专属默认配置文件

#### Scenario: device-service 启动指定 GF_GCFG_FILE
- **WHEN** 启动 `device-service` 且设置 `GF_GCFG_FILE`
- **THEN** 系统 MUST 使用指定配置文件并覆盖默认路径

### Requirement: 服务级覆盖变量 MUST 仅影响本服务

服务级环境变量覆盖（如数据库连接、监听地址）MUST 仅影响当前服务实例，不得通过同名变量隐式影响其他服务配置行为。

#### Scenario: 设置 VOICE_DB_LINK
- **WHEN** 部署仅设置 `VOICE_DB_LINK`
- **THEN** 系统 MUST 只影响 `voice-service` 数据库连接，不得改变 `history-service` 与 `device-service` 连接

---

### Requirement: 业务服务包 MUST NOT 互相 import 实现

`internal/services` 下各业务域包（至少：cash、voice、device、history、ucg、gatewayapp、simuser、mcpbridge、appstatus）MUST NOT 直接 import 另一业务域包中的实现代码。跨域协作 MUST 通过 `internal/clients/*` 出站调用、消息事件或 `contracts` 接口（实现由本域或 clients 装配）。共享无业务工具 MUST 放在 `internal/platform/**`（或同等 platform 包）。`aimodel` 与 `contracts` MAY 被各业务包 import，视为共享基础设施而非业务线包。

#### Scenario: voice 不再 import services/cash

- **WHEN** 检查 `internal/services/voice` 的 Go import
- **THEN** MUST NOT 出现 `hello/internal/services/cash`

#### Scenario: ucg 不再 import services/device 业务实现

- **WHEN** 检查 `internal/services/ucg` 的 Go import
- **THEN** MUST NOT 出现对 `hello/internal/services/device` 的依赖（设备 HTTP 经 clients；类型/密钥头经 platform 或 clients DTO）

#### Scenario: 允许 platform 与 contracts

- **WHEN** voice 需要解析内部 wxId 头或调用 DeviceAdmin 契约
- **THEN** MAY import `platform` 与 `contracts`，MUST NOT 因此 import device 业务实现包

### Requirement: 伪公共工具 MUST 下沉 platform

下列能力 MUST 从业务域包迁出至 `internal/platform`（或等价 platform 子包），供多进程共用：内部网关密钥头与校验、App/内部注入头常量（如 wxId/deviceNo/clientIP）、从请求头解析 wxId、常量时间比较等。迁出后，原业务包 MUST NOT 再作为他域获取这些工具的唯一来源。

#### Scenario: cash controller 不 import voice 解析 wxId

- **WHEN** cash 控制器需要解析 `X-Internal-Wx-Id`
- **THEN** MUST 使用 platform（或等价）解析函数，MUST NOT import `internal/services/voice`

#### Scenario: history/ucg 不借 device 包做密钥校验

- **WHEN** history 或 ucg 内部接口校验网关内部密钥
- **THEN** MUST 使用 platform 校验，MUST NOT 仅为校验而 import device 业务包

### Requirement: 仓库 MUST 提供跨域 import 门禁

仓库 MUST 提供可运行的检查脚本（路径以实现为准，如 `hack/check-service-import.sh`），用于检测业务服务包之间的禁止 import；检查失败时 MUST 以非零退出码表示（可先 warn 后在同一变更内收紧为 fail）。`AGENTS.md` 与 `openspec/project.md` MUST 记载源码包边界约定，与「跨库须走 HTTP」并列。

#### Scenario: 门禁检出 voice→cash

- **WHEN** 有人重新引入 `services/voice` → `services/cash` 的 import 并运行门禁
- **THEN** 检查 MUST 失败（收紧后）或明确告警（过渡配置下）

#### Scenario: clients 与 platform 不误杀

- **WHEN** `services/voice` 仅 import `clients/cash` 与 `platform`
- **THEN** 门禁 MUST 通过

---

### Requirement: 迁移执行 MUST 分批且可验证

全量迁移 MUST 以可回滚批次执行，每批完成后必须通过编译校验与关键服务启动校验后方可进入下一批。

#### Scenario: 批次完成校验
- **WHEN** 单个迁移批次完成
- **THEN** 系统 MUST 通过既定编译检查与启动健康检查

### Requirement: 迁移异常 MUST 支持按服务维度回滚

迁移引发异常时，系统 MUST 支持按受影响服务维度回退代码与配置，不要求全局回滚。

#### Scenario: 单服务回滚
- **WHEN** `voice-service` 批次迁移后出现运行异常
- **THEN** 团队 MUST 可仅回滚 `voice-service` 相关迁移批次并恢复可用

### Requirement: 收口验收 MUST 覆盖关键链路无回归

全量迁移收口时 MUST 验证 gateway/voice/device/history 关键链路，确保外部行为与迁移前一致。

#### Scenario: 收口链路验收
- **WHEN** 所有批次迁移完成并准备收口
- **THEN** 关键业务链路 MUST 通过无回归验证，且迁移结果可被文档化追踪

---

### Requirement: 域间出站客户端 MUST 位于中立 clients 包

系统 MUST 将调用他域服务的 HTTP/WS 出站客户端放置于 `internal/clients/{target}/`（`target` 为被调服务域名，如 `cash`、`device`、`history`、`ucg`、`voice`）。被调方业务包 `internal/services/{target}` MUST NOT 再向其他服务进程导出 `Remote*` 或等价「调本域 internal API」的客户端符号供其 `import`。调用方业务包 MUST 通过 `clients/{target}` 发起跨进程调用，MUST NOT 为调用他域而 `import` 他域 `internal/services/{other}` 业务实现包。

#### Scenario: voice 调 cash VIP/access

- **WHEN** voice-service 需要查询 VIP 或值得留意 access
- **THEN** 源码 MUST `import` `internal/clients/cash`（或等价 clients 路径），MUST NOT `import hello/internal/services/cash`

#### Scenario: cash 调 device/history

- **WHEN** cash-service 需要设备号或喂养日统计
- **THEN** 出站客户端 MUST 位于 `internal/clients/device` 与 `internal/clients/history`（或从原 `cash/*_client.go` 迁入后由此引用），MUST NOT 要求 device/history 业务包反向依赖 cash

#### Scenario: 非本仓域依赖可例外

- **WHEN** 出站目标为 Python AI、阿里云内容安全或外部 LLM
- **THEN** 客户端 MAY 保留在原业务包或非域名 clients 子目录，MUST NOT 因此允许域业务包互相 import

### Requirement: 出站客户端迁入 MUST NOT 改变内部 API 契约语义

迁入 clients 时，请求的 path、方法、内部密钥头与响应解析语义 MUST 与迁入前一致（允许包路径与符号名变更）。MUST NOT 借机修改 App 对外路径。

#### Scenario: CareAlert access 路径不变

- **WHEN** voice 经 clients/cash 调用值得留意 access
- **THEN** 请求 URL path MUST 仍为既有 `/cash/internal/api/care-alert/access`（或迁入前实际使用的同一 path）

---

### Requirement: 缓存键与 TTL 规范必须统一

所有服务 SHALL 遵循统一的 Redis key 命名空间规范和已文档化的 TTL 规则，覆盖缓存、守卫与幂等状态。

#### Scenario: 任一服务引入新缓存键
- **WHEN** 某服务为运行时状态新增 Redis key
- **THEN** 该 key SHALL 符合统一命名规范与 TTL 策略

### Requirement: 事件命名与投递语义必须统一

所有跨服务事件 SHALL 使用统一的 exchange/routing-key 命名规范，并遵循明确的投递失败语义。

#### Scenario: 服务发布跨服务事件
- **WHEN** 某服务发出跨服务处理所需的领域事件
- **THEN** 其 SHALL 通过 RabbitMQ 按统一 exchange/routing-key 规范发布，并执行既定发布失败行为

### Requirement: 本迁移阶段执行禁测文件策略

代码库 SHALL 删除现有 Go 测试文件，并在本迁移阶段 SHALL NOT 新增 Go 测试文件。

#### Scenario: 迁移阶段引入新代码
- **WHEN** 开发者在本迁移范围内新增或重构代码
- **THEN** 其 SHALL 通过运行时核验脚本与运维检查进行验证，而不是新增 `*_test.go` 文件

---

### Requirement: sim-user-service SHALL persist runtime task and interval configuration in database

enabled`、`max_sim_users` 外，sim 库 MUST 持久化以下运行时项（列或 JSON 字段，语义等价即可）：

- **taskSwitches**：`register`、`comment`、`postImage`、`postVideo`、`postDebate`、`debateComment`、`chat`、`follow`（bool）； **MUST NOT** 含 `videoPoll`
- **intervals**（秒或 duration）：`register`、`comment`、`postImage`、`postVideo`、`postDebate`、`debateComment`、`chat`、`follow`、`postVideoPollInterval`、`postVideoPollMaxWait`、`startupStaggerMax`、`ephemeralChatLoop`、`ephemeralChatWindow`； **MUST NOT** 含.

#### Scenario: DB 运行时覆盖 env

- **WHEN** DB 中 `interval_post_debate=43200` 且 env 兜底为 12h
- **THEN** 新启动的 T7 goroutine MUST 使用 DB 值

#### Scenario: 进程总闸仍读 env

- **WHEN** `SIM_USER_SERVICE_ENABLED=false` 且 DB `enabled=true`
- **THEN** scheduler MUST NOT 启动

### Requirement: sim-user-service SHALL support scheduler reload on runtime config save

`PUT /sim/admin/api/config` 持久化后：

- 变更 **taskSwitches**、**T4 调度周期**（`postVideo`）、**rateLimit**、ephemeral、**enabled** → MUST Reload scheduler
- 变更 **仅** `postVideoPollInterval` / `postVideoPollMaxWait` → MUST NOT 触发 scheduler Reload（进行中的 poll 使用启动时快照）
- 变更 **仅** `maxSimUsers` → MAY 跳过 Reload
- LLM lane PUT MUST NOT 触发 Reload

#### Scenario: 修改 postVideoPollMaxWait 不 reload

- **WHEN** PUT 仅变更 `postVideoPollMaxWait`
- **THEN** 响应 MUST 含 `scheduleReloaded=false`

#### Scenario: 修改 postVideo 调度周期触发 reload

- **WHEN** PUT 变更 `interval_post_video`
- **THEN** 响应 MUST 含 `scheduleReloaded=true`

### Requirement: sim config PUT response SHALL describe save effects

`PUT /sim/admin/api/config` 响应 MUST 扩展：

- `scheduleReloaded`（bool）
- `effects[]`：`kind`（如 `scheduler_reloaded`、`task_interval_changed`、`ephemeral_may_continue`）、可选 `task`、`message`
- `taskSchedule[]`：每任务 `name`、`enabled`、`intervalSec`、`lastRunAt`（来自 `sim_task_run` 若存在）、`nextRunHint`（人类可读，如「约 3h 后」或「保存后立即进入新周期等待」）

系统 MUST NOT 保证强杀进行中的 LLM 调用或已 spawn 的 E1 聊天 goroutine；`effects` MUST 可表达「进行中任务可能跑完旧配置」。

#### Scenario: 关闭 chat 任务的效果提示

- **WHEN** PUT 将 `task_chat=false`
- **THEN** 响应 `effects` MUST 含 chat 相关提示且 `taskSchedule` 中 chat MUST 反映 disabled

#### Scenario: 修改 interval 的下一跑提示

- **WHEN** PUT 缩短 postImage 周期且该任务有 `lastRunAt`
- **THEN** `taskSchedule` 中 postImage 的 `nextRunHint` MUST 基于 `lastRunAt + 新 interval` 估算

### Requirement: docker compose env SHALL omit DB-backed sim runtime and LLM variables

`manifest/docker/docker-compose.microservices.yml` 与 `.env.example` MUST 移除已 DB 化的 sim 变量，至少包括：`SIM_LLM_*`、`SIM_TASK_*`、`SIM_INTERVAL_*`、`SIM_EPHEMERAL_*`、`SIM_STARTUP_STAGGER_MAX`、`SIM_UCG_RATE_LIMIT_*`、`SIM_VIDEO_POLL_ENABLED`。MUST 保留：`SIM_DB_LINK`、`SIM_USER_SERVICE_ENABLED`、`SIM_ADMIN_PASSWORD`、API Key 类变量。

#### Scenario: compose 无 SIM_LLM env

- **WHEN** 运维查看 microservices compose 中 sim-user-service environment
- **THEN** MUST NOT 含 `SIM_LLM_TEXT_PROVIDER` 等 LLM lane env 块

---

### Requirement: sim-user-service SHALL run as an independent process with a master enable switch

系统 MUST 提供独立进程 `sim-user-service`（`cmd/sim-user-service`）。环境变量 `SIM_USER_SERVICE_ENABLED` 为 `false` 时，进程 MAY 启动健康检查但 MUST NOT 启动任何周期 ticker 或视频轮询。为 `true` 时 MUST 注册全部已声明背景任务。

#### Scenario: Service disabled

- **WHEN** `SIM_USER_SERVICE_ENABLED=false`
- **THEN** 进程 MUST NOT 执行 T1–T6 与 P1 任务

#### Scenario: Service enabled

- **WHEN** `SIM_USER_SERVICE_ENABLED=true` 且依赖服务可达
- **THEN** 进程 MUST 按各任务独立开关启动对应 ticker

### Requirement: T1 register task SHALL create simulated users every 24 hours

每 24 小时（±10% jitter）MUST 执行注册任务（`SIM_TASK_REGISTER_ENABLED`）。当当前 sim 用户数 `< maxSimUsers` 时，MUST 生成**随机**账号与**随机**密码，调用 device internal sim 注册，经 `simText` 生成昵称、`simImageGen` 生成头像，完成 UCG profile 与 media 上传并更新 nickname/avatarKey；**仅在 profile 写入成功后** MUST 写入 `sim_wx_credential`。profile 完成前任一步失败 MUST 回滚 wx（见 ADDED Requirement）。当已达 `maxSimUsers` 时 MUST 跳过本次执行。

MUST NOT 再分配 `ptest{N}` 序号账号。`sim_account_seq` MUST NOT 参与 T1 账号生成。

#### Scenario: Register under cap with random credentials

- **WHEN** sim 用户数为 99 且 `maxSimUsers=100` 且 T1 全流程成功
- **THEN** 系统 MUST 注册新 sim 用户并标记 `is_simulated=1`，且 MUST 写入 credential

#### Scenario: Skip at cap

- **WHEN** sim 用户数已达 `maxSimUsers`
- **THEN** T1 MUST NOT 调用注册接口

### Requirement: T2 comment task SHALL post AI comments every 6 hours

每 6 小时（±10% jitter，可由 `SIM_INTERVAL_COMMENT` 覆盖）MUST 随机选取一个 sim 用户，经 gateway 登录，调用 ucg internal **`POST /ucg/internal/api/posts/sample`**，`mode=random`，且 **`excludeMediaTypes` MUST 含 `2`（视频）**，以获取单条非视频已发布帖（MUST NOT 调用 `GET /ucg/app/api/feed/recommend`）。 取得帖子后 MUST 经单次 `LaneSimVision` 生成评论并 `POST` 发表。 T2 MUST NOT 对 `mediaType=2` 的帖子发表评论； 若 sample 仍返回视频帖（例如 ucg 未升级），MUST skip 并记任务失败或 warning，且 MUST NOT POST 评论。 允许评论的帖子类型 MUST 为 `mediaType=0`（纯文字）或 `mediaType=1`（图文）。 Green 审核.

#### Scenario: Sample request excludes video

- **WHEN** 代码评审 `RunCommentTask` 的 sample 请求体
- **THEN** MUST 含 `"excludeMediaTypes": [2]`（或等价语义）

#### Scenario: No comment on video post

- **WHEN** sample 返回帖的 `mediaType` 为 `2`
- **THEN** T2 MUST NOT 调用 LLM 或 POST 评论

#### Scenario: Text and image posts still commented

- **WHEN** sample 返回 `mediaType=0` 或 `1` 且 LLM 可用
- **THEN** T2 MUST 按现有多模态/纯文本规则生成并发表评论

#### Scenario: No eligible post after exclude

- **WHEN** sample 因 exclude 返回空 `list`
- **THEN** T2 MUST 记录失败（如「无已发布帖」）且 MUST NOT 发表评论

#### Scenario: No recommend feed call

- **WHEN** T2 任务执行
- **THEN** MUST NOT 请求 `/ucg/app/api/feed/recommend`

### Requirement: T3 image post task SHALL publish image posts every 3.5 hours

每 3.5 小时（±10% jitter）MUST 随机 sim 用户经 `simText` 生成母婴文案、`simImageGen` 生成配图，完成 OSS media 链路后以 `submit=true` 创建图文动态。

#### Scenario: Image post submitted

- **WHEN** 生图与上传成功
- **THEN** 帖子 MUST 进入 `pending_audit` 并触发正常审核 MQ

### Requirement: T4 and P1 SHALL submit and poll video posts without retry

`RunPostVideoSubmitTask`（调度与手动共用）在 `SubmitVideoGeneration` 成功且 `InsertVideoJob` 写入 `pending` 后 MUST 启动视频结果轮询。 轮询 MUST 调用智谱 `GET /paas/v4/async-result/{task_id}`（经 `aimodel. 轮询 MUST 使用 submit 阶段已获得的 `loginSession`，MUST NOT 再经分页 list 线性查 account。 )`

MUST NOT 在 submit 成功时单独写 `RecordTaskRun` success。

#### Scenario: Poll success posts video

- **WHEN** T4 提交后 async-result 返回 success 且上传发帖 OK
- **THEN** job MUST 为 `done` 且 `sim_task_run` MUST 记 post_video_submit 成功

#### Scenario: Poll timeout fails task

- **WHEN** 自 submit 起超过 `postVideoPollMaxWait` 仍为 processing
- **THEN** job MUST 为 `skipped` 且 `sim_task_run` MUST 记 post_video_submit 失败

### Requirement: T5 and E1 SHALL handle chat with real users

chat`（默认 1h，± jitter）T5（`chat_scan`）MUST 执行：

1. 若无 eligible 未读（`found=false`）→ MUST `RecordTaskRun(chat_scan, false, "无未读会话")` 并结束
4. 若有命中 → 结合 sample 返回的 `messages` 经 `simText`（`chat_reply` prompt）生成 **一条**回复，并 `POST /ucg/internal/api/chat/send`

每 tick MUST 最多回复 **一条**会话。 MUST NOT spawn detached goroutine 或临时聊天窗口循环。 MUST NOT 使用 App `GET /conversations` 扫描全量未读。 peer MUST 为真人（由 ucg sample 的 `peer NOT IN simWxIds` 保证）。

#### Scenario: Single reply per tick

- **WHEN** T5 tick 且 ucg sample 返回一条 eligible 未读
- **THEN** 系统 MUST 发送恰好一条 chat 消息并 MUST NOT 启动后台聊天 goroutine

#### Scenario: Skip when no unread

- **WHEN** 全量 simWxIds 非空但 ucg sample 返回 `found=false`
- **THEN** MUST 记录失败或 skip 语义（非 success 空消息）且 MUST NOT 调用 LLM

#### Scenario: Real peer only via sample

- **WHEN** 未读会话对端 wxId 属于 simWxIds
- **THEN** ucg sample MUST NOT 返回该会话，T5 MUST NOT 回复

#### Scenario: Full sim coverage

- **WHEN** sim 用户总数超过 200
- **THEN** T5 MUST 仍通过 ids 接口覆盖全部 sim 用户，MUST NOT 仅使用前 200 个 id

### Requirement: T6 follow task SHALL follow sim to sim every 7 hours

每 7 小时（±10% jitter）MUST 随机选取一个 sim 用户 A，经 gateway 登录； 调用 ucg internal **`POST /ucg/internal/api/posts/sample`**，`mode=random`，body **`excludeAuthorWxIds`** MUST 为 device internal 拉取的全量 sim wxId 列表； `authorWxId` MUST NOT 等于 A 的 wxId（不等则重试抽样，有界次数后仍失败则记 task 失败）。 MUST NOT 执行 sim→sim 关注； MUST NOT 使用 `pickTwoDistinctSimWx` 或两次 `sim/wx/random` 互关。 已关注 MUST 幂等跳过。 无 eligible 真人 author 时 MUST 记 task 失败（如「无真人作者」），MUST NOT 假 success。

#### Scenario: Sim follows real author

- **WHEN** 存在 published 帖且 author 非 sim，A 未关注 B
- **THEN** `POST /ucg/app/api/follow/{B}` MUST 成功或幂等

#### Scenario: No sim to sim follow

- **WHEN** T6 tick 执行
- **THEN** MUST NOT 选取两个 sim 用户互相关注

#### Scenario: No real author available

- **WHEN** 所有 published 帖作者均在 `excludeAuthorWxIds` 中
- **THEN** MUST 记 follow task 失败且 MUST NOT POST follow

### Requirement: sim-user-service MUST use HTTP contracts only

sim-user-service MUST NOT import 或查询 device/ucg 域 DAO。跨服务数据 MUST 经 gateway App API、device internal API、ucg internal API 与 `aimodel` 完成。

#### Scenario: No cross-domain DAO

- **WHEN** 代码评审检索 `internal/dao` import
- **THEN** sim-user-service 包路径下 MUST NOT 出现 device/ucg 业务表 DAO

### Requirement: AI queue full SHALL abort current task tick

当 `aimodel.Acquire` 返回队列满或调用超时时，当前任务 tick MUST 提前结束并记录日志，MUST NOT 阻塞其他任务 goroutine。

#### Scenario: Queue full

- **WHEN** `ErrQueueFull` 在评论任务中发生
- **THEN** 该次 T2 执行 MUST 结束且不重试至下一周期

### Requirement: sim task intervals SHALL be overridable via environment variables

各背景任务周期 MUST 支持环境变量覆盖，未设置或非法值时 MUST 回退下列默认值：`SIM_INTERVAL_REGISTER=24h`、`SIM_INTERVAL_COMMENT=6h`、`SIM_INTERVAL_POST_IMAGE=3h30m`、`SIM_INTERVAL_POST_VIDEO=6h30m`、`SIM_INTERVAL_POST_DEBATE=12h`、`SIM_INTERVAL_DEBATE_COMMENT=1h`、`SIM_INTERVAL_CHAT=1h`、`SIM_INTERVAL_FOLLOW=7h`。周期执行 MUST 保留 ±10% jitter。

#### Scenario: Default intervals preserved

- **WHEN** 未设置任何 `SIM_INTERVAL_*` 环境变量
- **THEN** T1–T8 名义周期 MUST 与上表一致

#### Scenario: Custom debate comment interval

- **WHEN** `SIM_INTERVAL_DEBATE_COMMENT=30m` 且服务已启动
- **THEN** T8 两次成功执行间隔 MUST 约为 30m（含 jitter）

### Requirement: P1 video poll SHALL use adaptive idle and active intervals

P1 MUST 根据是否存在 `pending`/`processing` 的 `sim_video_job` 选择下一等待间隔：`SIM_INTERVAL_VIDEO_POLL_IDLE`（默认 `10m`）与 `SIM_INTERVAL_VIDEO_POLL_ACTIVE`（默认 `2m`）。无 pending job 时 MUST NOT 调用智谱轮询或 UCG 发帖。

#### Scenario: Idle backoff

- **WHEN** `sim_video_job` 无 pending/processing 行
- **THEN** 下一次 P1 唤醒间隔 MUST 使用 idle 间隔（默认 10m）

#### Scenario: Active polling

- **WHEN** 存在至少一条 pending/processing job
- **THEN** 下一次 P1 唤醒间隔 MUST 使用 active 间隔（默认 2m）且 MUST 执行智谱 `PollVideoGeneration`

### Requirement: sim outbound App HTTP SHALL be globally rate limited

经 gateway-app 的 `appGet`/`appPost`/`appPut` MUST 经全局限速器；默认 `SIM_UCG_RATE_LIMIT_RPS=2`（token bucket，burst 默认 4）。限速 MUST 在发起 HTTP 前阻塞等待，MUST NOT 静默丢弃请求。

#### Scenario: Burst within limit

- **WHEN** 2 秒内发起不超过 4 次 App API 调用且 RPS=2
- **THEN** 调用 MUST 成功发出（不因限速失败）

#### Scenario: Over limit waits

- **WHEN** 持续超过 RPS 限额发起调用
- **THEN** 超额调用 MUST 等待至许可可用后再发送

### Requirement: scheduler SHALL stagger first task ticks after startup

`SIM_USER_SERVICE_ENABLED=true` 启动 scheduler 时，各任务 goroutine 在首次执行前 MUST 额外等待 `0` 至 `SIM_STARTUP_STAGGER_MAX`（默认 `30m`）的均匀随机延迟，以避免多任务同时首次齐射。

#### Scenario: Staggered startup

- **WHEN** 服务启动且启用全部任务
- **THEN** 各任务首次 tick 时间 MUST NOT 全部相同（在 stagger 窗口内随机分布）

### Requirement: sim chat temperature constants SHALL be documented for redeploy tuning

任务 temperature 常量 MUST 附中文注释说明业务用途；变更常量 MUST 通过 redeploy `sim-user-service` 生效。首期 MUST NOT 依赖 env 或 sim-admin 配置项。

#### Scenario: Constants live in one file

- **WHEN** 代码评审 sim temperature 定义
- **THEN** 全部任务常量 MUST 位于同一 Go 源文件（如 `task_llm_temp.go`）

### Requirement: sim tasks SHALL pick random simulated user via device random API

各调度任务（T2 评论、T3 图文、T4 视频、E1 聊天、T6 关注等）当需要随机模拟用户时，sim-user-service MUST 经 device internal **`GET /device/internal/api/sim/wx/random`** 取得单条 `{wxId, account}` 并完成 gateway 登录；MUST NOT 调用 `sim/wx/list` 拉取分页列表后在内存随机；MUST NOT 对同一选取流程重复请求 list。T6 关注 MUST 选取两个不同 wxId（在仅 1 个 sim 用户时 MUST 失败「sim 用户不足」）。

#### Scenario: Comment task uses random pick once

- **WHEN** `RunCommentTask` 需要随机 sim 用户
- **THEN** MUST 仅一次 random 调用取得 account 并登录，MUST NOT 先 list 再 random index

#### Scenario: Follow picks two distinct users

- **WHEN** `RunFollowTask` 且 simulated 用户数 ≥ 2
- **THEN** MUST 经 random（或等价 device 有界选取）得到两个不同 wxId 并完成关注

#### Scenario: No sim users

- **WHEN** random 返回无用户
- **THEN** 任务 MUST 失败且错误语义与现网「无模拟用户」一致

#### Scenario: Count still uses list total

- **WHEN** T1 注册判断 `maxSimUsers` 上限
- **THEN** MAY 继续使用 list `pageSize=1` 读取 `total`，MUST NOT 为此拉取 pageSize=200 全量

### Requirement: sim T4 video upload SHALL use ucg internal transcode API

sim-user-service 在上传 T4 视频至 UCG 时 MUST 经 HTTP 调用 ucg-service `POST /ucg/internal/api/media/upload-video`（与 device internal 鉴权方式一致），MUST 使用返回的 `objectKey`、`contentHash` 与 `transformVersion` 调用 App API `RegisterMedia`（`mediaKind=2`、`dedupHit=false`）。首版服务端不再转码时，响应 `transformVersion` MUST 为 `v1`，register MUST 使用同一版本。

MUST NOT 再使用 `transformVersion=sim-raw` 或 presign 直传未验真字节。MUST NOT 假设 internal 接口会将任意片源转码为 v2；T4 源片 MUST 已满足 ucg v1 验真规则。

#### Scenario: T4 registers after internal validate-and-put

- **WHEN** internal 上传成功且响应含 `transformVersion=v1`
- **THEN** 后续 register MUST 使用 `transformVersion=v1` 且 MUST 通过 ucg 侧 v1 验真

#### Scenario: Internal upload failure skips job

- **WHEN** internal 上传因验真失败或其它错误失败
- **THEN** T4 流水线 MUST 将 job 标为 skipped 且 MUST NOT 使用 presign 回退直传

### Requirement: T4 video pipeline SHALL be globally single-flight

任意时刻 sim-user-service 进程内 MUST 最多一条进行中的 T4 视频流水线（submit + poll + post）。`videoPostInFlight`（或等价机制）为 true 时：

- 调度 tick MUST skip 新 submit（MAY 记 success + 说明「video poll in progress」）
- 手动 `POST .../tasks/post_video_submit/run` MUST 拒绝或返回「任务正在执行中」（与 `manualBusy` 语义一致）

流水线结束（成功/失败/超时）后 MUST 清除 inFlight。

#### Scenario: Scheduler skips while poll active

- **WHEN** 上一 T4 流水线仍在轮询且调度 tick 到达
- **THEN** MUST NOT 再次 SubmitVideoGeneration

#### Scenario: Manual run rejected while poll active

- **WHEN** 管理员在流水线进行中再次点击 T4 手动执行
- **THEN** API MUST 返回任务忙错误

### Requirement: sim-user-service SHALL discard pending video jobs on startup

进程 scheduler 启动前（或等效启动钩子）MUST 将 `sim_video_job` 中 `status IN ('pending','processing')` 更新为 `skipped`。MUST NOT 为遗留 job 恢复轮询 goroutine。

#### Scenario: Startup clears stale jobs

- **WHEN** sim-user-service 重启且 DB 存在 pending job
- **THEN** 启动后这些 job MUST 为 skipped 且无恢复轮询

### Requirement: sim-user-service SHALL persist simulated user credentials for admin and task login

`sim-user-service` MUST 在 `SIM_DB_LINK` 库维护表 `sim_wx_credential`（或语义等价名），字段至少含：

- `wx_id`（BIGINT PRIMARY KEY）
- `account`（VARCHAR，非空）
- `password_plain`（VARCHAR，非空）
- `created_at`（BIGINT Unix 秒，非空）

T1 MUST 在 **UCG profile PUT 成功之后** INSERT credential（`created_at` 为 profile 写入完成时刻）。MUST NOT 在 profile 完成前写入 credential。`EnsureSchema` MUST 幂等创建该表。

注销 sim 用户时 MUST DELETE 对应 `wx_id` 行。MUST NOT 在响应日志中打印 `password_plain`。

#### Scenario: Credential written after profile success

- **WHEN** T1 profile PUT 成功完成
- **THEN** `sim_wx_credential` MUST 含对应 wx_id 与非空 `password_plain`

#### Scenario: Credential not written on profile failure

- **WHEN** T1 在 profile PUT 之前失败并已回滚 wx
- **THEN** MUST NOT 存在该 wx_id 的 credential 行

#### Scenario: Credential removed on admin deactivate

- **WHEN** admin deactivate 成功删除 wxId=2001
- **THEN** credential 行 MUST 不存在

### Requirement: sim task random user login SHALL use per-wxId credential password

当 T2、T3、T4、T5、T6 或手动任务需经 gateway `username_login` 登录某一 sim 用户时，sim-user-service MUST 按该用户 `wxId` 从 `sim_wx_credential` 读取 `password_plain` 作为登录密码。

若 credential 无记录（历史用户），MUST fallback 到运行时 `defaultPassword`（yaml/env `simUser.defaultPassword`，空则 `123456`）。

MUST NOT 假定所有 sim 用户共用同一注册密码（除上述历史 fallback）。

#### Scenario: Task login uses stored password

- **WHEN** T2 随机选中 wxId=2001 且 credential 存在
- **THEN** `username_login` MUST 使用 credential 中的 `password_plain`

#### Scenario: Legacy sim login fallback

- **WHEN** 随机选中历史 wxId 无 credential 行
- **THEN** login MUST 使用 defaultPassword fallback

### Requirement: T1 register SHALL rollback sim wx on profile setup failure

T1（`RunRegisterTask`，含 sim-admin 手动触发）在 device internal sim 注册成功获得 `wxId` 后，若 **未完成** UCG profile 写入（`PUT /ucg/app/api/profile/me` 含 nickname 与 avatarKey），则 MUST 视为注册失败。 下列任一步失败时 MUST 调用 device internal `POST /device/internal/api/sim/wx/{wxId}/deactivate` 回滚 wx（须 `is_simulated=1`），并 MUST 删除 sim 库 `sim_wx_credential` 中对应行（若存在）：

- `usernameLogin` 失败
- Prompt 加载失败
- 昵称 `simText` 调用失败，或生成 nickname trim 后为空
- 头像 `simImageGen` 调用失败，或返回 URL 为空
- 头像 upload / media 链路失败
- profile.

#### Scenario: Nickname AI failure rolls back wx

- **WHEN** T1 已 simRegister 成功但 `simText` 昵称生成返回错误
- **THEN** 系统 MUST 注销该 wxId 且 MUST NOT 保留 `sim_wx_credential` 行

#### Scenario: Avatar AI failure rolls back wx

- **WHEN** T1 昵称生成成功但 `simImageGen` 失败
- **THEN** 系统 MUST 注销该 wxId

#### Scenario: Empty nickname rolls back wx

- **WHEN** T1 昵称 AI 返回仅空白字符
- **THEN** 系统 MUST 视为失败并 MUST 注销该 wxId

#### Scenario: Profile PUT success commits registration

- **WHEN** T1 profile PUT 成功且 nickname 非空、avatarKey 非空
- **THEN** 系统 MUST 写入 `sim_wx_credential` 且 MUST NOT 回滚 wx

#### Scenario: Failed registration does not consume cap slot

- **WHEN** T1 因头像 AI 失败回滚 wx 且当前 sim 用户数为 99、`maxSimUsers=100`
- **THEN** 下一次 T1 MUST 仍可尝试注册新用户

### Requirement: T7 post_debate task SHALL publish debate posts every 12 hours

每 12 小时（±10% jitter，可由 `SIM_INTERVAL_POST_DEBATE` 覆盖）MUST 随机选取一个 sim 用户，经 gateway 登录，经 `LaneSimText`（prompt `post_debate_text`）生成 JSON：`content`（话题正文）、`debateLeft`、`debateRight`（各 ≤5 字，与 UCG `validateDebateLabel` 一致），并 `POST /ucg/app/api/posts`，body MUST 含 `content`、`debateLeft`、`debateRight`、`mediaType=0`、`submit=true`，且 MUST NOT 含 `media`。标签校验失败或 LLM 失败 MUST 记 `post_debate` task 失败。

#### Scenario: Debate post submitted for audit

- **WHEN** LLM 输出合法且 POST 成功
- **THEN** 帖子 MUST 为 `type=debate`（或等价双填标签）且 MUST 进入 `pending_audit`

#### Scenario: Invalid debate labels fail task

- **WHEN** LLM 输出任一方立场标签超过 5 字或为空
- **THEN** MUST 记 `post_debate` 失败且 MUST NOT POST

#### Scenario: Default interval twelve hours

- **WHEN** 未设置 `SIM_INTERVAL_POST_DEBATE` 且 DB 无覆盖
- **THEN** T7 名义周期 MUST 为 12h（含 jitter）

### Requirement: T8 debate_comment task SHALL vote and post argument every 1 hour

每 1 小时（±10% jitter，可由 `SIM_INTERVAL_DEBATE_COMMENT` 覆盖）MUST 随机选取一个 sim 用户，经 gateway 登录，调用 ucg internal **`POST /ucg/internal/api/posts/sample`**，`mode=random`，**`onlyDebate` MUST 为 `true`**，取得单条辩论帖（含 `debateLeft`、`debateRight`）。 若 `authorWxId` 等于当前 sim 用户 wxId，MUST 记失败（如「不可评自己的帖」）且 MUST NOT 投票或评论。 否则 MUST 经 `LaneSimText`（prompt `debate_comment`）生成 JSON：`side`（`left` 或 `right`）、`argument`（utf8 长度 ≤10），MUST 先 `POST /ucg/app/api/posts/{id}/vote` 再 `POST.

#### Scenario: Vote before comment

- **WHEN** T8 tick 且 sample 命中他人辩论帖
- **THEN** MUST 先 POST vote 成功再 POST comment

#### Scenario: Skip own debate post

- **WHEN** sample 返回帖 `authorWxId` 等于当前 sim wxId
- **THEN** MUST NOT POST vote 或 comment

#### Scenario: Argument length bound

- **WHEN** LLM 输出 argument 超过 10 字
- **THEN** 实现 MUST 截断至 10 字或记失败，且 POST 的 comment MUST NOT 超过 10 字

#### Scenario: Default interval one hour

- **WHEN** 未设置 `SIM_INTERVAL_DEBATE_COMMENT` 且 DB 无覆盖
- **THEN** T8 名义周期 MUST 为 1h（含 jitter）

#### Scenario: No debate posts available

- **WHEN** `onlyDebate=true` 且 sample 返回空 `list`
- **THEN** MUST 记 `debate_comment` 失败且 MUST NOT 调用 LLM

### Requirement: sim task intervals SHALL include T7 and T8 defaults

`SIM_INTERVAL_POST_DEBATE` 默认 MUST 为 `12h`；`SIM_INTERVAL_DEBATE_COMMENT` 默认 MUST 为 `1h`。对应 env 开关 `SIM_TASK_POST_DEBATE_ENABLED`、`SIM_TASK_DEBATE_COMMENT_ENABLED` 迁移期默认 SHOULD 为 true（与 DB runtime_json 合并后生效）。

#### Scenario: T7 T8 env defaults

- **WHEN** 未设置 T7/T8 interval env
- **THEN** LoadRuntimeFlags MUST 回退 12h 与 1h

---

### Requirement: Service DAO access SHALL use only default database

Service processes MUST access database via their own `database.default` connection and MUST NOT rely on multi-group routing fallback logic.

#### Scenario: Domain DB group resolver removed
- **WHEN** checking DAO infrastructure files
- **THEN** `internal/dao/domain_db.go` multi-group resolver MUST be removed

#### Scenario: Service reads only local default connection
- **WHEN** a service executes DAO operations
- **THEN** the resolved DB connection MUST come from the service-local `database.default` config

---

### Requirement: 核心领域值类型化枚举

系统 SHALL 为 `target_type`、`mode`、状态机状态、`event type` 提供类型化枚举定义，并通过统一常量与解析函数替代散落裸字符串匹配。

#### Scenario: 调用层使用枚举分支
- **WHEN** 业务代码需要按 `target_type` 或 `mode` 分支处理
- **THEN** 代码 MUST 使用枚举类型与常量进行判断，而不是直接比较裸字符串

#### Scenario: 非法值解析失败
- **WHEN** 输入字符串无法映射为合法枚举值
- **THEN** 系统 MUST 返回明确错误并记录可观测日志

### Requirement: 枚举与字符串双向兼容

系统 SHALL 提供枚举到字符串、字符串到枚举的双向转换能力，保证现有 DB 与消息协议字符串格式兼容。

#### Scenario: 入站字符串转换为枚举
- **WHEN** 系统从 DB/MQ/HTTP 读取字符串字段
- **THEN** 系统 MUST 通过统一 Parse 方法转换为枚举值后参与业务判断

#### Scenario: 出站枚举保持原协议字符串
- **WHEN** 系统写入 DB 或发布消息
- **THEN** 系统 MUST 输出与历史协议兼容的字符串值

---

### Requirement: Feed Redis 键 SHALL 经 cachekit builder 登记且禁止业务层字面量

UCG Feed 读写的 Redis 键 MUST 在 `internal/platform/cachekit/keys_ucg. go` 登记下列 builder，业务与 controller MUST 经 `cachekit. Cache`（含 `WithObserver`）访问，MUST NOT 使用 `g.

#### Scenario: 新增键仅经 platform 登记

- **WHEN** 实现需读写 Feed GEO 索引
- **THEN** 代码 MUST 调用 `UCGFeedGeoKey()` 且 MUST NOT 在 `internal/services/**` 出现 `ucg:feed:geo` 字面量

### Requirement: 帖子 snapshot SHALL 缓存 Feed 展示所需字段及 server-only 坐标

`ucg:post:snapshot:{postId}` JSON MUST 含客户端 Feed 所需字段：`id`、`content`、媒体 CDN URL、`authorWxId`、`likeCount`、`ipLocation`、`publishedAt`、`mediaType`。MAY 含 server-only `lat`、`lng` 供距离计算；该坐标 MUST NOT 出现在 App HTTP 响应体。

帖子 published 或更新时 MUST 写入/刷新 snapshot；unpublished/delete MUST DEL 键。

#### Scenario: snapshot 命中时 Feed 不查帖表

- **WHEN** `ListRecommendFeed` 组装 postId 列表且 snapshot 存在
- **THEN** 系统 MUST 从 Redis JSON 填充 `UcgPostItem` 字段且 MUST NOT 对该帖执行 `ucg_post` 单条 SELECT

#### Scenario: 坐标仅服务端使用

- **WHEN** snapshot 含 `lat`/`lng` 且 API 返回帖子
- **THEN** 响应 JSON MUST NOT 含 `lat`/`lng` 字段

### Requirement: 作者 profile snapshot SHALL 缓存公开 profile 字段

`ucg:profile:snapshot:{wxId}` JSON MUST 含 `wxId`、`nickname`、`bio`、`avatarUrl`、`avatarThumbnailUrl`。帖子 publish/update 时 MUST 刷新作者 snapshot；Feed 读路径 MUST 批量 GET snapshot 填充 `author`。

#### Scenario: Feed 作者信息来自 snapshot

- **WHEN** Feed 返回帖含 `authorWxId`
- **THEN** `author` 字段 MUST 优先来自 `ucg:profile:snapshot:{wxId}` 且 MUST NOT 在循环内逐条查 `ucg_profile`

### Requirement: 用户点赞 SET SHALL 支撑 likedByMe 批量判定

`ucg:user:{wxId}:liked-posts` SET member MUST 为 `postId` 字符串。like MUST SADD；unlike MUST SREM。Feed 读路径对当前页 postIds MUST 批量 SISMEMBER（pipeline）填充 `likedByMe`，MUST NOT 查 `ucg_post_like` 表。

#### Scenario: Feed likedByMe 走 Redis

- **WHEN** 已登录用户请求推荐 Feed 且 page 含 20 帖
- **THEN** 系统 MUST 经一次 pipeline SISMEMBER 判定 liked 状态且 MUST NOT 对每帖查询 MySQL like 表

### Requirement: Feed session SET SHALL 防止 cursor 分页重复下发

`ucg:feed:session:{sessionId}` MUST 为 SET，member 为已返回 `postId`；TTL MUST 为 30min（可配置）。每次 Feed 页返回前 MUST SADD 本页 postIds；候选集 MUST 排除 session 中已见 postId。

#### Scenario: 同 session 翻页无重复

- **WHEN** 客户端用合法 `nextCursor` 连续请求两页且 session 未过期
- **THEN** 两页 `list` 中 postId MUST 互斥（无交集）

#### Scenario: 刷新生成新 session

- **WHEN** 客户端请求 Feed 且不传 `cursor`
- **THEN** 系统 MUST 生成新 `sessionId` 且 MUST NOT 复用旧 session 的 seen SET

---

### Requirement: ucg-service SHALL run as dedicated microservice

The platform SHALL provide `ucg-service` process listening on `UCG_SERVICE_ADDR` default `:9804`, loading `manifest/config/config.ucg-service.yaml` when `GF_GCFG_FILE` is unset, mirroring `history-service` startup pattern.

#### Scenario: 启动与配置隔离
- **WHEN** 启动 ucg-service 且未设置 `GF_GCFG_FILE`
- **THEN** 进程 SHALL 加载 `config.ucg-service.yaml`，且 default DB SHALL 指向 `ai_voice_ucg`

#### Scenario: 依赖检查失败不监听
- **WHEN** MySQL 或 Redis 不可用且 fail-fast 启用
- **THEN** 进程 SHALL 退出且 SHALL NOT 进入监听态

---

### Requirement: 分发前必须执行路由键白名单校验

系统 MUST 在 outbox 发布与投影分发入口先执行路由键合法性校验；未注册路由键必须被拒绝并记录拒绝来源。

#### Scenario: 收到未注册路由键
- **WHEN** outbox 处理到不在注册表中的 `routing_key`
- **THEN** 系统必须拒绝该事件并输出包含来源模块的告警日志

### Requirement: 校验通过后必须按前缀分组分发

系统 SHALL 在路由键通过合法性校验后，基于前缀常量将事件分发到对应领域处理器，而非依赖逐项路由键枚举分支。

#### Scenario: 处理 history.record 前缀事件
- **WHEN** 事件路由键为 `history.record.*` 且已通过白名单校验
- **THEN** 系统必须将该事件分发给 history 投影处理器

### Requirement: 必须提供未知前缀默认保护

系统 MUST 为已注册但未映射分发处理器的前缀保留默认保护分支，避免静默忽略。

#### Scenario: 路由键合法但前缀未绑定处理器
- **WHEN** 事件通过白名单校验但其前缀没有配置分发处理器
- **THEN** 系统必须输出告警并按既定失败语义处理

---

### Requirement: Worker-service SHALL use dedicated configuration

`worker-service` MUST have a dedicated default config file and MUST load it when `GF_GCFG_FILE` is not explicitly provided.

#### Scenario: Worker starts without GF_GCFG_FILE
- **WHEN** `worker-service` starts and `GF_GCFG_FILE` is empty
- **THEN** runtime MUST default to `manifest/config/config.worker-service.yaml`

#### Scenario: Deployment manifest uses worker dedicated config
- **WHEN** compose/kustomize/dockerfile defines worker runtime env
- **THEN** worker `GF_GCFG_FILE` MUST point to `manifest/config/config.worker-service.yaml`

---

### Requirement: worker-service 后台任务角色 MUST 已废止

`worker-service` 进程 MUST NOT 部署；原由其独占的 domain outbox relay、voice task consumer 等后台任务 MUST NOT 存在。各域经 OpenSpec 批准的后台任务 MUST 在对应业务进程内启动（见 `background-loop-task-governance`）。

#### Scenario: 部署清单无 worker-service

- **WHEN** 审查 docker-compose 或 K8s manifest
- **THEN** MUST NOT 包含 `worker-service` 容器或等价部署单元

---

### Requirement: Apply 后包边界与 Redis 门禁仍生效

实现变更 MUST NOT 在 internal/services/** 互相 import 业务实现或直连 g.Redis()；跨域出站 MUST 经 internal/clients/*，Redis KV/PubSub MUST 经 cachekit/
edismsgkit（可用 hack/check-service-import 与 hack/check-redis-bypass 复核）。 The system SHALL keep platform import and Redis bypass gates mandatory after apply.

#### Scenario: 新增跨域 HTTP 调用

- **WHEN** voice 需要读取 device 数据
- **THEN** 实现 MUST 经 clients/device（或既有薄封装），MUST NOT import services/device 仅为拿 Remote

#### Scenario: 业务层禁止直连 Redis

- **WHEN** 评审检索 g.Redis() 于 internal/services 与 internal/controller
- **THEN** 期望命中为 0（platform 包除外）

