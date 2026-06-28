## MODIFIED Requirements

### Requirement: Clinic SHALL 强制执行 clinic_ai 月度额度（per wxId）

voice-service 在调用 Clinic LLM 前 MUST 使用 auth 已绑定的 `wxId>0`；`wxId≤0` MUST 返回 `error` code **40301**。LLM 调用前 MUST 经 device internal 对 feature `clinic_ai` 以该 wxId 执行 check。若 `allowed=true`，MUST 经 `LoadProfile(LaneClinic)` 调用 LLM；**仅** turn 以 **`answer_done` 成功结束** 时 MUST 以同一 wxId consume。**若 `allowed=false`（`used >= limit`）**，MUST **NOT** 返回 code **40302**；MUST 经 **degraded** 路径调用 LLM，强制 `DefaultSeedProfile(LaneClinic)`（智谱 **`glm-4.1v-thinking-flash`**），且 **`answer_done` 成功时 MUST NOT consume** `clinic_ai`。**cancelled**、**superseded**、**disconnected** 或 LLM/摘要失败而中断的 turn **MUST NOT** consume（含 degraded 路径）。

#### Scenario: 未登录

- **WHEN** wxId 解析为 0 且用户发送 `question`
- **THEN** WS SHALL 返回 40301 且 MUST NOT 调用 LLM

#### Scenario: clinic_ai 额度用尽 degraded 问答

- **WHEN** check 得到 clinic_ai used=limit
- **THEN** WS SHALL **NOT** 返回 40302
- **AND** SHALL 经 degraded 路径流式返回答案
- **AND** `answer_done` 后 MUST NOT consume clinic_ai

#### Scenario: 额度内成功完成扣减

- **WHEN** check 得到 used < limit 且 turn 完整流式结束并下发 `answer_done`
- **THEN** 系统 MUST consume `clinic_ai` 一次

#### Scenario: 用户 cancel 不扣额度

- **WHEN** turn 在流式过程中被 `cancel` 或 `superseded` 结束且未收到 `answer_done`
- **THEN** 系统 MUST NOT 对该 turn 调用 consume `clinic_ai`（含 degraded 路径）

## ADDED Requirements

### Requirement: Clinic App quota read API SHALL expose clinic_ai degraded flag

`GET /voice/app/api/ai-quota` 响应中 `clinicAi` 对象 MUST 含 **`degraded`** 布尔字段；当 `clinic_ai` 的 `used >= limit` 时 MUST 为 `true`。`voiceAi` 对象 MAY 含同名字段（`used >= limit` 时为 true），但 voice_ai 用尽时 WS 仍 MUST 返回 40302，行为不变。

#### Scenario: clinic 额度用尽 API 标记

- **WHEN** wxId=1001 的 clinic_ai used=10、limit=10
- **THEN** `clinicAi.degraded` SHALL 为 true

#### Scenario: clinic 额度内

- **WHEN** wxId=1001 的 clinic_ai used=3、limit=10
- **THEN** `clinicAi.degraded` SHALL 为 false
