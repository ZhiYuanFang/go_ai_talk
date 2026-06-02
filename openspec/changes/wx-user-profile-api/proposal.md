## Why

当前 device 用户域已有登录、绑定、注销等写能力，以及 `GET /device/app/api/user/detail` 仅返回设备号，但客户端无法在单次请求中获取「当前账号是否绑定微信、用户名账号、设备号」等账号状态，需在个人中心等场景额外拼装或多次调用。在用户名密码体系（`wx-username-auth`）落地后，补齐只读账号 profile 接口可满足客户端展示与引导绑定的需求。

## What Changes

- 新增 device-service 接口：`GET /device/app/api/user/profile`。
- 接口无需额外入参；从请求头 `X-Internal-Wx-Id`（由 gateway-app 从 access token `sub` 注入）定位 `wx` 表当前行。
- 响应包含：
  - `isWxBound`（bool，始终返回）：表示 `unionid` 是否非空；
  - `account`（string，有值时返回，空则省略）：用户名账号；
  - `deviceNo`（string，始终返回）：已绑定设备号，未绑定时为空字符串。
- 响应 SHALL NOT 返回 `unionid`、`password` 等敏感字段。
- 保留现有 `GET /device/app/api/user/detail`，新接口为 superset，非破坏性变更。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `device-wx-profile-apis`：补充「查询当前账号 profile」能力要求，定义 `/device/app/api/user/profile` 的鉴权来源、响应字段、空值/省略语义与错误语义。

## Impact

- 受影响 API：device app 用户域新增 `GET /device/app/api/user/profile`。
- 受影响代码：`api/v1/device_app_user_http.go`、`internal/controller/device_app_user.go`、`internal/services/device/wx.go`。
- 受影响数据：只读访问 `ai_voice_device.wx` 表（按主键 id）。
- 网关：无需改动（现有 Bearer 注入与反代已覆盖 `/device/app/api/user/*`）。
- 兼容性：新增接口，非破坏性。
