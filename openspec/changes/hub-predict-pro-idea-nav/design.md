## Context

`predict-pro-idea-hub` 在主 Hub 增加了 `predictProIdeaCard`（仅文案 +「打开管理页」），未登记 `ADMIN_MODULES`。其它运维页通过 `admin-modules.js` + `initAdminModuleLinks` 注入登录后顶栏。运维希望 Pro 想法入口与顶栏按钮一致，并去掉中间卡片。

## Goals / Non-Goals

**Goals:**

- 登录后顶栏出现「预测 Pro 想法」，跳转现有 `/device/admin/predict-pro-ideas`。
- 删除主 Hub 中间卡片入口。
- 管理页在有 Admin JWT（`sessionStorage.gatewayAdminToken`）时才拉 list，对齐 qa/feedback。

**Non-Goals:**

- 不把列表内嵌进 `admin.html`。
- 不改 Admin/App API、gateway 反代、Bearer 豁免表（路径已登记）。
- 不改 `device-service` 业务逻辑。

## Decisions

1. **入口唯一来源：`ADMIN_MODULES`**
   - 新增 `{ id: 'predict-pro-ideas', title: '预测 Pro 想法', pagePath: '/device/admin/predict-pro-ideas', showInNav: true }`。
   - 备选：仅改卡片文案 —— 仍与顶栏模式不一致，否决。

2. **删除 `predictProIdeaCard`**
   - 去掉 HTML 块、`predictProIdeaCard` 变量及登录成功后的 `remove('hidden')`，避免双入口。

3. **管理页 init 对齐**
   - 改为 `if (AdminCommon.requireAdmin()) { mainCard 显示; loadPage(1) }`，无 token 只跳转登录、不发 list。

## Risks / Trade-offs

- [Risk] 运维仍收藏旧卡片位置 → Mitigation：顶栏按钮更显眼；卡片移除后无死链（旧书签仍指向独立页）。
- [Trade-off] 顶栏按钮变多 → 与现有模块同一密度，可接受。

## Migration Plan

1. 合并静态改动，重部署 **gateway-app-server**。
2. 验证：Hub 登录 → 顶栏可见「预测 Pro 想法」→ 进入后有 list 请求；主页无想法卡片。
3. Rollback：回滚 gateway 静态即可。

## Open Questions

（无）
