## Purpose
本规格定义微服务平台中数据所有权边界、跨实例临时状态共享机制与缓存一致性治理原则，
用于保障服务拆分后系统在扩缩容、多实例路由与数据更新场景下的稳定性与一致性。

## Requirements

### Requirement: 数据所有权 SHALL 按服务边界划分
每个领域服务 SHALL 在其服务范围内独立拥有并管理核心数据模型，以降低跨服务耦合。

#### Scenario: 服务自有 schema 演进
- **WHEN** device service 需要更新设备相关 schema
- **THEN** 该变更仅在 device 所有的数据边界内完成，无需无关服务直接改动 schema

### Requirement: 共享临时状态 SHALL 使用 Redis 集群
系统 SHALL 将跨实例临时状态（包括会话状态、缓存条目与幂等键）存储在 Redis 集群中，而非进程本地内存。

#### Scenario: 扩容后会话连续性
- **WHEN** 同一设备的 voice-session 请求被不同服务实例处理
- **THEN** 新实例可从 Redis 集群读取一致会话状态并无上下文丢失地继续交互

### Requirement: 缓存一致性策略 SHALL 明确定义
系统 SHALL 为缓存数据定义明确的失效策略和 TTL 行为，避免出现陈旧数据或状态无界增长。

#### Scenario: 源数据更新后的缓存刷新
- **WHEN** 已缓存实体的源数据发生变化
- **THEN** 缓存条目会按配置的一致性策略与 TTL 被失效或刷新
