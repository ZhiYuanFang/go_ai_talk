## MODIFIED Requirements

### Requirement: Admin predict-pro-ideas page SHALL mirror feedback-records pattern

The system MUST serve `resource/public/predict-pro-ideas.html` at `/device/admin/predict-pro-ideas`. The page MUST load `admin-pages.css` and follow Hub 近全宽布局；宽屏若同时存在列表与编辑区 MUST 左右分栏或等价避免整页长滚。The page MUST use Admin JWT（`sessionStorage.gatewayAdminToken` / `AdminCommon.requireAdmin`），仅在已登录时调用 Admin predict-pro-ideas APIs，show a paginated table including wxId and baby name, allow edit of display text and delete, and link back to `/device/admin`.

Hub 入口 MUST 登记于 `resource/public/admin-modules.js`（`showInNav: true`），登录后出现在主 Hub 顶栏 `adminNavLinks`，与其它功能入口一致。`admin.html` MUST NOT 再保留独立的「预测 Pro 想法」中间卡片作为入口。gateway-app MUST 继续注册 HTML 于 `admin_static_pages` 且 Bearer 豁免 **HTML shell path only**。Admin **必须** 提供独立管理页与顶栏入口。

#### Scenario: Open management page from top nav

- **WHEN** 管理员已登录主 Hub 并点击顶栏「预测 Pro 想法」
- **THEN** 浏览器 SHALL 进入 `/device/admin/predict-pro-ideas`，且在有效 Admin token 下 SHALL 请求 list 并展示想法表

#### Scenario: No mid-page card entry

- **WHEN** 管理员查看主 Hub 登录后首页
- **THEN** 页面 MUST NOT 再展示「预测 Pro 想法」中间卡片入口；MUST 能通过顶栏按钮进入管理页
