## ADDED Requirements

### Requirement: 官网实体音箱区块 MUST 内嵌 promo 能力演示且不展示 tutorial

官网 `#xiaozhi`（或等价实体音箱连接能力区块，位置可保持现网中后段）MUST 在已配置 `promo` 槽时，将能力演示视频以页面内嵌 `<video>`（或等价）直接展示，MUST 随视口宽度缩放且画面整段可见（如 `width:100%` + `object-fit:contain`），MUST NOT 将 `promo` 藏在点击按钮后的弹层中才能观看。官网 MUST NOT 展示 `tutorial`（绑定教学）槽：即便匿名 `guide-videos` 列表含该槽，前端 MUST 忽略不渲染。未配置 `promo` 时 MUST 展示简短「准备中」或等价提示，MUST NOT 留下失效按钮。用户面文案 MUST 使用「实体音箱连接能力」或等价通俗表述，MUST NOT 以「小智 MCP」等难懂术语作为该区块主标题。

#### Scenario: 已配置 promo 时内嵌可见

- **WHEN** 匿名 guide-videos 返回含 `slot=promo` 且 `videoCdnUrl` 非空的项
- **THEN** 官网音箱区块 SHALL 直接内嵌可播放该 URL 的视频控件，且用户调整窗口宽度时视频画面 SHALL 保持完整可见（不得仅露出局部需横向裁切才能看懂）

#### Scenario: tutorial 不在官网出现

- **WHEN** guide-videos 同时返回 `promo` 与 `tutorial`
- **THEN** 官网 SHALL 仅展示 `promo` 内嵌视频，SHALL NOT 渲染绑定教学按钮、弹层入口或第二条内嵌 tutorial

#### Scenario: 文案为实体音箱连接能力

- **WHEN** 用户阅读官网导航、Hero 或音箱能力区块
- **THEN** 用户可见主表述 SHALL 为实体音箱连接能力或等价通俗说法，而非以「小智 MCP」为主标题

## MODIFIED Requirements

### Requirement: 官网展示三大产品能力柱

官网页面 SHALL 按现网优先级展示产品能力叙事，覆盖至少：**智能预测与提醒**、**宝宝信息记录（含事件可视化）**、**语音 AI 问答**、**实体音箱连接能力**（原「小智音箱」用户面叫法）。文案 MUST NOT 以同月龄社区发帖/关注/私信或「胖宝诊疗」作为主柱宣传；MUST NOT 出现付费、会员、邀请开通等营销表述。移动端 MUST 可正常阅读。实体音箱区块的视频呈现 MUST 遵守「内嵌 promo、不展示 tutorial」要求。

#### Scenario: 用户浏览产品能力

- **WHEN** 用户打开官网首页并浏览能力相关区块
- **THEN** 页面 SHALL 展示预测提醒、宝宝信息记录、语音 AI、实体音箱连接等相关说明，且 MUST NOT 将 UGC 社区或诊疗作为现行主能力宣传

#### Scenario: 官网禁止付费营销文案

- **WHEN** 用户阅读官网全文（含 Hero、能力区、弹层、页脚）
- **THEN** 页面 MUST NOT 出现付费开通、会员订阅、邀请码开通等营销文案

#### Scenario: 用户浏览产品能力区块
- **WHEN** 用户打开官网首页并滚动至 `#capabilities`
- **THEN** 页面 SHALL 展示上述三柱，且文案 MUST 涵盖语音记事件、同阶段妈妈社区（发帖/关注/私信）、AI 语音助手与智能喂养问答等等价能力描述

#### Scenario: 导航可跳转至产品能力
- **WHEN** 用户点击导航中的「产品能力」或等价链接
- **THEN** 页面 SHALL 滚动至 `#capabilities` 锚点

