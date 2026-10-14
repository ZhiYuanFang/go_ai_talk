## Context

`pangbao-theme.css` 的 `.page`≈1180px 被 Hub 与官网协议页共用。Hub 页另引 `admin-pages.css`，可在此覆盖宽度而不影响隐私协议/首页。`hub-stats-master-detail` 已引入 `.hub-md*` 与偏保守的 `.hub-page-wide`(1400)。

## Goals / Non-Goals

**Goals:**

- Hub（admin-pages）近全宽统一。
- 长页左右分栏，减少下滚。
- 纯 HTML/CSS，零 API。

**Non-Goals:**

- 不改协议页/官网宽度。
- 不逐页重做交互逻辑；ID/JS 尽量不动。
- 不强制所有单表页做 master–detail。

## Decisions

1. **宽度落点：`admin-pages.css` 覆盖 `.page` / `.card`**
   - `width: calc(100% - 32px)`；`.page-narrow`、`.hub-page-wide` 同规则（兼容旧 class）。
   - **不**改 theme，避免协议页变全宽。

2. **复用 `.hub-md*`**；多卡片页用 `.hub-cards-2col`（宽屏 grid 两列）。

3. **页级改造优先级**
   - cash-feature：左竖向 tile，右 hub-sec。
   - 10 个 vector-admin：左表右批量。
   - xiaozhi-mcp：左 A+B，右 C+D。
   - admin.html / history.html：卡片两列网格。

## Risks / Trade-offs

- [表单过宽] → 右栏内既有 form-grid；可去掉过严 max-width:920。
- [窄屏] → 既有 900px hub-md 回退；cards-2col 同步回退单列。

## Migration Plan

- 静态发布；Ctrl+F5。回滚 CSS + 相关 HTML。

## Open Questions

- （无。）
