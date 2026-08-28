## ADDED Requirements

### Requirement: Hub MUST 提供「开通功能管理」静态页

系统 MUST 在 `gateway-app-server` 注册付费功能管理静态页（建议 `/device/admin/cash-feature-admin.html`）。页面 MUST 使用 `AdminCommon.requireAdmin()` 与 `AdminCommon.adminFetch` 访问功能 Admin API，支持功能定义、解锁方式、功能 SKU 等 CRUD。浏览器 MUST NOT 发送 `X-Admin-Password`。

#### Scenario: 登录后可打开功能管理页

- **WHEN** 管理员已登录 Hub 并访问开通功能管理页
- **THEN** 页面 MUST 完成 Admin 校验并可通过 adminFetch 加载功能/SKU 列表

### Requirement: Hub MUST 提供独立「邀请码管理」静态页

系统 MUST 注册独立邀请码管理静态页（建议 `/device/admin/cash-invite-code-admin.html`）。页面 MUST 支持邀请码 CRUD（有效期、owner_wx_id、可开功能）及按码查看兑换明细（`wx_id` / 设备 / 功能 / 时间）。MUST 使用 AdminCommon，MUST NOT 在浏览器附带 `X-Admin-Password`。

#### Scenario: 可查看兑换明细

- **WHEN** 管理员在邀请码管理页打开某码的使用记录
- **THEN** 页面 MUST 展示该码成功兑换的用户与时间等信息

### Requirement: Admin Hub SHALL 登记两个模块入口

`admin-modules.js` MUST 增加至少两个模块：开通功能管理、邀请码管理（稳定 id、中文标题、`pagePath`、`showInNav: true`）。`RegisterAdminStaticPages` 与静态页 Bearer 白名单 MUST 包含上述 `pagePath`。

#### Scenario: 导航可见两个入口

- **WHEN** 管理员打开运维 Hub 导航
- **THEN** 模块列表 MUST 同时包含开通功能管理与邀请码管理入口

### Requirement: 管理页交互 MUST 对齐现有 cash-vip-admin 模式且可写

两页的登录、同源 API、错误提示与退出行为 MUST 对齐 Hub 惯例；MUST 为可写 CRUD，MUST NOT 误用 VIP 只读页的禁止写约束。

#### Scenario: 使用 AdminCommon 而非裸口令头

- **WHEN** 页面发起 Admin API 请求
- **THEN** 请求 MUST 经 Admin JWT / adminFetch，MUST NOT 在浏览器侧附带 `X-Admin-Password`
