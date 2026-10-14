## 1. 规范与全局 CSS

- [x] 1.1 更新 `openspec/project.md` / `AGENTS.md`：Hub 默认近全宽；长页左右分栏
- [x] 1.2 `admin-pages.css`：覆盖 `.page`/`.card` 为 `calc(100%-32px)`；`page-narrow`/`hub-page-wide` 同宽；增加 `.hub-cards-2col`；磁贴竖向辅助类

## 2. 长页左右分栏

- [x] 2.1 `cash-feature-admin.html`：左磁贴右表单（hub-md）
- [x] 2.2 全部 `*-vector-admin.html`：左列表右批量上传
- [x] 2.3 `cash-xiaozhi-mcp-admin.html`：左 A+B 右 C+D
- [x] 2.4 `admin.html`：登录后多卡片包进 `.hub-cards-2col`
- [x] 2.5 `history.html`：多卡片宽屏两列分组

## 3. 清理

- [x] 3.1 统计页可去掉冗余 `hub-page-wide`（行为已由全局覆盖）
- [x] 3.2 自检：无 API 改动；窄屏回退可用
