## MODIFIED Requirements

### Requirement: 隐私政策 MUST 披露运营主体与联系方式

`resource/public/privacy-policy.html` MUST 在显著位置说明：本应用「胖宝」由**杭州彩逗科技有限公司**（Hangzhou Caidou Technology Co., Ltd.）运营；联系邮箱 MUST 含 `pangbao@cuplay.top`。文档 MUST 更新修订日期（生效日期可保持不变）。路径 MUST 仍为 `/privacy-policy.html`。页面 MUST NOT 要求或展示公司注册地址（含「以市场主体公示为准」或具体地址文案）。

#### Scenario: 用户阅读主体信息

- **WHEN** 用户打开 `/privacy-policy.html`
- **THEN** 页面 SHALL 可见公司中文全称与客服邮箱 `pangbao@cuplay.top`，且 SHALL NOT 将运营者表述为未具名的「个人开发者」或「胖宝团队」作为唯一主体
- **AND** 页面 MUST NOT 出现注册地址相关表述

### Requirement: 隐私政策 MUST 按现网能力披露收集与处理范围

隐私政策 MUST 按**客户端现网可触达能力**披露收集与处理范围，MUST NOT 将已对用户隐藏的 UCG 社区入口或已退役 VIP/值得留意/成长轨迹/诊疗等写成现行收集项。至少 MUST 覆盖：微信登录仅 unionid 且不收集头像/昵称；Apple 仅匿名用户标识符；宝宝档案与成长/喂养记录；推送标识；预测临近提醒同步；精简商业化（订单/权益，MUST NOT 以现行 VIP 产品名主披露）；设备型号用于崩溃排查与运维统计、非广告跟踪；现网语音/AI 可配置多供应商。权限说明 MUST 与现网一致（麦克风喂养/预测；相机相册扫码/配图/头像等），MUST NOT 写成社区发帖或广场位置用途。

#### Scenario: 用户查阅登录与型号

- **WHEN** 用户阅读「我们收集哪些信息」
- **THEN** 文案 SHALL 说明微信不收集头像/昵称，并 SHALL 提及设备型号用于崩溃排查与运维统计、非广告跟踪

#### Scenario: 用户查阅权限与社区残留

- **WHEN** 用户阅读权限与全文
- **THEN** 文案 SHALL 含扫码/相册识码或记录配图相关的相机相册用途，且 MUST NOT 将现行能力描述为社区广场发帖或位置距离推荐

## ADDED Requirements

### Requirement: 隐私政策视觉 MUST 与官网同系

`resource/public/privacy-policy.html` 的视觉（色板、顶栏、字体气质）MUST 与同期 `pangbao-home.html` 高档浅色风格对齐。顶栏 MUST NOT sticky 置顶；品牌链接 MUST 使用 `a.brand` 且 MUST NOT 显示下划线。

#### Scenario: 法律页顶栏不置顶

- **WHEN** 用户在隐私政策页向下滚动
- **THEN** 顶栏 SHALL 随文档滚走，MUST NOT 固定吸附在视口顶部

#### Scenario: 品牌链无下划线

- **WHEN** 用户查看隐私政策顶栏品牌
- **THEN** `a.brand` SHALL 无下划线
