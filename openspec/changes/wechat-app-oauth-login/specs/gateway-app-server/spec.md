## MODIFIED Requirements

### Requirement: 登录与令牌仅由 gateway-app 签发

系统 SHALL 在 gateway-app-server 上暴露 `POST /device/app/api/login`，其通过 HTTP 调用 device-service 的 `POST /device/app/api/user/login`（Body 含 **`jsCode`** 与 **`platform`**；其中 **`jsCode`** 为微信开放平台授权临时 `code`，**`platform`** 为客户端传入的平台键如 `ios`/`android`/`web`，语义见 device 规格）获取业务字段后签发 access_token 与 refresh_token；其中 **access_token SHALL 为纯 JWT**，其载荷 **MUST** 包含标准 claim **`sub`** 且其值等于 wx 表主键 id（与 device 返回的 wxId 一致），**MUST** 包含 **`iat`** 与 **`exp`**；**refresh_token SHALL NOT** 为 JWT，SHALL 为高熵不透明串并与 Redis 会话绑定以便刷新与吊销。

请求与响应 JSON 字段名 SHALL 与现网 App 约定保持一致（**不得**将 `jsCode` 重命名为其他键名以强制客户端发版）。

#### Scenario: 登录成功

- **WHEN** 客户端（iOS、Android 或网页前端）调用 gateway-app 的登录接口且 device 返回有效业务结果（含 wxId 等，**不含**需保密的 unionid/openid）
- **THEN** 响应 SHALL 包含 access_token 与 refresh_token（及 device 返回的约定业务字段），且 access_token SHALL 可被验证为结构正确的 JWT，且 device-service SHALL NOT 在 `POST /device/app/api/user/login` 响应中返回 JWT 形式的 access_token

#### Scenario: 网页端复用同一登录契约

- **WHEN** 网页在 `qrconnect` 回调取得 `code` 后向 `POST /device/app/api/login` 提交 `jsCode` 与 `platform=web`
- **THEN** 网关 SHALL 原样转发至 device-service 并按与 App 相同的规则签发令牌
