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

- **WHEN** `oldDeviceNo` 未注册且 `newDeviceNo` 亦未注册
- **THEN** 系统 MUST 拒绝改名并返回可读错误

#### Scenario: 半失败后以新号幂等续跑

- **WHEN** `oldDeviceNo` 已无注册行且 `newDeviceNo` 已是本域宝宝号（本域段已完成）
- **THEN** 系统 MUST 跳过本域写并继续下游迁移/吊销（可读分步），MUST NOT 因「旧号不存在」整单拒绝

### Requirement: 宝宝号改名 MUST 先探再改

device-service 在执行本域宝宝号写改或下游 history/cash 迁移、gateway 会话吊销之前，MUST 先探测上述下游内部契约可达且相关库表可访问。探测 MUST NOT 改写业务分区键或 bump session epoch。任一探测失败时 MUST 中止并返回可读错误，且 MUST NOT 在首次路径上于探测失败后仍完成本域写改。

#### Scenario: 下游不可达则不改本域

- **WHEN** 管理员发起改名且 history 或 cash 或 gateway 内部探测失败
- **AND** 本域旧号仍注册（尚未迁到新号）
- **THEN** 系统 MUST NOT 更新本域 `user.device_no`，MUST 返回可读错误

#### Scenario: 探测通过后再编排

- **WHEN** history、cash、gateway 探测均成功
- **THEN** 系统 MAY 按既定顺序执行本域迁移与下游契约
