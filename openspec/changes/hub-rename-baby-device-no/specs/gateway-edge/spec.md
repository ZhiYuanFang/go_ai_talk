# Spec Delta

## ADDED Requirements

### Requirement: gateway-app SHALL 支持按账号吊销 App refresh 会话

gateway-app-server MUST 提供可信内部或管理路径，按 `wxId`（可批量）使该账号现有 App refresh 令牌失效，从而使客户端在不修改 Flutter 的前提下无法静默刷新并 MUST 走既有登录流。宝宝号改名编排 MUST 在数据迁移成功后对该宝宝当前绑定的全部 `wxId` 触发吊销。纯设备会话（`wxId=0`）若仍持有旧号 refresh，MUST 同样无法继续刷新成功。

#### Scenario: 改名后 refresh 失败

- **WHEN** 某 wx 曾绑定旧宝宝号且改名流程已对其执行会话吊销
- **THEN** 客户端使用该账号改名前签发的 refresh 调用 `POST /device/app/api/token/refresh` MUST 失败

#### Scenario: 重登获得新号

- **WHEN** 吊销后用户通过既有聚合登录成功且其 wx 已绑定新宝宝号
- **THEN** 登录响应 MUST 包含非空的新 `deviceNo` 与新的 access/refresh

## MODIFIED Requirements

### Requirement: refresh 重新签发的 access SHALL 同步 device_no 声明

gateway-app-server 在处理 **`POST /device/app/api/token/refresh`** 时，**SHALL** 在签发新 access JWT 时写入 **与当前 wx 会话权威一致的 `device_no` 声明**。当 `wxId>0` 时，权威 `device_no` MUST 来自 device-service 按 wx 主键查询的当前绑定（或网关侧同等单一规则），MUST NOT 仅复制 refresh 载荷中携带的旧 `device_no` 字符串。`wxId=0` 的纯设备会话 MUST 继续以 refresh 载荷中的设备号为准，且该号 MUST 仍能通过设备注册校验策略（若实现侧增加校验）。本策略用于避免换绑或宝宝号改名后长期持有错误 `device_no` claim。

#### Scenario: 刷新成功

- **WHEN** refresh_token 有效且旋转策略允许签发新 access，且 `wxId>0`、device 域当前绑定非空
- **THEN** 新 access JWT **MUST** 包含与 device 域当前绑定一致的 **`device_no` 声明**

#### Scenario: 换绑或改名后刷新拿到新号声明

- **WHEN** refresh 载荷仍携带旧 `device_no` 字符串，但 device 域该 wx 当前绑定已变为新号
- **THEN** 新 access（及新 refresh 载荷若写入 device_no）MUST 使用新号，MUST NOT 继续签发旧号声明
