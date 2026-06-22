## Context

- **现状**：合规路径仅 `device/cache_repo`、`device/wx`、`history/cache_repo`、`gatewayapp/refresh_store`、`voice/voice_chat`、`runtimecheck` 等少数模块使用 `cachekit.WithObserver`；其余 ~75 次 `g.Redis().Do` 分散在 ucg chat、usagestats、clinic、aimodel gate 等。
- **键命名**：`cachekit/keys.go` 定义 `domain:module:kind:identifier` 与 device/history 读模型 helper；大量键仍为业务包内 `fmt.Sprintf`（`ucg:chat:*`、`gw:usage:*`、`dev:wx:*`）。
- **Pub/Sub**：`history/realtime_notify.go` PUBLISH；`gatewayapp/history_subscriber.go` 用 `go-redis/v9` SUBSCRIBE（非 g.Redis）。
- **约束**：`openspec/project.md` 第 32 行；用户确认策略 **A**（收拢 builder、**不改键字符串**）；PUBLISH 归 messaging 不归 cachekit。

## Goals / Non-Goals

**Goals:**

- 业务与 controller 层零 `g.Redis()` / 零 Redis 键字面量；platform 层统一观测。
- `cachekit` 覆盖全仓 KV 命令集；`redismsgkit` 覆盖 Pub/Sub。
- 全量键/频道 builder 登记在 platform，附中文注释（TTL、共享域、失效语义）。
- `AGENTS.md` 升格为仓库级 AI/评审强制项；spec 去除 `g.Redis()` 字面量。

**Non-Goals:**

- **键重命名**（`dev→device`、`gw→gateway` 等）留待后续 change。
- 不新增 Redis 读缓存场景或新键空间。
- 不新增 `*_test.go`；不改动 RabbitMQ `eventkit` 路径。
- 不统一 Redis 客户端库（cachekit 继续 GoFrame `g.Redis`；redismsgkit 订阅端保留 go-redis cluster 能力）。

## Decisions

### D1：访问分层（采用）

| 层 | 包 | 用途 |
|---|---|---|
| 配置 | `rediscfg` | 启动前 `GF_REDIS_DEFAULT_ADDRESS` |
| 缓存/状态 KV | `cachekit.WithObserver(NewRedisCache(), LoggingObserver{})` | GET/SET/Hash/List/Set/ZSet/Incr/Eval 等 |
| 消息 Pub/Sub | `redismsgkit.WithObserver(..., LoggingObserver{})` | PUBLISH、SUBSCRIBE |

业务代码 **禁止** 直接 `g.Redis()` 或 `redis.NewClient`。

### D2：typed 方法扩展（采用，无 Raw Do 后门）

在 `cachekit.Cache` 新增（实现于 `RedisCache` + `observedCache`）：

| 方法 | Redis 命令 | 主要调用方 |
|---|---|---|
| `Decr` | DECR | aimodel gate, ai quota |
| `TTL(ctx, key) (int, error)` | TTL | clinic_session |
| `Set(ctx, key, value string)` | SET（无 EX） | clinic_session 边界 |
| `Persist(ctx, key)` | PERSIST | ucg chat |
| `HashIncrBy(ctx, key, field string, delta int64)` | HINCRBY | usagestats |
| `HashSet(ctx, key, field, value string)` | HSET | usagestats |
| `HashGet(ctx, key, field string) (string, bool, error)` | HGET | usagestats |
| `HashGetAll(ctx, key) (map[string]string, error)` | HGETALL | usagestats（内聚 flat `[]string` 解析） |
| `ListPush(ctx, key, value string)` | RPUSH | ucg chat |
| `ListLen(ctx, key) (int64, error)` | LLEN | ucg chat, audit |
| `ListRange(ctx, key, start, end int64) ([]string, error)` | LRANGE | ucg chat |
| `ListIndex(ctx, key, index int64) (string, error)` | LINDEX | audit_chat |
| `ListSet(ctx, key, index int64, value string)` | LSET | audit_chat |
| `SetAdd(ctx, key string, members ...string)` | SADD | profile_audit, sim wx |
| `SetIsMember(ctx, key, member string) (bool, error)` | SISMEMBER | simulated_wx |
| `SetMembers(ctx, key) ([]string, error)` | SMEMBERS | profile_audit |
| `SetRemove(ctx, key string, members ...string)` | SREM | profile_audit |
| `SortedSetAdd(ctx, key string, score float64, member string)` | ZADD | ucg chat |

已有方法继续用于：`Get/MGet/SetEX/SetNXEX/Exists/Del/Incr/IncrBy/Expire/Eval/Ping`。

可选便利：`cachekit.Default() Cache` 返回包级 `WithObserver(NewRedisCache(), LoggingObserver{})` 单例，减少各文件重复声明。

### D3：redismsgkit（采用）

```
internal/platform/redismsgkit/
  publisher.go      # Publish(ctx, channel, payload string) — 底层 g.Redis PUBLISH
  subscriber.go     # RunSubscriber(ctx, channel, handler) — 收编 go-redis standalone/cluster
  observer.go       # PublishObserver / SubscribeObserver + LoggingObserver
  channels.go       # ChannelAppHistoryNotify = "app:history:notify"
  errors.go
```

- Publisher 用于 `history.publishHistoryChange`。
- Subscriber 迁移 `gatewayapp/history_subscriber.go` 逻辑；`StartHistoryNotifySubscriber` 改为薄包装调用 `redismsgkit`。
- 观测：`OnPublish(ctx, channel, duration, err)`；订阅侧 `OnSubscribeMessage` 或等价 debug/warn。

### D4：键注册表 — 策略 A（采用）

按域拆分 builder 文件，**返回值必须与当前线上一致**：

| 文件 | 登记内容（示例） |
|---|---|
| `keys.go` | 扩展 `DomainUCG`、`DomainAI`；保留现有 device/history helper |
| `keys_device.go` | `WxIDToUnionKey(id)` → `dev:wx:id2union:{id}` 等 |
| `keys_history.go` | `PieceVerKey(deviceNo)`、`PieceDataKey(hash)` |
| `keys_gateway.go` | refresh、version、usage 日桶、sim wx set、`AppVersionLatestCacheKey(ctx)` |
| `keys_voice.go` | session（含 env prefix）、guard rate/idem、clinic session/rate/summary |
| `keys_ucg.go` | chat list/seq/unread/rebuild/user conv、profile pending、recommend throttle、ip throttle |
| `keys_ai.go` | `AIQuotaUsageKey(feature, wxID, month)`、`LLMGateWaitingKey(lane)`、`LLMGateInflightKey(lane)` |

**段规范（文档化，非强制 4 段）**：`{domain}:{module}:{...}`；builder 注释写清每段含义与 TTL。跨进程共享键（`ai:usage:`）仅在 `keys_ai.go` 定义。

删除业务侧：`ucg/chat_keys.go`、各文件内 `redis*Prefix` 常量（迁至 platform 后）。

### D5：业务注入模式（采用）

沿用 package 级变量：

```go
var cache = cachekit.Default() // 或 WithObserver(NewRedisCache(), LoggingObserver{})
var msgPub = redismsgkit.DefaultPublisher()
```

`recommend_mq_consumer.go` 的 `recommendThrottleCache` 补 `WithObserver`。

### D6：规范升格（采用）

**AGENTS.md** 新增两节（与「服务边界」同级）：

1. **Redis 访问（强制）**：cachekit + WithObserver；redismsgkit + WithObserver；禁止业务/controller 直连。
2. **Redis 键命名（强制）**：禁止业务/controller 键/频道字面量；新增键仅许在 platform builder 登记。

**project.md** 第 32 行扩展为访问 + 命名两句；第 95 行 UCG chat LIST 等同族模式补充「键经 platform builder」。

评审 grep（design 验收，可选 CI）：

```bash
rg 'g\.Redis\(\)' internal/services internal/controller --glob '*.go'
rg 'redis\.New(Client|ClusterClient)' internal/services internal/controller --glob '*.go'
```

platform 内允许：`cachekit/*.go`、`redismsgkit/*.go`、`rediscfg/*.go`。

### D7：spec 措辞（采用）

- `compose-redis-topology-2g`：Scenario 改为「经 platform Redis 客户端（cachekit Ping / redismsgkit）连接测试单机」。
- `gateway-app-api-usage-stats`：HGETALL Scenario 改为 `cachekit.HashGetAll`。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| GET 错误从静默变为 `ErrUnavailable` + warning | 与 platform-hardening「依赖失败显式」一致；piece 等路径在迁移时保留「miss 回源 MySQL」语义 |
| `HashGetAll` 行为回归 | 将 `usagestats/redisHashToMap` 原样迁入 cachekit；保留 v2.0.5 Scenario |
| 大 PR 冲突 | 单 change 单 PR；tasks 按域排序便于分工 |
| clinic_session `Set` 无 TTL 分支 | 保留 `Set` typed 方法，design 注释禁止滥用 |
| 订阅端 go-redis 与 cachekit g.Redis 双客户端 | 接受；redismsgkit 文档说明 cluster 订阅需 go-redis |

## Migration Plan

1. **Phase 1 — Platform**：扩展 cachekit + 新建 redismsgkit + keys_*.go + `Default()` helper。
2. **Phase 2 — 业务替换**：按 tasks.md 文件清单替换；删散落键文件。
3. **Phase 3 — 治理**：AGENTS.md、project.md、可选 `hack/check-redis-bypass.sh`。
4. **部署**：滚动发布各微服务；**无需** Redis FLUSH 或双读；键不变。
5. **回滚**：回退镜像；无数据兼容问题。

## Open Questions

（无阻塞项。）
