# service-migration-safety-and-rollback Specification

## Purpose
TBD - created by archiving change migrate-service-to-services-full-cutover. Update Purpose after archive.
## Requirements
### Requirement: 迁移执行 MUST 分批且可验证
全量迁移 MUST 以可回滚批次执行，每批完成后必须通过编译校验与关键服务启动校验后方可进入下一批。

#### Scenario: 批次完成校验
- **WHEN** 单个迁移批次完成
- **THEN** 系统 MUST 通过既定编译检查与启动健康检查

### Requirement: 迁移异常 MUST 支持按服务维度回滚
迁移引发异常时，系统 MUST 支持按受影响服务维度回退代码与配置，不要求全局回滚。

#### Scenario: 单服务回滚
- **WHEN** `voice-service` 批次迁移后出现运行异常
- **THEN** 团队 MUST 可仅回滚 `voice-service` 相关迁移批次并恢复可用

### Requirement: 收口验收 MUST 覆盖关键链路无回归
全量迁移收口时 MUST 验证 gateway/voice/device/history 关键链路，确保外部行为与迁移前一致。

#### Scenario: 收口链路验收
- **WHEN** 所有批次迁移完成并准备收口
- **THEN** 关键业务链路 MUST 通过无回归验证，且迁移结果可被文档化追踪

