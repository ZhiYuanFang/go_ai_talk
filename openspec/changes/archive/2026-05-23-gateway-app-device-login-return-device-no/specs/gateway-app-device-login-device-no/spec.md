## ADDED Requirements

### Requirement: 设备号业务登录响应须含非空 deviceNo

对 **POST `/device/app/api/user/device_login`**（device-service，业务成功 `code=0`），响应 `data` **MUST** 包含 JSON 字段 **`deviceNo`**，且为 **trim 后非空字符串**，表示本次完成登录校验的设备号（与请求入参在规范化后一致，或与库内绑定到该会话的权威 `device_no` 一致）。

#### Scenario: 纯设备会话（无 wx 行绑定）

- **WHEN** 客户端提交已注册设备号且 wx 表无对应行，业务校验成功
- **THEN** 响应 `data.deviceNo` **MUST** 为非空字符串，且与本次登录所用设备号一致

### Requirement: 网关聚合设备号登录响应须含非空 deviceNo

对 **POST `/device/app/api/device_login`**（gateway-app-server，业务成功 `code=0`），响应 `data` **MUST** 包含 **`deviceNo`**，且为 **trim 后非空字符串**。若下游 device 返回的 `data.deviceNo` 为空或缺失，网关 **SHALL** 使用本次请求体中的 `deviceNo`（trim 后）作为回包与 JWT 签发所用设备号；若兜底后仍为空，**MUST** 拒绝成功语义（沿用现有参数/内部错误路径）。

#### Scenario: 下游 data 缺 deviceNo 时网关兜底

- **WHEN** device `device_login` 返回 `code=0` 但 `data` 中无 `deviceNo` 或值为空白，且请求体 `deviceNo` 经 trim 后非空
- **THEN** 网关返回的聚合响应 `data.deviceNo` **MUST** 等于该 trim 后的请求 `deviceNo`，且签发的 access/refresh 所绑定的设备号与该值一致

#### Scenario: 请求与下游均无可用设备号

- **WHEN** 请求体 `deviceNo` trim 后为空，或业务失败
- **THEN** 网关 **MUST NOT** 返回 `code=0` 且带非空 `deviceNo` 的成功形态（保持现有错误语义）
