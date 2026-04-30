# service-runtime-standardization Specification

## Purpose
TBD - created by archiving change platform-hardening-redis-rabbitmq-service-split. Update Purpose after archive.
## Requirements
### Requirement: 缓存键与 TTL 规范必须统一
所有服务 SHALL 遵循统一的 Redis key 命名空间规范和已文档化的 TTL 规则，覆盖缓存、守卫与幂等状态。

#### Scenario: 任一服务引入新缓存键
- **WHEN** 某服务为运行时状态新增 Redis key
- **THEN** 该 key SHALL 符合统一命名规范与 TTL 策略

### Requirement: 事件命名与投递语义必须统一
所有跨服务事件 SHALL 使用统一的 exchange/routing-key 命名规范，并遵循明确的投递失败语义。

#### Scenario: 服务发布跨服务事件
- **WHEN** 某服务发出跨服务处理所需的领域事件
- **THEN** 其 SHALL 通过 RabbitMQ 按统一 exchange/routing-key 规范发布，并执行既定发布失败行为

### Requirement: 本迁移阶段执行禁测文件策略
代码库 SHALL 删除现有 Go 测试文件，并在本迁移阶段 SHALL NOT 新增 Go 测试文件。

#### Scenario: 迁移阶段引入新代码
- **WHEN** 开发者在本迁移范围内新增或重构代码
- **THEN** 其 SHALL 通过运行时核验脚本与运维检查进行验证，而不是新增 `*_test.go` 文件

