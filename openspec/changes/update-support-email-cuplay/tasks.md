## 1. 静态页邮箱替换

- [x] 1.1 更新 `resource/public/pangbao-home.html`：页脚客服邮箱与 `mailto:` 改为 `pangbao@cuplay.top`
- [x] 1.2 更新 `resource/public/privacy-policy.html`：全部 `tou_zy@foxmail.com` → `pangbao@cuplay.top`；将「修订日期」更新为发布日（生效日期不变）
- [x] 1.3 更新 `resource/public/user-agreement.html`：全部旧邮箱 → `pangbao@cuplay.top`；将「修订日期」更新为发布日（生效日期不变）

## 2. 验收与边界

- [x] 2.1 全仓 `rg 'tou_zy@foxmail.com'` 期望 0（除历史归档/旧版本基线文档若保留外；实现页与本 change 内规格不得残留）
- [x] 2.2 确认路径仍为 `/privacy-policy.html`、`/user-agreement.html`；**不新增** App HTTP 路由（gateway-app / usage N/A）、**不新增** Redis、**不新增** 背景循环
- [x] 2.3 发版前确认 `pangbao@cuplay.top` 可收信（运维/业务勾选）；抽检官网页脚与两份合规页展示新邮箱
