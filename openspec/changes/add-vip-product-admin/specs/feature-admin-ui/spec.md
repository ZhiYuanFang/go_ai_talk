## ADDED Requirements

### Requirement: 开通功能管理页 MUST 提供 VIP列表入口

`cash-feature-admin.html` MUST 提供按钮或等价控件，可见文案 MUST 为「VIP列表」，激活后 MUST 导航至 `/device/admin/cash-vip-admin.html`。该入口 MUST 替代 Hub 主导航上的「VIP 权益」模块作为常规运维路径。

#### Scenario: 按钮文案与跳转

- **WHEN** 已登录管理员打开开通功能管理页
- **THEN** 页面 MUST 展示文案为「VIP列表」的入口，激活后 MUST 进入 VIP 权益只读页
