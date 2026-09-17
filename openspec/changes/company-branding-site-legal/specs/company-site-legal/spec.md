## ADDED Requirements

### Requirement: 官网页脚 MUST 展示公司主体与备案信息

`resource/public/pangbao-home.html` 页脚 MUST 展示运营主体中文全称「杭州彩逗科技有限公司」、英文名「Hangzhou Caidou Technology Co., Ltd.」、ICP 备案号「浙ICP备2024101366号」（可点击链至工信部备案查询站）、客服邮箱 `tou_zy@foxmail.com`，以及隐私政策与用户协议链接。页脚 MUST NOT 再展示「官网事件 logo 与 Android 下载链路均来自系统权威数据」类运维旁白。

#### Scenario: 用户查看官网底部

- **WHEN** 用户打开官网首页并滚动至页脚
- **THEN** 页面 SHALL 可见公司中英文全称、浙ICP备2024101366号与客服邮箱，且可进入隐私政策与用户协议

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
