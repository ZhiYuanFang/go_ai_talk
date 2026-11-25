# Spec Delta

## ADDED Requirements

### Requirement: wx SHALL 具备账号级单值 role 且 parent 与 staff 互斥
`wx` MUST 持久化账号级 `role`，取值仅为 `parent` 或 `staff`（缺省新建 C 端用户为 `parent`）。系统 MUST NOT 允许同一 `wx` 行同时或先后兼具家长自用与门店 staff 语义：将 `parent` 改为 `staff` 或反向 MUST 拒绝（除非无宝宝成员且产品显式允许的空户迁移——一期 MUST 拒绝任何 role 变更）。商户创建子账户 MUST 以 `username/register` 等价路径新建且 `role=staff`。

#### Scenario: 商户创建 staff
- **WHEN** 商户提交未占用的用户名与密码创建子账户
- **THEN** 系统 MUST 新建 `wx` 且 `role=staff`，并返回可在 C 端登录的账号信息

#### Scenario: parent 不可改为 staff
- **WHEN** 对已有 `role=parent` 的 wx 请求改为 `staff`
- **THEN** 系统 MUST 拒绝

#### Scenario: staff 不可当家长自建互斥路径
- **WHEN** `role=staff` 的账号尝试以家长语义创建「非商户分发」的互斥流程被产品禁止的操作（若一期仅允许商户分发宝宝给 staff）
- **THEN** 系统 MUST 按 staff 规则校验成员来源且 MUST NOT 将其 `role` 改为 `parent`

### Requirement: 宝宝成员关系 SHALL 以多对多表为真源并保留焦点镜像
系统 MUST 以 `(wx_id, device_no)` 成员关系表作为账号可管理宝宝的真源。`wx.device_no` MUST 继续存在并表示 **焦点宝宝**（供旧客户端与 JWT claim）。登录与 `detail` 对旧客户端 MUST 仍返回标量 `deviceNo`（焦点）。新客户端 MUST 能列出该 wx 全部成员宝宝并切换焦点（换发或更新会话焦点）。

#### Scenario: 迁移后旧语义可读
- **WHEN** 仅有一个成员宝宝的 parent 用旧客户端登录
- **THEN** 响应 MUST 含该宝宝的 `deviceNo`，行为与「一账号一宝宝」可观察一致

#### Scenario: 新客户端列出多个宝宝
- **WHEN** staff 或 parent 拥有多个成员宝宝且新客户端请求宝宝列表
- **THEN** 系统 MUST 返回全部成员 `deviceNo` 及展示名

#### Scenario: 切换焦点
- **WHEN** 新客户端将焦点切换到其成员集合内的另一宝宝
- **THEN** 系统 MUST 更新焦点镜像与后续会话 `device_no`，且 MUST NOT 删除其它成员边

### Requirement: parent 新客户端 SHALL 可创建或绑定虚拟宝宝
`role=parent` 的新客户端 MUST 能在 App 内 **创建虚拟宝宝**（注册新 `device_no` 并加入成员表）或 **绑定** 已有虚拟/已注册 `device_no`（扫码或等价，经 preview 确认）。客户端 MUST NOT 提供解绑。旧客户端 MUST 仅支持从无宝宝到绑定一个宝宝；旧客户端无「再绑另一个」入口时，服务端对已有成员的旧式 bind 再绑其它号 MUST 拒绝或保持防御性错误。

#### Scenario: parent 创建第二个虚拟宝宝
- **WHEN** 已有一个成员宝宝的 parent 在新客户端创建另一虚拟宝宝
- **THEN** 系统 MUST 生成新 `device_no`、写入成员表，且焦点策略按产品默认（可保持原焦点或切到新宝宝，但 MUST 文档化）

#### Scenario: 旧客户端无宝宝首次绑定
- **WHEN** 无成员的旧客户端完成一次 bindwx
- **THEN** 系统 MUST 写入成员表与焦点镜像

#### Scenario: 客户端解绑被拒
- **WHEN** 任意 App 会话请求解绑宝宝成员
- **THEN** 系统 MUST 拒绝

### Requirement: 虚拟宝宝 SHALL 为无实体硬件的 device_no
商户或 parent 创建的宝宝 MUST 在设备域注册表登记合法 `device_no`（与现网随机号规则兼容），MUST NOT 要求实体硬件在线。画像（昵称/生日/性别等）MUST 允许商户后台写入，也 MUST 允许持有成员关系的子账户或 parent 经 App 既有/扩展画像 API 写入。

#### Scenario: 商户创建虚拟宝宝
- **WHEN** 商户在额度内创建宝宝并可选填写画像
- **THEN** 系统 MUST 注册 `device_no` 并归属该商户，画像字段按提交持久化

#### Scenario: 子账户补全画像
- **WHEN** staff 对已分发宝宝在 App 更新画像
- **THEN** 系统 MUST 更新该 `device_no` 画像且 MUST NOT 要求实体设备

### Requirement: staff 多宝宝时旧客户端 SHALL 仅见焦点宝宝
当 `role=staff` 拥有多个成员宝宝时，旧客户端 MUST 仅能通过标量 `deviceNo`/焦点访问其中一个。商户分发 MUST 能指定或默认该 staff 的焦点宝宝。系统 MUST NOT 因多成员而破坏旧客户端登录。

#### Scenario: 旧 App 登录 staff
- **WHEN** 拥有三个成员宝宝的 staff 用旧 App 登录
- **THEN** 系统 MUST 返回单一焦点 `deviceNo` 且旧 App 可继续读写该宝宝（在写闸允许范围内）
