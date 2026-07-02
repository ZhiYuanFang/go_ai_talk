## MODIFIED Requirements

### Requirement: device admin SHALL provide paginated wx account list

`device-service` MUST 暴露 `GET /device/admin/api/wx/list`，鉴权与现有 device admin 一致（Header `X-Admin-Password`）。查询参数：`page`（默认 1）、`pageSize`（默认 20，最大 100）、可选 `q`（模糊匹配 wx.id、deviceNo、unionid、account）。响应 MUST 包含 `list`、`total`、`page`、`pageSize`；`list` 每项 SHALL 至少含 `id`（wxId）、`deviceNo`、`unionid`、`platform`、`account`（用户名账号若有）、`createdAt`（wx 账号创建 Unix 秒，0 表示未知）。列表 MUST **仅**包含 `is_simulated=0` 的真实用户；`is_simulated=1` 的行 MUST NOT 出现在 `list` 或 `total` 计数中。

#### Scenario: 默认分页列表

- **WHEN** 管理员携带有效 `X-Admin-Password` 请求 `GET /device/admin/api/wx/list` 且未传分页参数
- **THEN** 响应 SHALL 返回第一页 wx 记录且 `page=1`
- **AND** `list` MUST NOT 含 `is_simulated=1` 的 wxId

#### Scenario: 关键字搜索

- **WHEN** 管理员请求 `q=138` 且存在 deviceNo 或 id 匹配的非模拟 wx 行
- **THEN** 响应 `list` SHALL 仅包含匹配的非模拟项

#### Scenario: 经 gateway-app 反代可达

- **WHEN** 管理员经 gateway-app 携带 Admin JWT 请求 `/device/admin/api/wx/list`
- **THEN** gateway-app SHALL 反代至 device-service 并成功返回列表

#### Scenario: createdAt on new account

- **WHEN** 列表项对应迁移后新注册的 wx 账号
- **THEN** 该项 `createdAt` MUST 大于 0

#### Scenario: createdAt zero displays as unknown

- **WHEN** 列表项 `createdAt` 为 0
- **THEN** 消费方展示层 MAY 显示「—」
