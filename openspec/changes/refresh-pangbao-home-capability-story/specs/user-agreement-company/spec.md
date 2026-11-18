## MODIFIED Requirements

### Requirement: 用户协议 MUST 以公司为运营与缔约主体

`resource/public/user-agreement.html` MUST 明确本协议由用户与**杭州彩逗科技有限公司**（Hangzhou Caidou Technology Co., Ltd.）订立；「胖宝」为该公司运营的应用程序。联系方式 MUST 提供客服邮箱 `pangbao@cuplay.top`。路径 MUST 仍为 `/user-agreement.html`。文档 MUST 更新修订日期（生效日期可保持不变）。页面 MUST NOT 要求或展示公司注册地址（含「以市场主体公示为准」或具体地址文案）。

#### Scenario: 用户阅读协议主体

- **WHEN** 用户打开 `/user-agreement.html`
- **THEN** 页面 SHALL 可见公司中文全称作为运营/缔约主体，客服邮箱 SHALL 为 `pangbao@cuplay.top`，且 MUST NOT 将知识产权或运营归属表述为单独的「胖宝团队」而无公司主体
- **AND** 页面 MUST NOT 出现注册地址相关表述

### Requirement: 用户协议 MUST 覆盖知识产权、付费与 AI 免责要点

用户协议 MUST：

- 将 App UI、LOGO、软件技术等知识产权归属表述为归杭州彩逗科技有限公司所有（或该公司享有相应权利）；用户宝宝记录数据归用户所有的既有精神 MUST 保留。
- 说明部分功能可能需付费或开通（以 App 内展示与支付渠道规则为准；退款遵循适用法律与平台规则）——表述须简洁；**官网营销文案不在本要求范围内**。
- 说明 AI 生成内容仅供参考，不能替代专业医疗诊断。
- 保留微信登录与 Apple 登录相关账号说明的正确边界：微信为账号标识（unionid），MUST 明确不收集微信头像、昵称；Apple 仅匿名用户标识符。
- MUST NOT 将 UGC 社区（发帖、评论、点赞、关注、私信等）写成现行服务内容或单独社区行为规范章节；MUST NOT 保留「社区发布内容授权」类现行条款。

服务内容描述 MUST 对齐现网能力（如宝宝/喂养记录、统计可视化、智能预测与提醒、语音 AI、小智音箱等），MUST NOT 以已关闭社区或已退役诊疗等为主列举。

#### Scenario: 用户阅读知识产权与 AI

- **WHEN** 用户阅读知识产权与服务说明相关章节
- **THEN** 文案 SHALL 将 IP 归属指向公司，并 SHALL 含 AI 不能替代医疗诊断的免责要点

#### Scenario: 用户阅读账号与社区残留

- **WHEN** 用户阅读账号与服务内容章节
- **THEN** 文案 SHALL 说明微信不收集头像/昵称，且 MUST NOT 将 UGC 社区列为现行服务内容或独立行为规范章

#### Scenario: 用户阅读付费要点

- **WHEN** 用户阅读付费或功能开通相关段落
- **THEN** 文案 SHALL 含以 App 内展示与支付渠道规则为准的极简说明

## ADDED Requirements

### Requirement: 用户协议视觉 MUST 与官网同系

`resource/public/user-agreement.html` 的视觉 MUST 与同期 `pangbao-home.html` 高档浅色风格对齐。顶栏 MUST NOT sticky 置顶；品牌链接 MUST 使用 `a.brand` 且 MUST NOT 显示下划线。

#### Scenario: 协议页顶栏不置顶

- **WHEN** 用户在用户协议页向下滚动
- **THEN** 顶栏 SHALL 随文档滚走，MUST NOT 固定吸附在视口顶部

#### Scenario: 品牌链无下划线

- **WHEN** 用户查看用户协议顶栏品牌
- **THEN** `a.brand` SHALL 无下划线
