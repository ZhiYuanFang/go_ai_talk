## ADDED Requirements

### Requirement: 系统 MUST 提供可按业务类型发送的公共可见推送

系统 MUST 提供跨域可调用的公共推送契约（HTTP internal 或等价），入参至少包含接收方 `wxId`、业务类型 `bizType`、可见文案（及可选深链/data）。当 `bizType` 为预测临近（实现常量名 MAY 为 `predict_imminent`）时，系统 MUST 向该用户已注册的推送 token（APNs/HMS/MiPush）发送可见通知，且 MUST NOT 使用 UCG 社区未读总和作为该通知的角标语义（MAY 省略 badge 或使用与社区无关的固定策略）。调用方（含 voice）MUST 经 `clients/*` 契约调用，MUST NOT 直接 import `internal/services/ucg` 业务包或直查 ucg 库表。

#### Scenario: 预测临近推送不套用社区角标

- **WHEN** voice 以预测临近 `bizType` 请求向某 `wxId` 发送可见推送
- **THEN** 下发载荷 MUST 为可见提醒，且 MUST NOT 把 UCG 会话/通知未读之和写入该推送的业务角标语义

### Requirement: 预测临近推送 MUST 扇出到宝宝全部绑定用户

当预测临近叫醒校验通过后，系统 MUST 解析该 `deviceNo` 下全部绑定 `wxId`，并对每一用户尝试公共推送。单个用户无有效 token 或发送失败 MUST NOT 阻止对其余用户的尝试；整体流程 MUST 仍按调度规格 Ack 消息。

#### Scenario: 两名看护均收到尝试

- **WHEN** 某 `deviceNo` 绑定两个 `wxId` 且均有有效推送 token，且叫醒校验通过
- **THEN** 系统 MUST 对两个 `wxId` 均发起可见推送调用

#### Scenario: 无 token 用户跳过

- **WHEN** 某绑定 `wxId` 没有可用推送 token
- **THEN** 系统 MUST 跳过该用户且 MUST 继续处理其他绑定用户
