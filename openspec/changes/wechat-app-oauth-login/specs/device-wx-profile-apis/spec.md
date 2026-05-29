## MODIFIED Requirements

### Requirement: 微信登录仅返回业务字段

device-service SHALL 提供 `POST /device/app/api/user/login`（设备 wx 业务登录，与网关聚合 `POST /device/app/api/login` 区分），接受 **`jsCode`**（微信开放平台授权返回的**临时** `code`：移动应用 SDK `SendAuth` 或网站应用 `qrconnect` 回调，**禁止**持久化）与 **`platform`**（客户端传入，与 device 配置 `wechat.platforms` 下键一致，用于选择该端的 `appId`/`appSecret`；产品约定至少包括 `ios`、`android`、`web`）。

系统 SHALL **使用服务端持有的凭据**调用微信开放平台 **`sns/oauth2/access_token`** 将 `jsCode` 换为 **openid** 与 **unionid**；**SHALL NOT** 调用微信小程序 `sns/jscode2session`；**SHALL NOT** 接受客户端直传的 openid/unionid 作为身份依据。持久化身份键为 **`wx` 表 `union_id`**（微信开放平台 **unionid**，多端统一）；若微信响应中 **unionid 为空**，SHALL 返回明确业务错误且 **SHALL NOT** 创建或匹配用户行。

若 `union_id` 在库中不存在则创建 wx 行（并记录请求中的 `platform` 至 `wx.platform`）；响应 SHALL 包含至少 wx 表主键 **id（wxId）**、是否新用户（**isNewUser**）、已绑定时的 **device_no**；响应 **SHALL NOT** 包含 `access_token`/`refresh_token`（由 gateway-app 签发），且 **SHALL NOT** 向客户端回传 **unionid**、**openid** 或微信 OAuth 侧令牌。

#### Scenario: 新用户注册业务结果

- **WHEN** 首次出现的 **unionid**（经 OAuth 换票得到）调用登录接口
- **THEN** 系统 SHALL 创建 wx 行并返回 is_new_user 为真（或等价字段），且 device_no 字段 MAY 为空

#### Scenario: 已存在用户

- **WHEN** **unionid** 已存在于 `wx.union_id`
- **THEN** 系统 SHALL 返回已有 wxId，并在已绑定设备时返回 device_no

#### Scenario: 移动应用 code 换票

- **WHEN** 请求 `platform` 为 `ios` 或 `android` 且 `jsCode` 为有效的移动应用授权 `code`
- **THEN** 系统 SHALL 使用移动应用凭据完成 OAuth 换票并取得 unionid

#### Scenario: 网站应用 code 换票

- **WHEN** 请求 `platform` 为 `web` 且 `jsCode` 为有效的网站应用授权 `code`
- **THEN** 系统 SHALL 使用网站应用凭据完成 OAuth 换票并取得 unionid
