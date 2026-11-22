## MODIFIED Requirements

### Requirement: 官网展示母婴喂养定位与事件卡片

官网页面 SHALL 以浅色克制高档视觉（非粉紫玻璃英雄风）展示品牌定位，并 SHALL 展示从数据库权威链路读取的事件列表（用于「宝宝信息记录」叠放展示等）。每个事件项 MUST 至少包含事件名与事件 logo；若 logo 为 path-only，MUST 可解析为当前站点可访问地址。页面能力叙事 MUST 对齐现网：**智能预测与提醒、宝宝信息记录与统计可视化、语音 AI 问答、小智音箱**；MUST NOT 以同月龄轻社交/UGC 社区、胖宝诊疗、「值得留意」等已关闭或退役能力作为主宣传。

#### Scenario: 官网首屏展示品牌定位

- **WHEN** 用户打开官网首页
- **THEN** 页面 SHALL 明确表达面向 0–18 月新手妈妈（或等价阶段化表述），以及记录、预测提醒、语音 AI、小智等现网价值，而不仅限于单一喂养工具或轻社交定位

#### Scenario: 官网展示事件 logo 与事件名

- **WHEN** 官网聚合到至少一条事件数据
- **THEN** 页面 SHALL 能将事件 logo 与事件名用于宝宝信息记录相关展示，且 logo 地址 MUST 可被当前官网域名直接访问

#### Scenario: 喂养事件区说明对齐小月龄场景
- **WHEN** 用户查看 `#events` 喂养事件区块
- **THEN** 区块说明文案 SHOULD 提及小月龄高频照护场景（如吃奶、睡觉、换尿布）及语音记录等等价表述

### Requirement: 官网提供匿名只读聚合数据接口

系统 SHALL 提供适用于官网匿名访问的只读聚合接口，由 `gateway-app-server` 统一返回官网所需的事件展示数据、Android 下载信息与 iOS 下载说明。该接口 MUST 通过服务契约或本进程已有能力获取数据，MUST NOT 让前端直接调用受保护业务接口或跨服务直连数据库。接口返回的 Hero 类文案字段若仍含旧社交叙事，前端展示 MUST 以现网文案为准（MAY 同步清洗后端默认串）。

#### Scenario: 匿名读取官网数据

- **WHEN** 未登录用户请求官网聚合接口
- **THEN** 系统 SHALL 返回成功响应，其中包含事件列表、Android 下载展示信息与 iOS 下载说明

#### Scenario: 官网数据来源遵守服务边界

- **WHEN** `gateway-app-server` 组装官网响应
- **THEN** 系统 MUST 通过现有服务契约读取事件权威数据，并复用本进程版本信息读取能力，MUST NOT 新增跨服务直连他域库表行为

### Requirement: 官网展示 Android 下载二维码与 iOS 指引

官网 SHALL 提供应用下载能力：Android MUST 基于最新下载链接生成二维码并提供可点击入口；iOS MUST 提示前往 App Store 搜索或直达「胖宝」。下载 UI MUST 以**弹层/模态**为主入口（可由导航或 Hero CTA 打开），MUST NOT 强制占据首屏大段双列下载区。当最新行为 OSS objectKey 时，聚合接口 MUST 返回 CDN 绝对 URL；历史 path-only MUST 仍拼出可访问绝对地址。

#### Scenario: Android 存在可下载版本（OSS）

- **WHEN** 版本表存在最新 Android 下载记录且 `download_url` 为 `apk/` objectKey
- **THEN** 官网聚合接口 SHALL 返回 CDN 绝对 Android 下载地址，用户打开下载弹层后页面 SHALL 生成对应二维码并展示下载入口

#### Scenario: iOS 下载说明固定展示

- **WHEN** 用户打开下载弹层
- **THEN** 页面 SHALL 展示前往 App Store 获取「胖宝」的指引或直达链接

#### Scenario: Android 存在可下载版本（历史 path）

- **WHEN** 版本表存在最新 Android 下载记录且 `download_url` 为 `/device/app/apk/` 路径且本地文件可访问
- **THEN** 官网聚合接口 SHALL 返回官网可直接使用的绝对下载地址（基于 `publicBaseUrl` 或请求 Host），页面 SHALL 生成对应二维码并展示下载入口

#### Scenario: Android 暂无可下载版本

- **WHEN** 版本表没有可用的 Android 下载记录
- **THEN** 页面 SHALL 不展示失效二维码，并 SHALL 展示明确的“Android 下载暂未开放”或等价提示

### Requirement: 官网展示三大产品能力柱

官网页面 SHALL 按现网优先级展示产品能力叙事（锚点可保留 `#capabilities` 或拆分为多节），覆盖至少：**智能预测与提醒**、**宝宝信息记录（含事件可视化）**、**语音 AI 问答**、**小智音箱**。文案 MUST NOT 以同月龄社区发帖/关注/私信或「胖宝诊疗」作为主柱宣传；MUST NOT 出现付费、会员、邀请开通等营销表述。移动端 MUST 可正常阅读。

#### Scenario: 用户浏览产品能力

- **WHEN** 用户打开官网首页并浏览能力相关区块
- **THEN** 页面 SHALL 展示预测提醒、宝宝信息记录、语音 AI、小智等相关说明，且 MUST NOT 将 UGC 社区或诊疗作为现行主能力宣传

#### Scenario: 官网禁止付费营销文案

- **WHEN** 用户阅读官网全文（含 Hero、能力区、弹层、页脚）
- **THEN** 页面 MUST NOT 出现付费开通、会员订阅、邀请码开通等营销文案

#### Scenario: 用户浏览产品能力区块
- **WHEN** 用户打开官网首页并滚动至 `#capabilities`
- **THEN** 页面 SHALL 展示上述三柱，且文案 MUST 涵盖语音记事件、同阶段妈妈社区（发帖/关注/私信）、AI 语音助手与智能喂养问答等等价能力描述

#### Scenario: 导航可跳转至产品能力
- **WHEN** 用户点击导航中的「产品能力」或等价链接
- **THEN** 页面 SHALL 滚动至 `#capabilities` 锚点

### Requirement: 官网面向 0-18 月新手妈妈定位

官网首屏（Hero）SHALL 明确面向 **0–18 月新手妈妈** 或等价表述。Hero MUST 传达记录与智能陪伴价值（预测/语音 AI/小智等），而非仅喂养工具或轻社交定位。顶栏 MUST NOT 在滚动时 sticky 置顶。

#### Scenario: 首屏展示阶段化定位

- **WHEN** 用户打开官网首页
- **THEN** Hero 区域 SHALL 包含 0–18 月或新手妈妈阶段化定位文案

#### Scenario: SEO meta 覆盖核心关键词

- **WHEN** 爬虫或浏览器读取页面 `<head>`
- **THEN** `<title>` 与 `<meta name="description">` MUST 包含喂养/记录相关表述，并 SHOULD 包含预测、AI、小智等现网关键词中的至少一项；MUST NOT 以社区/诊疗为唯一卖点

### Requirement: 官网 AI 能力合规表述

官网展示 AI 能力时 SHALL 使用智能问答/语音助手等非医疗主宣传措辞，MUST NOT 单独使用「诊疗」作为官网主宣传语。页面 MUST 包含简短免责声明：AI 回答仅供参考、不能替代医生诊断。

#### Scenario: AI 能力不含诊疗主宣传

- **WHEN** 用户阅读官网 AI 相关文案
- **THEN** 页面 SHALL NOT 以「胖宝诊疗」作为唯一或主标题对外展示

#### Scenario: 展示 AI 免责声明

- **WHEN** 用户浏览含 AI 描述的区块或页脚附近
- **THEN** 页面 SHALL 展示等价于「AI 回答仅供参考，不能替代医生诊断」的免责声明

#### Scenario: AI 能力柱不含诊疗主宣传
- **WHEN** 用户阅读 `#capabilities` 中 AI 相关文案
- **THEN** 页面 SHALL NOT 以「胖宝诊疗」作为唯一或主标题对外展示，SHALL 使用智能问答/AI 助手等表述

### Requirement: 官网 Footer 法律页链接

官网页脚 SHALL 提供指向隐私政策与用户协议的可见链接（`/privacy-policy.html`、`/user-agreement.html`），并 SHALL 展示公司中英文全称、客服邮箱 `pangbao@cuplay.top`、ICP 备案号「浙ICP备2024101366号」（可链至工信部）。页脚布局 MUST 相对紧凑；MUST NOT 展示运维旁白类文案。

#### Scenario: 页脚可访问法律文档与主体信息

- **WHEN** 用户查看官网页脚
- **THEN** 页面 SHALL 可见公司主体、ICP、邮箱以及可点击的隐私政策与用户协议链接

#### Scenario: 页脚可访问法律文档
- **WHEN** 用户查看官网页脚
- **THEN** 页面 SHALL 展示可点击的「隐私政策」与「用户协议」链接，且链接 MUST 在同源 gateway-app 域名下可访问

---

## ADDED Requirements

### Requirement: 官网 SHALL 匿名展示小智引导视频

官网在小智相关区块 MUST 通过匿名只读接口拉取已配置的小智引导视频列表（`promo`/`tutorial` 中已有视频者），并按返回的 `title` 与 `videoCdnUrl` 提供可播放入口。当列表为空时，MUST 隐藏视频入口（MUST NOT 展示空壳「暂无视频」主块）。

#### Scenario: 有配置则可播放

- **WHEN** 公开列表返回至少一项含 `videoCdnUrl` 的槽
- **THEN** 官网小智区块 SHALL 展示对应标题入口且用户可打开该 URL 播放

#### Scenario: 双空隐藏

- **WHEN** 公开列表为空
- **THEN** 官网 MUST 不渲染引导视频入口块
