## Why

官网能力演示视频藏在按钮弹层里且 CSS 未命中导致裁切，差异化能力不够可见；隐私政策仍把本地缓存的宝宝头像/成长记录配图写成「收集」，并暴露预测提醒实现细节、罗列多供应商大模型，与产品事实及披露策略不符。

## What Changes

- **官网 `#xiaozhi`（现位置）**：`promo` 槽能力演示改为页面内嵌 `<video>`，随窗口缩放、整段可见（`object-fit:contain`）；**官网 MUST NOT 展示 `tutorial`（绑定教学）**（前端过滤即可，API/Hub 双槽可保留）。
- **官网文案**：将难懂的「小智音箱 / MCP」用户面表述改为「实体音箱连接能力」等常人可懂说法（导航、Hero、能力区、meta 等对齐）；删除「点按钮看引导视频」类指向。
- **隐私政策**：成长记录删除视频相关描述；明确宝宝头像与成长记录配图**不收集**、仅用户本地；预测临近提醒改为一句「可能发送预测临近提醒」、**不提**客户端同步/服务端下发；第三方大模型**仅保留阿里云 DashScope**，去掉智谱/DeepSeek 与「以运营配置为准」类表述；更新修订日期。
- **规格**：修订 `gateway-app-official-site`（内嵌 promo、不展示 tutorial、实体音箱文案）与 `privacy-policy-company`（本地媒体、预测措辞、单一 DashScope）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `gateway-app-official-site`：音箱区块内嵌 promo；官网不展示 tutorial；用户面文案改为实体音箱连接能力。
- `privacy-policy-company`：本地头像/配图不收集；预测提醒不披露同步链路；AI 仅 DashScope。

## Impact

- **静态页**：`resource/public/pangbao-home.html`、`resource/public/privacy-policy.html`（经 `gateway-app-server` 对外）。
- **不必改**：`cash` guide-videos API、Hub 双槽运维、`user-agreement.html`（除非顺带统一叫法——本变更默认不动）、Flutter。
- **部署**：`gateway-app-server`（静态资源）；无需改 `.env`。
