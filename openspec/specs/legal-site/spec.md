# legal-site

## Purpose

隐私政策、用户协议、公司站点法律页与 App 法律文档的行为规格。 Public legal/compliance copy served to clients and official site.

<!-- migrated-from: app-legal-docs, company-site-legal, privacy-policy-company, user-agreement-company -->

## Requirements

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

### Requirement: 隐私政策 SHALL 披露胖宝 AI 与模型供应商

`resource/public/privacy-policy.html` MUST 说明喂养语音 AI、胖宝诊疗与润笔等功能可能将用户输入或摘要发送至 **可配置的第三方大模型服务**（包括但不限于智谱 GLM、DeepSeek、阿里云 DashScope，以运维后台实际配置为准）。文档 MUST 在 AI 相关章节说明 **胖宝诊疗**：为生成回答，系统会读取用户近 **7 天喂养记录聚合摘要**（非完整原始记录全文），并 MAY 在 App 内展示 AI **思考过程**（thinking）供用户参考。文档 MUST 包含更新后的生效日期。

#### Scenario: 用户阅读胖宝诊疗相关说明

- **WHEN** 用户打开 `/privacy-policy.html`
- **THEN** 页面 SHALL 包含 7 天喂养摘要与 thinking 展示说明，且用户向名称 SHALL 为「胖宝诊疗」

#### Scenario: 供应商描述与可配置 provider 一致

- **WHEN** 系统默认使用智谱模型且 Admin 可切回 DeepSeek/DashScope
- **THEN** 隐私政策 SHALL 表述为可配置多供应商，且 MUST NOT 声称仅使用单一固定供应商

---

### Requirement: 官网页脚 MUST 展示公司主体与备案信息

`resource/public/pangbao-home.html` 页脚 MUST 展示运营主体中文全称「杭州彩逗科技有限公司」、英文名「Hangzhou Caidou Technology Co., Ltd.」、ICP 备案号「浙ICP备2024101366号」（可点击链至工信部备案查询站）、客服邮箱 `pangbao@cuplay.top`，以及隐私政策与用户协议链接。页脚 MUST NOT 再展示「官网事件 logo 与 Android 下载链路均来自系统权威数据」类运维旁白。

#### Scenario: 用户查看官网底部

- **WHEN** 用户打开官网首页并滚动至页脚
- **THEN** 页面 SHALL 可见公司中英文全称、浙ICP备2024101366号与客服邮箱 `pangbao@cuplay.top`，且可进入隐私政策与用户协议

### Requirement: 官网 iOS 下载 MUST 提供 App Store 直达链接

官网 iOS 下载区 MUST 提供可点击链接，目标为  
`https://apps.apple.com/cn/app/%E8%83%96%E5%AE%9D/id6774418472`（或等价已编码的 App Store 胖宝应用页）。MAY 保留「亦可在 App Store 搜索胖宝」作为辅助说明，但 MUST NOT 仅依赖「打开 App Store 首页」而无应用直达。

#### Scenario: 用户点击 iOS 下载

- **WHEN** 用户在官网点击 iOS 下载/打开 App Store 主按钮
- **THEN** 浏览器 SHALL 打开胖宝应用的 App Store 详情页（应用 id 6774418472）

### Requirement: 官网 MUST 简要说明产品由公司运营

官网 MUST 在页脚或独立短段中说明「胖宝」由杭州彩逗科技有限公司运营（文案可简短）。产品能力展示 MUST 至少覆盖喂养记录、同月龄社区、AI 陪伴，并 SHOULD 体现预测/提醒或诊疗等现网能力中的至少一项，且 MUST 保留 AI 不能替代医疗诊断的免责提示。

#### Scenario: 用户阅读产品能力

- **WHEN** 用户查看官网产品能力区块
- **THEN** 页面 SHALL 展示多类产品能力要点，并 SHALL 含 AI 免责相关提示

---

### Requirement: 隐私政策 MUST 披露运营主体与联系方式

`resource/public/privacy-policy.html` MUST 在显著位置（开篇或独立「运营者」节）说明：本应用「胖宝」由**杭州彩逗科技有限公司**（Hangzhou Caidou Technology Co., Ltd.）运营；注册地址表述 MUST 为「以市场主体公示为准」或等价；联系邮箱 MUST 含 `pangbao@cuplay.top`。文档 MUST 更新修订日期（生效日期可保持不变）。路径 MUST 仍为 `/privacy-policy.html`。

#### Scenario: 用户阅读主体信息

- **WHEN** 用户打开 `/privacy-policy.html`
- **THEN** 页面 SHALL 可见公司中文全称与客服邮箱 `pangbao@cuplay.top`，且 SHALL NOT 将运营者表述为未具名的「个人开发者」或「胖宝团队」作为唯一主体

### Requirement: 隐私政策 MUST 按现网能力披露收集与处理范围

隐私政策 MUST 在既有登录、宝宝档案、喂养记录、社区、位置（可选）、AI 与内容审核披露基础上，补充或明确至少下列现网相关项（表述可与实现细节对齐，但 MUST NOT 隐瞒已上线能力）：

- **推送通知**：为向用户发送离线提醒等，可能收集并使用设备推送标识（如 APNs / 华为 / 小米等厂商 token）； MUST NOT 声称仅使用单一固定供应商。 若 App 展示 AI 思考过程，MUST 说明该展示性质。 既有 Apple 登录「仅匿名用户标识符、不存储 Apple 邮箱/姓名/凭证原文」要求 MUST 保持。

#### Scenario: 用户查阅推送与预测说明

- **WHEN** 用户阅读「我们收集哪些信息」或等价章节
- **THEN** 文案 SHALL 提及推送标识用于提醒，以及预测相关时间点可能同步服务端

#### Scenario: 用户查阅 AI 与主体

- **WHEN** 用户阅读 AI / 第三方章节与运营者信息
- **THEN** 文案 SHALL 含多供应商可配置表述，并 SHALL 含公司主体全称

---

### Requirement: 用户协议 MUST 以公司为运营与缔约主体

`resource/public/user-agreement.html` MUST 明确本协议由用户与**杭州彩逗科技有限公司**（Hangzhou Caidou Technology Co., Ltd.）订立；「胖宝」为该公司运营的应用程序。注册地址表述 MUST 为「以市场主体公示为准」或等价。联系方式 MUST 提供客服邮箱 `pangbao@cuplay.top`（或指向与隐私政策一致的联系渠道）。路径 MUST 仍为 `/user-agreement.html`。文档 MUST 更新修订日期（生效日期可保持不变）。

#### Scenario: 用户阅读协议主体

- **WHEN** 用户打开 `/user-agreement.html`
- **THEN** 页面 SHALL 可见公司中文全称作为运营/缔约主体，客服邮箱 SHALL 为 `pangbao@cuplay.top`，且 MUST NOT 将知识产权或运营归属表述为单独的「胖宝团队」而无公司主体

### Requirement: 用户协议 MUST 覆盖知识产权、付费与 AI 免责要点

用户协议 MUST：

- 将 App UI、LOGO、软件技术等知识产权归属表述为归杭州彩逗科技有限公司所有（或该公司享有相应权利）；用户宝宝记录数据归用户所有的既有精神 MUST 保留。
- 说明付费或会员/功能开通相关服务以 App 内展示与支付渠道规则为准；依法可退款情形遵循适用法律与平台规则（表述简洁即可）。
- 说明 AI 生成内容仅供参考，不能替代专业医疗诊断；保留社区行为规范与账号安全等既有合理条款。
- 保留微信登录与 Apple 登录相关账号说明的既有正确边界。

#### Scenario: 用户阅读知识产权与 AI

- **WHEN** 用户阅读知识产权与服务说明相关章节
- **THEN** 文案 SHALL 将 IP 归属指向公司，并 SHALL 含 AI 不能替代医疗诊断的免责要点

---
