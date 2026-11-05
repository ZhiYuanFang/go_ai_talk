## Context

两页使用统计已是 `.hub-md` 主从布局。历史权威 list 已有 `GET /device/admin/api/history/list`（无时间参数）。产品纠正：只做静态页嵌入，**禁止**为时间窗改服务端；表样式对齐 `history.html`；usage `days=3` 只改左栏默认活跃窗。

## Goals / Non-Goals

**Goals:**

- 点选用户后右侧默认历史 Tab；可切使用明细。
- 历史：`deviceNo` + `page` + `pageSize=50` 调既有 admin list；列对齐运维历史表（无编辑/删除）。
- 左栏默认：有宝宝名 / 100 / days=3。

**Non-Goals:**

- 任何 history-service / `api/v1` admin list 字段扩展。
- 按 usage days 过滤历史。
- 统计页内写历史。

## Decisions

1. **零服务端改动。** 前端只调现有 list；曾加的 startTime/endTime / ListV2 路径 MUST 回滚。
2. **历史不跟 days。** days 只影响 usage/wx-list；切 days 时重拉左栏与（若在）使用明细即可，历史可不按时间重滤（可选刷新同 list）。
3. **表样式。** 表头：ID、事件名、数量/时长、开始时间、结束时间、备注。数量/时长用行内 `eventNumber`/`eventUnit` 与起止时间推算（无事件目录时尽力展示，与 history.html 观感接近）。时间文案贴近运维页（如进行中、今日时分）。
4. **pageSize=50 固定**；链到 `history.html` 做完整运维。

## Risks / Trade-offs

- [无 eventOptions 时「数量/时长」略简于运维页] → 接受；深挖走完整页。
- [宝宝名仅当前页滤] → 保持原语义。

## Migration Plan

1. 回滚任何 admin list 服务端改动。
2. 两静态页按上线；gateway-app 发静态即可。

## Open Questions

- （已收敛）无服务端；无历史时间窗；表样式对齐 history.html 只读。
