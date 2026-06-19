## ADDED Requirements

### Requirement: wx table SHALL store simulated user flag

`device-service` 权威库表 `wx` MUST 新增列 `is_simulated`（TINYINT NOT NULL DEFAULT 0）。`1` 表示模拟用户；公开 App 注册路径 MUST NOT 允许客户端自行设置为 1。

#### Scenario: Default not simulated

- **WHEN** 用户经公开 `username/register` 注册
- **THEN** 新行 `is_simulated` MUST 为 0

### Requirement: device internal sim register SHALL wrap username register with flag

`POST /device/internal/api/sim/username/register` MUST 要求有效 `X-Device-Gateway-Internal-Secret`（或兼容 `X-Gateway-Internal-Secret`）。请求体 MUST 含 `account`、`password`。服务 MUST 调用 `WxUsernameRegister` 并在同一注册流程内将 `is_simulated` 设为 1。成功响应 MUST 含 `wxId`、`account`。账号已存在 MUST 返回业务错误。

#### Scenario: Internal sim register

- **WHEN** sim-service 携带有效密钥 POST `{account:"ptest5",password:"123456"}`
- **THEN** 响应 MUST 含 `wxId>0` 且该 wx 行 `is_simulated=1`

#### Scenario: Reject without secret

- **WHEN** 无内部密钥调用 sim register
- **THEN** HTTP MUST 为 403

### Requirement: device internal SHALL list and batch-query simulated users

device-service MUST 提供：

- `GET /device/internal/api/sim/wx/list` — 分页返回 `is_simulated=1` 的 wxId 列表（供随机选取）
- sim 批量查询 MUST 在现有 `POST /device/internal/api/ucg/wx/batch` 响应项中增加 `isSimulated` 字段，或提供等价的 sim 专用 batch 接口

#### Scenario: List sim users

- **WHEN** sim-service 请求 sim wx list
- **THEN** 返回列表 MUST 仅含 `is_simulated=1` 的 wxId

#### Scenario: Batch includes flag

- **WHEN** UCG 或 gateway 批量查询 wxId 展示字段
- **THEN** 每项 MUST 含 `isSimulated` 布尔值
