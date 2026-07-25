# 仓库级 AI 执行约定

## 文档语言
- OpenSpec 相关文档（proposal/design/specs/tasks）默认使用中文说明性文本。
- 环境变量、路径、接口、协议关键字、代码符号可保留英文原文。

## 代码注释
- **强制要求**：所有生成的代码必须包含详细的业务逻辑中文注释
- **注释层级**：
  - 文件级：模块说明、业务说明、设计思路、使用场景
  - 类级：业务说明、设计思路、核心属性
  - 方法/函数级：业务逻辑、Args、Returns、Side Effects
  - 行级：关键业务逻辑分支必须有中文注释
- **注释语言**：中文

## 服务边界（强制）
- 跨服务数据访问必须走服务接口契约（HTTP/RPC/事件），禁止在服务实现中直接访问他域 DAO 或数据库表。
- 迁移期若使用 `local|remote|canary` 双路径，必须保留显式配置与 failover 语义，并输出可观测日志。
- 代码评审时需显式检查：是否出现跨库直查、是否补齐契约路径与错误语义。
- `internal/services/history` 不得 import 或调用 `dao.User`、`dao.Event`、`dao.Suggest` 等他域表；history/device 读模型缓存由写路径同步 patch + 读 miss 回源 MySQL 保证，禁止跨库直查他域表。
- `internal/services/voice` 不得 import `hello/internal/dao` 中 **user/event/action** 等他域表 DAO；设备域访问 MUST 经 `voice.DeviceAdmin()`（HTTP 实现）或经批准的契约，禁止在 voice 进程内直连 device 库表。评审可检索：`grep -r \"internal/dao\" internal/services/voice`（应仅出现 suggest/qa 等本域表）。
- 服务默认配置必须按进程独立（`gateway`/`voice-service`/`device-service`/`history-service`），禁止回退到共享主配置承载他域业务项。
- `manifest/config/config.yaml` 仅允许保留网关与全局公共配置；评审时必须检查是否有 voice/device/history 专属字段回流。
- `voiceChat`（ASR/LLM/TTS/会话）必须维护在 `manifest/config/voice-chat.shared.yaml`，供 `voice-service` 与 `history-service` 共用；禁止在 `config.voice-service.yaml` 与 `config.history-service.yaml` 中重复整段 `voiceChat`（迁移期可仅依赖 `GF_GCFG` 中的 `voiceChat` 作兜底，不作为长期双源）。
- 代码目录边界与包边界必须一致：业务实现统一位于 `internal/services/**`，禁止新增实现文件到 `internal/service`。
- 评审时必须检查 import 路径，确保 `cmd`/`controller` 不再依赖 `hello/internal/service` 旧路径。
- 已有接口版本（v1、v2、v3…）永远不可修改结构；任何新增字段或逻辑变更，必须创建更高版本（v+1）接口，原版本代码保持完全不变，保证历史前端接口请求不受影响。
- 新增需求必须以最少的db操作来考虑实现方案。

## 测试文件
- 当前阶段不新增测试文件（包括 `*_test.go`、`*.spec.*`、`*.test.*`）。

## gateway-app 对外 App 接口（OpenSpec / 实现强制）

新增经 **gateway-app-server** 对外暴露的 App HTTP 接口（含 `api/v1`、`api/v2` 的 `g.Meta` 路由）时，**MUST** 在 propose/apply 与 PR 自检中逐项确认，避免「仅在领域服务注册、网关未放行」：

| 检查项 | 位置 / 说明 |
|--------|-------------|
| **领域服务路由** | 对应 `*-service` 的 `controller` + `api/v*` 的 `g.Meta`（path/method/summary） |
| **gateway 反代** | `installUcgProxyMiddleware` / `installDeviceProxyMiddleware` 等；UCG App 路径 `/ucg/app/api/*` 已 fuzzy 覆盖 v2 子路径，**无需**为 v2 单独 Bind，但 **MUST** 确认 path 前缀落在已绑定模式内 |
| **Bearer 白名单** | `internal/controller/gateway_app_auth_exempt.go`：若 v1 同语义接口可匿名访问，**v2 MUST 同步**添加精确 path（如 v1 `feed/recommend` ↔ v2 `v2/feed/recommend`） |
| **usage 统计** | 见下节；维护型排除见 `usagestats/maintenance_skip.go` |
| **apiregistry** | `api/v2` 路由须含 `g.Meta`，由 `apiregistry.Init` 自动加载 summary |

细则见 **`openspec/project.md`**「gateway-app 对外 App 接口约定」。

## App API 使用统计（OpenSpec / 实现强制）
- 新增经 **gateway-app** 对外的 **App HTTP 接口**时，**必须先向负责人询问是否计入 usage 统计**；未获明确答复前不得修改 `usagestats/maintenance_skip.go` 或假定统计策略。
- 细则见 **`openspec/project.md`**「App API 使用统计约定」；维护型排除列表在 **`internal/services/gatewayapp/usagestats/maintenance_skip.go`**。

## 背景循环任务（OpenSpec / 实现强制）
- 默认禁止在 `internal/services/**` 新增循环后台任务（`time.NewTicker`、扫 outbox/表 reconciler、HTTP Pull 轮询队列等）；新增 MUST 有 OpenSpec proposal/design 明确批准（任务名、宿主进程、周期、开关、失败语义）。
- RabbitMQ **broker push** consumer（`autoAck=false`）不视为 ticker 扫表，但仍须在变更中声明队列与宿主进程。
- 细则见 **`openspec/project.md`**「背景循环任务约定」；已批准的 UCG outbox relay / AMQP consumer 保留在 `ucg-service`。

## Redis 访问（OpenSpec / 实现强制）
- **KV 缓存/状态**：`internal/services/**` 与 `internal/controller/**` MUST 经 `cachekit.Cache` 访问 Redis，且 MUST 使用 `cachekit.WithObserver(..., cachekit.LoggingObserver{})` 或 `cachekit.Default()` 包装；**禁止**直接 `g.Redis()` / `g.Redis().Do(...)`。
- **Pub/Sub 消息**：PUBLISH / SUBSCRIBE MUST 经 `internal/platform/redismsgkit`，同样 MUST `WithObserver` 或 `DefaultPublisher()` / `StartSubscriber`；**禁止**业务层直连 PUBLISH 或使用 `github.com/redis/go-redis/v9` 创建客户端。
- **允许直连 `g.Redis()` 的位置**：仅 `internal/platform/cachekit/**`、`internal/platform/redismsgkit/**`、`internal/platform/rediscfg/**`。
- 评审 grep：`rg 'g\.Redis\(\)' internal/services internal/controller --glob '*.go'` 期望 **0**；`rg 'redis\.New(Client|ClusterClient)' internal/services internal/controller` 期望 **0**。可运行 **`hack/check-redis-bypass.sh`** 自动化检查。

## Redis 键命名（OpenSpec / 实现强制）
- Redis 键与 Pub/Sub 频道 **禁止**在 `internal/services/**`、`internal/controller/**` 以字面量或 `fmt.Sprintf("domain:...")` 拼写。
- **必须**使用 `internal/platform/cachekit/keys_*.go` 与 `internal/platform/redismsgkit/channels.go` 中的 builder；新增键 ONLY 在 platform 登记并附中文注释（TTL、失效、跨进程共享语义）。
- 与「Redis 读缓存约定」关系：命名规范管「怎么用键」；读缓存约定管「要不要加 Redis」。

## Redis 读缓存（OpenSpec / 实现强制）
- 新增或改造业务读路径时，**不得默认引入 Redis**；拟引入**新的** Redis 读缓存须先做收益率评估并**向负责人确认**；**负责人已确认要加的，实现阶段不得省略**。
- 沿用 **`cachekit`** 等同族既有模式时，在 design 说明即可；Redis 持久化格式与 HTTP 边界映射须与现有约定一致。
- 细则见 **`openspec/project.md`**「Redis 读缓存约定」。
