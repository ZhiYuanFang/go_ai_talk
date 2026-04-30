## ADDED Requirements

### Requirement: DeepSeek 历史读取 Redis 优先
系统在执行 DeepSeek 意图分析与对话补全前，历史上下文读取 SHALL 优先命中 Redis 历史读模型，并在不可用时回源到历史服务或数据库。

#### Scenario: 命中缓存快速构造上下文
- **WHEN** DeepSeek 请求前需要最近历史上下文且 Redis 命中
- **THEN** 系统 MUST 使用缓存数据完成上下文构造并减少对数据库/远程历史服务访问

#### Scenario: 未命中缓存自动回源
- **WHEN** DeepSeek 请求前历史缓存未命中
- **THEN** 系统 MUST 回源获取历史并回填缓存后继续调用 DeepSeek

### Requirement: 上下文读取一致性与降级可观测
系统 MUST 对 DeepSeek 上下文读取提供命中率、回源率与降级原因可观测性，并在异常时保持功能可用。

#### Scenario: 记录上下文读取指标
- **WHEN** 任一 DeepSeek 请求完成上下文装配
- **THEN** 系统 MUST 记录本次是否命中 Redis、是否发生回源与耗时指标

#### Scenario: Redis 异常时不阻断回复
- **WHEN** Redis 不可达或读取失败
- **THEN** 系统 MUST 降级回源并继续完成 DeepSeek 调用，同时输出结构化告警日志

### Requirement: 历史窗口语义一致
系统 SHALL 保证 Redis 历史读模型与权威源在“最近 N 小时”窗口语义上保持一致，避免因缓存截断导致模型上下文偏差。

#### Scenario: 缓存与权威窗口对齐
- **WHEN** 系统读取最近 N 小时历史用于 DeepSeek
- **THEN** 系统 MUST 返回与权威源相同窗口边界的数据集合（允许在异步延迟窗口内最终一致）
