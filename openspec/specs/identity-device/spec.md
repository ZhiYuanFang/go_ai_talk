# identity-device

## Purpose

登录身份、微信/Apple、设备绑定、邀请码、wx 画像与设备域账号的行为规格。 Device and account identity contracts shared by App and admin.

<!-- migrated-from: apple-sign-in-api, baby-bind-preview-api, device-app-device-login, device-satisfaction-survey, device-sim-user, device-wx-by-device-no-list, device-wx-profile-apis, growth-invite-per-user-only, growth-trajectory-user-gates, invite-code-identity, invite-group-qr-catalog, voice-device-profile-http-contract, wechat-oauth-platform-config, wx-created-at, wx-username-auth -->

## Requirements

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

device-service SHALL 提供 `POST /device/app/api/user/apple/login`（与网关聚合 `POST /device/app/api/apple_login` 区分），接受 JSON body 至少含 **`identityToken`**（字符串）与 **`platform`**（与客户端约定，iOS 为 `ios`）。 系统 SHALL 使用 Apple JWKS（`https://appleid. 校验失败时 SHALL 返回明确业务错误且 SHALL NOT 创建或匹配用户行。 本能力不强制以其换票，但 SHALL 在 API 定义中保留以便将来收紧校验。

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

gateway-app-server SHALL 提供 **`POST /device/app/api/apple_login`**，将请求体（至少 **`identityToken`**、**`platform`**）转发至 device-service 的 **`POST /device/app/api/user/apple/login`**； 当 device 返回成功且 `wxId > 0` 时，SHALL 签发 **`accessToken`**（JWT：`sub` 为 wx 主键，含 **`device_no` claim** 当 device 返回非空 `deviceNo`）与 **`refreshToken`**，响应字段 SHALL 与 `POST /device/app/api/login`、`POST /device/app/api/username_login` 对齐：`wxId`、`deviceNo`、`isNewUser`、`accessToken`、`refreshToken`。 该路径 SHALL 列入.

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

---

### Requirement: App SHALL preview baby nickname before bind without side effects

`POST /device/app/api/user/bind_preview` MUST accept JSON body `{ "deviceNo": "<string>" }` under a valid App session (gateway-injected `X-Internal-Wx-Id`, same class as `bindwx`). On success the response `data` MUST include the trimmed `deviceNo` echo and `babyName` (string; The handler MUST NOT update `wx. Empty trimmed `deviceNo` MUST fail with a parameter/business error. An unregistered or unknown `deviceNo` MUST fail with a business error and MUST NOT return a.

#### Scenario: Valid device returns nickname

- **WHEN** 已登录用户对已注册 `deviceNo` 调用 `bind_preview`
- **THEN** 系统 MUST 返回 `code=0` 且 `data.deviceNo` 为该号、`data.babyName` 为当前画像昵称（可空）
- **AND** 调用方 `wx.device_no` MUST NOT 因本次请求被改写

#### Scenario: Empty deviceNo rejected

- **WHEN** 请求 body 的 `deviceNo` 为空或仅空白
- **THEN** 系统 MUST 返回错误且 MUST NOT 查询并冒充成功

#### Scenario: Unknown device rejected

- **WHEN** `deviceNo` 在设备域未注册或不存在
- **THEN** 系统 MUST 返回业务错误
- **AND** MUST NOT 返回成功信封冒充有效宝宝

#### Scenario: Auth required

- **WHEN** 请求缺少有效登录会话 / 无法解析当前 wx
- **THEN** 系统 MUST 拒绝（与其它需 `X-Internal-Wx-Id` 的 user API 同级）
- **AND** MUST NOT 作为匿名可访问路径

---

### Requirement: device-service 提供设备号业务登录

device-service SHALL 提供 **`POST /device/app/api/user/device_login`**，从 JSON body 读取 **`deviceNo`**（字符串，trim 后非空）。系统 SHALL 校验该设备号已在设备域注册表中注册。若存在 **`wx` 表行**其 **`device_no`** 与该值一致，响应 **`data.wxId`** SHALL 为该 wx 主键；若无绑定 wx 行，**`wxId` SHALL 为 `0`**（仍返回 **`deviceNo`**）。**`isNewUser`** 在设备号登录场景 SHALL 为 `false`。响应 SHALL NOT 包含由 gateway-app 签发的 **`accessToken`/`refreshToken`**。

#### Scenario: 已注册且已绑定 wx 的设备登录成功

- **WHEN** `deviceNo` 对应已注册设备且 wx 行已绑定该 `device_no`
- **THEN** 系统 SHALL 返回 `code=0` 且 `data` 含非零 **`wxId`**、**`deviceNo`**，且 **`isNewUser` 为 false**

#### Scenario: 已注册但未绑定 wx

- **WHEN** 设备已注册但无 wx 行绑定该 `device_no`
- **THEN** 系统 SHALL 返回 `code=0` 且 **`wxId` 为 0**、**`deviceNo`** 正确、**`isNewUser` 为 false**

#### Scenario: 设备不存在

- **WHEN** `deviceNo` 在设备注册表中不存在
- **THEN** 系统 SHALL 返回非 0 业务码及明确错误语义，且 SHALL NOT 返回 token

### Requirement: gateway-app 聚合设备号登录并签发令牌

gateway-app-server SHALL 提供 **`POST /device/app/api/device_login`**，将请求体（至少 **`deviceNo`**）转发至 device-service 的 **`POST /device/app/api/user/device_login`**；当 device 返回成功时，SHALL 签发 **`accessToken`**（JWT：`sub` 为 wx 主键，**无 wx 时 `sub` 为 `"0"`**，且 MUST 含 **`device_no` claim**）与 **`refresh_token`**（wx 会话载荷为纯数字 wxId；**`sub`=0 的会话** SHALL 在 refresh 侧携带 **`device_no`** 以便旋转 refresh 时恢复 claim）。该路径 SHALL 列入 **Bearer 鉴权白名单**。

#### Scenario: 聚合登录成功（含 wxId=0）

- **WHEN** 客户端调用 **`POST /device/app/api/device_login`** 且 body 中 `deviceNo` 在 device 侧校验通过
- **THEN** 响应 SHALL 包含 **`accessToken`/`refreshToken`** 及与 device 返回一致的 **`wxId`、`deviceNo`、`isNewUser`**

#### Scenario: device 业务失败

- **WHEN** device 返回非 0 或缺少 **`deviceNo`**
- **THEN** 网关 SHALL NOT 签发 token，且 SHALL 向客户端返回明确错误语义

### Requirement: 联调页提供设备号登录调试

`resource/public/gateway-app-integration-test.html` SHALL 提供用户可触发的操作，向当前配置的网关基址发起 **`POST /device/app/api/device_login`**（`Content-Type: application/json`，body 含 **`deviceNo`**），并将响应中的 token 与业务字段展示在页面日志区（与现有登录区块并列或分区清晰）。

#### Scenario: 用户点击设备登录

- **WHEN** 用户填写 `deviceNo` 并触发设备登录操作
- **THEN** 页面 SHALL 发起上述 HTTP 请求并 SHALL 展示成功或失败的可读结果

---

### Requirement: 当前活动可配置且仅最新活动对 App 生效

device-service MUST 只维护一份满意度问卷活动及其选项行。App 读取 MUST 只针对这一份。Hub MUST 能编辑该份的开关、说明、Asia/Shanghai 开始与结束时间。Hub MUST NOT 提供「开启新一轮」或按活动切换。就地修改选项文案 MUST NOT 清除已有提交，也 MUST NOT 使已提交的 wxId 再次变为可展示。时间 MUST 以 unix 秒存储。保存开关与起止时，若 `end_at` 早于 `start_at` 或缺少起止，MUST 拒绝该次设置保存。选项行的新增与修改 MUST NOT 因起止无效而失败。

#### Scenario: 开关关闭则不展示

- **WHEN** 这份问卷 `enabled=0`，且当前 unix 秒落在其起止之内
- **THEN** App 读取结果 MUST 为不展示

#### Scenario: 超出上海时区窗口不展示

- **WHEN** 这份问卷开关为开，且当前 unix 秒早于 `start_at` 或晚于 `end_at`
- **THEN** App 读取结果 MUST 为不展示

#### Scenario: 不能靠新一轮让已提交用户再弹

- **WHEN** 运维改选项文案或重新保存开关与有效期
- **THEN** 已有提交的 wxId MUST 仍然不展示
- **AND** 系统 MUST NOT 再创建第二份活动

### Requirement: 已登录用户读取是否展示

`GET /device/app/api/survey/current` MUST 要求登录，MUST NOT 列入 Bearer 匿名白名单。身份 MUST 使用网关注入的 wxId。需要展示时响应 MUST 包含活动 id、说明、以及启用中的选项 id 与文案。无启用选项时，只要活动开关与时间窗允许且该用户没有有效提交，MUST 仍返回需要展示。该用户存在 `valid=1` 的提交行时 MUST 返回不展示。仅有 `valid=0` 的提交行时 MUST 仍可返回需要展示。

#### Scenario: 未提交且活动有效

- **WHEN** 已登录用户请求 current，问卷为开、处于有效期、该 wxId 无提交
- **THEN** 响应 MUST 表示需要展示，并带上启用选项

#### Scenario: 有效提交不再展示

- **WHEN** 同一 wxId 已存在 `valid=1` 的提交行
- **THEN** current MUST 表示不展示

#### Scenario: 无效提交可以再展示

- **WHEN** 同一 wxId 只有 `valid=0` 的提交行，且问卷为开、处于有效期
- **THEN** current MUST 表示需要展示

### Requirement: 提交可只写意见且只记一次

`POST /device/app/api/survey/submit` MUST 要求登录。 请求体 MUST 使用 `optionResults` 字符串数组与意见，MUST NOT 根据 `optionIds` 写入选项快照。 任一结果 trim 后为空、结果之间重复、条数超过 40，或任一条超过 203 个 Unicode 码位时，系统 MUST 拒绝，且 MUST NOT 写入或改写提交行。 每一条结果 MUST 等于该活动某条启用且文案非空的选项原文，再直接连接「未使用」「不好用」或「好用」之一； 对不上时 MUST 拒绝。 多个启用选项都能对上同一条结果时，系统 MUST 仍然接受，并保存客户端提交的那条字符串。 `optionResults` 与去除空白后的意见都为空时 MUST 拒绝，且 MUST NOT 写入或改写提交行。 至少有一条合法结果或非空意见时，若该 wxId 尚无行，MUST 写入一行且 `valid=1`，并把 `optionResults` 按提交顺序写入选项快照，同时保存会话 deviceNo.

#### Scenario: 只写意见

- **WHEN** 用户提交空的 `optionResults` 和非空意见，活动当前可展示，且该用户无有效提交
- **THEN** 系统 MUST 写入或覆盖提交，选项快照为空，`valid=1`
- **AND** 随后 current MUST 对该用户不展示

#### Scenario: 空卷不关闭

- **WHEN** 用户提交空的 `optionResults` 且意见为空白
- **THEN** 系统 MUST 拒绝
- **AND** MUST NOT 留下新的提交行，也 MUST NOT 把无效行改成有效

#### Scenario: 快照保存评价后缀

- **WHEN** 启用选项原文为「夜间唤醒」，用户提交的一条结果为「夜间唤醒好用」，且活动当前可展示
- **THEN** 该提交行的选项快照 MUST 包含「夜间唤醒好用」
- **AND** MUST NOT 只保存「夜间唤醒」

#### Scenario: 快照不被事后改文案影响

- **WHEN** 用户已提交某条结果，运维随后修改该选项文案
- **THEN** 该提交行保存的文案 MUST 仍为提交当时的结果字符串

#### Scenario: 非法结果不落库

- **WHEN** 用户提交的某条结果不是任一启用选项原文加上「未使用」「不好用」或「好用」
- **THEN** 系统 MUST 拒绝
- **AND** MUST NOT 写入或改写提交行

#### Scenario: 无效后再交卷覆盖同一行

- **WHEN** 该 wxId 已有 `valid=0` 的提交，用户再次提交合法结果或非空意见，且问卷可展示
- **THEN** 系统 MUST 更新这一行的快照、意见与时间
- **AND** 该行 `valid` MUST 变为 `1`
- **AND** MUST NOT 再插入第二行

#### Scenario: 不必交齐全部选项

- **WHEN** 活动有多条启用选项，用户只提交其中一条合法结果，意见为空，且问卷可展示
- **THEN** 系统 MUST 接受并只把这一条写入快照

### Requirement: 网关放行且不计入使用统计

gateway-app MUST 将 `/device/app/api/survey/*` 反代到 device-service。`GET /device/app/api/survey/current` 与 `POST /device/app/api/survey/submit` MUST 列入 usage 维护排除（精确方法与路径），不得计入功能使用统计。Hub 管理接口 MUST 使用 `/device/admin/api/survey/` 前缀。

#### Scenario: App 路径到达 device-service

- **WHEN** 已登录客户端请求上述 GET 或 POST
- **THEN** gateway-app MUST 按 device 反代规则转发，而不是当本机路由处理

#### Scenario: 成功调用不进使用统计

- **WHEN** 这两条 App 接口返回 2xx
- **THEN** usage 记录 MUST 跳过它们

### Requirement: Hub 可查看提交明细

Hub 问卷调查页 MUST 近全宽，宽屏下编辑区与结果区左右分栏。结果 MUST 展示用户 id、宝宝昵称、设备号、选项快照、意见与提交时间，并且 MUST 列出这一份问卷的全部提交，不得按活动分表或筛选。

#### Scenario: 查看某次提交

- **WHEN** 运维打开问卷调查页且已有提交
- **THEN** 页面 MUST 列出该条的 wxId、宝宝昵称、设备号、所选文案与意见

### Requirement: Hub 可修改提交是否有效

Hub MUST 能把某一条提交标为有效或无效，并且该操作 MUST 立即写入 device 库。标为无效 MUST NOT 删除该行。刷新页面后 MUST 仍显示这次设置的有效状态。

#### Scenario: 标成无效后仍留在列表

- **WHEN** 运维把一条有效提交改为无效并刷新页面
- **THEN** 该行 MUST 仍在，且显示为无效

### Requirement: 选项行可单独编辑并立即持久化

Hub MUST 能新增一行选项并立即写入 device 库，随后 MUST 能修改该行文案、是否启用，以及删除该行，且每次操作立即持久化。刷新页面后这些行 MUST 仍在。空文案行 MAY 存在，但 App current MUST NOT 把它当作可选项返回。删除选项 MUST NOT 改写已有提交的文案快照。

#### Scenario: 新增后刷新仍在

- **WHEN** 运维新增一行并写入文案，然后重新打开问卷调查页
- **THEN** 该行 MUST 仍显示刚写入的文案

#### Scenario: 空文案不对 App 展示

- **WHEN** 存在文案为空白的选项行，且问卷对某用户需要展示
- **THEN** current 的选项列表 MUST NOT 包含该行

---

### Requirement: wx table SHALL store simulated user flag

`device-service` 权威库表 `wx` MUST 新增列 `is_simulated`（TINYINT NOT NULL DEFAULT 0）。`1` 表示模拟用户；公开 App 注册路径 MUST NOT 允许客户端自行设置为 1。

#### Scenario: Default not simulated

- **WHEN** 用户经公开 `username/register` 注册
- **THEN** 新行 `is_simulated` MUST 为 0

### Requirement: device internal sim register SHALL wrap username register with flag

`POST /device/internal/api/sim/username/register` MUST 要求有效 `X-Device-Gateway-Internal-Secret`（或兼容 `X-Gateway-Internal-Secret`）。请求体 MUST 含 `account`、`password`。服务 MUST 调用 `WxUsernameRegister` 并在同一注册流程内将 `is_simulated` 设为 1。成功响应 MUST 含 `wxId`、`account`。账号已存在 MUST 返回业务错误。

调用方（sim-user-service T1）MUST 传入随机生成的 account/password，MUST NOT 依赖固定 `ptest{N}` 或固定默认密码作为长期约定。

#### Scenario: Internal sim register with random account

- **WHEN** sim-service 携带有效密钥 POST `{account:"s8k2m9xq4n",password:"Kp9#mX2vLq4n"}`
- **THEN** 响应 MUST 含 `wxId>0` 且该 wx 行 `is_simulated=1`

#### Scenario: Reject without secret

- **WHEN** 无内部密钥调用 sim register
- **THEN** HTTP MUST 为 403

### Requirement: device internal SHALL list and batch-query simulated users

device-service MUST 提供：

- `GET /device/internal/api/sim/wx/list` — 分页返回 `is_simulated=1` 的 wxId 与 account 列表（供 Admin、计数 total 等列举场景）
- `GET /device/internal/api/sim/wx/random` — 返回单条随机 simulated 用户（供 sim 任务随机选取）
- `GET /device/internal/api/sim/wx/ids` — 返回**全量** simulated wxId（供 T5 未读抽样等 MUST 覆盖全库 sim 的场景）
- sim 批量查询 MUST 在现有 `POST /device/internal/api/ucg/wx/batch` 响应项中增加 `isSimulated` 字段，或提供等价的 sim 专用 batch 接口

#### Scenario: List sim users

- **WHEN** sim-service 请求 sim wx list 分页
- **THEN** 返回列表 MUST 仅含 `is_simulated=1` 的 wxId

#### Scenario: Batch includes flag

- **WHEN** UCG 或 gateway 批量查询 wxId 展示字段
- **THEN** 每项 MUST 含 `isSimulated` 布尔值

#### Scenario: List not required for random pick

- **WHEN** sim-service 需随机选取一个模拟用户执行任务
- **THEN** MUST 使用 random 接口，MUST NOT 依赖 list 分页结果在客户端 `rand` 选 wxId

#### Scenario: T5 must not use list first page as full sim set

- **WHEN** sim-service T5 需要全库 sim wxId 集合
- **THEN** MUST 使用 ids 接口，MUST NOT 使用 `sim/wx/list?page=1&pageSize=200` 代替全集

### Requirement: device internal SHALL provide random single simulated wx pick

device-service MUST 提供 `GET /device/internal/api/sim/wx/random`，要求有效内部密钥（与 sim wx list 相同）。服务 MUST 在 `is_simulated=1` 集合上通过有界 ID 探测返回 **0 或 1** 条 `{wxId, account}`，MUST NOT 使用 `ORDER BY RAND()`。探测 MUST 覆盖全库 simulated 用户（非仅第一页）。锚点 MUST 在 `[minId, maxId]` 上 **均匀** 生成：`R = minId + floor((maxId - minId) * U)`（`U` 为 `(0,1)` 均匀随机），随后 `WHERE is_simulated=1 AND id >= R ORDER BY id ASC LIMIT 1`；锚点落空且 eligible 存在时 MUST 回退 `minId` 一条。MUST NOT 对 high-id / 新注册用户做幂次偏置。

#### Scenario: Random returns one sim user

- **WHEN** sim-service 携带有效密钥 GET random 且存在至少一条 `is_simulated=1`
- **THEN** 响应 MUST 含 `wxId>0` 与非空 `account`

#### Scenario: Random empty when no sim users

- **WHEN** 无 `is_simulated=1` 行
- **THEN** 响应 MUST 表示无结果（空或 found=false），且 MUST NOT 500

#### Scenario: Bounded SQL only

- **WHEN** 代码评审 random 实现
- **THEN** MUST 为 MIN/MAX 聚合 + `LIMIT 1` 探测，MUST NOT 全表加载或 `ORDER BY RAND()`

### Requirement: device internal SHALL provide full simulated wx id list

device-service MUST 提供 `GET /device/internal/api/sim/wx/ids`，要求有效内部密钥（与 sim wx list 相同）。服务 MUST 在单条有界 SQL 内返回 **全部** `is_simulated=1` 的 wxId 列表，响应 `{ ids: int64[], total: int }`。MUST NOT 分页截断（不得仅返回前 200 条）。当 `total` 超过 **10000** 时 MUST 返回 4xx 且 MUST NOT 返回部分 ids。

#### Scenario: Ids returns all simulated users

- **WHEN** 库中存在 350 条 `is_simulated=1` 且请求携带有效密钥
- **THEN** 响应 `total` MUST 为 350 且 `ids` 长度 MUST 为 350

#### Scenario: Ids empty when no sim users

- **WHEN** 无 `is_simulated=1` 行
- **THEN** 响应 MUST 为 `{ ids: [], total: 0 }` 且 MUST NOT 500

#### Scenario: Over limit rejected

- **WHEN** `is_simulated=1` 计数超过 10000
- **THEN** MUST 返回错误且 MUST NOT 返回 ids 数组

### Requirement: device internal SHALL deactivate simulated wx by wxId

`device-service` MUST 提供 `POST /device/internal/api/sim/wx/{wxId}/deactivate`，要求有效 `X-Device-Gateway-Internal-Secret`（与 sim register 相同）。

服务 MUST 校验目标 wx 行存在且 `is_simulated=1`；否则 MUST 返回 4xx 业务错误且 MUST NOT 删除。

校验通过后 MUST 调用 `WxDeactivateByID` 删除 wx 单行，并 MUST 失效 wx 相关 cachekit 键；MUST 从 `usage:sim_wx_ids` SET 移除该 wxId member（经 `cachekit.GatewayUsageSimWxSetKey` / `GatewayUsageSimWxMember`）。

MUST NOT 删除 ucg 或 user 域数据。

#### Scenario: Deactivate sim wx success

- **WHEN** sim-service 携带有效密钥 POST deactivate 且 wxId 为 `is_simulated=1`
- **THEN** wx 行 MUST 删除且 HTTP MUST 成功

#### Scenario: Reject real user deactivate

- **WHEN** wxId 对应 `is_simulated=0`
- **THEN** MUST 返回 4xx 且 wx 行 MUST 保持不变

#### Scenario: Reject without secret

- **WHEN** 无内部密钥调用 sim deactivate
- **THEN** HTTP MUST 为 403

---

### Requirement: device MUST 能按 device_no 列出全部绑定 wxId

device-service MUST 提供内部（或服务间）契约：输入非空 `deviceNo`，返回 `wx` 表中 `device_no` 等于该值的全部用户主键 `wxId` 列表（顺序不限）。MUST NOT 仅返回 `LIMIT 1` 单条作为「全部绑定用户」的语义。实现 MUST 对单次返回数量设上限（配置或常量）；超过上限时 MUST 截断并记录告警日志，或提供分页（一期可选截断+告警）。voice 等调用方 MUST 经 `clients/device` 访问，MUST NOT 直查 device 库。

#### Scenario: 同设备多账号全部返回

- **WHEN** 两个不同 `wx.id` 的行均绑定同一 `device_no`
- **THEN** 列表接口 MUST 返回包含这两个 `wxId` 的结果集

#### Scenario: 无绑定时返回空列表

- **WHEN** 没有任何 `wx` 行绑定该 `deviceNo`
- **THEN** 接口 MUST 成功返回空列表且 MUST NOT 报错为系统故障

---

### Requirement: 微信登录仅返回业务字段

device-service SHALL 提供 `POST /device/app/api/user/login`（设备 wx 业务登录，与网关聚合 `POST /device/app/api/login` 区分），接受 **`jsCode`**（微信开放平台授权临时 `code`）与 **`platform`**（与 device 配置平台键一致）。 系统 SHALL 使用服务端持有的微信凭据换取 `unionid` 并按 `unionid` 查找或创建 `wx` 行。 对于微信登录路径，若微信响应中 `unionid` 为空，系统 SHALL 返回明确业务错误且 SHALL NOT 创建或匹配用户行。 系统同时 SHALL 支持同表中的用户名账号记录（其 `unionid` MAY 为空），且 SHALL NOT 因存在 `unionid` 为空的用户名记录影响微信登录判定。 响应 SHALL 包含至少 `wxId`、`isNewUser`、已绑定时的 `deviceNo`； 响应 SHALL NOT 包含 gateway 签发令牌，也.

#### Scenario: 新微信用户登录成功
- **WHEN** 首次出现的 `unionid` 调用登录接口
- **THEN** 系统 SHALL 创建 wx 行并返回 `isNewUser=true`

#### Scenario: 既有微信用户登录成功
- **WHEN** `unionid` 已存在于 `wx` 表
- **THEN** 系统 SHALL 返回已有 `wxId` 与已绑定 `deviceNo`（若有）

### Requirement: 绑定设备与 wx

device-service SHALL 提供 `POST /device/app/api/user/bindwx`，从请求头读取 **`X-Internal-Wx-Id`**（值为 `wx.id`，由 gateway 从 access token `sub` 注入），从 JSON body 读取 `deviceNo`，并将设备号绑定到对应 `wx` 行。

#### Scenario: 绑定成功
- **WHEN** 头部包含有效 `X-Internal-Wx-Id` 且 `deviceNo` 合法并已注册
- **THEN** 系统 SHALL 持久化绑定关系并返回成功语义

#### Scenario: 头部无效
- **WHEN** 缺失或提供非法 `X-Internal-Wx-Id`
- **THEN** 系统 SHALL 返回参数错误，且 SHALL NOT 更新绑定关系

### Requirement: 自动保存画像

device-service SHALL 提供 `POST /device/app/api/user/auto_save`，从请求头读取 **`X-Internal-Wx-Id`**，从 body 读取 `birthday` 与 `sex`，并 SHALL 返回 `device_no`。当目标 `wx` 行尚未绑定设备时，系统 SHALL 生成全局唯一、6 位大写字母 `device_no`，完成设备注册与绑定后写入画像；当已绑定时，系统 SHALL 仅更新画像并返回原 `device_no`。

#### Scenario: 无设备号时创建并绑定
- **WHEN** `wxId` 有效且当前 wx 行未绑定 `device_no`
- **THEN** 系统 SHALL 生成并绑定唯一 `device_no`，保存画像后返回该值

#### Scenario: 已绑定设备仅更新画像
- **WHEN** `wxId` 有效且 wx 已绑定 `device_no`
- **THEN** 系统 SHALL 仅更新画像并返回原 `device_no`

#### Scenario: 候选设备号冲突重试
- **WHEN** 随机候选 `device_no` 与现有数据冲突
- **THEN** 系统 SHALL 重试生成直到成功或达到最大重试上限

### Requirement: 按 unionid 查询设备号

device-service SHALL 提供 `GET /device/app/api/user/detail`，并以 **`X-Internal-Wx-Id`** 识别当前账号主体，返回该主体绑定的 `device_no`（未绑定时返回约定空值或错误语义）。

#### Scenario: 已绑定返回设备号
- **WHEN** `X-Internal-Wx-Id` 对应记录已绑定 `device_no`
- **THEN** 响应 SHALL 包含该 `device_no`

#### Scenario: 未绑定返回空语义
- **WHEN** `X-Internal-Wx-Id` 对应记录未绑定设备
- **THEN** 响应 SHALL 返回空 `device_no` 或约定未绑定语义

### Requirement: 按主键 id 解析 unionid（内部）

device-service SHALL 提供仅供内网或网关调用的只读接口（例如 `GET /device/app/api/user/internal/by-id`），根据 wx 表主键 id 返回对应 **`union_id`（响应字段 unionId）**，以便 gateway-app 在仅持有 access 内 id 时解析 **unionid** 并写入 **`X-Internal-Wx-Union-Id`**；该接口 SHALL 不对外网匿名开放（依赖部署网络或额外共享密钥策略）。

#### Scenario: 有效 id

- **WHEN** 网关使用有效 id 调用内部解析接口
- **THEN** 响应 SHALL 包含与该 id 对应的 unionId

#### Scenario: 无效 id

- **WHEN** id 不存在或非法
- **THEN** 系统 SHALL 返回明确错误且 SHALL NOT 泄露其他行信息

### Requirement: Redis 缓存与失效

device-service 对高频读路径（含 `wxId -> unionid`、`wxId -> deviceNo`）SHALL 可选使用 Redis 缓存；在绑定设备、注销、或任何影响映射关系的写操作成功后，系统 SHALL 失效相关缓存键，确保后续读取一致。

#### Scenario: 写后缓存一致性
- **WHEN** bindwx、auto_save 或 deactivate 成功完成
- **THEN** 与该 `wxId` 相关缓存 SHALL 被删除或失效

### Requirement: 设备画像读接口 SHALL 返回宝宝名字

系统在读取设备画像时 MUST 同时返回 `babyName`、`birthday`、`sex` 三个字段；其中 `babyName` 为可选字符串，未设置时返回空串。
该要求适用于 device 画像接口与历史页面画像接口的统一读取语义。

#### Scenario: 读取画像返回完整字段
- **WHEN** 调用方使用有效 `deviceNo` 请求画像读取接口
- **THEN** 响应 SHALL 包含 `babyName`、`birthday`、`sex`
- **AND** 当数据库中 `baby_name` 为空时，`babyName` SHALL 返回空串

### Requirement: 设备画像写接口 SHALL 支持宝宝名字更新

系统在保存设备画像时 MUST 接受 `babyName` 字段，并与 `birthday`、`sex` 一并持久化到 `user` 表画像字段集合。
该要求适用于 `/device/app/api/user/save`、`/device/app/api/user/auto_save` 以及历史页面画像保存链路。

#### Scenario: 仅修改宝宝名字
- **WHEN** 调用方提交合法 `deviceNo` 与 `babyName`，且未变更生日/性别
- **THEN** 系统 SHALL 更新 `user.baby_name`
- **AND** 系统 SHALL 保持 `birthday`、`sex` 原值不变

#### Scenario: 同时修改名字与性别生日
- **WHEN** 调用方提交 `babyName`、`birthday`、`sex`
- **THEN** 系统 SHALL 在一次保存语义内持久化三项画像字段

### Requirement: 账号注销删除 wx 记录

系统 MUST 提供 `POST /device/app/api/user/deactivate`。接口 SHALL 从请求头读取 `X-Internal-Wx-Id`，并按该主键删除 `wx` 表中的对应单条记录。删除成功后，系统 SHALL 使该 `wxId` 相关缓存映射失效，避免后续读取命中陈旧数据。

#### Scenario: 注销成功删除单条记录
- **WHEN** 请求头包含有效的 `X-Internal-Wx-Id` 且该 `wx` 记录存在
- **THEN** 系统 SHALL 删除该主键对应的一条 `wx` 记录并返回成功语义

#### Scenario: 请求头缺失或无效
- **WHEN** `X-Internal-Wx-Id` 缺失、非整数或小于等于 0
- **THEN** 系统 SHALL 返回参数错误，且 SHALL NOT 执行删除

#### Scenario: 目标记录不存在
- **WHEN** `X-Internal-Wx-Id` 合法但对应 `wx` 记录不存在
- **THEN** 系统 SHALL 返回明确业务错误语义（已注销或不存在），且 SHALL NOT 影响其他记录

### Requirement: 查询当前账号 profile

device-service SHALL 提供 `GET /device/app/api/user/profile`。接口 SHALL NOT 要求额外 query 或 body 入参；SHALL 从请求头读取 **`X-Internal-Wx-Id`**（值为 `wx.id`，由 gateway 从 access token `sub` 注入）定位当前 `wx` 行，并返回账号状态字段。

响应 SHALL 包含：
- **`isWxBound`**（bool，始终返回）：当且仅当该行 `unionid` 非空时为 `true`；
- **`account`**（string）：该行用户名账号；当账号为空时，响应 SHALL 省略该字段（JSON `omitempty`）；
- **`deviceNo`**（string，始终返回）：该行已绑定设备号；未绑定时 SHALL 返回空字符串。

响应 SHALL NOT 包含 `unionid`、`password`、`openid` 或微信令牌明文。

#### Scenario: 纯微信用户已绑设备
- **WHEN** `X-Internal-Wx-Id` 有效且对应 `wx` 行 `unionid` 非空、`account` 为空、已绑定 `device_no`
- **THEN** 响应 SHALL 包含 `isWxBound=true`、`deviceNo` 为已绑定值，且 SHALL NOT 包含 `account` 字段

#### Scenario: 纯用户名用户未绑微信
- **WHEN** `X-Internal-Wx-Id` 有效且对应 `wx` 行 `account` 非空、`unionid` 为空
- **THEN** 响应 SHALL 包含 `isWxBound=false`、`account` 为对应用户名，以及 `deviceNo`（未绑设备时为 `""`）

#### Scenario: 用户名与微信均已绑定
- **WHEN** `X-Internal-Wx-Id` 有效且对应 `wx` 行 `unionid` 与 `account` 均非空
- **THEN** 响应 SHALL 包含 `isWxBound=true`、`account` 与 `deviceNo`

#### Scenario: 请求头缺失或无效
- **WHEN** `X-Internal-Wx-Id` 缺失、非整数或小于等于 0
- **THEN** 系统 SHALL 返回参数错误，且 SHALL NOT 返回 profile 数据

#### Scenario: wx 记录不存在
- **WHEN** `X-Internal-Wx-Id` 合法但对应 `wx` 记录不存在
- **THEN** 系统 SHALL 返回明确错误语义（如 404），且 SHALL NOT 泄露其他行信息

---

### Requirement: 成长轨迹邀请按人一次且允许同机多码

对 `growth_trajectory_predict`，系统在邀请码兑换时 MUST NOT 因同一 `device_no` 已对本功能邀请开通而拒绝（MUST NOT 再适用 `InviteOncePerDevice` / MUST NOT 读写本功能的 `feature_invite_device_grant` 作为拒绝依据）。系统 MUST 继续：若同一 `redeemer_wx_id` 已对任意邀请码成功开通过本功能，则拒绝再次邀请开通；MUST 保留人×码×功能去重与同宝宝禁兑；MUST NOT 再要求自用禁兑（`owner_wx_id == redeemer_wx_id`）。`care_alert_smart_remind` 的设备邀请一次约束 MUST 保持不变。

#### Scenario: 同宝宝第二人兑不同码

- **WHEN** 家长 A 已用码 X 邀请开通成长轨迹，家长 B（不同 wx、同一 `device_no`）使用另一好友码 Y 兑换成长轨迹且 B 从未邀请开通过本功能，且码主人 Y 当前绑定设备与 B 的 `device_no` 不同
- **THEN** 系统 MUST 允许兑换并写入 B 的 user 权益

#### Scenario: 同一人第二次邀请拒绝

- **WHEN** 同一 `redeemer_wx_id` 已对成长轨迹邀请开通成功后再次用任意码兑换本功能
- **THEN** 系统 MUST 拒绝

#### Scenario: 值得留意仍同机一次

- **WHEN** 某 `device_no` 已邀请开通 `care_alert_smart_remind` 后，同机另一账号用另一码再兑该功能
- **THEN** 系统 MUST 拒绝

---

### Requirement: 成长轨迹智能分析 MUST 校验账号开通或 VIP

`POST /device/api/growth-trajectory/turn`（及 voice 内等价入口）在开流前 MUST 要求有效登录 `wxId`，并 MUST 经 cash 合成：该 `wxId` 对 `growth_trajectory_predict` 存在未过期 user 权益 **或** 该账号为 VIP。未满足 MUST 拒绝且 MUST NOT 调用 Python turn。合成 MUST NOT 仅因同 `device_no` 上他人已开通而放行。

#### Scenario: 未开通非 VIP 拒绝 turn

- **WHEN** 已登录用户无成长轨迹 user 权益且非 VIP 请求 turn
- **THEN** 系统 MUST 返回未开通类错误且不开 SSE 业务流

#### Scenario: VIP 免开通可 turn

- **WHEN** 用户无 user 权益但是 VIP 请求 turn
- **THEN** 系统 MUST 允许进入分析流（仍受日限等其它闸门约束）

### Requirement: 成长轨迹最新结果 MUST 按设备读取且免开通

`GET /device/api/growth-trajectory/latest` MUST 要求登录，MUST 按请求 `deviceNo` 返回该设备最新一条结果（无则空字段），MUST NOT 校验功能开通或 VIP。存储仍为按设备最新一条，MUST NOT 本变更引入多版本历史列表。

#### Scenario: 未开通用户可读最新报告

- **WHEN** 已登录但未开通成长轨迹的用户请求 latest 且该 deviceNo 已有结果
- **THEN** 响应 MUST 返回已有 `resultMarkdown`（或等价字段）且 MUST NOT 因未开通失败

#### Scenario: 未登录拒绝 latest

- **WHEN** 缺少有效登录 wxId 请求 latest
- **THEN** 系统 MUST 拒绝

### Requirement: 成长轨迹日限 MUST 按用户计数

成长轨迹每日成功结果次数上限 MUST 以登录 `wxId` + 上海日历日为键计数（经 cachekit 键构建器），MUST NOT 再以 `deviceNo` 作为日限主键。达上限时 turn MUST 在开流前拒绝。

#### Scenario: 同人换设备共享日限

- **WHEN** 同一 wxId 在设备 D1 已用尽当日次数后在 D2 再请求 turn
- **THEN** 系统 MUST 因日限拒绝

### Requirement: 成长轨迹邀请 MUST 同时人一次与设备一次

对 `growth_trajectory_predict` 邀请开通，系统 MUST 在既有人×码×功能去重之外：若该 `redeemer_wx_id` 已对任意邀请码成功开通过本功能，MUST 拒绝再次邀请开通；若该 `device_no` 已对本功能邀请开通过，MUST 拒绝。设备一次闸 MUST 继续写入/校验 `feature_invite_device_grant`（或等价）。值得留意等其它 InviteOncePerDevice 功能 MUST NOT 因本变更被迫加上「人一次」除非另行配置。

#### Scenario: 同一人第二次兑码成长轨迹

- **WHEN** 用户已用码1开通成长轨迹后用码2再兑同一功能
- **THEN** 系统 MUST 拒绝且 MUST NOT 延长 user 权益

#### Scenario: 同设备第二人兑码成长轨迹

- **WHEN** 设备 D 已有任意账号邀请开通过成长轨迹后另一账号再兑
- **THEN** 系统 MUST 因设备已开通过而拒绝

### Requirement: cash 成长轨迹 access MUST 以账号权益为准

内部 `GET /cash/internal/api/growth-trajectory/access` MUST 以 `wxId` 的 user 权益 ∨ VIP 合成 `allowed`/`featureActive`，MUST NOT 将设备维 `feature_entitlement` 上的成长轨迹行视为有效开通（本变更后）。

#### Scenario: 仅有历史 device 行不算开通

- **WHEN** 某 device 上仍有旧成长轨迹 device entitlement，但当前 wx 无 user 权益且非 VIP
- **THEN** access MUST 返回不允许

---

### Requirement: 用户邀请码由 cash Ensure 持有

系统 MUST 在 cash 库为每个有效 `owner_wx_id` 最多保留一条邀请码（一用户一码）。系统 MUST 在用户注册完成后触发 Ensure，或在 App 首次读取「我的邀请码」时懒创建。邀请码 MUST NOT 写入 device `wx` 表。邀请码 MUST NOT 要求有效期、码级可开功能范围或 grant 时长才能用于兑换。

#### Scenario: 首次读取生成

- **WHEN** 某 wx 尚无邀请码且请求我的邀请码
- **THEN** 系统创建唯一码并返回，且 `redeemedCount` 为 0

#### Scenario: 重复 Ensure 幂等

- **WHEN** 同一 wx 再次 Ensure
- **THEN** 系统返回同一码且 MUST NOT 插入第二行

### Requirement: 兑码去重为人×码×功能

系统兑换邀请码时 MUST **NOT** 仅因码主人 `owner_wx_id` 与兑换者 `redeemer_wx_id` 相同而拒绝（已取消自用禁兑）。 系统 MUST 拒绝码主人**当前**绑定 `device_no` 与兑换者请求 `device_no` 相同且均非空的兑换（同一宝宝下账号互兑，含兑换者兑自己的码且双方同机）。 系统 MUST 以 `(redeemer_wx_id, code, feature_id)` 唯一约束防止重复兑换。 系统 MUST 允许同一兑换者使用不同 owner 的码（且兑换者与各 owner 当前绑定设备号不同时）。 系统 MUST NOT 再执行「一家锁定」（redeemer 绑定单一 owner）。 系统 MUST 仅当目标 `feature_def` 启用且 `unlock_methods` 含 `invite_code` 时允许兑换。 码主人设备号 MUST 经 device-service 契约按 `owner_wx_id` 查询，MUST NOT 由 cash 直查 device 库。.

#### Scenario: 多好友码累加预测

- **WHEN** 用户依次兑换好友 A、B 的码且 featureId 为 `prediction_unlock`，且兑换者与 A、B 当前绑定设备号均不同
- **THEN** 两次均成功且该设备预测永久条数各 +1

#### Scenario: 不同设备互兑

- **WHEN** A 兑 B 的码开通某功能且 B 兑 A 的码开通同一功能，且双方当前绑定 `device_no` 不同
- **THEN** 两次均成功

#### Scenario: 同设备拒绝兑码

- **WHEN** 兑换者当前 `device_no` 与码主人当前绑定 `device_no` 相同且均非空
- **THEN** 系统 MUST 拒绝兑换，MUST NOT 写入开通或获客原力

#### Scenario: 同人异机可兑自己的码

- **WHEN** 兑换者 `wx_id` 等于码的 `owner_wx_id`，且主人当前绑定 `device_no` 与兑换者请求 `device_no` 不同（或主人未绑机为空），且其它校验通过
- **THEN** 系统 MUST NOT 因「自用」拒绝，MUST 按同宝宝及其余规则继续

#### Scenario: 主人未绑机可兑

- **WHEN** 码主人经 device 查询得到空 `device_no`，且其它校验通过
- **THEN** 系统 MUST NOT 因同设备规则拒绝

#### Scenario: device 查询失败 fail-closed

- **WHEN** cash 调用 device 按主人 wxId 取 `device_no` 失败
- **THEN** 系统 MUST 拒绝本次兑换

#### Scenario: 同码同功能重复

- **WHEN** 同一用户对同一码同一 featureId 再次兑换
- **THEN** 系统拒绝

### Requirement: App 可读我的码与邀请列表

系统 MUST 提供 App 接口返回当前用户邀请码与获客数量。系统 MUST 提供 App 接口返回成功使用该码的用户列表（含 UCG 展示昵称与兑换时间）。系统 MUST NOT 再提供面向运营的邀请码造码 Admin 页面。

#### Scenario: 邀请列表

- **WHEN** owner 查询 invitees
- **THEN** 每条含昵称与 `redeemedAt`，且仅含成功兑换记录

---

### Requirement: Catalog returns inviteGroupQrUrl only while valid

When synthesizing `GET /cash/app/api/feature/catalog`, the system MUST include top-level `inviteGroupQrUrl` only when `expires_at > 0` and `expires_at > now`. The URL MUST point at the apk download path for `er_code.png` and SHOULD include a cache-busting query derived from `updated_at`. When invalid, the field MUST be omitted or empty.

#### Scenario: Valid window

- **WHEN** expiry is in the future and metadata exists
- **THEN** catalog data contains non-empty `inviteGroupQrUrl` referencing `er_code.png`

#### Scenario: Expired or zero

- **WHEN** `expires_at` is 0 or not after now
- **THEN** catalog MUST NOT present a usable invite group QR URL to the App

---

### Requirement: Voice MUST 通过 Device 服务契约获取设备画像数据

`voice-service` 在涉及设备信息、用户生日、性别、注册状态等 device/profile 领域数据时 MUST 通过 `device-service` 暴露的内部 HTTP 接口获取，MUST NOT 直接访问 `user/device` 领域数据库表。

#### Scenario: 通用问答需要设备画像
- **WHEN** voice 在生成通用问答提示词时需要生日或性别等画像信息
- **THEN** voice MUST 调用 device 内部接口获取画像数据，并将结果用于提示词构建

#### Scenario: 设备信息接口不可达
- **WHEN** voice 调用 device 内部画像接口超时或网络失败
- **THEN** voice MUST 返回可观测的错误语义，并按配置决定是否执行迁移期兜底

### Requirement: Device 内部画像接口 MUST 提供一致错误结构

device 内部画像接口 MUST 对参数错误、设备不存在、服务异常返回统一错误结构，供 voice 侧做稳定错误映射。

#### Scenario: 设备不存在
- **WHEN** voice 传入的 `deviceNo` 在 device 服务中不存在
- **THEN** device MUST 返回可区分的业务错误码，voice MUST 返回可理解的失败信息

#### Scenario: 参数缺失
- **WHEN** voice 调用画像接口时缺失关键参数
- **THEN** device MUST 返回参数错误结构，voice MUST 记录请求参数异常日志

### Requirement: 画像链路的「本地」实现 MUST 不依赖进程内 user DAO

即使存在 `localDeviceProfileAdapter` 类实现，`voice-service` 在生产单库模式下 MUST 使用 **HTTP 远程实现**（或指向 device 基址的 HTTP local）；MUST NOT 依赖在 voice 进程内对 `dao.User` 的查询作为获取画像的主路径。

#### Scenario: 生产 voice 配置

- **WHEN** 部署 `voice-service` 且 `database.default` 仅含 voice 域表
- **THEN** 设备画像 MUST 通过 `device-service` 的 HTTP 接口获取；若误配为进程内 local 适配器，MUST 视为配置错误并暴露启动或运行期检查（若实现）

---

### Requirement: 按 platform 加载微信开放平台凭据

device-service SHALL 从配置 `wechat.platforms` 读取各 `platform` 键对应的 `appId` 与 `appSecret`。系统 SHALL 至少支持以下键名：`ios`、`android`、`web`。当请求中的 `platform` 在配置中不存在或 `appId`/`appSecret` 任一为空时，SHALL 返回明确配置错误且 SHALL NOT 调用微信 API。

`ios` 与 `android` SHALL 映射到**同一微信开放平台移动应用**的 `appId`/`appSecret`（部署时两键配置相同值）。`web` SHALL 映射到**微信开放平台网站应用**的独立 `appId`/`appSecret`。

生产环境 SHALL 通过环境变量或挂载配置覆盖 `appSecret`，且 SHALL NOT 将真实密钥提交到版本库。

#### Scenario: 移动应用 platform 解析凭据

- **WHEN** 登录请求 `platform` 为 `ios` 或 `android` 且对应配置项已填写有效 `appId`/`appSecret`
- **THEN** 系统 SHALL 使用该移动应用凭据调用微信 OAuth 换票 API

#### Scenario: 网站应用 platform 解析凭据

- **WHEN** 登录请求 `platform` 为 `web` 且 `wechat.platforms.web` 已填写有效 `appId`/`appSecret`
- **THEN** 系统 SHALL 使用该网站应用凭据调用微信 OAuth 换票 API

#### Scenario: 未配置的 platform

- **WHEN** 登录请求 `platform` 在 `wechat.platforms` 中不存在或凭据不完整
- **THEN** 系统 SHALL 返回明确错误且 SHALL NOT 创建或匹配 wx 行

---

### Requirement: wx table SHALL store account creation timestamp

`device-service` 权威库表 `wx` MUST 新增列 `created_at`（`BIGINT NOT NULL DEFAULT 0`，Unix 秒）。该字段 MUST 表示 wx 账号行**首次创建**时刻。历史行在迁移后 `created_at` MAY 为 0；展示层对 0 MUST 显示「—」。

#### Scenario: New wx row gets created_at on insert

- **WHEN** 用户经微信 OAuth、Apple 登录或用户名注册**首次**创建 wx 行
- **THEN** 插入时 `created_at` MUST 为当前 Unix 秒且大于 0

#### Scenario: Legacy row defaults to zero

- **WHEN** 迁移脚本执行后存在迁移前已创建的 wx 行且未回填
- **THEN** 该行 `created_at` MUST 为 0

### Requirement: wx creation write paths SHALL set created_at

下列 device-service 写路径在 **Insert** 新 wx 行时 MUST 写入 `created_at`：

- `WxLogin`（unionid 首次出现）
- Apple 首次登录 Insert
- `WxUsernameRegister`

模拟用户经 `SimUsernameRegister` 创建时 MUST 同样写入 `created_at`（与真实用户相同写路径）。

#### Scenario: WeChat first login

- **WHEN** `WxLogin` 为新的 unionid 插入 wx 行
- **THEN** 新行 `created_at` MUST 大于 0

#### Scenario: Username register

- **WHEN** `WxUsernameRegister` 成功插入新账号
- **THEN** 新行 `created_at` MUST 大于 0

---

### Requirement: 用户名注册写入 wx 账号

系统 SHALL 提供用户名注册接口，并在 `ai_voice_device.wx` 新建账号行；该行以 `wx.id` 作为账号主键，`unionid` MAY 为空，`user_name` MUST 全局唯一，`password` MUST 以不可逆哈希密文保存。

#### Scenario: 注册成功
- **WHEN** 客户端提交合法且未占用的 `userName` 与 `password`
- **THEN** 系统 SHALL 新建一条 `wx` 记录并返回 `wxId`，且数据库中的 `password` SHALL NOT 为明文

#### Scenario: 用户名冲突
- **WHEN** 客户端提交的 `userName` 已被其他 `wx` 记录占用
- **THEN** 系统 SHALL 返回“用户名已存在”冲突错误，且 SHALL NOT 新建记录

### Requirement: 用户名密码登录

系统 SHALL 提供用户名登录接口，按 `user_name` 定位 `wx` 记录并校验哈希密码；校验通过后 SHALL 返回 `wxId` 与账号业务信息供网关签发令牌。

#### Scenario: 登录成功
- **WHEN** `userName` 存在且密码校验通过
- **THEN** 系统 SHALL 返回对应 `wxId`，并标识登录成功

#### Scenario: 登录失败
- **WHEN** `userName` 不存在或密码校验失败
- **THEN** 系统 SHALL 返回认证失败错误，且 SHALL NOT 泄露是“用户名不存在”还是“密码错误”的内部细节

### Requirement: 用户名账号绑定微信

系统 SHALL 提供用户名账号绑定微信接口，将微信 `unionid` 绑定到指定 `wx.id` 账号；同一 `unionid` MUST NOT 同时绑定多个账号。

#### Scenario: 绑定成功
- **WHEN** 当前账号未绑定微信，且目标 `unionid` 未被其他账号占用
- **THEN** 系统 SHALL 将该 `unionid` 写入当前账号并返回成功

#### Scenario: 微信已被占用
- **WHEN** 目标 `unionid` 已绑定在其他 `wx.id`
- **THEN** 系统 SHALL 返回“微信已绑定其他账号”错误，且 SHALL NOT 覆盖原绑定

### Requirement: 用户名账号绑定设备号

系统 SHALL 提供用户名账号绑定设备号接口，绑定前 MUST 校验设备号已在设备域注册，绑定后 SHALL 维护 `wx.device_no` 一致性并失效相关缓存。

#### Scenario: 绑定成功
- **WHEN** `deviceNo` 已注册且请求主体账号合法
- **THEN** 系统 SHALL 更新 `wx.device_no` 并返回成功

#### Scenario: 设备号未注册
- **WHEN** 提交的 `deviceNo` 未在设备域注册
- **THEN** 系统 SHALL 返回业务校验失败，且 SHALL NOT 更新绑定关系

### Requirement: 修改用户名密码

系统 SHALL 提供修改密码接口；调用方 MUST 提供旧密码并通过校验后方可写入新密码哈希。

#### Scenario: 改密成功
- **WHEN** 旧密码校验通过且新密码满足格式策略
- **THEN** 系统 SHALL 将 `password` 更新为新哈希并返回成功

#### Scenario: 旧密码错误
- **WHEN** 旧密码校验失败
- **THEN** 系统 SHALL 返回认证失败错误，且 SHALL NOT 修改数据库密码

### Requirement: 微信账号下创建用户名密码

系统 SHALL 提供“微信账号创建用户名密码”接口，使已存在微信账号（`unionid` 已绑定）补齐 `user_name` 与 `password`；若账号已存在用户名，系统 MUST 拒绝重复创建。

#### Scenario: 创建成功
- **WHEN** 微信账号存在且尚未设置 `user_name`
- **THEN** 系统 SHALL 写入唯一用户名与密码哈希，并返回成功

#### Scenario: 已存在用户名
- **WHEN** 当前微信账号已设置 `user_name`
- **THEN** 系统 SHALL 返回“账号已存在用户名密码”错误，且 SHALL NOT 覆盖原值

---
