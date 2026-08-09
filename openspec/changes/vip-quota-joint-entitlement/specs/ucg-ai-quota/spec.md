## ADDED Requirements

### Requirement: polish 额度与 VIP 共同策略

ucg 润笔（`polish` feature）MUST 使用与 voice 相同的 VIP∪额度共同策略：VIP 时 Allowed=true 且成功不计次；非 VIP 额度内使用 `polish` lane 正式模型并在成功后 consume；非 VIP 额尽时 MUST 使用该 lane 的 free 配置，free 为空时 MUST NOT 使用硬编码 `DegradedPolishProfile` 作为真相源（按 design 以无覆盖/默认上游语义处理，且 MUST 可观测）。

#### Scenario: VIP 润笔不计次

- **WHEN** VIP 用户润笔成功返回正文
- **THEN** `polish` 月用量 MUST NOT 增加，响应 MUST NOT 再以额度用尽为由失败

#### Scenario: 非 VIP 额尽走 free

- **WHEN** 非 VIP 且 polish 额度用尽，且 Admin 配置了 polish free 模型
- **THEN** 润笔 MUST 使用 free 模型完成（成功时可不计次），MUST NOT 返回 40302 阻断
