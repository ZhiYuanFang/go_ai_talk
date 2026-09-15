## ADDED Requirements

### Requirement: App MUST 能注册与注销推送 token

系统 MUST 经 gateway-app 对外提供中性路径注册/注销接口（如 `POST /app/api/push/register` 与 `POST /app/api/push/unregister`），请求体含 `channel`、`token`、`deviceKey`（注销可按既有字段）。接口 MUST 使用登录用户 `wxId`（网关注入内部头），MUST NOT 要求微信 unionid 绑定作为前提。系统 MUST NOT 再提供 `/ucg/app/api/push/register` 与 `/ucg/app/api/push/unregister` 作为有效推送注册入口。

#### Scenario: 登录用户注册成功

- **WHEN** 已登录用户以合法 channel/token/deviceKey 调用注册接口
- **THEN** push-service MUST 持久化该 token，且后续按该 wxId 发送可见推送时可投递到对应通道

#### Scenario: 旧 UCG 注册路径移除

- **WHEN** 客户端请求已删除的 `/ucg/app/api/push/register`
- **THEN** 系统 MUST NOT 再将其作为推送注册成功路径（404 或网关未反代至 push 实现）

### Requirement: Internal MUST 支持按 bizType 发送且 badge 由调用方提供

push-service MUST 提供经内部密钥鉴权的发送接口，入参至少含 `wxId`、`bizType`、可见文案（或静默标志）与可选 `badge`、`data`。当请求携带 `badge` 时，下发载荷 MUST 使用该值；未携带时 MUST 默认 `0`。push-service MUST NOT 为填充 badge 去查询 UCG 未读。无有效 token 时 MUST 跳过发送且 MUST NOT 视为调用方硬失败（可返回成功空投递或等价 best-effort 语义，并在实现中固定）。

#### Scenario: UCG 传入未读角标

- **WHEN** 调用方传入 `badge=5` 与非空 alert
- **THEN** 厂商载荷中的角标 MUST 为 5

#### Scenario: 预测类默认零角标

- **WHEN** 调用方不传 badge 或显式传 0
- **THEN** 厂商载荷角标 MUST 为 0

### Requirement: push 注册注销 MUST 不计入 App usage 统计

经 gateway-app 暴露的 push 注册与注销 App 接口 MUST 列入 usage 维护型排除（`maintenance_skip` 精确 METHOD+path）。负责人结论：不统计。旧 ucg push 排除项 MUST 在路由删除后移除或替换为新 path，避免残留无效键。

#### Scenario: 注册不写入使用统计

- **WHEN** 客户端成功或失败调用 `POST /app/api/push/register`
- **THEN** 该次调用 MUST NOT 计入 App API 使用统计报表
