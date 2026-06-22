## ADDED Requirements

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
