## ADDED Requirements

### Requirement: 官网展示三大产品能力柱
官网页面 SHALL 在 Hero 与喂养事件区块之间展示独立的「产品能力」区块（锚点 `#capabilities`），以玻璃拟态卡片呈现三大能力柱：**智能喂养记录**、**同月龄轻社交**、**AI 智能陪伴**。每柱 MUST 至少包含标题与 2 条场景化说明；移动端 MUST 可正常阅读（单列或等价布局）。

#### Scenario: 用户浏览产品能力区块
- **WHEN** 用户打开官网首页并滚动至 `#capabilities`
- **THEN** 页面 SHALL 展示上述三柱，且文案 MUST 涵盖语音记事件、同阶段妈妈社区（发帖/关注/私信）、AI 语音助手与智能喂养问答等等价能力描述

#### Scenario: 导航可跳转至产品能力
- **WHEN** 用户点击导航中的「产品能力」或等价链接
- **THEN** 页面 SHALL 滚动至 `#capabilities` 锚点

### Requirement: 官网面向 0-18 月新手妈妈定位
官网首屏（Hero）SHALL 明确面向 **0–18 月新手妈妈** 或等价表述（如 kicker 含「0–18 月」）。Hero MUST 同时传达「喂养记录」「同月龄连接」「AI 问答」三类价值，而非仅喂养工具定位。

#### Scenario: 首屏展示阶段化定位
- **WHEN** 用户打开官网首页
- **THEN** Hero 区域 SHALL 包含 0–18 月或新手妈妈阶段化定位文案，且副标题或 metric 卡片 MUST 提及社交与 AI 能力

#### Scenario: SEO meta 覆盖核心关键词
- **WHEN** 爬虫或浏览器读取页面 `<head>`
- **THEN** `<title>` 与 `<meta name="description">` MUST 包含喂养记录相关表述，并 SHOULD 包含 0–18 月、AI、同月龄或社区等关键词中的至少两项

### Requirement: 官网 AI 能力合规表述
官网展示 AI 能力时 SHALL 使用「智能喂养问答」或等价非医疗措辞，MUST NOT 单独使用「诊疗」作为官网主宣传语。页面 MUST 包含简短免责声明，说明 AI 回答仅供参考、不能替代医生诊断。

#### Scenario: AI 能力柱不含诊疗主宣传
- **WHEN** 用户阅读 `#capabilities` 中 AI 相关文案
- **THEN** 页面 SHALL NOT 以「胖宝诊疗」作为唯一或主标题对外展示，SHALL 使用智能问答/AI 助手等表述

#### Scenario: 展示 AI 免责声明
- **WHEN** 用户浏览含 AI 描述的区块或页脚
- **THEN** 页面 SHALL 展示等价于「AI 回答仅供参考，不能替代医生诊断」的免责声明

### Requirement: 官网 Footer 法律页链接
官网页脚 SHALL 提供指向隐私政策与用户协议的可见链接，分别对应 `/privacy-policy.html` 与 `/user-agreement.html`（或 gateway-app 注册的等价路由）。

#### Scenario: 页脚可访问法律文档
- **WHEN** 用户查看官网页脚
- **THEN** 页面 SHALL 展示可点击的「隐私政策」与「用户协议」链接，且链接 MUST 在同源 gateway-app 域名下可访问

## MODIFIED Requirements

### Requirement: 官网展示母婴喂养定位与事件卡片
官网页面 SHALL 以玻璃拟态风格展示品牌定位文案，并 SHALL 展示从数据库权威链路读取的事件列表。每个事件项 MUST 至少包含事件名与事件 logo；若 logo 为 path-only 资源，前端或聚合接口 MUST 能将其解析为当前站点可访问的同源地址。除喂养事件卡片外，页面 SHALL 包含独立的产品能力区块（见「官网展示三大产品能力柱」），首屏定位 MUST 覆盖 0–18 月新手妈妈的喂养、轻社交与 AI 价值。

#### Scenario: 官网首屏展示品牌定位
- **WHEN** 用户打开官网首页
- **THEN** 页面 SHALL 明确表达面向 0–18 月新手妈妈（或等价阶段化表述），以及「喂养记录更便捷、同月龄轻社交、AI 智能陪伴」等核心信息，而不仅限于单一喂养工具定位

#### Scenario: 官网展示事件 logo 与事件名
- **WHEN** 官网聚合到至少一条事件数据
- **THEN** 页面 SHALL 为每条事件渲染事件 logo 与事件名，且 logo 地址 MUST 可被当前官网域名直接访问

#### Scenario: 喂养事件区说明对齐小月龄场景
- **WHEN** 用户查看 `#events` 喂养事件区块
- **THEN** 区块说明文案 SHOULD 提及小月龄高频照护场景（如吃奶、睡觉、换尿布）及语音记录等等价表述

### Requirement: 官网提供匿名只读聚合数据接口
系统 SHALL 提供一个适用于官网匿名访问的只读聚合接口，由 `gateway-app-server` 统一返回官网所需的事件展示数据、Android 下载信息与 iOS 下载说明。该接口 MUST 通过服务契约或本进程已有能力获取数据，MUST NOT 让前端直接调用受保护业务接口或跨服务直连数据库。接口返回的 `heroTitle`、`heroSubtitle`、`serviceSummary` MUST 与官网 0–18 月新手妈妈及「喂养 + 轻社交 + AI」三角定位一致。

#### Scenario: 匿名读取官网数据
- **WHEN** 未登录用户请求官网聚合接口
- **THEN** 系统 SHALL 返回成功响应，其中包含事件列表、Android 下载展示信息、iOS 下载说明，以及已更新语义的 Hero 文案字段

#### Scenario: 官网数据来源遵守服务边界
- **WHEN** `gateway-app-server` 组装官网响应
- **THEN** 系统 MUST 通过现有服务契约读取事件权威数据，并复用本进程版本信息读取能力，MUST NOT 新增跨服务直连他域库表行为
