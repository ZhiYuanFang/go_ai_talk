## MODIFIED Requirements

### Requirement: 邀请码兑换 MUST 按单功能开通并遵守一家绑定与人维去重

cash-service MUST 提供 `POST /cash/app/api/feature/invite-codes/redeem`（计入 usage）。请求 MUST 含邀请码与目标 `featureId`。主体：`device_no` 与 `wx_id` 均只信网关注入头；`wx_id` MUST > 0。

规则 MUST 包括：

- 一码可覆盖其配置的（或所有支持邀请码的）多个功能，但单次请求 MUST 只开通所请求的一个 `featureId`，MUST NOT 自动开通其余功能。
- 同一 `(redeemer_wx_id, code, feature_id)` 仅能成功一次（人×码×功能去重）。
- 系统 MUST NOT 再执行「一家锁定」（redeemer 绑定单一 owner）；MUST NOT 因已兑过其他 owner 的码而拒绝新 owner（其它功能级 InviteOncePerUser 等约束除外）。
- 系统 MUST NOT 仅因 `redeemer_wx_id` 等于 `owner_wx_id` 而拒绝（已取消自用禁兑）。
- 系统 MUST 拒绝码主人当前绑定 `device_no` 与兑换者请求 `device_no` 相同且均非空的兑换（同宝宝禁兑）。
- 成功后 MUST 写设备权益或 `allowedCount`，`unlockMethod` MUST 为 `invite_code`，并写入兑换流水（含 owner/redeemer/device/feature/time）。

#### Scenario: 指定功能兑换成功

- **WHEN** 合法用户提交有效码与尚未码开的 `featureId` 且未违反同宝宝及其余人维规则
- **THEN** 系统 MUST 仅开通该功能（或增加对应数量），并占用相应兑换记录

#### Scenario: 同码再次开通另一功能

- **WHEN** 用户已用某码成功开通功能 A，再次提交同一码与功能 B（B 尚未码开且码支持 B）
- **THEN** 系统 MUST 允许开通 B，且 MUST NOT 因已开 A 而自动开其它功能

#### Scenario: 不同 owner 码在未触达功能级一次限制时可兑

- **WHEN** 用户已成功使用 owner=X 的码开通某功能后，再提交 owner=Y 的码开通**另一**仍允许邀请开通的功能，且未违反同宝宝与 InviteOncePerUser 等
- **THEN** 系统 MUST NOT 仅因「换一家」而拒绝

#### Scenario: 同宝宝拒兑

- **WHEN** 兑换者 `device_no` 与码主人当前绑定 `device_no` 相同且均非空
- **THEN** 系统 MUST 拒绝兑换

#### Scenario: 同人异机可兑自己的码

- **WHEN** 兑换人 `wx_id` 等于码的 `owner_wx_id`，且主人当前 `device_no` 与兑换者不同或主人未绑机，且其它校验通过
- **THEN** 系统 MUST NOT 因自用拒绝

#### Scenario: 同一功能不可再次码开（功能级一次时）

- **WHEN** 某 `wx_id` 已通过邀请码开通过某适用 InviteOncePerUser 的 `featureId`
- **THEN** 再次用任意邀请码开通同一 `featureId` MUST 被拒绝
