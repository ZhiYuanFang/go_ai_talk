## ADDED Requirements

### Requirement: 成长轨迹智能分析 MUST 校验账号开通或 VIP

`POST /device/api/growth-trajectory/turn`（及 voice 内等价入口）在开流前 MUST 要求有效登录 `wxId`，并 MUST 经 cash 合成：该 `wxId` 对 `growth_trajectory_predict` 存在未过期 user 权益 **或** 该账号为 VIP。未满足 MUST 拒绝且 MUST NOT 调用 Python turn。合成 MUST NOT 仅因同 `device_no` 上他人已开通而放行。

#### Scenario: 未开通非 VIP 拒绝 turn

- **WHEN** 已登录用户无成长轨迹 user 权益且非 VIP 请求 turn
- **THEN** 系统 MUST 返回未开通类错误且不开 SSE 业务流

#### Scenario: VIP 免开通可 turn

- **WHEN** 用户无 user 权益但是 VIP 请求 turn
- **THEN** 系统 MUST 允许进入分析流（仍受日限等其它闸门约束）

### Requirement: 成长轨迹最新结果 MUST 按设备读取且免开通

`GET /device/api/growth-trajectory/latest` MUST 要求登录，MUST 按请求 `deviceNo` 返回该设备最新一条结果（无则空字段），MUST NOT 校验功能开通或 VIP。存储仍为按设备最新一条，MUST NOT 本变更引入多版本历史列表。

#### Scenario: 未开通用户可读最新报告

- **WHEN** 已登录但未开通成长轨迹的用户请求 latest 且该 deviceNo 已有结果
- **THEN** 响应 MUST 返回已有 `resultMarkdown`（或等价字段）且 MUST NOT 因未开通失败

#### Scenario: 未登录拒绝 latest

- **WHEN** 缺少有效登录 wxId 请求 latest
- **THEN** 系统 MUST 拒绝

### Requirement: 成长轨迹日限 MUST 按用户计数

成长轨迹每日成功结果次数上限 MUST 以登录 `wxId` + 上海日历日为键计数（经 cachekit 键构建器），MUST NOT 再以 `deviceNo` 作为日限主键。达上限时 turn MUST 在开流前拒绝。

#### Scenario: 同人换设备共享日限

- **WHEN** 同一 wxId 在设备 D1 已用尽当日次数后在 D2 再请求 turn
- **THEN** 系统 MUST 因日限拒绝

### Requirement: 成长轨迹邀请 MUST 同时人一次与设备一次

对 `growth_trajectory_predict` 邀请开通，系统 MUST 在既有人×码×功能去重之外：若该 `redeemer_wx_id` 已对任意邀请码成功开通过本功能，MUST 拒绝再次邀请开通；若该 `device_no` 已对本功能邀请开通过，MUST 拒绝。设备一次闸 MUST 继续写入/校验 `feature_invite_device_grant`（或等价）。值得留意等其它 InviteOncePerDevice 功能 MUST NOT 因本变更被迫加上「人一次」除非另行配置。

#### Scenario: 同一人第二次兑码成长轨迹

- **WHEN** 用户已用码1开通成长轨迹后用码2再兑同一功能
- **THEN** 系统 MUST 拒绝且 MUST NOT 延长 user 权益

#### Scenario: 同设备第二人兑码成长轨迹

- **WHEN** 设备 D 已有任意账号邀请开通过成长轨迹后另一账号再兑
- **THEN** 系统 MUST 因设备已开通过而拒绝

### Requirement: cash 成长轨迹 access MUST 以账号权益为准

内部 `GET /cash/internal/api/growth-trajectory/access` MUST 以 `wxId` 的 user 权益 ∨ VIP 合成 `allowed`/`featureActive`，MUST NOT 将设备维 `feature_entitlement` 上的成长轨迹行视为有效开通（本变更后）。

#### Scenario: 仅有历史 device 行不算开通

- **WHEN** 某 device 上仍有旧成长轨迹 device entitlement，但当前 wx 无 user 权益且非 VIP
- **THEN** access MUST 返回不允许
