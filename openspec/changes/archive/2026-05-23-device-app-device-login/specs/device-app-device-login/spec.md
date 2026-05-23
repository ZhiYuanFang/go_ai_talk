## ADDED Requirements

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
