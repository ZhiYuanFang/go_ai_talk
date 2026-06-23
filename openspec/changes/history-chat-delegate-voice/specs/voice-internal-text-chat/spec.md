## ADDED Requirements

### Requirement: voice-service 提供 internal 文本对话 API

voice-service SHALL 提供 `POST /voice/internal/api/text/chat`，接受 JSON body `deviceNo`、`transcript`；调用方 MUST 在 Header 携带 `X-Device-Gateway-Internal-Secret`（与 `DEVICE_GATEWAY_INTERNAL_SECRET` 一致）。若需喂养 AI 额度预检，调用方 SHOULD 携带 `X-Internal-Wx-Id`（正整数 wxId）。

成功时响应 data MUST 含 `reply` 字符串。失败时 MUST 使用与 App 网关一致的 business code（含 40301 未登录、40302 额度用尽）。

#### Scenario: 合法 internal 请求成功

- **WHEN** secret 正确且 `deviceNo`、`transcript` 合法且额度允许
- **THEN** 接口 MUST 返回 `code=0` 且 `reply` 为非空或允许的空串回复

#### Scenario: 未登录 wxId

- **WHEN** 未携带有效 `X-Internal-Wx-Id` 且链路需要额度预检
- **THEN** 接口 MUST 返回 business code 40301

#### Scenario: 额度用尽

- **WHEN** wxId 有效且当月 voice_ai 额度已用尽
- **THEN** 接口 MUST 返回 business code 40302

#### Scenario: secret 无效

- **WHEN** `X-Device-Gateway-Internal-Secret` 缺失或错误
- **THEN** 接口 MUST 拒绝请求且 MUST NOT 执行 TextChat
