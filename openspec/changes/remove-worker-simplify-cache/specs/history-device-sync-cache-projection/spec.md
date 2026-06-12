## ADDED Requirements

### Requirement: history 写操作 MUST 同步更新 Redis 读模型

`history-service`（及 voice adapter 本地 patch 路径）在 `history` 表 insert/update/delete **成功提交后**，MUST 在同请求内调用读模型 patch（`patchHistoryOnAdd` / `patchHistoryOnUpdate` / `patchHistoryOnDelete`）或等价逻辑更新 `history:record:list:*` 与 `history:record:latest:*`；MUST NOT 依赖 worker-service 或 domain_outbox 异步投影作为唯一更新路径。

#### Scenario: AddHistory 成功后列表与 latest 可读

- **WHEN** `AddDeviceHistory` 事务提交成功
- **THEN** 系统 MUST 尝试同步 patch Redis；后续 `GetLatestHistory` 在 cache miss 或 patch 成功时 MUST 返回含新记录的正确数据（以 MySQL 为准）

#### Scenario: 冷缓存新增记录仍更新 latest

- **WHEN** 设备 history 列表 Redis key 不存在且新增一条记录
- **THEN** 系统 MUST 至少写入 `history:record:latest:{deviceNo}`；列表 MAY 在首次 `ListHistory` miss 时从 MySQL 全量回填

### Requirement: list patch 失败 MUST 避免长期 stale hit

当 `setHistoryList` 失败时，系统 SHOULD best-effort 删除对应 list key，使下次读取走 MySQL 回源；MUST NOT 假设 worker 会异步修复。

#### Scenario: Redis 写入失败后读路径自愈

- **WHEN** prepend 列表缓存时 Redis 返回错误
- **THEN** 系统 MUST 记录 warning；SHOULD 删除 list key；读路径在 miss 时 MUST 从 MySQL 重建列表

### Requirement: device 域字典与画像 MUST 在写路径同步重建缓存

device admin 变更 `event`/`action`/`user` 后，MUST 在写库成功后调用 `refreshEventOptionsCacheAfterMutate`、`RebuildActionCache` 或 `setUserProfile` 等同步重建；MUST NOT 依赖 worker HTTP 回调作为唯一刷新路径。

#### Scenario: 后台新增事件后主数据可读

- **WHEN** 管理端新增事件并提交成功
- **THEN** 下一次 `ListEvents` MUST 能读到新事件（Redis hit 或 miss 回源 MySQL 后回填）

### Requirement: 读路径 MUST 在 cache miss 时回源 MySQL 并回填

`ListHistory`、`GetLatestHistory`、device `ListEvents` 等在 Redis miss 或不可用时，MUST 查询权威 MySQL 并回填 Redis；数据 correctness MUST 以 MySQL 为准。

#### Scenario: 列表 cache miss

- **WHEN** `getHistoryList` 未命中
- **THEN** adapter MUST 查库并 `setHistoryList` 回填

## REMOVED Requirements

### Requirement: 写入后异步更新缓存投影（domain_outbox + worker relay）

**Reason**: worker-service 与 domain_outbox relay 删除；投影改为写路径同步 patch + 读 miss 重建。

**Migration**: 删除 `EnqueueDomainOutbox`、`OUTBOX_RELAY_ENABLED`、`WORKER_SERVICE_URL`；删除 `async/domain_outbox.go`；保留 `ApplyProjection` 逻辑仅当无引用则删除。

### Requirement: 失败补偿与可重建（worker outbox 重试与 projection repair ticker）

**Reason**: outbox relay 与 `CACHE_PROJECTION_REPAIR` 删除；重建改为读 miss 全量回填与运维手动 `RebuildHistoryCacheByDevice`（无自动 ticker）。

**Migration**: 文档标注手动 rebuild 命令/HTTP（若保留管理端点）；监控同步 patch 失败日志。
