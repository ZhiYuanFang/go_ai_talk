## 1. 顶栏入口

- [x] 1.1 在 `admin-modules.js` 增加 `predict-pro-ideas`（title「预测 Pro 想法」、`pagePath: /device/admin/predict-pro-ideas`、`showInNav: true`）
- [x] 1.2 从 `admin.html` 删除 `predictProIdeaCard` 及相关 JS（变量、登录后 unhide）

## 2. 管理页 init

- [x] 2.1 `predict-pro-ideas.html` 改为 `if (AdminCommon.requireAdmin()) { 显示 mainCard; loadPage(1) }`，对齐 qa/feedback

## 3. 部署说明

- [x] 3.1 回复列出重部署：`gateway-app-server`；不必重部署 `device-service`（无后端变更）
