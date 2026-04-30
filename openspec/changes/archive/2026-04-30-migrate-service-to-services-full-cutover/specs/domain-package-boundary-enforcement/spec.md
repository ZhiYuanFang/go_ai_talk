## ADDED Requirements

### Requirement: 领域包边界 MUST 与服务边界一致
迁移后代码 MUST 按领域归属放置在 `internal/services/voice`、`internal/services/device`、`internal/services/history` 等目录，且包语义必须与目录一致。

#### Scenario: 包语义审查
- **WHEN** 审查迁移后的领域目录
- **THEN** 目录内代码包语义 MUST 体现对应领域职责，不得继续使用统一 `service` 包承载多域逻辑

### Requirement: 共享目录准入 MUST 可审计
`internal/shared` MUST 仅容纳无领域语义的通用能力；含领域流程或领域模型耦合的实现 MUST 禁止进入共享目录。

#### Scenario: 共享目录准入检查
- **WHEN** 有文件计划迁入 `internal/shared`
- **THEN** 评审 MUST 给出“无领域语义”依据，否则该文件 MUST 回到对应领域目录

### Requirement: 新增代码 MUST 禁止回流到 `internal/service`
迁移完成后，新增实现文件 MUST 不得再放入 `internal/service`。

#### Scenario: 新增文件路径检查
- **WHEN** 提交包含新增实现文件
- **THEN** 若目标路径为 `internal/service`，该提交 MUST 视为不符合边界规范
