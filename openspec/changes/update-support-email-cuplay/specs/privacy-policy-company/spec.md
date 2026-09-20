## MODIFIED Requirements

### Requirement: 隐私政策 MUST 披露运营主体与联系方式

`resource/public/privacy-policy.html` MUST 在显著位置（开篇或独立「运营者」节）说明：本应用「胖宝」由**杭州彩逗科技有限公司**（Hangzhou Caidou Technology Co., Ltd.）运营；注册地址表述 MUST 为「以市场主体公示为准」或等价；联系邮箱 MUST 含 `pangbao@cuplay.top`。文档 MUST 更新修订日期（生效日期可保持不变）。路径 MUST 仍为 `/privacy-policy.html`。

#### Scenario: 用户阅读主体信息

- **WHEN** 用户打开 `/privacy-policy.html`
- **THEN** 页面 SHALL 可见公司中文全称与客服邮箱 `pangbao@cuplay.top`，且 SHALL NOT 将运营者表述为未具名的「个人开发者」或「胖宝团队」作为唯一主体
