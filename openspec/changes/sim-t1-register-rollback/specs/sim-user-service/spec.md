## ADDED Requirements

### Requirement: T1 register SHALL rollback sim wx on profile setup failure

T1（`RunRegisterTask`，含 sim-admin 手动触发）在 device internal sim 注册成功获得 `wxId` 后，若 **未完成** UCG profile 写入（`PUT /ucg/app/api/profile/me` 含 nickname 与 avatarKey），则 MUST 视为注册失败。

下列任一步失败时 MUST 调用 device internal `POST /device/internal/api/sim/wx/{wxId}/deactivate` 回滚 wx（须 `is_simulated=1`），并 MUST 删除 sim 库 `sim_wx_credential` 中对应行（若存在）：

- `usernameLogin` 失败
- Prompt 加载失败
- 昵称 `simText` 调用失败，或生成 nickname trim 后为空
- 头像 `simImageGen` 调用失败，或返回 URL 为空
- 头像 upload / media 链路失败
- profile `PUT` 失败

回滚后 MUST `RecordTaskRun(register, false, ...)`，且该 wx MUST NOT 计入 sim 用户数。回滚 device 调用失败时 MUST 记录 warning 日志，任务仍记失败。

手动 T1 与 scheduler T1 MUST 共用上述语义。

#### Scenario: Nickname AI failure rolls back wx

- **WHEN** T1 已 simRegister 成功但 `simText` 昵称生成返回错误
- **THEN** 系统 MUST 注销该 wxId 且 MUST NOT 保留 `sim_wx_credential` 行

#### Scenario: Avatar AI failure rolls back wx

- **WHEN** T1 昵称生成成功但 `simImageGen` 失败
- **THEN** 系统 MUST 注销该 wxId

#### Scenario: Empty nickname rolls back wx

- **WHEN** T1 昵称 AI 返回仅空白字符
- **THEN** 系统 MUST 视为失败并 MUST 注销该 wxId

#### Scenario: Profile PUT success commits registration

- **WHEN** T1 profile PUT 成功且 nickname 非空、avatarKey 非空
- **THEN** 系统 MUST 写入 `sim_wx_credential` 且 MUST NOT 回滚 wx

#### Scenario: Failed registration does not consume cap slot

- **WHEN** T1 因头像 AI 失败回滚 wx 且当前 sim 用户数为 99、`maxSimUsers=100`
- **THEN** 下一次 T1 MUST 仍可尝试注册新用户

## MODIFIED Requirements

### Requirement: sim-user-service SHALL persist simulated user credentials for admin and task login

`sim-user-service` MUST 在 `SIM_DB_LINK` 库维护表 `sim_wx_credential`（或语义等价名），字段至少含：

- `wx_id`（BIGINT PRIMARY KEY）
- `account`（VARCHAR，非空）
- `password_plain`（VARCHAR，非空）
- `created_at`（BIGINT Unix 秒，非空）

T1 MUST 在 **UCG profile PUT 成功之后** INSERT credential（`created_at` 为 profile 写入完成时刻）。MUST NOT 在 profile 完成前写入 credential。`EnsureSchema` MUST 幂等创建该表。

注销 sim 用户时 MUST DELETE 对应 `wx_id` 行。MUST NOT 在响应日志中打印 `password_plain`。

#### Scenario: Credential written after profile success

- **WHEN** T1 profile PUT 成功完成
- **THEN** `sim_wx_credential` MUST 含对应 wx_id 与非空 `password_plain`

#### Scenario: Credential not written on profile failure

- **WHEN** T1 在 profile PUT 之前失败并已回滚 wx
- **THEN** MUST NOT 存在该 wx_id 的 credential 行

#### Scenario: Credential removed on admin deactivate

- **WHEN** admin deactivate 成功删除 wxId=2001
- **THEN** credential 行 MUST 不存在

### Requirement: T1 register task SHALL create simulated users every 24 hours

每 24 小时（±10% jitter）MUST 执行注册任务（`SIM_TASK_REGISTER_ENABLED`）。当当前 sim 用户数 `< maxSimUsers` 时，MUST 生成**随机**账号与**随机**密码，调用 device internal sim 注册，经 `simText` 生成昵称、`simImageGen` 生成头像，完成 UCG profile 与 media 上传并更新 nickname/avatarKey；**仅在 profile 写入成功后** MUST 写入 `sim_wx_credential`。profile 完成前任一步失败 MUST 回滚 wx（见 ADDED Requirement）。当已达 `maxSimUsers` 时 MUST 跳过本次执行。

MUST NOT 再分配 `ptest{N}` 序号账号。`sim_account_seq` MUST NOT 参与 T1 账号生成。

#### Scenario: Register under cap with random credentials

- **WHEN** sim 用户数为 99 且 `maxSimUsers=100` 且 T1 全流程成功
- **THEN** 系统 MUST 注册新 sim 用户并标记 `is_simulated=1`，且 MUST 写入 credential

#### Scenario: Skip at cap

- **WHEN** sim 用户数已达 `maxSimUsers`
- **THEN** T1 MUST NOT 调用注册接口
