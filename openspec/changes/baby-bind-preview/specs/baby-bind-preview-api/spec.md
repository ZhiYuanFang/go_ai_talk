## ADDED Requirements

### Requirement: App SHALL preview baby nickname before bind without side effects

`POST /device/app/api/user/bind_preview` MUST accept JSON body `{ "deviceNo": "<string>" }` under a valid App session (gateway-injected `X-Internal-Wx-Id`, same class as `bindwx`). On success the response `data` MUST include the trimmed `deviceNo` echo and `babyName` (string; MAY be empty). The handler MUST NOT update `wx.device_no` or otherwise bind the caller to the device. Empty trimmed `deviceNo` MUST fail with a parameter/business error. An unregistered or unknown `deviceNo` MUST fail with a business error and MUST NOT return a successful empty preview that looks like a valid baby. The endpoint MUST NOT be auth-exempt. 已登录会话 MUST 可按 `deviceNo` 只读预览宝宝昵称；MUST NOT 产生绑定副作用；无效/未注册 MUST 业务失败。

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
