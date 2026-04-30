## Why

当前 `history/action/event/user` 的读取路径仍会频繁访问数据库或远程 history-service，导致高并发下响应抖动，并放大 DeepSeek 请求前置数据装配的延迟。随着服务实例扩缩容和异步链路增多，需要建立稳定的 Redis 读模型与异步更新机制，在保障语义一致性的前提下降低读时延和数据库压力。

## What Changes

- 为 `history/action/event/user` 建立 Redis 读模型，读请求默认优先命中 Redis，失败或缺失时按策略回源并回填。
- 将历史记录及相关元数据的缓存更新从“写后删缓存”调整为“写后异步补丁更新缓存”，通过低延迟异步消费者维护缓存投影。
- 为 DeepSeek 依赖的历史上下文读取接入 Redis 优先路径，减少构造 prompt 时对 DB/远程服务的直接访问。
- 引入缓存版本、乱序保护与失败补偿机制，确保多实例和异步重试场景下的缓存可恢复与最终一致。
- 补充缓存可观测性指标与运维操作规范（命中率、回源率、补丁失败率、修复任务状态）。

## Capabilities

### New Capabilities
- `redis-read-model-cache`: 定义 history/action/event/user 的 Redis 读模型、Key 规范、回源与回填行为。
- `async-cache-projection-sync`: 定义写入成功后通过异步事件更新缓存投影、版本控制、失败补偿与修复机制。
- `deepseek-history-redis-prefer`: 定义 DeepSeek 历史上下文获取时 Redis 优先读取与降级行为。

### Modified Capabilities
- （无）

## Impact

- Affected code:
  - `internal/services/history/**`
  - `internal/services/device/**`
  - `internal/services/voice/**`
  - `internal/platform/cachekit/**`
  - `internal/platform/eventkit/**`
  - `internal/services/async/**`
- Affected systems:
  - Redis cluster（新增读模型 key 空间、异步更新与修复负载）
  - RabbitMQ/outbox（新增缓存投影事件与消费链路）
  - MySQL（读压下降，但需保留权威回源能力）
- Ops/Runtime:
  - 新增缓存一致性与回源告警指标
  - 新增缓存修复/重建运行手册
