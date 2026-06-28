## MODIFIED Requirements

### Requirement: UCG polish SHALL pre-check quota and consume only on DashScope success

`POST /ucg/app/api/posts/polish` MUST 在调用上游 LLM（经 **ucg-service** `LanePolish` / aimodel）前于本进程执行 polish 额度 check。若 `allowed=true`，MUST 经 `LoadProfile(LanePolish)` 调用上游；上游成功返回有效正文后 MUST 于本进程 consume；响应 MUST 含 `quotaDegraded: false`（或省略，客户端视为 false）。若 `allowed=false`（`used >= limit`），MUST **NOT** 返回 code **40302**；MUST 经 **degraded** 路径调用上游，强制使用 `DefaultSeedProfile(LanePolish)`（智谱 **`glm-4.6v-flash`**），MUST **NOT** consume 月度额度；成功响应 MUST 含 **`quotaDegraded: true`**。两种路径在调用上游前 MUST 经 polish lane 闸门 `Acquire`；队列满 MUST 返回 **50301** 且 MUST NOT 调用上游、MUST NOT consume。参数错误、未配置 AI、上游失败、50301 MUST NOT 调用 consume。

#### Scenario: 额度内正常润笔

- **WHEN** 用户润笔 check 得到 used=2、limit=5
- **THEN** API SHALL 经 Admin 配置的 LanePolish profile 调用上游
- **AND** 成功返回润色正文且 SHALL consume 一次 polish 额度
- **AND** `quotaDegraded` SHALL 为 false 或省略

#### Scenario: 额度用尽 degraded 润笔

- **WHEN** 用户润笔 check 得到 used=5、limit=5
- **THEN** API SHALL **NOT** 返回 40302
- **AND** SHALL 经 degraded 路径调用智谱 `glm-4.6v-flash` 并返回润色正文
- **AND** SHALL **NOT** consume polish 额度
- **AND** `quotaDegraded` SHALL 为 true

#### Scenario: 上游失败不扣减

- **WHEN** check 通过（allowed=true）但上游返回 5xx
- **THEN** 系统 SHALL NOT 调用 consume 且 used SHALL 不变

#### Scenario: degraded 上游失败不扣减

- **WHEN** check 得到 allowed=false 且 degraded 路径上游失败
- **THEN** 系统 SHALL NOT 调用 consume 且 used SHALL 不变

#### Scenario: 队列满不扣减

- **WHEN** check 通过但 polish lane 闸门返回队列满
- **THEN** API SHALL 返回 50301 且 MUST NOT 调用 consume

### Requirement: Flutter client SHALL display quota exhaustion dialog

App 客户端（flutter_ai_talk）在收到 HTTP 40302 或 WS 错误帧 code=40302 MUST 弹框展示 **「本月额度已用完」**，**但 polish HTTP 与 clinic WS 在仅因月度额度用尽（degraded 路径）时 MUST NOT 再返回 40302**。40301 MUST 引导用户登录，MUST NOT 使用额度用尽文案。额度展示 MUST 从分域 API 获取：voice 域（voiceAi、clinicAi）与 ucg 域（polish），MUST NOT 依赖 `/device/app/api/ai-quota`。当额度 API 返回 `degraded=true` 时，polish 与 clinicAi MUST 展示降速文案（见 `app-ai-quota-degraded-ui`），MUST NOT 仅展示「剩余 0 次」。

#### Scenario: 润笔 degraded 不弹 40302

- **WHEN** 用户额度用尽且 polish API 成功返回正文与 `quotaDegraded=true`
- **THEN** App SHALL **NOT** 弹框「本月额度已用完」

#### Scenario: 喂养 AI 超额弹框

- **WHEN** voice WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

#### Scenario: 胖宝 degraded 不弹 40302

- **WHEN** 用户 clinic_ai 额度用尽且 clinic WS 正常流式返回答案
- **THEN** App SHALL **NOT** 弹框「本月额度已用完」

### Requirement: Voice feeding AI SHALL require wxId and enforce quota around LLM calls

voice-service 在即将调用 LLM（**voiceUnderstanding** lane，含母婴理解、casual 流式、成长建议、历史问答等全部喂养 voice LLM 路径）前 MUST 解析 wxId>0（优先 `X-Internal-Wx-Id`，否则 device **user 域** internal API 由 deviceNo 反查，MUST NOT 经 ai-quota API）。wxId≤0 MUST 返回 code **40301** 与登录引导文案，MUST NOT 调用 LLM。LLM 调用前 MUST 于 **voice-service 进程内**对 feature **`voice_ai`** 执行 check；`allowed=false` MUST 返回 WS 错误帧 code **40302** message **「本月额度已用完」**（**无 degraded fallback**）。额度 check 通过后、调用上游前 MUST 经 voiceUnderstanding lane 闸门 `Acquire`；队列满 MUST 返回 **50301** 且 MUST NOT 调用 LLM、MUST NOT consume。LLM 成功完成后 MUST consume。模式切换、规则回复、LLM 失败兜底、纯 ASR、50301 MUST NOT check 或 consume（50301 在额度 check 之后短路，不 consume）。

#### Scenario: 未登录不可用喂养 AI

- **WHEN** WS 会话 wxId 解析为 0 且用户 utterance 将触发 LLM
- **THEN** 系统 SHALL 返回 40301 且 SHALL NOT 调用 LLM

#### Scenario: 喂养 AI 额度用尽

- **WHEN** check 得到 voice_ai used=limit
- **THEN** WS SHALL 返回 40302「本月额度已用完」且 SHALL NOT 调用 LLM

#### Scenario: 模式切换不扣减

- **WHEN** 用户发送模式切换指令且不触发 LLM
- **THEN** 系统 SHALL NOT 调用 check 或 consume

#### Scenario: 队列满不扣减喂养额度

- **WHEN** voice_ai check 通过但 voiceUnderstanding 闸门队列满
- **THEN** WS SHALL 返回 50301 且 MUST NOT consume voice_ai

### Requirement: Voice clinic AI SHALL enforce clinic_ai quota locally

voice-service 在处理 `/voice/clinic/ws` 的 `question` 并即将调用 LLM（**clinic** lane）前 MUST 解析 wxId>0。LLM 调用前 MUST 于 **voice-service 进程内**对 feature **`clinic_ai`** 执行 check。若 `allowed=true`，MUST 经 `LoadProfile(LaneClinic)` 调用上游；**仅** turn 以 **`answer_done` 成功结束** 时 MUST consume。**若 `allowed=false`（`used >= limit`）**，MUST **NOT** 返回 WS **40302**；MUST 经 **degraded** 路径调用上游，强制使用 `DefaultSeedProfile(LaneClinic)`（智谱 **`glm-4.1v-thinking-flash`**），且 **`answer_done` 成功时 MUST NOT consume** `clinic_ai`。业务限流 42901 检查后、调用上游前 MUST 经 clinic lane 闸门 `Acquire`；队列满 MUST 返回 **50301** 且 MUST NOT 调用 LLM、MUST NOT consume。摘要刷新失败、rate limit、参数校验失败、50301 MUST NOT consume。

#### Scenario: 胖宝额度用尽 degraded 仍可问答

- **WHEN** check 得到 clinic_ai used=limit
- **THEN** WS SHALL **NOT** 返回 40302
- **AND** SHALL 经 degraded 路径调用 LLM 并正常流式返回答案
- **AND** `answer_done` 后 MUST NOT consume clinic_ai

#### Scenario: 胖宝额度内正常扣减

- **WHEN** check 得到 clinic_ai used < limit 且 turn 以 `answer_done` 成功结束
- **THEN** 系统 MUST consume clinic_ai 一次

#### Scenario: 队列满不扣减胖宝额度

- **WHEN** clinic_ai check 与 42901 检查通过但 clinic 闸门队列满
- **THEN** WS SHALL 返回 50301 且 MUST NOT consume clinic_ai
