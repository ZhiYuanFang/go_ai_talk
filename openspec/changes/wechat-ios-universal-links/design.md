## Context

当前对外官网由 `gateway-app-server` 承载，公网入口现已切到 `https://www.pangbao.cuplay.top/` 根路径；但 Apple Universal Links 的 AASA 文件校验无论站点前缀如何变化，都必须直接访问主机根路径 `https://www.pangbao.cuplay.top/apple-app-site-association` 或 `https://www.pangbao.cuplay.top/.well-known/apple-app-site-association`。因此当前实现应直接围绕根路径站点与 `/wx/ulink/` 业务前缀组织，保证 HTTPS、无重定向、正确 `Content-Type` 与可重复验证，并显式拒绝把 HTTP 入口作为 Universal Links 配置源。

本次变更同时存在两项现实约束：

1. Bundle ID 已确定为 `com.fzy.pangbao`，但 Team ID 尚未提供，因此仓库不能提交一个“看起来正式但实际不可验证”的硬编码生产值。
2. iOS 应用不是人工在本地 Xcode 打包上架，而是通过 GitHub 执行构建与发布，因此文档必须覆盖证书、描述文件、entitlements、微信 SDK `universalLink` 与 GitHub Actions 环境变量/密钥的配合方式。

## Goals / Non-Goals

**Goals:**
- 让服务端具备可托管 Apple AASA 文件的能力，满足微信开放平台对 iPhone 应用 Universal Links 的前置要求。
- 在 Team ID 暂缺时仍能落地变更：仓库内保留模板与配置接口，部署时补齐 Team ID 后无需再次改代码。
- 为 `https://www.pangbao.cuplay.top/` 根路径站点明确 Universal Links 路径规范，避免微信后台、AASA `components`、iOS entitlements 和代码中的 `universalLink` 各写各的。
- 输出面向 GitHub 打包发布的操作文档，覆盖 Associated Domains、签名配置、微信开放平台填写与线上验证步骤。

**Non-Goals:**
- 不在本仓库中实现 iOS 客户端工程或新增 IPA 上传/分发能力。
- 不自动代管 Apple Developer 账号、证书或描述文件，只提供需要的配置说明与仓库侧接入点。
- 不将 Team ID 写死在仓库提交中；若未提供 Team ID，线上验证不视为完成。

## Decisions

### 1. Universal Links 主机使用 `www.pangbao.cuplay.top`，业务路径前缀使用 `/wx/ulink/`
**Decision**：Associated Domains 以 `applinks:www.pangbao.cuplay.top` 为准；微信开放平台与 iOS 客户端中的 Universal Links 基址统一使用 `https://www.pangbao.cuplay.top/wx/ulink/`；AASA `components` 只放行 `/wx/ulink/*`。

**Why**：Universal Links 的域名维度只看 host，不看路径前缀；而当前公网入口已确定为 `https://www.pangbao.cuplay.top/`。将业务前缀固定在 `/wx/ulink/` 可以直接匹配根路径站点部署，减少反向代理改写与额外前缀的认知成本。即使业务上临时暴露了 HTTP，Apple 与微信校验仍只接受 HTTPS。

**Alternatives considered:**
- 保持旧版 `/app/wx/ulink/`：与当前根路径公网入口不一致，会让文档、AASA 与真实部署脱节。
- 为 iOS 单独申请子域名：更清晰，但当前用户已给定正式域名，新增子域会扩大证书与部署面。

### 2. AASA 同时暴露在根路径与 `/.well-known/`，且由仓库内能力统一生成/托管
**Decision**：实现应同时支持 `GET /apple-app-site-association` 与 `GET /.well-known/apple-app-site-association`，两者输出一致；当前公网站点既然已直接挂在根路径，边缘代理/Nginx 应直接把根路径流量转发到 `gateway-app-server`，无需再保留 `/app/` 前缀假设。

**Why**：Apple 文档接受两个位置，双入口能提升兼容性；而 AASA 本来就必须在 host 根路径可读，因此直接以根路径站点组织部署更符合当前实际环境。

**Alternatives considered:**
- 仅提供 `/.well-known/`：理论可行，但现场排障时不如双入口直观。
- 将 AASA 文件放到 `/app/.well-known/`：Apple 不会读取，无法满足要求。

### 3. Team ID 通过部署配置注入，缺失时返回显式不可用语义
**Decision**：实现阶段应通过独立配置项或环境变量（如 `GATEWAY_APP_IOS_TEAM_ID`）注入 Team ID；当该值为空时，AASA 端点不得伪造生产内容，而应返回显式错误或仅在仓库保留模板文件，文档中要求部署前补齐。

**Why**：Team ID 属于环境相关信息，当前用户也明确表示稍后补充。把它做成部署配置既能让本次变更继续推进，又能避免把错误值固化进版本库或让 Apple 抓到无效 `appIDs`。

**Alternatives considered:**
- 直接提交占位值到正式端点：最省事，但会导致微信/Apple 校验失败且不易排查。
- 等 Team ID 到位后再做全部改动：会阻塞当前服务端、文档与路径方案准备。

### 4. 仓库内同时提供“机器可用端点”和“人工可读 runbook”
**Decision**：变更应包含两部分输出：一是实际对外可访问的 AASA 文件托管能力；二是新增/更新 runbook，明确微信开放平台、Apple Associated Domains、GitHub Actions 打包签名、entitlements 与验证命令。

**Why**：本仓库只能覆盖服务端与文档，无法直接修改外部 iOS 私有工程；因此必须把跨系统依赖清楚写入 runbook，降低后续接入遗漏。

**Alternatives considered:**
- 只交付 AASA 文件，不写文档：很容易在 GitHub 打包链路里漏掉 entitlements 或 profile 能力开关。
- 只写文档，不做端点：无法满足微信开放平台实际校验。

## Risks / Trade-offs

- **[Risk] 文档继续保留旧 `/app/` 前缀导致配置错填** → 将 runbook、AASA 路径规则与服务端实现统一切到根路径版本。
- **[Risk] Team ID 缺失导致变更上线后仍不可验证** → 端点在缺失配置时返回显式不可用语义，并在 runbook 中将 Team ID 补齐列为发布前门槛。
- **[Risk] 反向代理把 AASA 请求改写到其它路径或添加下载头** → 在任务中明确校验状态码、`Content-Type`、无跳转与 HTTPS 直出。
- **[Risk] GitHub Actions 成功打包但 Associated Domains 未进入最终签名产物** → runbook 中加入 entitlements、Provisioning Profile 能力和归档后二次验签检查步骤。

## Migration Plan

1. 在仓库中新增 Universal Links 相关 capability、AASA 模板/端点方案与 runbook。
2. 在 `gateway-app-server` 或边缘代理配置中打通根路径的两个 AASA 入口，使其落到统一内容源。
3. 部署前补齐 Team ID，并将微信开放平台的 Universal Links 设置为 `https://www.pangbao.cuplay.top/wx/ulink/`。
4. 在 iOS GitHub 打包链路中启用 `Associated Domains` 与 `applinks:www.pangbao.cuplay.top`，同步更新微信 SDK `universalLink` 常量/配置。
5. 用 `curl` 与真机/测试包进行联调验证；若需回滚，可先撤回微信开放平台配置并移除 AASA 暴露，官网主体仍可独立保留。

## Open Questions

- Team ID 的最终值是什么，以及是否计划通过环境变量还是私有配置文件注入。
- 当前公网入口 `https://www.pangbao.cuplay.top/` 前面是否还有 Nginx/CDN，需要在哪一层放通 `/.well-known/apple-app-site-association`，并确保外部访问不会降级成 HTTP。
- iOS 私有仓库/工作流当前使用的是 Fastlane、原生 `xcodebuild` 还是其它封装脚本，runbook 是否需要补具体示例。