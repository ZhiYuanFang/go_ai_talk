## ADDED Requirements

### Requirement: sim-user-service SHALL persist simulated user credentials for admin and task login

`sim-user-service` MUST 在 `SIM_DB_LINK` 库维护表 `sim_wx_credential`（或语义等价名），字段至少含：

- `wx_id`（BIGINT PRIMARY KEY）
- `account`（VARCHAR，非空）
- `password_plain`（VARCHAR，非空）
- `created_at`（BIGINT Unix 秒，非空）

T1 在 device internal 注册成功且获得 `wxId` 后 MUST 立即 INSERT 一行 credential（`created_at` 为注册时刻）。`EnsureSchema` MUST 幂等创建该表。

注销 sim 用户时 MUST DELETE 对应 `wx_id` 行。MUST NOT 在响应日志中打印 `password_plain`。

#### Scenario: Credential written on register

- **WHEN** T1 成功注册新 sim 用户 wxId=2001
- **THEN** `sim_wx_credential` MUST 含 wx_id=2001 与非空 `password_plain`

#### Scenario: Credential removed on admin deactivate

- **WHEN** admin deactivate 成功删除 wxId=2001
- **THEN** credential 行 MUST 不存在

## MODIFIED Requirements

### Requirement: T1 register task SHALL create simulated users every 24 hours

每 24 小时（±10% jitter）MUST 执行注册任务（`SIM_TASK_REGISTER_ENABLED`）。当当前 sim 用户数 `< maxSimUsers` 时，MUST 生成**随机**账号与**随机**密码（均满足 device `WxUsernameRegister` 规则），调用 device internal sim 注册，写入 `sim_wx_credential`，经 `simText` 生成昵称、`simImageGen` 生成头像，完成 UCG profile 与 media 上传，并更新 nickname/avatarKey。当已达 `maxSimUsers` 时 MUST 跳过本次执行。

MUST NOT 再分配 `ptest{N}` 序号账号，MUST NOT 对全部新用户使用固定默认密码注册。`sim_account_seq` MUST NOT 参与 T1 账号生成。

#### Scenario: Register under cap with random credentials

- **WHEN** sim 用户数为 99 且 `maxSimUsers=100`
- **THEN** 系统 MUST 注册新 sim 用户并标记 `is_simulated=1`，且账号 MUST NOT 匹配 `^ptest\d+$` 固定序号模式（应为随机生成）

#### Scenario: Skip at cap

- **WHEN** sim 用户数已达 `maxSimUsers`
- **THEN** T1 MUST NOT 调用注册接口

### Requirement: sim task random user login SHALL use per-wxId credential password

当 T2、T3、T4、T5、T6 或手动任务需经 gateway `username_login` 登录某一 sim 用户时，sim-user-service MUST 按该用户 `wxId` 从 `sim_wx_credential` 读取 `password_plain` 作为登录密码。

若 credential 无记录（历史用户），MUST fallback 到运行时 `defaultPassword`（yaml/env `simUser.defaultPassword`，空则 `123456`）。

MUST NOT 假定所有 sim 用户共用同一注册密码（除上述历史 fallback）。

#### Scenario: Task login uses stored password

- **WHEN** T2 随机选中 wxId=2001 且 credential 存在
- **THEN** `username_login` MUST 使用 credential 中的 `password_plain`

#### Scenario: Legacy sim login fallback

- **WHEN** 随机选中历史 wxId 无 credential 行
- **THEN** login MUST 使用 defaultPassword fallback
