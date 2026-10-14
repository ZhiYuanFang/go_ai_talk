## 1. 全局规范

- [x] 1.1 在 `openspec/project.md` 增加「Hub 运维页布局约定（强制）」：宽屏列表+详情左右分栏、禁止默认整页纵向堆叠主从、窄屏可回退、后续 proposal/design 须说明空间利用
- [x] 1.2 在 `AGENTS.md` 增加对应短条，与 project.md 一致，便于 AI/评审默认遵守

## 2. 共享样式

- [x] 2.1 在 `resource/public/admin-pages.css` 增加 `.hub-md` / `.hub-md-master` / `.hub-md-detail`（宽屏横向、窄屏约 900px 纵向回退；右侧可独立滚动）
- [x] 2.2 增加统计页加宽类（如 `.hub-page-wide`，相对默认 ~1080px 放宽），供两统计页使用

## 3. 客户端使用统计页

- [x] 3.1 `client-usage-stats.html`「按用户」：列表与用户汇总/时间线改为左右 master–detail，沿用现有 API 与选中逻辑
- [x] 3.2 同页「按功能」：`featDetailPanel` 迁到功能表右侧分栏
- [x] 3.3 应用加宽页壳类；窄屏堆叠可验收

## 4. App API 使用统计页

- [x] 4.1 `api-usage-stats.html`「按用户」：列表与用户 API 明细改为左右 master–detail
- [x] 4.2 同页「按功能」：API 明细面板迁到右侧分栏
- [x] 4.3 应用加宽页壳类；确认 query/path 与改前一致

## 5. 验收

- [x] 5.1 宽屏：两页按用户/按功能均无需整页下滚即可对照列表与详情
- [x] 5.2 窄屏：上下堆叠仍可用
- [x] 5.3 规范两文件可检索到 Hub 布局约定
