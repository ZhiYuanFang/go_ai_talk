## ADDED Requirements

### Requirement: App 额度状态与 VIP 有额度语义一致

App 所依赖的 voice/ucg `ai-quota` API 在账号为 VIP 时，MUST 对相关 feature 返回 `allowed=true`，MUST NOT 再以 40302「本月额度已用完」阻断 VIP 用户的 premium 路径。非 VIP 用户在额度用尽走 free/omit 时，API 可继续返回 `degraded=true`（或等价），但文案/弹框策略 SHOULD 与「免费模型降速」而非「功能不可用」对齐；Flutter 完整文案改版可作为跨仓跟随，本仓 MUST 保证字段语义可供客户端区分 VIP 与降速。

#### Scenario: VIP 拉取 voice ai-quota

- **WHEN** VIP 用户 GET `/voice/app/api/ai-quota`
- **THEN** `voiceAi`/`clinicAi`（及若暴露的 `careAlert`）MUST `allowed=true`

#### Scenario: 非 VIP 额尽仍可 degraded

- **WHEN** 非 VIP 用户某 feature used 已达 limit
- **THEN** 快照 MUST 标明不可再走 premium（如 `allowed=false` 且 `degraded=true`），客户端 MUST NOT 将其理解为 VIP
