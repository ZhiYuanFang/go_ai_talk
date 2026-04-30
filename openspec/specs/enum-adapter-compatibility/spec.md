# enum-adapter-compatibility Specification

## Purpose
定义字符串到枚举迁移期兼容策略，保证旧入口可用并可验证关键路径完成枚举化收敛。

## Requirements
### Requirement: 渐进迁移兼容层
系统 MUST 在迁移期间保留字符串入口的兼容适配层，并通过统一适配函数将旧字符串路径映射到新枚举实现。

#### Scenario: 旧入口继续可用
- **WHEN** 调用方仍传入历史字符串值
- **THEN** 系统 MUST 通过兼容适配层完成转换并保持行为一致

#### Scenario: 兼容层输出弃用提示
- **WHEN** 旧入口被调用
- **THEN** 系统 SHOULD 输出弃用告警日志，提示迁移到枚举入口

### Requirement: 枚举化迁移可验证
系统 SHALL 提供可验证迁移清单，确保关键模块不再新增裸字符串匹配。

#### Scenario: 核心模块迁移完成检查
- **WHEN** 执行迁移验收
- **THEN** 系统 MUST 能确认 outbox、consumer、voice 关键路径已使用枚举匹配
