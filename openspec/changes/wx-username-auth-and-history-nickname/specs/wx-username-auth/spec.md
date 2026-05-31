## ADDED Requirements

### Requirement: 用户名注册写入 wx 账号
系统 SHALL 提供用户名注册接口，并在 `ai_voice_device.wx` 新建账号行；该行以 `wx.id` 作为账号主键，`unionid` MAY 为空，`user_name` MUST 全局唯一，`password` MUST 以不可逆哈希密文保存。

#### Scenario: 注册成功
- **WHEN** 客户端提交合法且未占用的 `userName` 与 `password`
- **THEN** 系统 SHALL 新建一条 `wx` 记录并返回 `wxId`，且数据库中的 `password` SHALL NOT 为明文

#### Scenario: 用户名冲突
- **WHEN** 客户端提交的 `userName` 已被其他 `wx` 记录占用
- **THEN** 系统 SHALL 返回“用户名已存在”冲突错误，且 SHALL NOT 新建记录

### Requirement: 用户名密码登录
系统 SHALL 提供用户名登录接口，按 `user_name` 定位 `wx` 记录并校验哈希密码；校验通过后 SHALL 返回 `wxId` 与账号业务信息供网关签发令牌。

#### Scenario: 登录成功
- **WHEN** `userName` 存在且密码校验通过
- **THEN** 系统 SHALL 返回对应 `wxId`，并标识登录成功

#### Scenario: 登录失败
- **WHEN** `userName` 不存在或密码校验失败
- **THEN** 系统 SHALL 返回认证失败错误，且 SHALL NOT 泄露是“用户名不存在”还是“密码错误”的内部细节

### Requirement: 用户名账号绑定微信
系统 SHALL 提供用户名账号绑定微信接口，将微信 `unionid` 绑定到指定 `wx.id` 账号；同一 `unionid` MUST NOT 同时绑定多个账号。

#### Scenario: 绑定成功
- **WHEN** 当前账号未绑定微信，且目标 `unionid` 未被其他账号占用
- **THEN** 系统 SHALL 将该 `unionid` 写入当前账号并返回成功

#### Scenario: 微信已被占用
- **WHEN** 目标 `unionid` 已绑定在其他 `wx.id`
- **THEN** 系统 SHALL 返回“微信已绑定其他账号”错误，且 SHALL NOT 覆盖原绑定

### Requirement: 用户名账号绑定设备号
系统 SHALL 提供用户名账号绑定设备号接口，绑定前 MUST 校验设备号已在设备域注册，绑定后 SHALL 维护 `wx.device_no` 一致性并失效相关缓存。

#### Scenario: 绑定成功
- **WHEN** `deviceNo` 已注册且请求主体账号合法
- **THEN** 系统 SHALL 更新 `wx.device_no` 并返回成功

#### Scenario: 设备号未注册
- **WHEN** 提交的 `deviceNo` 未在设备域注册
- **THEN** 系统 SHALL 返回业务校验失败，且 SHALL NOT 更新绑定关系

### Requirement: 修改用户名密码
系统 SHALL 提供修改密码接口；调用方 MUST 提供旧密码并通过校验后方可写入新密码哈希。

#### Scenario: 改密成功
- **WHEN** 旧密码校验通过且新密码满足格式策略
- **THEN** 系统 SHALL 将 `password` 更新为新哈希并返回成功

#### Scenario: 旧密码错误
- **WHEN** 旧密码校验失败
- **THEN** 系统 SHALL 返回认证失败错误，且 SHALL NOT 修改数据库密码

### Requirement: 微信账号下创建用户名密码
系统 SHALL 提供“微信账号创建用户名密码”接口，使已存在微信账号（`unionid` 已绑定）补齐 `user_name` 与 `password`；若账号已存在用户名，系统 MUST 拒绝重复创建。

#### Scenario: 创建成功
- **WHEN** 微信账号存在且尚未设置 `user_name`
- **THEN** 系统 SHALL 写入唯一用户名与密码哈希，并返回成功

#### Scenario: 已存在用户名
- **WHEN** 当前微信账号已设置 `user_name`
- **THEN** 系统 SHALL 返回“账号已存在用户名密码”错误，且 SHALL NOT 覆盖原值
