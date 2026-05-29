## ADDED Requirements

### Requirement: 按 platform 加载微信开放平台凭据

device-service SHALL 从配置 `wechat.platforms` 读取各 `platform` 键对应的 `appId` 与 `appSecret`。系统 SHALL 至少支持以下键名：`ios`、`android`、`web`。当请求中的 `platform` 在配置中不存在或 `appId`/`appSecret` 任一为空时，SHALL 返回明确配置错误且 SHALL NOT 调用微信 API。

`ios` 与 `android` SHALL 映射到**同一微信开放平台移动应用**的 `appId`/`appSecret`（部署时两键配置相同值）。`web` SHALL 映射到**微信开放平台网站应用**的独立 `appId`/`appSecret`。

生产环境 SHALL 通过环境变量或挂载配置覆盖 `appSecret`，且 SHALL NOT 将真实密钥提交到版本库。

#### Scenario: 移动应用 platform 解析凭据

- **WHEN** 登录请求 `platform` 为 `ios` 或 `android` 且对应配置项已填写有效 `appId`/`appSecret`
- **THEN** 系统 SHALL 使用该移动应用凭据调用微信 OAuth 换票 API

#### Scenario: 网站应用 platform 解析凭据

- **WHEN** 登录请求 `platform` 为 `web` 且 `wechat.platforms.web` 已填写有效 `appId`/`appSecret`
- **THEN** 系统 SHALL 使用该网站应用凭据调用微信 OAuth 换票 API

#### Scenario: 未配置的 platform

- **WHEN** 登录请求 `platform` 在 `wechat.platforms` 中不存在或凭据不完整
- **THEN** 系统 SHALL 返回明确错误且 SHALL NOT 创建或匹配 wx 行
