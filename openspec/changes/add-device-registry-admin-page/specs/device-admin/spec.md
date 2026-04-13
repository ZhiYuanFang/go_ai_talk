## ADDED Requirements
### Requirement: 设备管理网页口令访问
系统 SHALL 提供设备管理网页入口，并在显示管理内容前要求用户输入固定口令 `a521521521` 完成校验。

#### Scenario: 口令正确后进入管理页面
- **WHEN** 用户访问设备管理网页并输入口令 `a521521521`
- **THEN** 系统允许进入管理页面并可执行设备注册与列表查看操作

#### Scenario: 口令错误时拒绝访问
- **WHEN** 用户输入非 `a521521521` 的口令
- **THEN** 系统拒绝显示管理功能并提示口令错误

### Requirement: 设备注册写入 user 表
系统 SHALL 提供设备注册能力，注册时 MUST 仅校验 `deviceNo` 唯一性；注册成功后 MUST 向 `user` 表写入 `device_no` 与 `active_time`。

#### Scenario: 注册新设备成功
- **WHEN** 管理页提交一个未存在的 `deviceNo`
- **THEN** 系统创建 `user` 记录，写入该 `device_no` 与当前 `active_time`

#### Scenario: 重复设备号注册失败
- **WHEN** 管理页提交一个已存在于 `user.device_no` 的 `deviceNo`
- **THEN** 系统返回设备号重复错误
- **AND** 不创建新记录

### Requirement: 设备激活与最后对话信息列表展示
系统 SHALL 提供设备列表查询能力，返回全部设备的 `device_no`、`active_time`、`last_talk_time`、`last_talk_ask`、`last_talk_answer` 以供网页展示。

#### Scenario: 查看设备列表
- **WHEN** 用户通过口令校验后打开设备列表
- **THEN** 系统返回所有设备记录
- **AND** 每条记录包含 `device_no`、`active_time`、`last_talk_time`、`last_talk_ask`、`last_talk_answer`
