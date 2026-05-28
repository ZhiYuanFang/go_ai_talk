## Why

微信开放平台在创建 iPhone 应用时要求填写 Universal Links，但当前仓库尚未提供 Apple AASA（apple-app-site-association）验证文件的托管方案，也没有针对 GitHub 打包上架链路的配置说明。现在对外域名已调整为 `https://www.pangbao.cuplay.top/`、Bundle ID 已确定为 `com.fzy.pangbao`，需要尽快把服务端托管、路径约束、Team ID 待补位策略和 GitHub 发布文档一并沉淀，避免后续 iOS 接入与审核反复返工。

## What Changes

- 在对外站点增加 iOS Universal Links 所需的 `apple-app-site-association` 托管能力，覆盖根路径与 `/.well-known/` 两个 Apple 兼容访问入口。
- 基于已确定的 Bundle ID `com.fzy.pangbao` 提供可部署的 AASA 模板，并为暂缺的 Team ID 预留明确占位与替换规则。
- 约束微信开放平台使用的 Universal Links 前缀，使其与站点真实可访问路径、AASA `components` 放行路径和 iOS 工程配置保持一致。
- 补充部署与运维文档，说明当官网基址为 `https://www.pangbao.cuplay.top/` 时，如何通过网关或边缘代理确保 AASA 仍可从 `https://www.pangbao.cuplay.top/.well-known/apple-app-site-association` 获取，并明确 `http://www.pangbao.cuplay.top/` 不能作为 Universal Links 最终配置值。
- 补充 GitHub 打包/上架链路说明，覆盖 Associated Domains、entitlements、Provisioning Profile、微信 SDK `universalLink` 配置及发布后验证步骤。

## Capabilities

### New Capabilities
- `wechat-ios-universal-links`: 为胖宝 iOS 应用提供微信开放平台所需的 Universal Links 验证文件托管、路径约束与 GitHub 发布配置说明。

### Modified Capabilities
- 无

## Impact

- Affected code：`gateway-app-server` 静态文件/路由、可能新增的 `resource/public` 或专用静态资源目录、与站点入口相关的反向代理配置说明。
- Affected docs：`docs/runbooks/` 下新增或更新 Universal Links / iOS 发布说明，明确 GitHub Actions 打包而非本地 Xcode 手工打包的操作约束。
- Affected APIs / public paths：新增公开访问的 `/.well-known/apple-app-site-association` 与 `/apple-app-site-association`；约束微信 Universal Links 前缀建议值。
- External dependencies：Apple Universal Links 校验规则、微信开放平台 iOS 应用配置、GitHub Actions 中的 iOS 签名/打包流程。
