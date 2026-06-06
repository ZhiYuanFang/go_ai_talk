## Why

iOS 已上线「通过 Apple 登录」（`apple-sign-in-api`），App Store 审核与国内合规要求隐私政策与用户协议如实披露第三方登录所收集的数据。当前两份 HTML 仍仅描述微信登录，与后端实际能力（仅持久化 Apple JWT `sub`，不存邮箱/姓名）不一致。

## What Changes

- 更新 `resource/public/privacy-policy.html`：在「收集信息」中补充 Apple 登录说明（仅匿名用户标识符 `sub`；不存储 Apple 邮箱、姓名或登录凭证原文）；更新生效日期。
- 更新 `resource/public/user-agreement.html`：在「账号注册与安全」中补充 Apple 登录作为 iOS 可选登录方式；更新生效日期。
- **范围（方案 A）**：仅补充 Apple 登录，不扩展用户名/设备号登录表述；不改动 gateway 路由或 URL。

## Capabilities

### New Capabilities

- `app-legal-docs`: 定义 App 合规文档（隐私政策、用户协议）对 Apple 登录的数据收集与账号条款披露要求。

### Modified Capabilities

（无）

## Impact

- **静态资源**：`resource/public/privacy-policy.html`、`resource/public/user-agreement.html`
- **访问路径不变**：`/privacy-policy.html`、`/user-agreement.html`（gateway-app 已注册）
- **无 API / 数据库 / 配置变更**
