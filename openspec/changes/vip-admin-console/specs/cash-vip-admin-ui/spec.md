## ADDED Requirements

### Requirement: Hub SHALL 提供 VIP 权益只读管理页

系统 MUST 在 `gateway-app-server` 注册静态页 `/device/admin/cash-vip-admin.html`（静态文件 `resource/public/cash-vip-admin.html`）。页面 MUST 使用 `resource/public/admin-common.js` 的 `AdminCommon.requireAdmin()` 与 `AdminCommon.adminFetch`（或等价封装）初始化。页面 MUST 为只读：MUST NOT 提供改价、退款、手工开通或编辑权益的表单/按钮。浏览器 MUST NOT 发送 `X-Admin-Password`。

#### Scenario: 已登录 Hub 打开页面

- **WHEN** 管理员已在运维 Hub 登录并访问 `/device/admin/cash-vip-admin.html`
- **THEN** 主内容区 MUST 可见且 MUST 能发起对 `/cash/admin/api/vip/entitlements` 的加载请求

#### Scenario: 未登录访问

- **WHEN** 管理员未登录即打开该静态页
- **THEN** 页面 MUST 按 admin-common 惯例引导登录或隐藏需鉴权的主内容，MUST NOT 在浏览器侧携带运维口令头

### Requirement: Admin Hub SHALL 登记 cash-vip-admin 模块

`resource/public/admin-modules.js` MUST 增加模块入口：`id: cash-vip-admin`（或等价稳定 id）、导航至 `/device/admin/cash-vip-admin.html`、`showInNav: true`。`RegisterAdminStaticPages`（或等价单点注册表）MUST 包含该 `pagePath`。PR MUST NOT 仅在主网关注册该页而 App 网关不可见。

#### Scenario: Hub 导航可见 VIP 模块

- **WHEN** 运维已登录 `/device/admin` Hub
- **THEN** 模块列表 MUST 包含 VIP 权益入口且链接至 `/device/admin/cash-vip-admin.html`

### Requirement: VIP 管理页 SHALL 展示权益与激活金额列

`cash-vip-admin.html` MUST 将 `GET /cash/admin/api/vip/entitlements` 结果以表格或等价列表展示，列至少含：`wxId`、是否仍有效、到期时间、剩余时间（过期时 MUST 展示「已过期」或等价文案，而非伪装为正剩余）、激活金额、支付渠道、最近支付时间。页面 MUST 支持分页；宜提供按 `wxId` 查询。激活金额展示 MUST 与 API 的最近一次 paid `amount_fen` 语义一致（页内须能识别单位为分或换算为元）。

#### Scenario: 列表含已过期行

- **WHEN** API 返回 `isVip=false` 且剩余秒数 ≤0 的行
- **THEN** UI MUST 仍展示该行，且剩余时间文案 MUST 表明已过期

#### Scenario: 展示最近实付金额

- **WHEN** API 返回某行 `lastPaidAmountFen` 为非零
- **THEN** UI MUST 在激活金额列展示对应该值的金额（不得展示商品现价替代）
