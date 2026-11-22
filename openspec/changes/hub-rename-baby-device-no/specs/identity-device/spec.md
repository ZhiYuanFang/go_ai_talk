# Spec Delta

## ADDED Requirements

### Requirement: device-service SHALL 提供宝宝号原地改名编排

device-service MUST 提供仅管理凭证可调用的宝宝号改名能力：将已注册 `oldDeviceNo` 原地迁移为未占用的 `newDeviceNo`，并编排本域与下游契约迁移，使业务数据在新号下连续。改名主语 MUST 为宝宝身份键，MUST NOT 依赖「某一单个 wx」作为迁移主键。成功后 `oldDeviceNo` MUST 不再通过 `EnsureRegistered`。

#### Scenario: 改名成功

- **WHEN** 管理员提交合法且未占用的 `newDeviceNo`，且 `oldDeviceNo` 已注册
- **THEN** 系统 MUST 将本域宝宝注册行与所有指向旧号的绑定引用更新为新号，MUST 调用下游迁移契约，且随后对 `oldDeviceNo` 的 `EnsureRegistered` MUST 失败

#### Scenario: 新号冲突

- **WHEN** `newDeviceNo` 已被其他宝宝注册
- **THEN** 系统 MUST 拒绝改名且 MUST NOT 修改旧号注册行

#### Scenario: 旧号不存在

- **WHEN** `oldDeviceNo` 未注册
- **THEN** 系统 MUST 拒绝改名并返回可读错误
