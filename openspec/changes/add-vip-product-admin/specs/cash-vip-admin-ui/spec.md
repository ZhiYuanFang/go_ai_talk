## MODIFIED Requirements

### Requirement: Admin Hub SHALL 登记 cash-vip-admin 模块

`resource/public/admin-modules.js` MUST 保留模块登记：`id: cash-vip-admin`（或等价稳定 id）、`pagePath` 为 `/device/admin/cash-vip-admin.html`、`showInNav: false`。Hub 主导航 MUST NOT 展示「VIP 权益」独立入口。`RegisterAdminStaticPages`（或等价单点注册表）MUST 仍包含该 `pagePath`。PR MUST NOT 仅在主网关注册该页而 App 网关不可见。权益只读页的运维入口 MUST 为开通功能管理页文案为「VIP列表」的按钮（或等价控件）并导航至上述 `pagePath`。

#### Scenario: Hub 导航不展示 VIP 模块

- **WHEN** 运维已登录 `/device/admin` Hub
- **THEN** 模块列表 MUST NOT 以独立导航项展示「VIP 权益」

#### Scenario: 静态页仍可直达

- **WHEN** 已登录管理员打开 `/device/admin/cash-vip-admin.html`
- **THEN** 主内容区 MUST 可见且 MUST 能发起对 `/cash/admin/api/vip/entitlements` 的加载请求
