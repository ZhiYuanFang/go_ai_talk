## ADDED Requirements

### Requirement: 本变更 MUST NOT 回退广告开通清理

功能开通域广告通道拒绝、Hub 无广告勾选、catalog 不依赖广告字段等行为 MUST 保持；本变更回退推送门禁与基础版时 MUST NOT 恢复 `channel=ad` 授予或 Hub 广告开通配置入口。

#### Scenario: Activate ad 仍拒绝

- **WHEN** 调用方以 Channel=`ad` 请求 ActivateFeature
- **THEN** 系统 MUST 仍返回错误且 MUST NOT 变更权益
