## ADDED Requirements

### Requirement: Apple AASA 文件 SHALL 在 `www.pangbao.cuplay.top` 主机根路径可访问
系统 SHALL 为胖宝 iOS 应用提供 Apple `apple-app-site-association` 文件，并且 MUST 同时支持 `GET /apple-app-site-association` 与 `GET /.well-known/apple-app-site-association`。两条路径返回的内容 MUST 等价、响应状态 MUST 为 `200`、传输协议 MUST 为 HTTPS，且响应不得被改写到任何其它业务路径后才可获取。

#### Scenario: Apple 从根路径获取 AASA 文件
- **WHEN** 运维或校验程序请求 `https://www.pangbao.cuplay.top/apple-app-site-association`
- **THEN** 系统返回 `200` 和可解析的 AASA JSON 内容，且不发生 301/302 到其它路径

#### Scenario: Apple 从 well-known 路径获取 AASA 文件
- **WHEN** 运维或校验程序请求 `https://www.pangbao.cuplay.top/.well-known/apple-app-site-association`
- **THEN** 系统返回与根路径等价的 AASA JSON 内容，并使用适合 JSON/AASA 的响应头

### Requirement: AASA 内容 SHALL 与微信 Universal Links 前缀保持一致
AASA 内容 MUST 使用 `appIDs = ["<TEAM_ID>.com.fzy.pangbao"]` 的结构，其中 `com.fzy.pangbao` 为固定 Bundle ID，`<TEAM_ID>` 为部署时注入的真实 Apple Team ID。AASA `components` 或等价路径约束 MUST 放行 `https://www.pangbao.cuplay.top/wx/ulink/` 对应的 `/wx/ulink/*` 路径，使微信开放平台填写值、iOS 客户端 `universalLink` 和服务端声明保持一致。

#### Scenario: Team ID 已配置时生成正式 AASA 内容
- **WHEN** 部署环境已经提供真实 Team ID
- **THEN** AASA 中的 `appIDs` 使用 `<真实TeamID>.com.fzy.pangbao`，且放行路径覆盖 `/wx/ulink/*`

#### Scenario: 微信后台使用推荐的 Universal Links 前缀
- **WHEN** 接入人员在微信开放平台填写 Universal Links
- **THEN** 文档与服务端约束均指向 `https://www.pangbao.cuplay.top/wx/ulink/`

### Requirement: Team ID 缺失时系统 SHALL 提供显式待配置语义
在 Team ID 尚未提供的阶段，仓库 MUST 保留明确的 AASA 模板或配置占位说明；正式对外端点在未配置 Team ID 时 MUST 返回显式不可验证语义或运维可识别的失败提示，而不是伪造一个看似可用的生产 `appIDs`。

#### Scenario: 生产配置缺少 Team ID
- **WHEN** AASA 端点所在环境未设置 Team ID
- **THEN** 系统返回显式错误或不可用提示，并在日志/文档中指向需要补充的配置项

#### Scenario: 仓库中保留待补位模板
- **WHEN** 开发人员阅读仓库内 Universal Links 相关资源
- **THEN** 可以看到 Team ID 待补位规则，以及 `com.fzy.pangbao` 已固定、仅 Team ID 需要在部署前补齐

### Requirement: 仓库 SHALL 提供 GitHub 打包上架的 Universal Links 操作文档
仓库 MUST 提供面向 GitHub 打包链路的 runbook，明确 iOS 工程需要开启 `Associated Domains`、加入 `applinks:www.pangbao.cuplay.top`、保证 Provisioning Profile 启用该能力，并在微信 SDK 注册配置中使用与 AASA 一致的 `https://www.pangbao.cuplay.top/wx/ulink/`。文档 MUST 说明该流程适用于 GitHub Actions / CI 打包，不要求人工在本地 Xcode 界面逐步操作才能理解。文档 MUST 同时明确 `http://www.pangbao.cuplay.top/` 不能作为 Universal Links 或 AASA 校验地址。

#### Scenario: GitHub Actions 打包配置指引可读
- **WHEN** 维护者按照 runbook 配置 GitHub 打包环境
- **THEN** 可以明确知道需要准备哪些证书/描述文件/Secrets、如何确认 entitlements 被正确签入产物

#### Scenario: 发布后可执行 Universal Links 验证
- **WHEN** 维护者完成部署与打包
- **THEN** runbook 提供 `curl`、Apple/微信侧检查项或真机验证步骤，以确认 Universal Links 已生效
