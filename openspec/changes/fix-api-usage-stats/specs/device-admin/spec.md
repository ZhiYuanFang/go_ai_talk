## ADDED Requirements

### Requirement: device admin SHALL provide paginated wx account list

`device-service` MUST 暴露 `GET /device/admin/api/wx/list`，鉴权与现有 device admin 一致（Header `X-Admin-Password`）。查询参数：`page`（默认 1）、`pageSize`（默认 20，最大 100）、可选 `q`（模糊匹配 wx.id、deviceNo、unionid、account）。响应 MUST 包含 `list`、`total`、`page`、`pageSize`；`list` 每项 SHALL 至少含 `id`（wxId）、`deviceNo`、`unionid`、`platform`、`account`（用户名账号若有）。

#### Scenario: 默认分页列表

- **WHEN** 管理员携带有效 `X-Admin-Password` 请求 `GET /device/admin/api/wx/list` 且未传分页参数
- **THEN** 响应 SHALL 返回第一页 wx 记录且 `page=1`

#### Scenario: 关键字搜索

- **WHEN** 管理员请求 `q=138` 且存在 deviceNo 或 id 匹配的 wx 行
- **THEN** 响应 `list` SHALL 仅包含匹配项

#### Scenario: 经 gateway-app 反代可达

- **WHEN** 管理员经 gateway-app 携带 Admin JWT 请求 `/device/admin/api/wx/list`
- **THEN** gateway-app SHALL 反代至 device-service 并成功返回列表
