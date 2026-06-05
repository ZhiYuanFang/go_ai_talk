## ADDED Requirements

### Requirement: wx 表 SHALL 持久化 Apple JWT sub

系统 MUST 在 `wx` 表新增 `apple_sub` 列，用于存储经校验的 Apple `identityToken` 内 `sub` 字段。系统 SHALL NOT 将 Apple 邮箱、姓名或 `identityToken` 原文写入业务库。`apple_sub` SHALL 建立唯一索引；允许多行 `apple_sub` 为 NULL（兼容既有微信/用户名记录）。

#### Scenario: 首次 Apple 登录创建 wx 行
- **WHEN** 校验通过的 `sub` 在 `wx` 表中不存在
- **THEN** 系统 SHALL 插入新 `wx` 行并仅设置 `apple_sub` 与 `platform`（及系统默认字段）
- **AND** 响应 `isNewUser` SHALL 为 `true`
- **AND** 系统 SHALL NOT 写入 `unionid` 或 Apple 邮箱

#### Scenario: 既有 Apple 用户再次登录
- **WHEN** 校验通过的 `sub` 已存在于某 `wx.apple_sub`
- **THEN** 系统 SHALL 返回该行的 `wxId` 与已绑定 `deviceNo`（若有）
- **AND** 响应 `isNewUser` SHALL 为 `false`

### Requirement: device-service SHALL 校验 Apple identityToken

device-service SHALL 提供 `POST /device/app/api/user/apple/login`（与网关聚合 `POST /device/app/api/apple_login` 区分），接受 JSON body 至少含 **`identityToken`**（字符串）与 **`platform`**（与客户端约定，iOS 为 `ios`）。系统 SHALL 使用 Apple JWKS（`https://appleid.apple.com/auth/keys`）验证 JWT 签名，并校验 **`iss`** 为 `https://appleid.apple.com`、**`aud`** 等于配置项 `apple.ios.bundleId`（`com.fzy.pangbao`）、**`exp`** 未过期。校验失败时 SHALL 返回明确业务错误且 SHALL NOT 创建或匹配用户行。body 中的 **`authorizationCode`** MAY 为可选字段；本能力不强制以其换票，但 SHALL 在 API 定义中保留以便将来收紧校验。

#### Scenario: 有效 identityToken 登录成功
- **WHEN** 客户端提交未过期且 `aud` 匹配的 `identityToken`
- **THEN** 系统 SHALL 提取 `sub` 并完成查/建 `wx` 行
- **AND** 响应 SHALL 包含 `wxId`、`isNewUser`、已绑定时的 `deviceNo`
- **AND** 响应 SHALL NOT 包含 `accessToken`、`refreshToken` 或 `apple_sub` 明文

#### Scenario: 无效或过期 token
- **WHEN** `identityToken` 签名校验失败、`aud` 不匹配或已过期
- **THEN** 系统 SHALL 返回业务校验失败语义
- **AND** SHALL NOT 创建或更新 `wx` 行

#### Scenario: identityToken 为空
- **WHEN** `identityToken` trim 后为空
- **THEN** 系统 SHALL 返回参数错误
- **AND** SHALL NOT 调用 Apple JWKS

### Requirement: gateway-app SHALL 聚合 Apple 登录并签发令牌

gateway-app-server SHALL 提供 **`POST /device/app/api/apple_login`**，将请求体（至少 **`identityToken`**、**`platform`**）转发至 device-service 的 **`POST /device/app/api/user/apple/login`**；当 device 返回成功且 `wxId > 0` 时，SHALL 签发 **`accessToken`**（JWT：`sub` 为 wx 主键，含 **`device_no` claim** 当 device 返回非空 `deviceNo`）与 **`refreshToken`**，响应字段 SHALL 与 `POST /device/app/api/login`、`POST /device/app/api/username_login` 对齐：`wxId`、`deviceNo`、`isNewUser`、`accessToken`、`refreshToken`。该路径 SHALL 列入 Bearer 鉴权白名单。

#### Scenario: 聚合登录成功签发 JWT
- **WHEN** 客户端调用 **`POST /device/app/api/apple_login`** 且 device 业务返回成功与有效 `wxId`
- **THEN** gateway SHALL 返回 `accessToken` 与 `refreshToken`
- **AND** 响应 `data` SHALL 含 `wxId`、`deviceNo`、`isNewUser`

#### Scenario: device 业务失败
- **WHEN** device 返回非零 `code` 或 `wxId` 无效
- **THEN** gateway SHALL 返回业务失败语义
- **AND** SHALL NOT 签发 JWT

### Requirement: Apple 与微信登录 SHALL 为独立查/建路径

Apple 登录 SHALL 以 `apple_sub` 查/建 `wx` 行；微信登录 SHALL 以 `unionid` 查/建。用户仅使用单一方式登录且未绑定第二方式时，SHALL 对应单条 `wx` 行。用户分别以两种方式**各独立登录一次且从未在同一行绑定**时，SHALL 产生两条独立 `wx` 行。

#### Scenario: 仅 Apple 登录未绑定微信
- **WHEN** 用户通过 `apple_login` 登录且当前 `wx` 行仅有 `apple_sub`
- **THEN** 系统 SHALL 维护单条 `wx` 记录
- **AND** 用户 MAY 继续使用而无需绑定微信

#### Scenario: 两次独立登录产生两行
- **WHEN** 同一自然人先后以 Apple、微信各登录一次，且每次登录时均未将第二方式绑定到当前行
- **THEN** 系统 SHALL 维护两条独立 `wx` 记录（各含对应标识）
- **AND** SHALL NOT 自动合并

### Requirement: 已登录用户 SHALL 可将第二登录方式绑定到当前 wx 行

系统 SHALL 提供 Bearer 绑定能力，向**当前**已登录 `wx` 行 UPDATE 第二标识符（`apple_sub` 或 `unionid`）。绑定 SHALL 在目标标识符未被**不同** `wxId` 占用时成功。系统 MUST NOT 合并两条已独立存在的完整 `wx` 行（各 `wxId` 已分别通过独立登录创建且含对应标识）。

#### Scenario: Apple 用户绑定微信成功
- **WHEN** 已登录 `wx` 行仅有 `apple_sub`、`unionid` 为空，且 `POST /device/app/api/user/wx/bindwx` 提交的 `jsCode` 换得 `unionid` 未被其他 `wxId` 占用
- **THEN** 系统 SHALL 将 `unionid` 写入当前 `wx` 行
- **AND** 当前会话 `wxId` SHALL 不变

#### Scenario: 微信用户绑定 Apple 成功
- **WHEN** 已登录 `wx` 行仅有 `unionid`、`apple_sub` 为空，且 `POST /device/app/api/user/apple/bind` 提交的 `identityToken` 校验通过且 `sub` 未被其他 `wxId` 占用
- **THEN** 系统 SHALL 将 `apple_sub` 写入当前 `wx` 行
- **AND** 当前会话 `wxId` SHALL 不变

#### Scenario: apple_sub 已被其他 wx 行占用
- **WHEN** 用户尝试 `apple/bind`，但校验得的 `sub` 已存在于 `wxId' != 当前 wxId` 的行
- **THEN** 系统 SHALL 返回 `ErrAppleSubTakenByOther`
- **AND** SHALL NOT 修改任何 `wx` 行

#### Scenario: unionid 已被其他 wx 行占用
- **WHEN** 用户尝试 `wx/bindwx`，但 `unionid` 已存在于 `wxId' != 当前 wxId` 的行
- **THEN** 系统 SHALL 返回 `ErrUnionIDTakenByOther`
- **AND** SHALL NOT 修改任何 `wx` 行

#### Scenario: 尝试合并两条已独立完整账号
- **WHEN** 用户曾分别以 Apple、微信各登录并各产生独立 `wx` 行，现以其中一行登录并尝试绑定另一行已占用的 `apple_sub` 或 `unionid`
- **THEN** 系统 SHALL 返回 `ErrAppleSubTakenByOther`、`ErrUnionIDTakenByOther` 或 `ErrAccountMergeConflict`
- **AND** SHALL NOT 合并、删除或迁移另一条 `wx` 行的数据

#### Scenario: 绑定端点须 Bearer 鉴权
- **WHEN** 客户端调用 `apple/bind` 或 `wx/bindwx` 且无有效 Bearer access token
- **THEN** 系统 SHALL 拒绝请求
- **AND** SHALL NOT 写入 `wx` 行

### Requirement: bind 端点契约

device-service SHALL 暴露 **`POST /device/app/api/user/apple/bind`**（body 至少含 **`identityToken`**、**`platform`**）与 **`POST /device/app/api/user/wx/bindwx`**（body 至少含 **`jsCode`**、**`platform`**）。两路径 SHALL 从请求头 **`X-Internal-Wx-Id`**（gateway 从 JWT `sub` 注入）定位当前 `wx` 行。`wx/bindwx` SHALL 泛化绑定语义，不限于用户名账号；既有 **`POST /device/app/api/user/username/bindwx`** SHALL 保留。

#### Scenario: apple/bind 请求校验
- **WHEN** 已登录用户提交有效 `identityToken` 且 `sub` 可绑定到当前行
- **THEN** 系统 SHALL 更新当前行 `apple_sub` 并返回成功
- **AND** SHALL NOT 返回新 JWT（会话 `wxId` 不变）

#### Scenario: wx/bindwx 供 Apple-only 账号使用
- **WHEN** 当前 `wx` 行仅有 `apple_sub`、无用户名密码，且 `jsCode` 有效
- **THEN** `wx/bindwx` SHALL 允许绑定微信
- **AND** SHALL NOT 要求先创建用户名密码

### Requirement: profile SHALL 可选暴露绑定状态

profile 读接口（如 `GET /device/app/api/user/detail` 或等价）SHALL 在响应中提供 **`isAppleBound`**（当且仅当 `apple_sub` 非空时为 `true`）与 **`authProviders`**（按当前行已配置身份派生的最小列表，如 `apple`、`wechat`、`username`）。SHALL NOT 在 profile 中暴露 `apple_sub` 或 `unionid` 明文。

#### Scenario: 双绑账号 profile
- **WHEN** 当前 `wx` 行同时含非空 `apple_sub` 与 `unionid`
- **THEN** 响应 SHALL 含 `isAppleBound=true`、`isWxBound=true`
- **AND** `authProviders` SHALL 同时包含 `apple` 与 `wechat`

### Requirement: 配置 SHALL 提供 iOS Bundle ID

device-service 配置 SHALL 包含 `apple.ios.bundleId`，默认值 SHALL 为 `com.fzy.pangbao`，供 `identityToken` 的 `aud` 校验使用。生产环境 SHALL 可通过配置文件或等价覆盖机制修改，且 SHALL NOT 将 Bundle ID 硬编码在多处业务逻辑中。

#### Scenario: aud 与配置一致时通过
- **WHEN** token 的 `aud` 等于当前配置的 `apple.ios.bundleId`
- **THEN** `aud` 校验 SHALL 通过

#### Scenario: aud 与配置不一致时拒绝
- **WHEN** token 的 `aud` 不等于配置的 `apple.ios.bundleId`
- **THEN** 登录 SHALL 失败并返回明确错误

### Requirement: 联调页 SHALL 支持 Apple 登录探测

`resource/public/gateway-app-integration-test.html` SHALL 提供用户可触发的操作，向当前配置的网关基址发起 **`POST /device/app/api/apple_login`**（`Content-Type: application/json`，body 含 **`identityToken`** 与 **`platform`**），并将响应中的 token 与业务字段展示在页面日志区（与现有微信/用户名登录区块并列或分区清晰）。

#### Scenario: 联调页发起 apple_login
- **WHEN** 运维在联调页填入有效 `identityToken` 并触发 Apple 登录
- **THEN** 页面 SHALL 展示 HTTP 状态与响应体中的 `accessToken`、`refreshToken`、`wxId`、`deviceNo`、`isNewUser`
