## ADDED Requirements

### Requirement: 事件字典 MUST 支持预约标志且默认非预约

系统 MUST 在全局事件字典（`event`）持久化预约标志，对外 JSON 字段名为 `isAppointment`（或与现有 camelCase 约定一致的布尔字段）。新建事件时若未显式指定，该标志 MUST 为 false（非预约）。该标志 MUST NOT 复用或改写既有 `eventType`（`number`|`time`|`one`）语义。

#### Scenario: 新建事件默认非预约

- **WHEN** 管理员新增事件且未勾选预约
- **THEN** 持久化记录的预约标志 MUST 为 false，且后续列表/options 返回中该事件 `isAppointment` 为 false

#### Scenario: 与 eventType 正交

- **WHEN** 某事件 `eventType` 为 `one` 且预约标志为 false
- **THEN** 系统 MUST 仍按非预约事件暴露，MUST NOT 仅因 `one` 而视为预约

### Requirement: 后台 MUST 可编辑预约标志

设备域后台新增与更新事件时，MUST 允许设置预约标志。更新成功后，系统 MUST 使事件字典读模型与既有「变更后重建 EventOptions 缓存」约定一致，以便后续 options/list 读到新值。

#### Scenario: 后台将事件标为预约

- **WHEN** 管理员将已有事件更新为预约（`isAppointment=true`）且保存成功
- **THEN** 之后经事件列表或 options 读取该事件时 MUST 返回 `isAppointment=true`

### Requirement: 事件 options 与列表 MUST 附带预约标志

凡返回 `entity.Event`（或等价事件字典项）的 App/内部 options 与后台列表路径，MUST 包含预约标志字段，供 Flutter 判断是否按预约逻辑处理。字段缺失 MUST NOT 作为「非预约」的隐式协议以外的行为；实现 MUST 保证新列映射到响应（旧缓存淘汰后）。

#### Scenario: Flutter 可从 options 识别预约事件

- **WHEN** 客户端请求历史事件 options（或等价字典接口）且字典中存在已标记预约的事件
- **THEN** 响应列表中对应项 MUST 包含 `isAppointment=true`
