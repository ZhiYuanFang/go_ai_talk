## ADDED Requirements

### Requirement: 隐私政策 SHALL 披露 Apple 登录数据收集边界

`resource/public/privacy-policy.html` MUST 在「我们收集哪些信息」中说明：使用 Apple 登录时，为建立与识别账户，系统收集并存储 Apple 提供的匿名用户标识符；系统 SHALL NOT 在文案中声称存储 Apple 邮箱、姓名或登录凭证原文。文档 MUST 包含更新后的生效日期。

#### Scenario: 用户阅读隐私政策中的 Apple 登录说明

- **WHEN** 用户通过 App WebView 或浏览器打开 `/privacy-policy.html`
- **THEN** 页面 SHALL 包含 Apple 登录相关收集说明
- **AND** 说明 SHALL 与后端仅持久化 `apple_sub` 的行为一致

### Requirement: 用户协议 SHALL 披露 Apple 登录账号方式

`resource/public/user-agreement.html` MUST 在「账号注册与安全」中说明：iOS 用户可选用「通过 Apple 登录」建立账户。文档 MUST 包含更新后的生效日期。

#### Scenario: 用户阅读用户协议中的 Apple 登录说明

- **WHEN** 用户通过 App WebView 或浏览器打开 `/user-agreement.html`
- **THEN** 页面 SHALL 包含 Apple 登录作为可选登录方式的说明

### Requirement: 合规文档 URL SHALL 保持不变

gateway-app 暴露的合规文档路径 MUST 仍为 `/privacy-policy.html` 与 `/user-agreement.html`，客户端无需因本次修订修改加载 URL。

#### Scenario: 客户端加载合规文档

- **WHEN** 客户端请求既有隐私政策或用户协议 URL
- **THEN** gateway SHALL 返回更新后的 HTML 内容
- **AND** 路径 SHALL 与修订前相同
