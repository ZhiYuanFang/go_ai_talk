## Why

Hub 运维页默认约 1180px，宽屏两侧大量留白；此前仅统计页加到 1400 仍偏窄。多页（开通功能、向量飞轮、设备历史、Hub 首页多卡片等）列表与表单纵向堆叠，运维需长时间下滚。需一次纯前端统一近全宽，并把明显「上列表下表单/多段」页改为左右分栏。

## What Changes

- **全部**加载 `admin-pages.css` 的 Hub 运维页统一近全宽（`calc(100% - 32px)`，无 1400 封顶）；`page-narrow` / `hub-page-wide` 与默认一致。
- 不改 `pangbao-theme.css` 的 `.page`（隐私协议/首页等非 Hub 页保持原宽）。
- 纵向过长的页改为左右：`cash-feature-admin`（磁贴 | 表单）、全部 `*-vector-admin`（列表 | 批量上传）、`cash-xiaozhi-mcp-admin`（配置 | 列表/视频）、`admin.html` / `history.html` 多卡片宽屏两列。
- 更新 `project.md` / `AGENTS.md`：Hub 默认近全宽 + 继续强调 master–detail。
- **不**改任何 API。

## Capabilities

### New Capabilities

- `hub-near-fullwidth`: Hub 运维页统一近全宽壳层。
- `hub-long-pages-master-detail`: 长页左右分栏改造（功能管理、向量页、MCP、Hub 首页与设备历史多卡片）。

### Modified Capabilities

- （无基线 capability 名变更；布局约定文案增量见 design。）

## Impact

- `resource/public/admin-pages.css`、多份 Hub HTML；`openspec/project.md`、`AGENTS.md`。
- 静态资源发布即可；无后端发版。
