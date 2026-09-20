## MODIFIED Requirements

### Requirement: 用户协议 MUST 以公司为运营与缔约主体

`resource/public/user-agreement.html` MUST 明确本协议由用户与**杭州彩逗科技有限公司**（Hangzhou Caidou Technology Co., Ltd.）订立；「胖宝」为该公司运营的应用程序。注册地址表述 MUST 为「以市场主体公示为准」或等价。联系方式 MUST 提供客服邮箱 `pangbao@cuplay.top`（或指向与隐私政策一致的联系渠道）。路径 MUST 仍为 `/user-agreement.html`。文档 MUST 更新修订日期（生效日期可保持不变）。

#### Scenario: 用户阅读协议主体

- **WHEN** 用户打开 `/user-agreement.html`
- **THEN** 页面 SHALL 可见公司中文全称作为运营/缔约主体，客服邮箱 SHALL 为 `pangbao@cuplay.top`，且 MUST NOT 将知识产权或运营归属表述为单独的「胖宝团队」而无公司主体
