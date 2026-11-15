## ADDED Requirements

### Requirement: 预测 imminent 推送 MUST NOT 依赖功能开通或 VIP

向绑定用户发送普通可见 `predict_imminent` 推送时，系统 MUST NOT 查询或要求 `predict_capability_basic`、`predict_capability_pro` 人权益，MUST NOT 要求 isVip 作为发送前提。既有非商业门禁（pending、去重、历史窗口、限流等）MAY 继续生效。

#### Scenario: 未开通非 VIP 仍可收 imminent

- **WHEN** 目标 wx 无 Pro/基础版权益且非 VIP，imminent 触发且其它既有门禁通过
- **THEN** 系统 MUST 允许向该 wx 发送可见 imminent 推送

### Requirement: 统一 Live Activity 预测段 MUST NOT 因未开通而剥离

`RefreshBabyLiveActivity`（或等价统一投影）在扇出 upsert 时 MUST NOT 因缺少预测能力开通或 VIP 而过滤 `kind=imminent` 项。空投影时 end/清屏行为 MUST 保持。系统 MUST NOT 调用已删除的预测能力 access 合成作为扇出前提。

#### Scenario: 未开通用户仍可投影 imminent 段

- **WHEN** 无 Pro 权益且非 VIP 的用户设备有进窗预测 pending，触发统一 Refresh
- **THEN** ContentState 仍可含 imminent 项并 upsert（在开关开启等既有条件下）
