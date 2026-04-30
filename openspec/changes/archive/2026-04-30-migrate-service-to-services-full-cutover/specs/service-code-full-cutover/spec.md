## ADDED Requirements

### Requirement: `internal/service` 实现文件 MUST 全量迁移
系统 MUST 将 `internal/service` 中实现文件按领域归属迁移到 `internal/services/*` 或 `internal/shared/*`，迁移完成后不得遗留可编译业务实现文件。

#### Scenario: 全量迁移完成
- **WHEN** 执行迁移收口检查
- **THEN** `internal/service` 中不得再存在业务实现文件，且对应实现已在目标目录可追踪

### Requirement: 迁移后调用路径 MUST 指向新目录
所有服务入口、控制器和内部调用方 MUST 使用迁移后的包路径，不得继续依赖旧 `internal/service` 路径。

#### Scenario: 调用路径校验
- **WHEN** 对迁移范围执行 import 路径审查
- **THEN** 迁移后的调用引用 MUST 全部指向 `internal/services/*` 或 `internal/shared/*`
