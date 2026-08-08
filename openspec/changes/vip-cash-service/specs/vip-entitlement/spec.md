## ADDED Requirements

### Requirement: VIP 权益以 ai_voice_cash 表为准

账号是否 VIP MUST 由 `ai_voice_cash` 中的权益记录判定，主键为 `wx.id`（`wx_id`）。MUST NOT 使用 `ai_voice_device.wx.is_vip` 或 `device_no` 作为 VIP 真相源。当且仅当权益 `expire_at` 晚于当前时间时，该账号 MUST 被视为 VIP。

#### Scenario: 无权益记录

- **WHEN** 查询某 `wxId` 且不存在权益行或 `expire_at` 已过期
- **THEN** 系统 MUST 判定 `isVip=false`

#### Scenario: 未过期权益

- **WHEN** 查询某 `wxId` 且权益 `expire_at` 大于当前时间
- **THEN** 系统 MUST 判定 `isVip=true`

### Requirement: 内部按 wxId 查询 VIP

cash-service MUST 提供经内部密钥鉴权的 HTTP 接口 `GET /cash/internal/api/vip/by-wx-id`（query `wxId`），响应至少包含 `wxId` 与 `isVip`（宜含 `expireAt`）。voice-service 等跨域调用方 MUST 经此契约读取 VIP，MUST NOT 直查 `ai_voice_cash`。

#### Scenario: 合法内部查询

- **WHEN** 调用方携带正确内部密钥请求正整数 `wxId`
- **THEN** cash-service MUST 返回该账号当前 VIP 布尔结果

#### Scenario: 缺少内部密钥

- **WHEN** 请求未携带正确内部密钥
- **THEN** cash-service MUST 拒绝该请求

#### Scenario: 无效 wxId

- **WHEN** `wxId` 缺失或 ≤0
- **THEN** cash-service MUST 返回参数错误，MUST NOT 视为 VIP
