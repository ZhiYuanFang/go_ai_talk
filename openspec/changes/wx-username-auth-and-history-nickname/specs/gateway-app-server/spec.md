## MODIFIED Requirements

### Requirement: Bearer 鉴权与内部头注入
系统 SHALL 对除白名单外的受保护 HTTP 路径校验 `Authorization: Bearer <access_token>`，其中 `access_token` MUST 为合法 JWT。系统 SHALL 在校验签名与过期时间后，从 `sub` 解析 `wx.id`（允许 `0` 表示纯设备会话）并向下游注入 **`X-Internal-Wx-Id`**；当 access 含 `device_no` 声明时，系统 SHALL 同步注入 **`X-Internal-Device-No`**。

#### Scenario: 鉴权通过并注入头
- **WHEN** Bearer 为合法未过期 JWT，且 `sub` 与 `device_no` 组合满足会话规则
- **THEN** 网关 SHALL 设置 `X-Internal-Wx-Id`，并在有值时设置 `X-Internal-Device-No`

#### Scenario: 鉴权失败
- **WHEN** Bearer 缺失、签名错误、已过期或会话字段非法
- **THEN** 网关 SHALL 返回未授权错误，且 SHALL NOT 注入内部头

### Requirement: 登录与令牌仅由 gateway-app 签发
系统 SHALL 在 gateway-app-server 提供并维护两类聚合登录：
1. `POST /device/app/api/login`（微信聚合登录，转发 device 微信业务登录）
2. 用户名聚合登录接口（路径位于 `/device/app/api/` 前缀下，转发 device 用户名登录业务接口）

两类聚合登录在成功后 SHALL 统一由 gateway 签发 access/refresh；access MUST 为 JWT，`sub` MUST 等于目标 `wx.id`；refresh SHALL 为不透明随机串并绑定 Redis 会话。

#### Scenario: 用户名聚合登录成功
- **WHEN** 客户端调用用户名聚合登录且 device 返回有效 `wxId`
- **THEN** 网关 SHALL 返回 accessToken 与 refreshToken，且 access `sub` SHALL 等于该 `wxId`

#### Scenario: 微信聚合登录保持兼容
- **WHEN** 客户端调用既有微信聚合登录
- **THEN** 网关 SHALL 按现有语义返回 token 与业务字段，不因新增用户名能力破坏兼容性

### Requirement: 鉴权白名单
系统 SHALL 将无需 Bearer 的入口纳入白名单，至少包含：微信聚合登录、用户名聚合登录、refresh、公开版本检查（若启用）及 WebSocket 握手路径。

#### Scenario: 无令牌访问用户名登录
- **WHEN** 客户端无 Authorization 头调用用户名聚合登录接口
- **THEN** 请求 SHALL 进入对应处理器且 SHALL NOT 被 Bearer 中间件拦截
