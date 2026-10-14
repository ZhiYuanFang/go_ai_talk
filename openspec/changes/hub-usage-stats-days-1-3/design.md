## Context

Hub 两页使用统计（`client-usage-stats.html`、`api-usage-stats.html`）顶栏 `daysSelect` 仅「近 7 天」「近 30 天」，默认 7。请求已把 `days` 传给既有 Admin API；client-usage / api-usage 的 `NormalizeDays`（或等价）将天数规范到 `[1,30]`。仓库内仅此两页含该时间范围控件。

约束：只改静态 HTML；不新增 Redis 键；不改排序语义；Admin 不碰 App usage skip。

## Goals / Non-Goals

**Goals:**

- 两页时间范围选项为：近 1 / 3 / 7 / 30 天。
- 首次打开默认「近 1 天」。

**Non-Goals:**

- 自填任意正整数天数。
- wx 列表全量 lastAt 排序或 ZSET。
- 放宽 Redis 30 天窗口上限。
- 改 Go API DTO 或 handler。

## Decisions

### D1：仅改两页 `<select>`，不改 JS 传参逻辑

`daysParam()` / `qSuffix()` 已读 `daysSelect.value`；新增 `<option value="1">`、`value="3"`，并将 `selected` 移到 `value="1"`。

**备选**：自定义输入框 → 否决（用户明确只要预设）。只改 client 页 → 否决（用户要求其它统计页一起）。

### D2：选项顺序

`1 → 3 → 7 → 30`，由短到长，默认 1。

### D3：部署

仅 gateway-app 静态资源；强刷/no-cache 页已 `noCache: true`。

## Risks / Trade-offs

- [运维习惯默认 7 天] → 可手动改回 7；默认 1 为本次明确要求
- [误以为解决了按用户全表 lastAt 全局排序] → 文案/验收不宣称排序变更；本 change 仅窗口预设

## Migration Plan

1. 发 gateway-app。
2. 目视两页下拉含 1/3/7/30，默认 1；切换后请求 `days=` 与选项一致。
3. 回滚：还原两 HTML 选项即可。

## Open Questions

（无。）
