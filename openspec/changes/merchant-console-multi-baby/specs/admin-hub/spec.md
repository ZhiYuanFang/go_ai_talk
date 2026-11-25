# Spec Delta

## ADDED Requirements

### Requirement: Hub SHALL 提供商户订阅运维页
系统 MUST 在运维 Hub 注册商户运维静态页（或等价模块），使用 `AdminCommon.requireAdmin()` 与 `adminFetch`。页面 MUST 列出商户：标识/店名、是否曾开通、付费来源、`expire_at`、剩余或欠费状态。浏览器 MUST NOT 发送 `X-Admin-Password`。

#### Scenario: 已登录管理员打开列表
- **WHEN** 管理员已登录 Hub 并打开商户运维页
- **THEN** 页面 MUST 展示商户订阅状态列表

#### Scenario: 未登录重定向
- **WHEN** 未登录访问该页
- **THEN** MUST 按 Hub 惯例要求登录

### Requirement: Hub SHALL 支持选定商户设置绝对到期日
商户运维页 MUST 允许选择某一商户并提交 **绝对到期时间**，调用 care-commerce 授期 API。成功后 MUST 刷新该行状态。MUST 提供商户月费（默认 1000 元/月）的读取与更新入口（可同页或链到商品配置）。

#### Scenario: 授期成功
- **WHEN** 管理员为商户提交合法未来绝对到期日并保存
- **THEN** 列表 MUST 显示更新后的 `expire_at` 且非欠费

#### Scenario: 改月费
- **WHEN** 管理员将商户连续订阅月费改为非默认分值并保存
- **THEN** 后续商户读价 MUST 反映新 `price_fen`
