# Spec Delta

## ADDED Requirements

### Requirement: history-service SHALL 支持按宝宝号迁移历史分区键

history-service MUST 提供仅内部/管理可信调用方可访问的迁移接口：将 `history` 与 `suggest` 表中 `device_no=oldDeviceNo` 的行更新为 `newDeviceNo`。迁移 MUST NOT 经跨库直查由他域进程执行。

#### Scenario: 迁移成功

- **WHEN** 可信调用方请求将旧宝宝号迁移到新宝宝号
- **THEN** 该旧号下原有 history 与 suggest 行 MUST 在新号下可查，且旧号下 MUST 不再保留这些业务行

### Requirement: 喂养历史写入前 MUST 校验宝宝号已注册

history-service 在新增喂养历史（含 batch create）时 MUST 经 device 域契约确认请求 `deviceNo` 已注册；未注册 MUST 拒绝写入并返回可读错误。MUST NOT 仅因 Redis 缓存存在旧键而放行或改写落库 `device_no`。

#### Scenario: 未注册号拒绝写入

- **WHEN** 客户端或上游以未注册的 `deviceNo` 请求新增历史
- **THEN** 系统 MUST 不插入历史行并 MUST 返回表示设备未注册的错误

#### Scenario: 已注册号允许写入

- **WHEN** 客户端以已注册的 `deviceNo` 请求新增合法历史
- **THEN** 系统 MUST 按既有规则写入且落库 `device_no` MUST 等于请求号
