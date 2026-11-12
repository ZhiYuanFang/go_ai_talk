## ADDED Requirements

### Requirement: 功能开通域 MUST 拒绝广告开通通道

系统 MUST NOT 再通过广告完成功能权益授予。`POST /cash/app/api/feature/ad/complete`（若仍注册）MUST 返回明确业务错误。`ActivateFeature` 在 `channel=ad` 时 MUST 拒绝。Admin 保存功能定义时，若 `unlockMethods` 含 `ad` MUST 拒绝。EnsureSchema MUST 将仍含 `ad` 的 care/growth/预测相关种子或强制回写路径剥离为 `payment,invite_code`（与现网 care/growth 一致并覆盖新预测功能）。

#### Scenario: App 广告 complete 失败

- **WHEN** 旧客户端调用功能广告 complete
- **THEN** 系统 MUST 失败且 MUST NOT 写入或续期任何功能权益

#### Scenario: Admin 勾选广告被拒

- **WHEN** 运维保存功能定义且 unlockMethods 含 ad
- **THEN** 系统 MUST 拒绝保存

### Requirement: Hub 与 App 目录 MUST 去掉广告开通展示依赖

`cash-feature-admin.html` MUST NOT 再提供「广告 ad」开通方式勾选，邀请天数文案 MUST NOT 再捆绑「广告授予天数」作为开通方式配置入口。App 合成 catalog 项 MUST NOT 再依赖 `adDurationDays` 作为开通能力展示（字段省略或恒为无意义值且文档标明废弃；推荐省略）。

#### Scenario: Hub 无广告勾选

- **WHEN** 运维打开开通功能管理编辑任一功能
- **THEN** 页面 MUST NOT 展示可勾选的广告开通方式
