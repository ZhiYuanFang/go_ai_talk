## Why

当前 device-service 微信登录实现基于微信小程序 `sns/jscode2session` 换票，与胖宝实际产品形态不符：客户端为 iOS/Android 移动应用（微信开放平台「移动应用」）及官网网页（「网站应用」扫码登录），且 Universal Links 等客户端前置能力已就绪。继续使用小程序换票 API 将导致移动应用与网页端 `code` 无法校验通过。需要在**不改动 App 已约定的登录入参**（保留 `jsCode`/`platform` 与 `POST /device/app/api/login`）的前提下，将服务端换票统一为微信开放平台 OAuth `sns/oauth2/access_token`，并按 `platform` 选择凭据。

## What Changes

- device-service 将 `POST /device/app/api/user/login` 的换票实现从 `jscode2session` 改为 `oauth2/access_token`；入参仍接受 `jsCode`（语义为微信授权临时 `code`）与 `platform`（客户端传入，用于选择配置凭据）。
- 配置由 `wechatMp.platforms` 调整为 `wechat.platforms`（或等价命名），支持至少三个 platform 键：`ios`、`android`（共用同一移动应用 `appId`/`appSecret`）与 `web`（网站应用独立 `appId`/`appSecret`）。
- 移除对微信小程序 `jscode2session` 的依赖与相关错误/注释文案；**不**新增平行登录 HTTP 路径。
- 持久化身份仍以 `wx.union_id` 为唯一键；`unionid` 为空时拒绝登录；JWT 签发与网关聚合形态不变。
- 更新联调页与 OpenAPI 字段说明：`jsCode` 注释改为移动应用/网站应用授权 `code`，非 `wx.login`。
- 补充 runbook：网页端 `qrconnect` 回调与 `platform=web` 调登录的约定（回调路径由前端实现，本变更以服务端换票与配置为主）。

## Capabilities

### New Capabilities

- `wechat-oauth-platform-config`: device-service 按 `platform` 加载微信开放平台凭据（移动应用 ios/android、网站应用 web），并约定部署与密钥注入方式。

### Modified Capabilities

- `device-wx-profile-apis`: 登录换票 API 由小程序 `jscode2session` 改为开放平台 `oauth2/access_token`；`platform` 支持 `ios`/`android`/`web`；配置键名与凭据选择语义更新。
- `gateway-app-server`: 聚合登录接口入参字段名不变，规格与文档中 `jsCode`/`platform` 语义与 device 对齐。

## Impact

- Affected code：`internal/services/device/wechat_mp.go`（或重命名）、`internal/services/device/wx.go`、`manifest/config/config.device-service.yaml`、`api/v1/device_app_user_http.go`、`api/v1/gateway_app_http.go`、`resource/public/gateway-app-integration-test.html`（文案）。
- Affected APIs：`POST /device/app/api/login`、`POST /device/app/api/user/login`（路径与 JSON 字段名不变，换票行为变更）。
- External dependencies：微信开放平台移动应用与网站应用凭据；网页端需自行实现 `qrconnect` 与授权回调（不在本变更强制新增网关路由）。
- Breaking（行为）：若环境仍按小程序 `jsCode` 联调，登录将失败；产品侧已确认不需要微信小程序登录。
