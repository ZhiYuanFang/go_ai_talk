## ADDED Requirements

### Requirement: device internal SHALL deactivate simulated wx by wxId

`device-service` MUST 提供 `POST /device/internal/api/sim/wx/{wxId}/deactivate`，要求有效 `X-Device-Gateway-Internal-Secret`（与 sim register 相同）。

服务 MUST 校验目标 wx 行存在且 `is_simulated=1`；否则 MUST 返回 4xx 业务错误且 MUST NOT 删除。

校验通过后 MUST 调用 `WxDeactivateByID` 删除 wx 单行，并 MUST 失效 wx 相关 cachekit 键；MUST 从 `usage:sim_wx_ids` SET 移除该 wxId member（经 `cachekit.GatewayUsageSimWxSetKey` / `GatewayUsageSimWxMember`）。

MUST NOT 删除 ucg 或 user 域数据。

#### Scenario: Deactivate sim wx success

- **WHEN** sim-service 携带有效密钥 POST deactivate 且 wxId 为 `is_simulated=1`
- **THEN** wx 行 MUST 删除且 HTTP MUST 成功

#### Scenario: Reject real user deactivate

- **WHEN** wxId 对应 `is_simulated=0`
- **THEN** MUST 返回 4xx 且 wx 行 MUST 保持不变

#### Scenario: Reject without secret

- **WHEN** 无内部密钥调用 sim deactivate
- **THEN** HTTP MUST 为 403

## MODIFIED Requirements

### Requirement: device internal sim register SHALL wrap username register with flag

`POST /device/internal/api/sim/username/register` MUST 要求有效 `X-Device-Gateway-Internal-Secret`（或兼容 `X-Gateway-Internal-Secret`）。请求体 MUST 含 `account`、`password`。服务 MUST 调用 `WxUsernameRegister` 并在同一注册流程内将 `is_simulated` 设为 1。成功响应 MUST 含 `wxId`、`account`。账号已存在 MUST 返回业务错误。

调用方（sim-user-service T1）MUST 传入随机生成的 account/password，MUST NOT 依赖固定 `ptest{N}` 或固定默认密码作为长期约定。

#### Scenario: Internal sim register with random account

- **WHEN** sim-service 携带有效密钥 POST `{account:"s8k2m9xq4n",password:"Kp9#mX2vLq4n"}`
- **THEN** 响应 MUST 含 `wxId>0` 且该 wx 行 `is_simulated=1`

#### Scenario: Reject without secret

- **WHEN** 无内部密钥调用 sim register
- **THEN** HTTP MUST 为 403
