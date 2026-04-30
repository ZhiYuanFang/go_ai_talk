## ADDED Requirements

### Requirement: 核心领域值类型化枚举
系统 SHALL 为 `target_type`、`mode`、状态机状态、`event type` 提供类型化枚举定义，并通过统一常量与解析函数替代散落裸字符串匹配。

#### Scenario: 调用层使用枚举分支
- **WHEN** 业务代码需要按 `target_type` 或 `mode` 分支处理
- **THEN** 代码 MUST 使用枚举类型与常量进行判断，而不是直接比较裸字符串

#### Scenario: 非法值解析失败
- **WHEN** 输入字符串无法映射为合法枚举值
- **THEN** 系统 MUST 返回明确错误并记录可观测日志

### Requirement: 枚举与字符串双向兼容
系统 SHALL 提供枚举到字符串、字符串到枚举的双向转换能力，保证现有 DB 与消息协议字符串格式兼容。

#### Scenario: 入站字符串转换为枚举
- **WHEN** 系统从 DB/MQ/HTTP 读取字符串字段
- **THEN** 系统 MUST 通过统一 Parse 方法转换为枚举值后参与业务判断

#### Scenario: 出站枚举保持原协议字符串
- **WHEN** 系统写入 DB 或发布消息
- **THEN** 系统 MUST 输出与历史协议兼容的字符串值
