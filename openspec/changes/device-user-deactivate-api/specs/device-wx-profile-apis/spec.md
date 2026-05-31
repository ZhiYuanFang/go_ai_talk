## ADDED Requirements

### Requirement: 账号注销删除 wx 记录
系统 MUST 提供 `POST /device/app/api/user/deactivate`。接口 SHALL 从请求头读取 `X-Internal-Wx-Id`，并按该主键删除 `wx` 表中的对应单条记录。删除成功后，系统 SHALL 使该 `wxId` 相关缓存映射失效，避免后续读取命中陈旧数据。

#### Scenario: 注销成功删除单条记录
- **WHEN** 请求头包含有效的 `X-Internal-Wx-Id` 且该 `wx` 记录存在
- **THEN** 系统 SHALL 删除该主键对应的一条 `wx` 记录并返回成功语义

#### Scenario: 请求头缺失或无效
- **WHEN** `X-Internal-Wx-Id` 缺失、非整数或小于等于 0
- **THEN** 系统 SHALL 返回参数错误，且 SHALL NOT 执行删除

#### Scenario: 目标记录不存在
- **WHEN** `X-Internal-Wx-Id` 合法但对应 `wx` 记录不存在
- **THEN** 系统 SHALL 返回明确业务错误语义（已注销或不存在），且 SHALL NOT 影响其他记录
