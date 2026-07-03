# 项目上下文

## 项目目标
本项目用于构建智能语音与设备管理相关能力，并持续从单体架构演进为以 gateway + 领域服务为核心的微服务平台。当前重点是提升代码可读性、运行时一致性与服务边界清晰度。

## 技术栈
- Go（Golang）
- GoFrame（HTTP 服务、路由、中间件、配置）
- Redis（缓存、幂等等运行时状态）
- RabbitMQ（跨服务事件消息）
- Docker / Kustomize（部署与环境编排）
- OpenSpec（变更提案、设计、任务与规格管理）

## 项目约定

### 文档语言约定（强制）
- 所有后续由 OpenSpec 生成的文档（包括 proposal、design、specs、tasks）必须使用中文撰写。
- 面向团队协作的说明、约束、验收标准、任务拆解均使用中文表述。
- 仅在必须保留英文标识时使用英文（例如环境变量名、接口路径、技术关键字）。

### 代码与架构约定
- 基础设施能力以单一后端为目标：缓存统一走 Redis，跨服务消息统一走 RabbitMQ。
- gateway-service 逐步收敛为流量与策略层，不承载领域业务实现。
- voice、device、history 以领域边界划分服务职责。
- 业务实现代码统一放置于 `internal/services/**`，禁止新增实现文件回流到 `internal/service`。
- 配置边界必须与服务边界一致：`voice-service`、`device-service`、`history-service`、`gateway-service`、`gateway-app-server`、`ucg-service` 使用各自默认配置文件（见 `manifest/config/config.*.yaml`）。
- 主配置 `manifest/config/config.yaml` 仅承载主网关与全局公共项，禁止承载 voice/device/history 领域专属配置，**禁止**出现 `database` 段。
- 微服务架构遵循“一服务一数据库”原则；各业务服务配置文件仅维护本域库，一般以 `database.default` 为主。**例外**：`gateway-app-server` 仅允许配置 **`database.app`**（`ai_voice_app`，用于版本等只读场景），**不得**用其 `default` 直连他域库。
- 当服务需要他域数据时，必须通过跨服务 API 获取，禁止跨库直查。
- 跨服务数据访问必须走服务契约（HTTP/RPC/事件），禁止在服务内直接访问他域 DAO/数据表。
- 若迁移期保留 `local|remote|canary` 双路径，必须显式配置 failover 语义并记录命中日志，避免隐式跨库回流。
- Redis KV 访问 MUST 经 `internal/platform/cachekit` 且 MUST 使用 `cachekit.WithObserver(...)`（或 `cachekit.Default()`）；Redis Pub/Sub MUST 经 `internal/platform/redismsgkit` 且 MUST 使用 `WithObserver` / `DefaultPublisher()`。业务与 controller 层禁止直接 `g.Redis()`。
- Redis 键/频道 MUST 经 platform builder（`cachekit/keys_*.go`、`redismsgkit/channels.go`）构造，禁止业务层键字面量（策略 A：builder 返回值与线上一致，本变更不重命名键空间）。

### 数据库连接与部署实例约定（强制）
- 任意需求变更若**新增、调整或迁移**某进程对 MySQL 的访问，OpenSpec **proposal / design / tasks** 中必须写清：**进程名**、**库与表域**、GoFrame **配置组**（如 `default`、`app`）、以及推荐覆盖方式（yaml 内 `*.link` 或 **`HISTORY_DB_LINK` / `DEVICE_DB_LINK` / `VOICE_DB_LINK` / `APP_DB_LINK` / `UCG_DB_LINK`** 等与 `cmd/*-service/main.go` 一致的环境变量名）。
- **`gateway-app-server`** 连 `ai_voice_app` 时**仅**通过 **`APP_DB_LINK`**（写入 `GF_DATABASE_APP_LINK`）或 `config.gateway-app-server.yaml` 的 **`database.app.link`**；不得与主网关 `gateway-service`（无 DB）混淆。
- **新建或扩容部署实例**（新 Compose 栈、新集群、新环境）时，同一变更必须同步检查并更新 **`manifest/docker/.env.example`**、**`manifest/docker/docker-compose.microservices.yml`**（各服务 `environment` 中的 `${*_DB_LINK:-}` / `${APP_DB_LINK:-}` 等注入是否齐全）与 **`docs/runbooks/release-deploy-and-run.md`**，避免仅改代码或仅改 yaml 占位 DSN，导致实例上**未设置 `*_DB_LINK` / `APP_DB_LINK`** 仍连占位地址或漏配库；K8s 路径须在 overlay 或 Secret 挂载中同等落实。
- **tasks 验收**：若本变更引入或可影响库连接，须勾选或描述「`.env.example` / runbook / 服务主配置顶部注释是否已反映新连接约定」；归档前评审须核对此项。

### 代码注释约定（强制）
- 后续生成或修改代码时，必须尽可能补充中文注释，重点覆盖：复杂流程、跨服务调用、重试/降级/错误语义、关键配置项。
- 公共导出函数、结构体与关键字段优先添加中文注释，保证阅读代码时无需反复跳转上下文。
- 允许保留英文符号名、库 API 与协议字段名，但解释说明必须使用中文。
- 禁止添加“废话注释”；注释应解释“为什么这样做”与“边界条件”，而非逐行复述代码。

### 测试文件约定（强制）
- 在后续代码生成与实现过程中，不允许新建任何测试文件。
- 禁止创建或新增以下文件类型：`*_test.go`、`*.spec.*`、`*.test.*`。
- 若历史测试文件存在，按当前变更策略逐步清理，不作为新实现的交付前置条件。

### OpenSpec 基线参考约定（强制）
- 后续 AI 发起任何新变更前，必须先读取并对照 **`openspec/specs/v2.0.24/spec.md`**（v2.0.24 合并基线规格），再生成 proposal/design/tasks。
- 若本次需求涉及已有 capability，必须在 proposal 中明确标注“复用/变更了哪些已有 spec 边界（见 v2.0.24 目录章节）”；禁止脱离历史 spec 直接重写同类能力。
- AI 在实现阶段（apply）必须以 v2.0.24 基线中的 Requirement/Scenario 作为验收约束，若实现与基线冲突，需先更新变更规格再改代码。
- AI 在评审或总结阶段必须说明“本次改动对应的 spec 依据”，至少列出相关 capability 名称（v2.0.24 目录中的章节名）。
- 新版本发版时可运行 `hack/merge-openspec-specs.py`（或等价流程）合并为 `openspec/specs/vX.Y.Z/spec.md`；若行为发生变更而未同步到 specs，不允许归档为完成状态（除非明确记录 `--skip-specs` 原因并经人工确认）。

## 重要约束
- Redis 与 RabbitMQ 作为关键依赖，允许在运行时采用失败即显式报错策略。
- 对于依赖不可用场景，优先采用启动期/请求期明确失败，不引入隐式降级分支。
- 评审检查项必须包含“是否存在跨服务直查数据库”与“契约路径是否完整可观测”两项硬性检查。
- 评审检查项必须包含“主配置是否回流服务专属字段”与“服务默认配置是否仍指向共享主配置”两项硬性检查。
- 评审检查项必须包含“是否存在 `hello/internal/service` 旧导入路径”与“`internal/service` 是否新增实现文件”两项硬性检查。
- 评审检查项必须包含“运行文档是否同步更新”检查：凡涉及运行/发布/DAO 边界变更，必须同时更新 `docs/runbooks/dao-sync-by-domain.md` 与 `docs/runbooks/release-deploy-and-run.md`；若涉及**新进程或新库连接**，还须核对 **`manifest/docker/.env.example`** 与相关 **`manifest/config/config.*.yaml`** 顶部说明是否已包含对应 **`*_DB_LINK` / `APP_DB_LINK`** 约定（见上文「数据库连接与部署实例约定」）。
- 评审检查项必须包含“是否引用并遵循 **`openspec/specs/v2.0.24/spec.md`** 基线”检查：涉及行为变更时必须可追溯到对应 Requirement/Scenario。
- 评审检查项必须包含“**Redis platform 访问**”检查：业务/controller 是否 bypass `cachekit`/`redismsgkit`（见 `AGENTS.md` 与 `hack/check-redis-bypass.sh`）。
- 评审检查项必须包含“**Redis 读缓存**”检查：涉及新读路径或 Redis 键变更时，是否已在 proposal/design 完成收益率评估、负责人确认结论，且实现与 design 一致（见「Redis 读缓存约定」）。
- 在没有特别要求的情况下，不用生成关于当前变更需求的md文件。当有要求生成文档时，文档必须在`docs/`文件夹内生成md文件，不要生成到`docs/runbooks/`

### gateway-app 对外 App 接口约定（强制）

- **适用范围**：App 客户端经 **gateway-app-server** 访问的 **新增 HTTP 接口**（`api/v1` / `api/v2` 的 `g.Meta`，以及 gateway 上 `BindHandler` 注册的 App 路径）。领域服务（如 `ucg-service`）单独注册路由 **不等于** 已对 App 开放，须完成下列 gateway 侧检查。
- **反代覆盖**：UCG App 流量由 `installUcgProxyMiddleware` 绑定 `/ucg/app/api/*`（fuzzy，**含** `/ucg/app/api/v2/...`）。device/voice/history 等域同理，新增 path **MUST** 落在对应 `install*ProxyMiddleware` 已绑定前缀下；若新增顶层前缀，**MUST** 扩展 proxy 绑定。
- **Bearer 鉴权白名单**：匿名可访问的读接口（如 UCG 推荐 Feed）在 `gateway_app_auth_exempt.go` 维护。**新增 v2 且与 v1 同语义时，MUST 同步登记 v2 精确 path**，不得假设「v1 已豁免则 v2 自动生效」。
- **与 usage 统计、apiregistry 的关系**：本节约束「能否经 gateway 到达」；usage 统计见下节；中文 summary 依赖 `api/v*` 的 `g.Meta` 嵌入 `apiregistry`。
- **proposal / tasks 检查项**：变更新增 App HTTP 路由时，MUST 包含「gateway-app 登记自检」勾选（反代前缀、auth exempt 同步、usage 结论）。

### App API 使用统计约定（强制）

- **适用范围**：经 **gateway-app** 对外暴露、且可能被 App 客户端或运维页调用的 **新增 HTTP 接口**（含 `api/v1` 中 `g.Meta` 路由，以及 `gateway-app-server` 上 `BindHandler` 注册的 App 路径）。**不适用**：`/device/internal/*`、`/device/admin/api/*`、WebSocket 升级、静态资源与 HTML 壳页（已由 `usagestats` 结构性 skip 排除，无需重复询问）。
- **新增接口时必须询问产品负责人**：该接口 **是否计入 App API 使用统计**（运维页「功能使用统计」）。**AI 在 propose / apply 阶段 MUST 向用户确认**，不得默认假设；用户未明确答复前，不得擅自将新接口加入或移出统计 denylist。
- **用户确认「统计」**：确保 `api/v1` 已登记 `path`/`method`/`summary`（供 apiregistry 归一化与中文展示）；**不得**写入 `internal/services/gatewayapp/usagestats/maintenance_skip.go`。
- **用户确认「不统计」**（维护型/探测型/运维型，如 token 刷新、版本探测）：在 **`maintenance_skip.go`** 增加精确 `METHOD + path` 或 path 前缀，并在变更 **proposal/tasks** 中列出排除项。
- **proposal / tasks 检查项**：若变更新增 App HTTP 路由，MUST 包含一条「已与负责人确认是否计入 usage 统计」及结论（统计 / 不统计 + denylist 变更说明）。

### Redis 读缓存约定（强制）

- **默认立场**：业务数据以 **MySQL 为唯一权威**；Redis 用于读加速、临时态、纯统计或幂等等场景，**不是**新读路径的默认依赖。

### 背景循环任务约定（强制）

- **默认禁止**：在 `internal/services/**` 业务实现中新增循环后台任务，包括 `time.NewTicker` / `time.Tick` 轮询、常驻 goroutine 扫描 MySQL/Redis/outbox 表、定时 reconciler、HTTP Management API Pull 轮询队列等。
- **新增 MUST 经 OpenSpec 批准**：proposal 与 design 须写明任务名称、宿主进程、周期/触发条件、环境开关、失败语义、关闭方式，以及为何不能在请求内同步完成或使用 AMQP push consumer 替代。
- **AMQP push consumer 例外**：经 RabbitMQ broker push（`autoAck=false`）的 consumer 不视为 ticker 扫表，但 MUST 在 OpenSpec 变更中声明队列名、routing key 与宿主进程（如 `ucg-service` 审核/推荐 consumer）。
- **评审检查项**：PR 含 `Start*Reconciler`、ticker 扫表或 Pull 消费且无 OpenSpec 变更引用时 MUST 要求补充已批准变更或删除任务。
- **新引入 Redis 读缓存时 MUST 经负责人确认**：变更新增或改造读路径，且拟引入**新的** Redis 键空间、缓存粒度或整页/列表 JSON 快照时，**AI 在 propose / apply 阶段 MUST 向产品负责人完成收益率评估并确认**；未获明确答复前 **MUST NOT** 擅自加 Redis 键、TTL 或失效逻辑。
- **负责人已确认「加 Redis」时 MUST 实现**：proposal/design/tasks 已记录确认结论（含键语义、TTL、失效方式）的，**实现阶段不得**以「默认不用 Redis」为由省略；与 usage 统计约定同理，约束是双向的。
- **proposal / design 收益率评估（新 Redis 读缓存时）**须简要覆盖：
  - **读特征**：是否热路径、读多写少程度（可估 QPS / P99）
  - **共享度**：全站共享 vs 按用户/设备/分页/viewer 个性化
  - **失效复杂度**：写操作触发点、整表重建 vs 细粒度 patch、跨服务共用键语义是否一致
  - **替代方案**：索引、批量查询、SQL 优化是否应先做
  - **建议结论**：加 / 不加 / 延后（及理由）
- **可沿用既有模式、design 说明即可（免重复询问）**：与现有 **`internal/platform/cachekit`** 读模型（如 event/action options、history list + patch + version）、gateway refresh token、voice session、usage 统计、UCG 聊天 LIST 等**同族**的扩展；访问 MUST 经 `cachekit.WithObserver` / `redismsgkit`，键 MUST 经 platform builder；**但**须在 design 中写明沿用哪一模式，且 **Redis 内持久化格式须与 device/history 约定一致**（如 objectKey vs CDN URL 边界映射，见 event logo 教训）。
- **倾向不加 Redis（须强理由 + 负责人确认才可加）**：分页列表（page/filter 组合多）、强个性化（如 `likedByMe`、关注 Feed）、写多读变/审核态频繁、整页 PostDTO 级大 JSON 缓存。
- **倾向可考虑 Redis（仍须负责人确认或 proposal 已写明）**：小体积全站字典、设备维度热读且失效清晰、短 TTL 限流/幂等、会话与 token、纯统计类数据。
- **proposal / tasks 检查项**：若变更新增或改造读路径且涉及 Redis，MUST 包含「已与负责人确认 Redis 策略」及结论（不加 / 沿用既有模式 / 加哪一层 + TTL + 失效）；若 design 已确认加 Redis，tasks 须含对应实现项且不得遗漏。

## 外部依赖
- Redis
- RabbitMQ
- （按环境）容器编排与网关入口能力
