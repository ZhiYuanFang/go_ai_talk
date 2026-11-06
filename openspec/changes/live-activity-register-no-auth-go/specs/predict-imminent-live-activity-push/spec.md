## MODIFIED Requirements

### Requirement: Live Activity tokens SHALL use a dedicated App path and storage

The twin Go service MUST expose App-facing `POST /app/api/push/live-activity/register` (separate from `POST /app/api/push/register`) for registering and updating Live Activity Push-to-Start and per-activity update tokens. Those tokens MUST be stored in a push-domain table (or equivalent) distinct from `push_device` device-token uniqueness (`uk_token`). The request body MUST include a non-zero `wxId` together with `deviceKey`, `tokenKind` (`push_to_start`|`activity`), `attributesKind` (`imminent`|`timing`), and `token`; when `tokenKind` is `activity`, `activityId` MUST be non-empty. Successful upsert MUST use body `wxId` as the owning user id and MUST NOT require a valid Bearer/JWT (or Bearer-derived `X-Internal-Wx-Id`) as a prerequisite. The existing device push register/unregister paths MUST continue to require Bearer and MUST store only vendor/device alert tokens. Live Activity tokens MUST NOT be written into `push_device.token` under the global device-token unique-key semantics.

Go MUST 提供并列于 `/app/api/push/register` 的 `POST /app/api/push/live-activity/register`，用于注册/更新 Live Activity 的 Push-to-Start 与 per-activity update token；MUST 落在 push 域分表（或等价），MUST NOT 写入 `push_device` 的 device `uk_token` 语义。请求体 MUST 含非零 `wxId` 以及既有字段；`activity` 时 `activityId` MUST 非空。成功 upsert MUST 以 body.`wxId` 为归属用户，MUST NOT 以有效 Bearer/JWT（或由其注入的 `X-Internal-Wx-Id`）为成功前置条件。既有 device register/unregister MUST 仍要求 Bearer，且 MUST 仅继续承载厂商/设备 alert token。

#### Scenario: Unauthenticated register with body wxId succeeds

- **WHEN** 客户端（含 iOS 原生唤醒路径）POST `/app/api/push/live-activity/register`，无 Authorization，body 含非零 `wxId` 与合法既有字段
- **THEN** gateway-app MUST NOT 仅因缺少 Bearer 返回 401
- **AND** Go MUST 将该 token upsert 到该 `wxId` 名下的 LA 专用存储

#### Scenario: Missing wxId rejected

- **WHEN** 请求缺少非零 `wxId`（即使带有有效 Bearer 与内部头）
- **THEN** 服务端 MUST 拒绝该次 register（参数错误类）
- **AND** MUST NOT 静默用内部头 `X-Internal-Wx-Id` 顶替 body.`wxId`（本变更默认不兼容旧客户端）

#### Scenario: LA token not stored as device push token

- **WHEN** 客户端上报 Push-to-Start 或 per-activity update token
- **THEN** Go MUST 经 live-activity register path 持久化到 LA 专用存储
- **AND** MUST NOT 将该 token 作为 `push_device` 全局唯一 device token 写入

#### Scenario: Device register unchanged for alerts and still requires Bearer

- **WHEN** 客户端调用既有 `/app/api/push/register` 或 `unregister` 上报/注销 device alert token
- **THEN** 行为 MUST 保持 device alert 注册语义与既有 Bearer 要求
- **AND** MUST NOT 因本变更被鉴权豁免
- **AND** MUST NOT 要求该调用携带 LA token 才能工作

## ADDED Requirements

### Requirement: Gateway-app SHALL exempt live-activity register from mandatory Bearer

The gateway-app Bearer middleware MUST treat `POST /app/api/push/live-activity/register` as an auth-exempt exact POST path (same maintenance style as other entries in `gatewayAppAuthExemptExactPOST`). Spoofed internal identity headers MUST still be stripped before handling. If a non-admin user Bearer is present on an exempt request, the gateway MAY still inject internal headers, but Live Activity register persistence identity MUST follow body `wxId` rules. The gateway MUST NOT exempt the entire `/app/api/push/` prefix solely to satisfy this requirement.

gateway-app Bearer 中间件 MUST 将 `POST /app/api/push/live-activity/register` 列为精确 POST 鉴权豁免（维护方式同 `gatewayAppAuthExemptExactPOST`）。伪造内部身份头 MUST 仍被剥离。豁免请求若携带非 admin 用户 Bearer，网关 MAY 仍注入内部头，但 LA register 落库身份 MUST 遵循 body.`wxId`。MUST NOT 仅为此需求豁免整个 `/app/api/push/` 前缀。

#### Scenario: Exact path exempt listed

- **WHEN** 检查 gateway-app 鉴权豁免配置
- **THEN** `POST /app/api/push/live-activity/register` MUST 出现在精确 POST 豁免集合中
- **AND** MUST NOT 通过豁免整个 `/app/api/push/` 前缀来实现
