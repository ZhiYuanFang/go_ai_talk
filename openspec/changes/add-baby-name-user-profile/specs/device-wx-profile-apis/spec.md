## ADDED Requirements

### Requirement: 设备画像读接口 SHALL 返回宝宝名字
系统在读取设备画像时 MUST 同时返回 `babyName`、`birthday`、`sex` 三个字段；其中 `babyName` 为可选字符串，未设置时返回空串。
该要求适用于 device 画像接口与历史页面画像接口的统一读取语义。

#### Scenario: 读取画像返回完整字段
- **WHEN** 调用方使用有效 `deviceNo` 请求画像读取接口
- **THEN** 响应 SHALL 包含 `babyName`、`birthday`、`sex`
- **AND** 当数据库中 `baby_name` 为空时，`babyName` SHALL 返回空串

### Requirement: 设备画像写接口 SHALL 支持宝宝名字更新
系统在保存设备画像时 MUST 接受 `babyName` 字段，并与 `birthday`、`sex` 一并持久化到 `user` 表画像字段集合。
该要求适用于 `/device/app/api/user/save`、`/device/app/api/user/auto_save` 以及历史页面画像保存链路。

#### Scenario: 仅修改宝宝名字
- **WHEN** 调用方提交合法 `deviceNo` 与 `babyName`，且未变更生日/性别
- **THEN** 系统 SHALL 更新 `user.baby_name`
- **AND** 系统 SHALL 保持 `birthday`、`sex` 原值不变

#### Scenario: 同时修改名字与性别生日
- **WHEN** 调用方提交 `babyName`、`birthday`、`sex`
- **THEN** 系统 SHALL 在一次保存语义内持久化三项画像字段

## MODIFIED Requirements

### Requirement: 自动保存画像

device-service SHALL 提供 `POST /device/app/api/user/auto_save`，从请求头读取 **`X-Internal-Wx-Union-Id`**，从 body 读取 `babyName`、`birthday` 与 `sex`，并 SHALL 返回 `device_no`。设备域中 `device_no` SHALL 在**全表范围内唯一**（与现有 `user`/设备表唯一约束一致）。当该 wx 尚未绑定设备时，系统 SHALL 在设备域注册表中 **新建一条设备记录**，其 `device_no` 为 **6 个字符**，且每个字符均为 **大写英文字母 A–Z** 的随机取值，且该值 SHALL **不与任何已存在的** `device_no` 重复；将该 `device_no` 与当前 wx 绑定后再写入画像（宝宝名字、生日、性别），语义与现有用户画像保存能力对齐。当该 wx **已绑定** `device_no` 时，系统 SHALL **仅更新**画像并返回**已有**的 `device_no`，且 SHALL NOT 更换设备号。

#### Scenario: 无设备号时创建并绑定

- **WHEN** **unionid** 有效且当前 wx 行未绑定 device_no，且画像字段合法
- **THEN** 系统 SHALL 生成符合「6 位大写 A–Z」规则的 device_no、在设备表中创建新设备、与 wx 绑定、保存画像（含 `babyName`），且响应 SHALL 包含该新 device_no，且持久化后的 device_no SHALL 与库内其它行的 device_no 均不相同

#### Scenario: 已绑定设备仅更新画像

- **WHEN** **unionid** 有效且 wx 已绑定 device_no
- **THEN** 系统 SHALL 仅更新画像字段（含 `babyName`）且响应 SHALL 返回原 device_no，且 SHALL NOT 生成新的随机 device_no

#### Scenario: 设备号与已有数据冲突时重试直至唯一

- **WHEN** 随机生成的候选 device_no 已存在于设备表中（或由唯一索引拒绝插入）
- **THEN** 系统 SHALL 放弃该候选并重新生成，重复直至插入成功或达到实现约定的最大重试次数；成功时持久化的 device_no SHALL 满足全局唯一约束
