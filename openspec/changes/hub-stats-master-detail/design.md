## Context

`client-usage-stats.html` / `api-usage-stats.html` 在「按用户」Tab 中：上方为 wx 分页表，下方为选中用户的汇总与时间线/明细（`userDetailTitle`、功能表、timeline）。「按功能」Tab 同理：功能表下方挂 `featDetailPanel` / `apiDetailPanel`。页壳 `admin-pages.css` 内容区约 `max-width: 1080px`，宽屏两侧留白大、纵向滚动长。

本变更为纯前端布局 + 全局文档约定；统计 HTTP 契约不变。

## Goals / Non-Goals

**Goals:**

- 宽屏下列表与明细左右并列，选中行后右侧刷新，减少整页下滚。
- 两页「按用户」「按功能」明细面板行为一致。
- 适度放宽统计页内容宽度以容纳双栏。
- 将「Hub 优先左右分栏」写入 `openspec/project.md` 与 `AGENTS.md`，约束后续 Hub 页设计。

**Non-Goals:**

- 不改任何 usage / client-usage / device admin API。
- 不改 days 预设、分页 size、排序语义。
- 不重构全部 Hub 旧页为双栏（仅本两页落地 + 规范约束新/改页）。
- 不引入前端框架或新静态依赖。

## Decisions

1. **布局：CSS Grid / Flex master–detail**
   - Tab 内容区改为 `.hub-md`：左 `.hub-md-master`（表+分页），右 `.hub-md-detail`（原下方 detail 块迁入）。
   - 断点约 `max-width: 900px` 时 `flex-direction: column`，窄屏仍上下堆叠。
   - **备选**：仅加 sticky 详情在底——仍占纵向，否决。

2. **样式落点：共享类进 `admin-pages.css`**
   - 新增 `.hub-md*`、可选 `.hub-page-wide`（本两页 shell 使用更宽 `max-width`，如 `min(1400px, 96vw)`），避免两份 HTML 复制大段 CSS。
   - **备选**：仅内联 style——难复用、难与规范对齐，否决。

3. **交互保持现有 JS 数据流**
   - 仍点行 → 调现有 detail / timeline / feat-detail API → 填右侧 DOM；不预取全量。
   - 未选中时右侧占位文案（「未选择用户/功能」）；切 Tab / 换 days / 换页时清空或保持当前选中策略与现逻辑一致（换页可清空选中以免错位）。

4. **全局规范文案**
   - `openspec/project.md` 新增「Hub 运维页布局约定（强制）」小节；`AGENTS.md` 增对应短条，便于 AI/评审默认遵守。
   - 要点：列表+详情 MUST 宽屏左右；禁止默认整页纵向堆叠主从内容；窄屏可回退；新 Hub 页 proposal/design 须说明左右利用方式。

## Risks / Trade-offs

- [窄屏/笔记本] 双栏挤压 → 断点回退上下堆叠。
- [右侧内容很长] 时间线仍可能内滚 → 右侧 `overflow:auto` + 限制视口高度（如 `max-height: calc(100vh - …)`），左侧列表独立滚动。
- [仅改两页] 其它 Hub 页仍纵向 → 靠规范约束后续；本变更不强制扫全站改造。

## Migration Plan

- 静态资源随 gateway-app（或托管 `resource/public` 的进程）发布即可；无 DB/API 迁移。
- 回滚：还原两 HTML + `admin-pages.css` + 两份规范文档。

## Open Questions

- （无。）按用户明确要求：两统计页 + 全局规范；默认宽屏左右、窄屏堆叠。
