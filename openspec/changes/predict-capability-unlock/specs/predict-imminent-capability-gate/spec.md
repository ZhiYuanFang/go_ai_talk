## ADDED Requirements

### Requirement: imminent 推送与 Live Activity 写入 MUST 校验预测能力开通或 VIP

当系统即将向某绑定用户成功发出 **imminent** 类**写入型**通知时，MUST 以该接收用户的 `wx_id` 校验：

`predict_capability_basic` 人权益未过期 **∨** `predict_capability_pro` 人权益未过期 **∨** 当前 isVip。

写入型包含：普通可见 `predict_imminent` 推送；Imminent / 统一投影路径上 Live Activity 的 **start** 与 **update**（向锁屏写入或刷新预测相关内容）。未通过时 MUST NOT 向该用户执行该次写入型 fanout，MUST 记录可观测跳过原因。权威判定 MUST 在 cash-service（internal API + clients），voice/history/push MUST NOT 直查 cash 库表。功能定义下架（`status=0`）MUST NOT 单独阻断已有有效权益或 VIP 的写入型 imminent（停售不停用）。

#### Scenario: 无权益非 VIP 跳过 imminent 写入

- **WHEN** imminent 触发且目标 wx 无基础版/Pro 有效权益且非 VIP
- **THEN** 系统 MUST NOT 向该 wx 发送该次 imminent 可见推送，MUST NOT 对其 Live Activity 执行 start/update

#### Scenario: 仅基础版可推

- **WHEN** 目标 wx 仅有基础版有效权益
- **THEN** 在其它既有门禁通过时 MUST 允许 imminent 写入型 fanout

#### Scenario: 仅 Pro 可推

- **WHEN** 目标 wx 仅有 Pro 有效权益
- **THEN** 在其它既有门禁通过时 MUST 允许 imminent 写入型 fanout

#### Scenario: 仅 VIP 可推

- **WHEN** 目标 wx 无预测功能权益但是 VIP
- **THEN** 在其它既有门禁通过时 MUST 允许 imminent 写入型 fanout

### Requirement: Live Activity 清屏 MUST NOT 受预测能力开通门禁

Live Activity 的 **end / dismissal（清屏）**——含统一 Refresh 在投影为空时下发的 end、以及任意为摘除锁屏实时活动而发送的 end——MUST NOT 因缺少基础版/Pro/VIP 或权益已过期而跳过。目的：预测能力到期后客户端不得因服务端拒发 end 而常驻实时活动。实现 MUST NOT 在「无开通」时短路整段 Refresh 以致连 end 也不发。

#### Scenario: 权益到期后仍可清屏

- **WHEN** 用户预测能力已过期且非 VIP，锁屏仍有此前预测相关 Live Activity，系统判定应 end/清屏
- **THEN** 系统 MUST 仍向该用户发送 Live Activity end（含有效 dismissal），MUST NOT 因未开通而跳过

#### Scenario: 无开通时空投影 Refresh 仍 end

- **WHEN** 无基础版/Pro 且非 VIP 的用户触发统一 Live Activity Refresh，且投影结果为空需 end
- **THEN** 系统 MUST 执行 end/清屏，MUST NOT 因能力门禁整段跳过 Refresh

### Requirement: timing Live Activity MUST NOT 受预测能力开通门禁

Timing Live Activity 的 start/update/end（及仅服务于 timing 的路径）MUST NOT 因缺少基础版/Pro/VIP 而拒绝。本变更 MUST NOT 将预测能力门禁挂到 timing 专用路径上。

#### Scenario: 未开通预测仍可更新 timing LA

- **WHEN** 用户无基础版/Pro 且非 VIP，系统需刷新 timing Live Activity
- **THEN** 系统 MUST NOT 仅因预测能力未开通而跳过 timing 更新
