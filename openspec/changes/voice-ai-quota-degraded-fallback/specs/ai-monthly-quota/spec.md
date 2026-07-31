## MODIFIED Requirements

### Requirement: Voice feeding AI SHALL require wxId and enforce quota around LLM calls

voice-service 在即将调用 LLM（**voiceUnderstanding** lane，含母婴理解、流式意图落地、成长建议、历史问答等全部喂养 voice LLM 路径）前 MUST 解析 wxId>0（优先 `X-Internal-Wx-Id`，否则 device **user 域** internal API 由 deviceNo 反查，MUST NOT 经 ai-quota API）。wxId≤0 MUST 返回 code **40301** 与登录引导文案，MUST NOT 调用 LLM。LLM 调用前 MUST 于 **voice-service 进程内**对 feature **`voice_ai`** 执行 check 并得到 snapshot。若 `allowed=true`，MUST 经 `LoadProfile(LaneVoiceUnderstanding)`（或等价）构造传给 Python 的模型配置，且意图/LLM **成功完成后** MUST consume。**若 `allowed=false`（`used >= limit`）**，MUST **NOT** 返回 code **40302**；MUST 经 **degraded** 路径继续调用（强制 `DefaultSeedProfile(LaneVoiceUnderstanding)` / `DegradedVoiceUnderstandingProfile`，智谱 **`glm-4.7-flash`**），且成功时 MUST NOT consume `voice_ai`。额度 check 通过或进入 degraded 后、调用上游前 MUST 经 voiceUnderstanding lane 闸门语义（若该路径适用 `Acquire`）；队列满 MUST 返回 **50301** 且 MUST NOT 调用 LLM、MUST NOT consume。模式切换、规则回复、LLM 失败兜底、纯 ASR、50301、**澄清续聊免计**（发起前本地已有有效 `conversation_id`）MUST NOT consume；澄清续聊 MUST NOT 因额度用尽被阻断。

#### Scenario: 未登录不可用喂养 AI

- **WHEN** WS 会话 wxId 解析为 0 且用户 utterance 将触发 LLM
- **THEN** 系统 SHALL 返回 40301 且 SHALL NOT 调用 LLM

#### Scenario: 喂养 AI 额度用尽 degraded 继续

- **WHEN** check 得到 voice_ai used=limit
- **THEN** WS/对话路径 SHALL **NOT** 返回 40302
- **AND** SHALL 经 degraded 路径调用意图/LLM（强制种子智谱模型）
- **AND** 成功完成后 MUST NOT consume voice_ai

#### Scenario: 额度内成功扣减

- **WHEN** check 得到 used < limit 且意图/LLM 成功完成（非澄清免计）
- **THEN** 系统 MUST consume voice_ai 一次

#### Scenario: 模式切换不扣减

- **WHEN** 用户发送模式切换指令且不触发 LLM
- **THEN** 系统 SHALL NOT 调用 check 或 consume

#### Scenario: 队列满不扣减喂养额度

- **WHEN** voice_ai check 允许继续（含 allowed 或 degraded）但 voiceUnderstanding 闸门队列满
- **THEN** WS SHALL 返回 50301 且 MUST NOT consume voice_ai

#### Scenario: 澄清续聊免计且额度用尽仍可续聊

- **WHEN** 发起意图前本地已有有效澄清 conversation_id 且 voice_ai used=limit
- **THEN** 系统 SHALL 继续调用意图分析且 MUST NOT 返回 40302
- **AND** MUST NOT consume voice_ai

### Requirement: Flutter client SHALL display quota exhaustion dialog

App 客户端（flutter_ai_talk）在收到 HTTP 40302 或 WS 错误帧 code=40302 MUST 弹框展示 **「本月额度已用完」**。**polish HTTP、clinic WS 与喂养 voice 对话路径在仅因月度额度用尽（degraded 路径）时 MUST NOT 再返回 40302**。40301 MUST 引导用户登录，MUST NOT 使用额度用尽文案。额度展示 MUST 从分域 API 获取：voice 域（voiceAi、clinicAi）与 ucg 域（polish），MUST NOT 依赖 `/device/app/api/ai-quota`。当额度 API 返回 `degraded=true` 时，polish、clinicAi 与 **voiceAi** MUST 展示降速文案（见 `app-ai-quota-degraded-ui`），MUST NOT 仅展示「剩余 0 次」。

#### Scenario: 润笔 degraded 不弹 40302

- **WHEN** 用户额度用尽且 polish API 成功返回正文与 `quotaDegraded=true`
- **THEN** App SHALL **NOT** 弹框「本月额度已用完」

#### Scenario: 喂养 AI degraded 不弹 40302

- **WHEN** 用户 voice_ai 额度用尽且喂养对话经 degraded 路径正常返回回复
- **THEN** App SHALL **NOT** 弹框「本月额度已用完」
- **AND** hint 区域 SHALL 反映 voiceAi degraded（降速文案）

#### Scenario: 胖宝 degraded 不弹 40302

- **WHEN** 用户 clinic_ai 额度用尽且 clinic WS 正常流式返回答案
- **THEN** App SHALL **NOT** 弹框「本月额度已用完」
