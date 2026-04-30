## ADDED Requirements

### Requirement: Redis 优先读取历史与元数据
系统对 `history/action/event/user` 的读取 SHALL 默认优先从 Redis 读模型获取；当缓存缺失、反序列化失败或依赖异常时，系统 MUST 回源权威数据源并在成功后回填 Redis。

#### Scenario: 缓存命中直接返回
- **WHEN** 读取 `history/action/event/user` 请求命中 Redis 且数据有效
- **THEN** 系统 MUST 直接返回缓存结果且不访问数据库

#### Scenario: 缓存缺失触发回源回填
- **WHEN** 读取请求未命中 Redis
- **THEN** 系统 MUST 回源数据库或契约服务获取数据并回填到 Redis 后返回

#### Scenario: 缓存损坏自动降级
- **WHEN** Redis 中对应 key 数据格式错误或反序列化失败
- **THEN** 系统 MUST 降级为回源读取并覆盖修复该缓存 key

### Requirement: 统一缓存键空间与版本语义
系统 SHALL 为 `history/action/event/user` 定义统一域内 key 规则与版本键规则，并 MUST 在读取时识别版本语义以支持后续乱序保护和修复。

#### Scenario: 键命名符合域规范
- **WHEN** 任一模块写入或读取缓存 key
- **THEN** key MUST 满足统一格式（domain:module:kind:identifier）且可由领域缓存仓储一致生成

#### Scenario: 版本键可用于一致性判断
- **WHEN** 读取方发现实体数据键与版本键不一致
- **THEN** 系统 MUST 触发回源修复或异步修复流程并避免返回明显过期快照

### Requirement: DeepSeek 历史上下文读取复用读模型
语音链路在构造 DeepSeek prompt 时，历史与画像读取 MUST 复用 Redis 读模型优先路径，不得绕过读模型直接形成新的 DB 热点通道。

#### Scenario: 构造 prompt 时命中 Redis 历史
- **WHEN** 语音链路需要读取最近历史记录
- **THEN** 系统 MUST 优先读取 Redis 中历史读模型并用于 prompt 构造

#### Scenario: Redis 不可用时语音链路可用
- **WHEN** Redis 短时不可用
- **THEN** 系统 MUST 回源获取历史并继续完成 prompt 构造，同时记录降级日志与指标
