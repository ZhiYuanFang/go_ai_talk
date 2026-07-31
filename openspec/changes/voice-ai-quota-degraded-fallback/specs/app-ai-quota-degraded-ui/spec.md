## MODIFIED Requirements

### Requirement: AiQuotaRemainingHint SHALL 展示降速文案

`AiQuotaRemainingHint`（`app/lib/ui/widgets/ai_quota_remaining_hint.dart`）在 **`remaining=0` 且 `degraded=true`** 时 MUST 展示醒目降速文案，MUST NOT 仅展示「剩余 0 次」：

- **polish**：**「本月润笔额度已用完，已降速」**
- **clinicAi**：**「本月胖宝诊疗额度已用完，已降速」**
- **voiceAi**：**「本月 AI 对话额度已用完，已降速」**（或等价「喂养 AI」表述）

降速文案 SHOULD 使用区别于普通 hint 的视觉权重（如略高对比度或 accent 色），具体样式由实现决定。

#### Scenario: 润笔降速 hint

- **WHEN** polish 额度 API 返回 used=5、limit=5、degraded=true
- **THEN** compose 页 hint SHALL 展示「本月润笔额度已用完，已降速」

#### Scenario: 胖宝降速 hint

- **WHEN** clinicAi 额度 API 返回 used=limit、degraded=true
- **THEN** 胖宝诊疗页 hint SHALL 展示「本月胖宝诊疗额度已用完，已降速」

#### Scenario: 喂养 AI 降速 hint

- **WHEN** voiceAi used=limit、degraded=true
- **THEN** hint SHALL 展示「本月 AI 对话额度已用完，已降速」（或等价文案）
- **AND** SHALL NOT 仅展示「剩余 0 次」

### Requirement: Flutter SHALL 收窄 40302 弹框至非 degraded 额度错误

`ai_quota_errors.dart` 及全局 40302 处理 MUST：**polish HTTP**、**clinic WS** 与 **喂养 voice 对话** 在仅额度用尽（degraded）场景不再收到 40302，MUST NOT 弹「本月额度已用完」。若仍收到其它路径的 40302，App MAY 保留通用弹框。40301 行为不变。

#### Scenario: 喂养 degraded 不弹框

- **WHEN** 用户 voice_ai 用尽且喂养对话成功返回（无 40302）
- **THEN** App SHALL NOT 弹框「本月额度已用完」

#### Scenario: polish 成功 degraded 不弹框

- **WHEN** polish 返回 200 且 body 含 `quotaDegraded=true`
- **THEN** App SHALL NOT 弹 40302 额度用尽框

## ADDED Requirements

### Requirement: 喂养对话页 SHALL 刷新 voiceAi degraded 状态

喂养/语音对话相关 UI MUST 使用含 `degraded` 的额度 provider 驱动 `AiQuotaRemainingHintFeature.voiceAi`（或等价）；额度用尽后用户仍可对话时，hint MUST 展示降速文案而非依赖 40302 弹框。

#### Scenario: 额度用尽后仍可喂养对话

- **WHEN** voiceAi degraded=true 且用户继续语音/文字喂养对话并收到正常回复
- **THEN** UI SHALL NOT 展示 40302 弹框
- **AND** hint SHALL 展示降速文案
