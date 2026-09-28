## Context

`api-usage-stats.html` / `client-usage-stats.html` 用户维度调用 `usage/wx-list`、`client-usage/wx-list`（已支持 `page`/`pageSize`/`total`），但前端写死 `page=1&pageSize=50` 且无 pager。宝宝名筛选与 lastAt 排序为当前页内存操作；默认 wxId 降序已在页面落地。

## Goals / Non-Goals

**Goals:**

- 两页用户 wx 表：服务端翻页 + 可调 pageSize（默认 20）+ 页码跳转。
- 序号页内 1…N；换页/改条数重新请求并保留 `days`/`q`。

**Non-Goals:**

- 不改 Go API / device wx/list。
- 不做宝宝名/lastAt 的全库服务端筛选排序。
- 不改「按 API / 按功能」侧既有客户端分页。

## Decisions

### 1. 纯前端消费既有分页

- 状态：`wxPage`、`wxPageSize`（默认 20）、`wxTotal`；请求 `page`/`pageSize`/`days`/`q`。
- pageSize 选项：20 / 50 / 100（与 device 上限 100 对齐）。
- UI：每页条数下拉、页码输入+跳转、上一页/下一页、`第 p / tp 页，共 total 条`。
- 改 pageSize → 回到第 1 页再请求；非法页码夹到 `[1, totalPages]`。

### 2. 与本页筛选/排序共存

- 换页或改 pageSize 后：重置 `wxSortByLastAt=false`，按 wxId 降序渲染本页；宝宝名筛选项可保留，作用于新页数据。
- 序号：渲染结果上的 1…N（筛完后的可视序）。

### 3. 对齐既有 pager 样式

- 复用 `.pager` / `link-btn` 等现有样式，不引入新框架。

## Risks / Trade-offs

- [本页筛选导致「本页 0 条」但 total 仍大] → hint 已说明筛选仅当前页；翻页可继续找。
- [pageSize=100 时 lastAt 批量 HGETALL 变慢] → 可接受；默认 20。

## Migration Plan

仅静态页；强刷即可。回滚回退 HTML。

## Open Questions

- （无）pageSize 默认 20、两页、页内序号已确认。
