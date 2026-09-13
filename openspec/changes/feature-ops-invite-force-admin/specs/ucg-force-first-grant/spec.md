## ADDED Requirements

### Requirement: 原力首次增量可写入空账户

当 `ucg_user_force` 尚无该 `wx_id` 行时，系统执行原力增量（含获客 `invite_acquisition` +100 与辩论自投 +1）MUST 成功插入余额行并写入 `ucg_force_ledger`，MUST NOT 因空结果集 `Scan`/`sql.ErrNoRows` 失败。已有行时 MUST 在事务内累加并写流水。

#### Scenario: 首次获客加分

- **WHEN** cash 通知 ucg 对无原力行的码主人执行获客加分且入参合法
- **THEN** 该用户 `force_value` 为 100（或既有 0+100），且 ledger 存在 `invite_acquisition` 记录

#### Scenario: 已有余额再加分

- **WHEN** 用户已有 `force_value=100` 再次获客成功加分
- **THEN** `force_value` 变为 200 且新增一条 ledger

### Requirement: 我的资料返回原力值

`GET` 当前用户 profile（如 `/ucg/app/api/profile/me`）成功时 MUST 填充与本域 `ucg_user_force` 一致的 `forceValue`（无行视为 0），MUST NOT 因未 enrich 而长期省略真实非零原力。

#### Scenario: 有原力时 me 可见

- **WHEN** 用户 `ucg_user_force.force_value=100` 且请求我的资料
- **THEN** 响应 MUST 含 `forceValue` 为 100
