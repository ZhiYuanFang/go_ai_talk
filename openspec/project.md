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
- 配置边界必须与服务边界一致：`voice-service`、`device-service`、`history-service`、`gateway-service` 使用各自默认配置文件。
- 配置边界必须与服务边界一致：`voice-service`、`device-service`、`history-service`、`worker-service`、`gateway-service` 使用各自默认配置文件。
- 主配置 `manifest/config/config.yaml` 仅承载网关与全局公共项，禁止承载 voice/device/history 领域专属配置。
- 微服务架构遵循“一服务一数据库”原则；每个服务配置文件仅允许维护本服务的 `database.default`。
- 当服务需要他域数据时，必须通过跨服务 API 获取，禁止跨库直查。
- 跨服务数据访问必须走服务契约（HTTP/RPC/事件），禁止在服务内直接访问他域 DAO/数据表。
- 若迁移期保留 `local|remote|canary` 双路径，必须显式配置 failover 语义并记录命中日志，避免隐式跨库回流。

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
- 后续 AI 发起任何新变更前，必须先读取并对照 `openspec/specs/**/spec.md` 历史基线规格，再生成 proposal/design/tasks。
- 若本次需求涉及已有 capability，必须在 proposal 中明确标注“复用/变更了哪些已有 spec 边界”；禁止脱离历史 spec 直接重写同类能力。
- AI 在实现阶段（apply）必须以历史 spec 的 Requirement/Scenario 作为验收约束，若实现与历史 spec 冲突，需先更新变更规格再改代码。
- AI 在评审或总结阶段必须说明“本次改动对应的 spec 依据”，至少列出相关 capability 路径。
- 若行为发生变更而未同步到 specs，不允许归档为完成状态（除非明确记录 `--skip-specs` 原因并经人工确认）。

## 重要约束
- Redis 与 RabbitMQ 作为关键依赖，允许在运行时采用失败即显式报错策略。
- 对于依赖不可用场景，优先采用启动期/请求期明确失败，不引入隐式降级分支。
- 评审检查项必须包含“是否存在跨服务直查数据库”与“契约路径是否完整可观测”两项硬性检查。
- 评审检查项必须包含“主配置是否回流服务专属字段”与“服务默认配置是否仍指向共享主配置”两项硬性检查。
- 评审检查项必须包含“是否存在 `hello/internal/service` 旧导入路径”与“`internal/service` 是否新增实现文件”两项硬性检查。
- 评审检查项必须包含“运行文档是否同步更新”检查：凡涉及运行/发布/DAO 边界变更，必须同时更新 `docs/runbooks/dao-sync-by-domain.md` 与 `docs/runbooks/release-deploy-and-run.md`。
- 评审检查项必须包含“是否引用并遵循历史 `openspec/specs` 基线”检查：涉及行为变更时必须可追溯到对应 Requirement/Scenario。
- 在没有特别要求的情况下，不用生成关于当前变更需求的md文件。当有要求生成文档时，文档必须在`docs/`文件夹内生成md文件，不要生成到`docs/runbooks/`

## 外部依赖
- Redis
- RabbitMQ
- （按环境）容器编排与网关入口能力
