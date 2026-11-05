## Context

History 实体与 admin/app list 已含可选 `expectStartTime`（见 `historyListSelectFields`）。Hub `history.html` 主表未展示；使用统计右侧历史 Tab（`hub-usage-history-side-panel`）列集对齐主表但同样无预计列。产品选定：只加「预计开始」，并同步嵌入 Tab。

## Goals / Non-Goals

**Goals:**

- 三处表格增加「预计开始」列，格式与现有「开始时间」可读风格一致（复用或对齐 `formatHistoryDateTime`；无值显示 `-`）。
- 列位置建议：在「开始时间」之前，便于「预计 → 实际」对照。

**Non-Goals:**

- 服务端改动、expectEndTime/expectAmount 列、编辑表单写入预计字段、准确率卡改版。

## Decisions

1. **仅静态页。** list 已回传则直接读 `it.expectStartTime`；无值 `-`。
2. **三页同一列序与文案。** history 主表与两页 usage 嵌入表一致。
3. **colspan / 空态** 随新列数更新（history 原 7 列含操作 → 8；usage 嵌入 6 → 7）。

## Risks / Trade-offs

- [库中尚无 expect 数据时大量 `-`] → 接受；有值后自然可见。
- [指针/数字 JSON 形态] → 前端对 number 与可解析字符串兼容。

## Migration Plan

1. 改三静态页；gateway-app 发版。
2. 回滚：去掉列即可。

## Open Questions

- （已收敛）仅预计开始；同步 usage 右侧历史 Tab。
