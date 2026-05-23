# database-unix-timestamp-storage Specification

## Purpose
TBD - created by archiving change database-unix-timestamp-storage. Update Purpose after archive.
## Requirements
### Requirement: 时间类字段落库形态

除经架构评审明确豁免的「纯日历日期」字段外，凡表示**时刻**（事件发生时间、创建/更新时间、最后活跃时间等）的数据库列 **MUST** 以 **Unix 时间戳秒** 数值存储，MySQL 类型 **MUST** 为可表达该范围的整数类型（推荐 `BIGINT`）。**MUST NOT** 将本地墙钟格式化的日期时间字符串作为权威落库值。

#### Scenario: 新表创建

- **WHEN** 新建包含「时刻」语义的表或列
- **THEN** 该列类型为整数型时间戳秒且注释标明 UTC 纪元秒，且应用写入路径使用 UTC 纪元秒（如 `time.Time.Unix()`）而非格式化字符串

### Requirement: API 与 JSON 契约

对外 HTTP JSON 中代表「时刻」的字段 **MUST** 使用数字类型（Unix 秒），与数据库存储单位一致；字段文档或 OpenAPI **MUST** 标明单位为秒。若迁移期需兼容旧客户端，**MUST** 在变更说明中定义弃用截止条件，且服务端权威值仍为数字戳。

#### Scenario: 客户端解析

- **WHEN** 客户端接收代表事件发生时刻的字段
- **THEN** 该值为 JSON number（Unix 秒），客户端在展示给用户时自行按目标时区转换，不依赖服务端返回本地日历字符串作为权威

### Requirement: 迁移与数据完整性

对已有「非数字时刻」列的迁移 **MUST** 提供可重复执行的回填策略，并在切换读写前完成行数一致性与抽样校验。**MUST** 定义 NULL/非法旧值的处置规则（拒绝写入、置 0 或置哨兵值须文档化且经评审）。

#### Scenario: 回填后校验

- **WHEN** 执行从旧列到新秒级列的回填脚本
- **THEN** 存在自动化或清单式校验（行数、非 NULL 比例、时间范围合理性）且通过后应用才切换为只读新列

### Requirement: 服务边界

各服务 **MUST** 仅修改本服务拥有库内的表与 DAO；跨服务时间语义通过契约（HTTP/RPC/事件）传递，传递值 **MUST** 为 Unix 秒或与契约显式声明的单位一致。

#### Scenario: history 与 device 分库

- **WHEN** 在 history-service 所属库中迁移时间列
- **THEN** 不修改 device-service 所属库的表结构于同一提交中混写；各自变更独立可发布

