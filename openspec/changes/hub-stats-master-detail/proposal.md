## Why

Hub「客户端使用统计 / App API 使用统计」按用户时，明细与时间线堆在列表下方，宽屏两侧空白却要长时间下滚才能对照；运维体验差。同时后续 Hub 新页若继续「整页纵向堆叠」，同类问题会反复出现，需把「优先左右分栏」写进全局规范。

## What Changes

- 两页「按用户」改为 master–detail：左侧 wx 列表与分页，右侧展示选中用户的汇总/明细（client：功能汇总 + 操作时间线；api：API 调用明细）。
- 统计页局部放宽内容宽度，便于左右栏并存；窄屏回退为上下堆叠。
- 「按功能」Tab 中表下明细（功能下用户列表等）同步改为列表右侧，与按用户一致。
- 在 `openspec/project.md` 与 `AGENTS.md` 增加 **Hub 运维页布局约定（强制）**：新增/改造 Hub 静态页时 MUST 优先左右利用空间，禁止默认把主列表与详情整页纵向堆叠（窄屏除外）。
- **不**改统计 API、分页、排序或 days 语义。

## Capabilities

### New Capabilities

- `hub-stats-master-detail`：两页使用统计按用户/按功能明细改为左右分栏 master–detail，并适度放宽页宽。
- `hub-layout-horizontal-first`：全局 Hub 运维页「优先左右分栏」规范写入 project.md / AGENTS.md。

### Modified Capabilities

- （无。）

## Impact

- **静态页**：`client-usage-stats.html`、`api-usage-stats.html`；可选共享样式写入 `admin-pages.css`。
- **规范**：`openspec/project.md`、`AGENTS.md`。
- **部署**：gateway-app（静态）；无后端发版要求。
- **usage**：Admin 页，不改 `maintenance_skip`。
