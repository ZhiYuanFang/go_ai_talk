## ADDED Requirements

### Requirement: 微信登录仅返回业务字段

device-service SHALL 提供 `POST /device/app/api/user/login`（设备 wx 业务登录，与网关聚合 `POST /device/app/api/login` 区分），接受 **`jsCode`**（微信小程序 `wx.login` 返回的**临时**登录凭证）与 **`platform`**（与 device 配置 `wechatMp.platforms` 下键一致，用于选择该端的 `appId`/`appSecret`）。

系统 SHALL **使用服务端持有的凭据**调用微信 **`sns/jscode2session`** 将 `jsCode` 换为 **openid** 与 **unionid**；**SHALL NOT** 接受客户端直传的 openid/unionid 作为身份依据。持久化身份键为 **`wx` 表 `union_id`**（微信开放平台 **unionid**，多端统一）；若微信响应中 **unionid 为空**（未绑定开放平台等），SHALL 返回明确业务错误且 **SHALL NOT** 创建或匹配用户行。

若 `union_id` 在库中不存在则创建 wx 行；响应 SHALL 包含至少 wx 表主键 **id（wxId）**、是否新用户（**isNewUser**）、已绑定时的 **device_no**；响应 **SHALL NOT** 包含 `access_token`/`refresh_token`（由 gateway-app 签发），且 **SHALL NOT** 向客户端回传 **unionid**、**openid**、**session_key**。

#### Scenario: 新用户注册业务结果

- **WHEN** 首次出现的 **unionid**（经换票得到）调用登录接口
- **THEN** 系统 SHALL 创建 wx 行并返回 is_new_user 为真（或等价字段），且 device_no 字段 MAY 为空

#### Scenario: 已存在用户

- **WHEN** **unionid** 已存在于 `wx.union_id`
- **THEN** 系统 SHALL 返回已有 wxId，并在已绑定设备时返回 device_no

### Requirement: 绑定设备与 wx

device-service SHALL 提供 `POST /device/app/api/user/bindwx`，从请求头读取 **`X-Internal-Wx-Union-Id`**（值为 **unionid**，由 gateway-app 根据 JWT `sub` 解析后注入），从 JSON body 读取 `deviceNo`，并将二者写入 ai_voice_device 库中的 wx 表（或等价绑定语义）。

#### Scenario: 绑定成功

- **WHEN** 请求头包含有效 **unionid** 且 body 中 deviceNo 合法
- **THEN** 系统 SHALL 持久化绑定关系并返回成功语义

### Requirement: 自动保存画像

device-service SHALL 提供 `POST /device/app/api/user/auto_save`，从请求头读取 **`X-Internal-Wx-Union-Id`**，从 body 读取 `birthday` 与 `sex`，并 SHALL 返回 `device_no`。设备域中 `device_no` SHALL 在**全表范围内唯一**（与现有 `user`/设备表唯一约束一致）。当该 wx 尚未绑定设备时，系统 SHALL 在设备域注册表中 **新建一条设备记录**，其 `device_no` 为 **6 个字符**，且每个字符均为 **大写英文字母 A–Z** 的随机取值，且该值 SHALL **不与任何已存在的** `device_no` 重复；将该 `device_no` 与当前 wx 绑定后再写入画像（生日、性别），语义与现有用户画像保存能力对齐。当该 wx **已绑定** `device_no` 时，系统 SHALL **仅更新**画像并返回**已有**的 `device_no`，且 SHALL NOT 更换设备号。

#### Scenario: 无设备号时创建并绑定

- **WHEN** **unionid** 有效且当前 wx 行未绑定 device_no，且 birthday/sex 合法
- **THEN** 系统 SHALL 生成符合「6 位大写 A–Z」规则的 device_no、在设备表中创建新设备、与 wx 绑定、保存画像，且响应 SHALL 包含该新 device_no，且持久化后的 device_no SHALL 与库内其它行的 device_no 均不相同

#### Scenario: 已绑定设备仅更新画像

- **WHEN** **unionid** 有效且 wx 已绑定 device_no
- **THEN** 系统 SHALL 仅更新画像字段且响应 SHALL 返回原 device_no，且 SHALL NOT 生成新的随机 device_no

#### Scenario: 设备号与已有数据冲突时重试直至唯一

- **WHEN** 随机生成的候选 device_no 已存在于设备表中（或由唯一索引拒绝插入）
- **THEN** 系统 SHALL 放弃该候选并重新生成，重复直至插入成功或达到实现约定的最大重试次数；成功时持久化的 device_no SHALL 满足全局唯一约束

### Requirement: 按 unionid 查询设备号

device-service SHALL 提供 `GET /device/app/api/user/detail`，从请求头读取 **`X-Internal-Wx-Union-Id`**，并 SHALL 返回绑定的 `device_no`（若未绑定则返回约定空或错误语义）。

#### Scenario: 已绑定返回设备号

- **WHEN** **unionid** 对应记录已绑定 device_no
- **THEN** 响应 SHALL 包含该 device_no

### Requirement: 按主键 id 解析 unionid（内部）

device-service SHALL 提供仅供内网或网关调用的只读接口（例如 `GET /device/app/api/user/internal/by-id`），根据 wx 表主键 id 返回对应 **`union_id`（响应字段 unionId）**，以便 gateway-app 在仅持有 access 内 id 时解析 **unionid** 并写入 **`X-Internal-Wx-Union-Id`**；该接口 SHALL 不对外网匿名开放（依赖部署网络或额外共享密钥策略）。

#### Scenario: 有效 id

- **WHEN** 网关使用有效 id 调用内部解析接口
- **THEN** 响应 SHALL 包含与该 id 对应的 unionId

#### Scenario: 无效 id

- **WHEN** id 不存在或非法
- **THEN** 系统 SHALL 返回明确错误且 SHALL NOT 泄露其他行信息

### Requirement: Redis 缓存与失效

device-service 对上述只读或高频读路径（含 **id→unionid**）SHALL 可选使用 Redis 缓存；在更新 wx 绑定或影响 **unionid** 映射的写操作成功后，SHALL 使相关缓存键失效。

#### Scenario: 绑定后缓存一致性

- **WHEN** bindwx 成功完成
- **THEN** 与该 wx 行相关的 id→unionid 或 detail 相关缓存 SHALL 被删除或失效，使得后续读取获得新状态
