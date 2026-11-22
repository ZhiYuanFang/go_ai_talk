## MODIFIED Requirements

### Requirement: 隐私政策 MUST 披露运营主体与联系方式

`resource/public/privacy-policy.html` MUST 在显著位置（开篇或独立「运营者」节）说明：本应用「胖宝」由**杭州彩逗科技有限公司**（Hangzhou Caidou Technology Co., Ltd.）运营；联系邮箱 MUST 含 `pangbao@cuplay.top`。文档 MUST 更新修订日期（生效日期可保持不变）。路径 MUST 仍为 `/privacy-policy.html`。页面 MUST NOT 要求或展示公司注册地址（含「以市场主体公示为准」或具体地址文案）。

#### Scenario: 用户阅读主体信息

- **WHEN** 用户打开 `/privacy-policy.html`
- **THEN** 页面 SHALL 可见公司中文全称与客服邮箱 `pangbao@cuplay.top`，且 SHALL NOT 将运营者表述为未具名的「个人开发者」或「胖宝团队」作为唯一主体
- **AND** 页面 MUST NOT 出现注册地址相关表述

### Requirement: 隐私政策 MUST 按现网能力披露收集与处理范围

隐私政策 MUST 按**客户端现网可触达能力**披露收集与处理范围（以 Flutter 产品开关与已退役能力为准），MUST NOT 将已对用户隐藏的 UCG 社区入口或已退役 VIP/值得留意/成长轨迹/诊疗等写成现行收集项。 至少 MUST 覆盖下列现网相关项（表述可与实现细节对齐）：

- **微信登录**：为登录建立账号，系统使用微信授权获得的账号标识（unionid）； MUST 明确**不收集**微信头像、昵称。 MUST NOT 声称存储 Apple 邮箱、姓名或凭证原文（既有要求保持）。 MUST NOT 将已退役的「VIP」作为现行产品名披露。 MUST 说明非广告跟踪。 MUST NOT 声称仅使用单一固定供应商。 MUST NOT 将已下线/对用户隐藏的社区润笔、胖宝诊疗、值得留意、成长轨迹等写成现行采集项（若提及 MUST 标明已下线或不展示）。 权限说明 MUST 与现网一致：麦克风用于喂养/预测语音； MUST NOT 将相机/相册/位置写为社区发帖或广场距离用途； MUST NOT.

#### Scenario: 用户查阅登录与型号

- **WHEN** 用户阅读「我们收集哪些信息」
- **THEN** 文案 SHALL 说明微信不收集头像/昵称，并 SHALL 提及设备型号用于崩溃排查与运维统计、非广告跟踪

#### Scenario: 用户查阅权限与社区残留

- **WHEN** 用户阅读权限与全文
- **THEN** 文案 SHALL 含扫码/相册识码或记录配图相关的相机相册用途，且 MUST NOT 将现行能力描述为社区广场发帖或位置距离推荐

#### Scenario: 用户查阅商业化与 AI

- **WHEN** 用户阅读商业化与第三方 AI 相关段落
- **THEN** 文案 SHALL 含精简的订单/权益表述且 MUST NOT 以现行 VIP 产品为名；AI 表述 SHALL 为可配置多供应商，且 MUST NOT 将已退役值得留意/成长轨迹/诊疗写为现行功能

#### Scenario: 用户查阅推送与预测说明

- **WHEN** 用户阅读「我们收集哪些信息」或等价章节
- **THEN** 文案 SHALL 提及推送标识用于提醒，以及预测相关时间点可能同步服务端

#### Scenario: 用户查阅 AI 与主体

- **WHEN** 用户阅读 AI / 第三方章节与运营者信息
- **THEN** 文案 SHALL 含多供应商可配置表述，并 SHALL 含公司主体全称

---

