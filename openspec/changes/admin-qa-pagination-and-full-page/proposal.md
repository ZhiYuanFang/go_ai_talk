# 提案：管理端问答库分页与全量页

## 背景

设备管理后台「问答库」此前通过 `GET /device/admin/api/qa/list` 一次拉取全表（`id` 升序），数据量大时加载慢且无法在 UI 删除记录。`qa` 表权威在 **voice-service** 库，device 须经 HTTP 委派。

## 目标

- 管理端首页问答库卡片标题为 **「问答库」**，默认展示 **10** 条、`id` **倒序**。
- 列表 API 支持 `page` / `pageSize` 分页，响应含 `total`。
- `total > 10` 时显示 **「展开更多」**，跳转独立页 `/device/admin/qa-records`，支持分页浏览与 **删除**。
- 删除经 device admin API → voice internal API，device 不直连 `qa` 表。

## 非目标

- 不按设备过滤（`qa` 为全局表）。
- 不改动 `domain_outbox.event_type` 等 worker 语义。

## 影响范围

- `voice-service`：`ListQaPage`、`DeleteQa`、内部 HTTP。
- `device-service`：admin 分页/删除委派、静态页路由。
- `gateway` / `gateway-app`：静态页 `qa-records.html` 路由。
