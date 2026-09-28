## Why

功能使用统计与客户端使用统计的「按用户」wx 列表写死 `page=1&pageSize=50` 且无分页控件，账号总数超过一页时无法翻页，也无法调整每页条数，运维只能看到最新一批用户。

## What Changes

- 两页（`api-usage-stats.html`、`client-usage-stats.html`）用户维度 wx 列表改为**服务端分页**：请求携带可变 `page` / `pageSize`。
- UI 提供：上一页 / 下一页、页码信息、可改每页条数、可输入页码跳转；**默认 pageSize=20**（可选 20/50/100，不超过接口上限 100）。
- 序号仍为**当前页内** 1…N（非全局偏移）。
- 宝宝名筛选、最近使用时间排序仍仅作用于当前页已加载行（与既有 `admin-hub-usage-ux` 约定一致）；换页或改 pageSize 时重新请求并重置为本页默认 wxId 降序。
- **不改**后端 wx-list 契约（已支持 page/pageSize）；无 App 接口变更。

## Capabilities

### New Capabilities

- `admin-usage-wx-list-pager`：运维使用统计两页「按用户」列表的分页与每页条数交互。

### Modified Capabilities

- （无）——不修改 `gateway-app-api-usage-stats` / `client-feature-usage-stats` 的 API 需求；仅前端消费既有分页参数。

## Impact

- **静态页**：`resource/public/api-usage-stats.html`、`client-usage-stats.html`。
- **API**：沿用既有 `GET .../usage/wx-list` 与 `.../client-usage/wx-list` 的 `page`/`pageSize`/`total`；无需改 Go handler。
- **usage 统计 / Bearer / Redis**：无影响。
