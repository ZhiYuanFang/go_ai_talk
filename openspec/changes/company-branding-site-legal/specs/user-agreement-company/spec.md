## ADDED Requirements

### Requirement: 用户协议 MUST 以公司为运营与缔约主体

`resource/public/user-agreement.html` MUST 明确本协议由用户与**杭州彩逗科技有限公司**（Hangzhou Caidou Technology Co., Ltd.）订立；「胖宝」为该公司运营的应用程序。注册地址表述 MUST 为「以市场主体公示为准」或等价。联系方式 MUST 提供客服邮箱 `tou_zy@foxmail.com`（或指向与隐私政策一致的联系渠道）。路径 MUST 仍为 `/user-agreement.html`。文档 MUST 更新生效或修订日期。

#### Scenario: 用户阅读协议主体

- **WHEN** 用户打开 `/user-agreement.html`
- **THEN** 页面 SHALL 可见公司中文全称作为运营/缔约主体，且 MUST NOT 将知识产权或运营归属表述为单独的「胖宝团队」而无公司主体

### Requirement: 用户协议 MUST 覆盖知识产权、付费与 AI 免责要点

用户协议 MUST：

- 将 App UI、LOGO、软件技术等知识产权归属表述为归杭州彩逗科技有限公司所有（或该公司享有相应权利）；用户宝宝记录数据归用户所有的既有精神 MUST 保留。
- 说明付费或会员/功能开通相关服务以 App 内展示与支付渠道规则为准；依法可退款情形遵循适用法律与平台规则（表述简洁即可）。
- 说明 AI 生成内容仅供参考，不能替代专业医疗诊断；保留社区行为规范与账号安全等既有合理条款。
- 保留微信登录与 Apple 登录相关账号说明的既有正确边界。

#### Scenario: 用户阅读知识产权与 AI

- **WHEN** 用户阅读知识产权与服务说明相关章节
- **THEN** 文案 SHALL 将 IP 归属指向公司，并 SHALL 含 AI 不能替代医疗诊断的免责要点
