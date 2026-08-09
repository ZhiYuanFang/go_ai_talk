## ADDED Requirements

### Requirement: voice 域额度与 VIP 共同策略

`voice_ai` 与 `clinic_ai` 的 check/快照/App 状态 MUST 与 VIP 共同决策：VIP 时 MUST `Allowed=true` 且成功路径不计次。非 VIP 时保持月度 used/limit 语义；用尽后 MUST 允许继续走非 premium 选模（free/omit），MUST NOT 再依赖硬编码 Degraded 智谱种子作为唯一降级模型。

#### Scenario: 非 VIP 额度内喂养

- **WHEN** 非 VIP 且 `voice_ai` Allowed，完成喂养 LLM 成功
- **THEN** 系统 MUST 使用 `voiceUnderstanding` 正式模型，并 consume `voice_ai`

#### Scenario: 非 VIP 喂养额尽

- **WHEN** 非 VIP 且 `voice_ai` 用尽
- **THEN** 对话 MUST 仍可继续，选模 MUST 走 `voiceUnderstanding` free（或 omit），且该次成功 MUST NOT consume

### Requirement: tip 挂靠 clinic 额度与 lane

小贴士（TipStream）MUST 使用与 polish/诊疗相同的 VIP∪额度策略，feature 为 **`clinic_ai`**，lane 为 **`clinic`**（含 free）。MUST NOT 在无权益判定的情况下无脑传正式模型。

#### Scenario: tip 与诊疗共享 clinic_ai

- **WHEN** 非 VIP 用户 `clinic_ai` 仍有额度并成功生成 tip
- **THEN** 系统 MUST 按 clinic 正式模调用 Python，并 consume `clinic_ai`

#### Scenario: tip 在 clinic_ai 用尽时

- **WHEN** 非 VIP 且 `clinic_ai` 用尽后请求 tip
- **THEN** 系统 MUST 按 clinic free/omit 调用，MUST NOT 因 40302 单独拒绝 tip（与 clinic degraded 放行策略对齐）

### Requirement: 硬件语音通道特权

经 gateway 设备直连的 `/voice/chat/ws` 以及 MCP/internal text chat 设备通道，MUST 标记硬件特权并按 premium 选模（`voiceUnderstanding` 正式模），MUST NOT 计次。该特权 MUST NOT 自动授予 clinic WS、care-alert、tip、polish。

#### Scenario: MCP 文本对话走正式模不计次

- **WHEN** mcp-service 以配置 deviceNo 调用 internal text chat 成功
- **THEN** 意图分析 MUST 使用 `voiceUnderstanding` 正式模型，MUST NOT consume `voice_ai`
