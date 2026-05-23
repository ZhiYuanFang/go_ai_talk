# Spec: 管理端问答库

## REQ-QA-001 分页列表

管理端 MUST 通过分页接口获取问答库，默认 `pageSize=10`，按 `id` 降序。

#### Scenario: 首页预览

- **WHEN** 管理员登录设备管理页
- **THEN** 问答库卡片展示最多 10 条最新记录

#### Scenario: 分页参数

- **WHEN** 请求 `GET /device/admin/api/qa/list?page=2&pageSize=20`
- **THEN** 响应包含 `list`、`total`、`page`、`pageSize`

## REQ-QA-002 展开更多

当 `total > 10` 时，管理端首页 MUST 显示「展开更多」链接至全量页。

## REQ-QA-003 删除

全量页 MUST 支持按 `id` 删除问答库行；删除 MUST 经 voice 内部接口落库，device 不直连 `qa` 表。

#### Scenario: 删除成功

- **WHEN** 管理员确认删除并提交 `POST /device/admin/api/qa/delete`
- **THEN** 对应行从列表消失且 voice 库记录已删除
