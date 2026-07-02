## ADDED Requirements

### Requirement: wx table SHALL store account creation timestamp

`device-service` 权威库表 `wx` MUST 新增列 `created_at`（`BIGINT NOT NULL DEFAULT 0`，Unix 秒）。该字段 MUST 表示 wx 账号行**首次创建**时刻。历史行在迁移后 `created_at` MAY 为 0；展示层对 0 MUST 显示「—」。

#### Scenario: New wx row gets created_at on insert

- **WHEN** 用户经微信 OAuth、Apple 登录或用户名注册**首次**创建 wx 行
- **THEN** 插入时 `created_at` MUST 为当前 Unix 秒且大于 0

#### Scenario: Legacy row defaults to zero

- **WHEN** 迁移脚本执行后存在迁移前已创建的 wx 行且未回填
- **THEN** 该行 `created_at` MUST 为 0

### Requirement: wx creation write paths SHALL set created_at

下列 device-service 写路径在 **Insert** 新 wx 行时 MUST 写入 `created_at`：

- `WxLogin`（unionid 首次出现）
- Apple 首次登录 Insert
- `WxUsernameRegister`

模拟用户经 `SimUsernameRegister` 创建时 MUST 同样写入 `created_at`（与真实用户相同写路径）。

#### Scenario: WeChat first login

- **WHEN** `WxLogin` 为新的 unionid 插入 wx 行
- **THEN** 新行 `created_at` MUST 大于 0

#### Scenario: Username register

- **WHEN** `WxUsernameRegister` 成功插入新账号
- **THEN** 新行 `created_at` MUST 大于 0
