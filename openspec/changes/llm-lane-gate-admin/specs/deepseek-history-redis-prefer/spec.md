## MODIFIED Requirements

### Requirement: DeepSeek 历史读取 Redis 优先

系统在执行 **voiceUnderstanding lane** 所触发的 LLM 意图分析与对话补全前，历史上下文读取 SHALL 优先命中 Redis 历史读模型，并在不可用时回源到历史服务或数据库。

#### Scenario: LLM 请求前 Redis 命中

- **WHEN** voiceUnderstanding 路径请求前需要最近历史上下文且 Redis 命中
- **THEN** 系统 MUST 使用缓存历史组装 prompt

#### Scenario: 缓存未命中回源

- **WHEN** voiceUnderstanding 路径请求前历史缓存未命中
- **THEN** 系统 MUST 回源获取历史并回填缓存后继续调用 LLM

### Requirement: 历史上下文读取可观测性

系统 MUST 对 **voiceUnderstanding** LLM 上下文读取提供命中率、回源率与降级原因可观测性，并在异常时保持功能可用。

#### Scenario: 请求完成可观测

- **WHEN** 任一 voiceUnderstanding LLM 请求完成上下文装配
- **THEN** 系统 SHOULD 输出结构化日志含缓存命中/回源标记

#### Scenario: Redis 不可用降级

- **WHEN** Redis 读模型暂时不可用
- **THEN** 系统 MUST 降级回源并继续完成 LLM 调用，同时输出结构化告警日志

### Requirement: 历史时间窗口读取

系统读取最近 N 小时历史用于 **voiceUnderstanding** LLM 时 MUST 遵守既有 history 读模型契约，MUST NOT 新增跨库直查他域表。

#### Scenario: 历史问答使用读模型

- **WHEN** 历史问答等价路径加载 12 小时历史
- **THEN** MUST 经 Redis 优先读模型路径
