## Why

预测 Pro 想法在主 Hub（`admin.html`）上以独立卡片 +「打开管理页」呈现，与其它运维能力「登录后顶栏 `adminNavLinks` 按钮」不一致；入口难找，且与登录态顶栏入口割裂，运维体验差。

## What Changes

- 将「预测 Pro 想法」登记进 `admin-modules.js`，`showInNav: true`，登录后出现在注册设备卡片右侧顶栏，与其它功能入口一致。
- 移除 `admin.html` 中的 `predictProIdeaCard` 卡片及登录后展示逻辑。
- 保留独立管理页 `/device/admin/predict-pro-ideas` 与现有 Admin API；管理页 init 对齐 qa/feedback：`if (requireAdmin())` 后再拉 list。
- 不改 App API、表结构、软删语义。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `predict-pro-idea-admin`: Hub 入口由主页卡片改为顶栏 `ADMIN_MODULES` 导航按钮；管理页仍为独立 HTML。

## Impact

- 静态：`resource/public/admin-modules.js`、`resource/public/admin.html`、可选微调 `predict-pro-ideas.html` init
- 进程：重部署 **`gateway-app-server`**（静态资源）；不必重部署 `device-service`（本变更无后端逻辑）
- Admin/App HTTP 契约路径不变
