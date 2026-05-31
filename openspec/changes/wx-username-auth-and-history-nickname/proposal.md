## Why

当前账号体系仅覆盖微信授权登录与设备号登录，无法满足“用户名+密码”登录诉求，也无法让已存在微信用户补建本地账号凭据。与此同时，历史记录页面目前只展示性别，不展示昵称，影响用户识别与体验一致性，需要在同一轮改造中补齐。

## What Changes

- 在 `ai_voice_device.wx` 表（以 `wx.id` 为统一账号主键）新增并启用用户名密码体系，不新建独立用户表。
- 新增用户名相关业务接口：用户名注册、用户名登录、用户名绑定微信号、用户名绑定设备号、修改密码、微信账号下创建用户名密码。
- 约束账号关系：`unionid` 允许为空；一个微信（`unionid`）只能绑定一个用户名账号；用户名全局唯一。
- 密码字段改为哈希密文存储与校验（不可逆），禁止明文入库与明文比较。
- 网关聚合登录补充用户名登录路径，沿用现有 JWT/refresh 机制，`sub` 继续使用 `wx.id`。
- 历史页面画像信息补充 `nickname` 展示，前后端接口同步扩展返回字段。
- **BREAKING**：`device-wx-profile-apis` 中对 `unionid` 的“必须存在”语义将被修订为“微信登录路径需要，用户名路径可为空”，并统一按 `wx.id` 识别会话主体。

## Capabilities

### New Capabilities
- `wx-username-auth`: 基于 `wx` 表实现用户名密码账号生命周期（注册/登录/绑定/改密）与冲突语义。
- `history-profile-nickname`: 历史页面与画像读取接口增加 `nickname` 字段并完成前端展示。

### Modified Capabilities
- `device-wx-profile-apis`: 修订 `wx` 账号模型与绑定语义（`unionid` 可为空、用户名与微信绑定规则、会话标识统一为 `wx.id`）。
- `gateway-app-server`: 增加用户名聚合登录能力并保持 access/refresh 签发规则与现有链路一致。

## Impact

- Affected code:
  - `api/v1/device_app_user_http.go`
  - `internal/controller/device_app_user.go`
  - `internal/services/device/wx.go`
  - `internal/controller/gateway_app_ctrl.go`
  - `api/v1/gateway_app_http.go`
  - `api/v1/device_history_http.go`
  - `internal/controller/device_history.go`
  - `internal/services/history/adapter.go`
  - `resource/public/history.html`
- Affected database:
  - `ai_voice_device.wx`（使用既有新增列 `user_name`、`password`；需要唯一性与绑定冲突语义校验）
- Affected APIs:
  - 新增 `/device/app/api/user/*` 用户名相关接口
  - 新增或扩展 `/device/app/api/*` 网关聚合用户名登录接口
  - 扩展 `/device/history/api/birthday` 返回 `nickname`
- Dependencies:
  - 密码哈希库（例如 bcrypt）
  - 既有 JWT 与 refresh token 存储机制复用
