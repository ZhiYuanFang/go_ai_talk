## ADDED Requirements

### Requirement: Flutter SHALL 解析额度 API degraded 字段

`AiQuotaFeatureStatus`（`app/lib/data/ai_quota_models.dart`）MUST 扩展 **`degraded`** 布尔字段（JSON `degraded`，默认 false）。`VoiceAiQuotaStatus` 与 `PolishAiQuotaStatus` 的 fromJson MUST 解析各 feature 的 `degraded`。当 `degraded=true` 时，`remaining` 计算逻辑不变（仍为 `limit - used`）。

#### Scenario: 解析 polish degraded

- **WHEN** `/ucg/app/api/ai-quota` 返回 `polish: { used: 5, limit: 5, degraded: true }`
- **THEN** `PolishAiQuotaStatus.polish.degraded` SHALL 为 true

#### Scenario: 解析 clinic degraded

- **WHEN** `/voice/app/api/ai-quota` 返回 `clinicAi: { used: 10, limit: 10, degraded: true }`
- **THEN** `VoiceAiQuotaStatus.clinicAi.degraded` SHALL 为 true

### Requirement: AiQuotaRemainingHint SHALL 展示降速文案

`AiQuotaRemainingHint`（`app/lib/ui/widgets/ai_quota_remaining_hint.dart`）在 **`remaining=0` 且 `degraded=true`** 时 MUST 展示醒目降速文案，MUST NOT 仅展示「剩余 0 次」：

- **polish**：**「本月润笔额度已用完，已降速」**
- **clinicAi**：**「本月胖宝诊疗额度已用完，已降速」**
- **voiceAi**：保持现有 **「本月 AI 对话剩余 N 次」** 文案（`remaining=0` 时展示「剩余 0 次」）；degraded 字段不影响 voice hint 文案

降速文案 SHOULD 使用区别于普通 hint 的视觉权重（如略高对比度或 accent 色），具体样式由实现决定。

#### Scenario: 润笔降速 hint

- **WHEN** polish 额度 API 返回 used=5、limit=5、degraded=true
- **THEN** compose 页 hint SHALL 展示「本月润笔额度已用完，已降速」

#### Scenario: 胖宝降速 hint

- **WHEN** clinicAi 额度 API 返回 used=limit、degraded=true
- **THEN** 胖宝诊疗页 hint SHALL 展示「本月胖宝诊疗额度已用完，已降速」

#### Scenario: voice_ai 额度用尽 hint 不变

- **WHEN** voiceAi used=limit、degraded=true
- **THEN** hint SHALL 仍展示「本月 AI 对话剩余 0 次」（或等价 remaining 文案）

### Requirement: Flutter SHALL 收窄 40302 弹框至 voice_ai

`ai_quota_errors.dart` 及全局 40302 处理 MUST：**polish HTTP** 与 **clinic WS** 在仅额度用尽场景不再收到 40302，MUST NOT 弹「本月额度已用完」。**voice_ai WS** 收到 40302 MUST 仍弹框 **「本月额度已用完」**。40301 行为不变。

#### Scenario: voice 喂养 40302 仍弹框

- **WHEN** voice chat WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

#### Scenario: polish 成功 degraded 不弹框

- **WHEN** polish 返回 200 且 body 含 `quotaDegraded=true`
- **THEN** App SHALL NOT 弹 40302 额度用尽框

### Requirement: ucg_compose_screen SHALL 处理 quotaDegraded 可选提示

`ucg_compose_screen.dart` 在润笔成功响应 **`quotaDegraded=true`** 时 MAY 展示一次性 snackbar/toast（如「已切换至降速模式」）；`quotaDegraded=false` 或省略时 MUST NOT 展示该提示。该提示 MUST NOT 替代 `AiQuotaRemainingHint` 的持久降速文案。

#### Scenario: degraded 润笔可选 toast

- **WHEN** 用户触发润笔且响应 `quotaDegraded=true`
- **THEN** App MAY 展示一次性降速提示
- **AND** hint 区域 SHALL 同步反映 degraded 状态（刷新 quota provider 或本地标记）

### Requirement: pangbao_ai_screen SHALL 刷新 clinic 额度 degraded 状态

胖宝诊疗页（`pangbao_ai_screen.dart`）MUST 使用含 `degraded` 的额度 provider 驱动 `AiQuotaRemainingHintFeature.clinicAi`；额度用尽后用户仍可提问时，hint MUST 展示降速文案而非触发 40302 弹框。

#### Scenario: 额度用尽后仍可提问

- **WHEN** clinic_ai degraded=true 且用户发送 question 并成功收到 answer_done
- **THEN** UI SHALL NOT 展示 40302 弹框
- **AND** hint SHALL 展示「本月胖宝诊疗额度已用完，已降速」
