## Why

当前约 80% 的 Redis 读写仍直接调用 `g.Redis().Do(...)`，键名散落在各业务包（`dev:`/`gw:`/`usage:` 等前缀并存），Pub/Sub 与 KV 混用且无统一观测。`openspec/project.md` 虽已约定经 `cachekit.WithObserver`，但未在 `AGENTS.md` 升格、无 CI/评审 grep，且 `cachekit.Cache` 接口未覆盖 Hash/List/Set 等命令，导致规范不可执行。需在**不改变线上键字符串**（策略 A）的前提下，一次性全仓迁移至 platform 抽象，并收拢键 builder。

## What Changes

- 扩展 `internal/platform/cachekit`：补齐全仓所需的 typed Redis 方法（含 `HashGetAll` 内聚 GoFrame flat `[]string` 解析）；所有方法经 `observedCache` 上报 `LoggingObserver`。
- 新增 `internal/platform/redismsgkit`：Redis Pub/Sub **Publish/Subscribe**（PUBLISH 不算 cache）；收编 `history/realtime_notify` 发布端与 `gatewayapp/history_subscriber` 订阅端（含 go-redis cluster 分支）。
- 收拢 Redis 键/频道 builder 至 `cachekit/keys_*.go` 与 `redismsgkit/channels.go`：**策略 A — builder 返回值与线上一致，不重命名键空间**（含保留 `dev:`、`gw:`、`usage:sim_wx_ids` 等历史前缀）。
- 一次性迁移 `internal/services/**` 与 `internal/controller/**` 内全部 `g.Redis()` 直连（约 18 业务文件 + 半合规 `recommend_mq_consumer`）；删除域内散落键 helper（如 `ucg/chat_keys.go`）。
- 升格仓库级强制项：更新 `AGENTS.md` 与 `openspec/project.md`（Redis 访问 + 键命名）；新增评审 grep 清单（可选 `hack/check-redis-bypass.sh`）。
- 规格增量：新增 `redis-platform-access` capability；修正基线中引用 `g.Redis()` 字样的 Scenario（测试 Redis 连接、usage HGETALL）。

## Capabilities

### New Capabilities

- `redis-platform-access`：业务代码 MUST 经 `cachekit.WithObserver` / `redismsgkit.WithObserver` 访问 Redis；键/频道 MUST 经 platform builder；禁止业务层键字面量与 `g.Redis()` 直连。

### Modified Capabilities

- `compose-redis-topology-2g`：测试微服务 Redis 连接 Scenario 改为描述 platform 封装后的客户端行为，去除 `g.Redis()` 字面量。
- `gateway-app-api-usage-stats`：`HGETALL` 解析 Scenario 改为经 `cachekit.HashGetAll` 表述；对外 Admin API 行为不变。

## Impact

- **代码**：`internal/platform/cachekit/**`（扩展）、新建 `internal/platform/redismsgkit/**`；`internal/services/**` 多域（ucg/voice/history/gatewayapp/device/aimodel）、`internal/controller/gateway_app_ctrl.go`；`AGENTS.md`、`openspec/project.md`。
- **API/行为**：对外 HTTP/WS 契约不变；Redis 键字符串不变（策略 A）；运行时错误语义统一为 `cachekit.ErrUnavailable` 等（个别路径由静默 miss 变为显式 warning，见 design）。
- **部署**：无 DB migration；无 Redis 数据迁移；部署后既有 Redis 数据继续有效。
- **Redis 读缓存约定**：本变更不新增 Redis 使用场景，仅统一访问路径与键注册表；不涉及新键空间收益率评估。
- **App API 使用统计**：无新增接口。
