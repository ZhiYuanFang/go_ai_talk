## MODIFIED Requirements

### Requirement: voice App quota read API SHALL expose voiceAi and clinicAi

`GET /voice/app/api/ai-quota` MUST 要求有效 Bearer 且 `X-Internal-Wx-Id > 0`（经 gateway-app 注入）。响应 MUST 为 `{ voiceAi: { used, limit, degraded }, clinicAi: { used, limit, degraded } }`，对应当月上海时区桶。**`voiceAi.degraded` MUST 为 `used >= limit`**（与 clinicAi 对称）。本接口 MUST NOT 返回 `polish` 字段。

#### Scenario: 登录用户查询 voice 域额度

- **WHEN** wxId=1001 请求 `/voice/app/api/ai-quota` 且当月胖宝已用 5、上限 30
- **THEN** `clinicAi.used` SHALL 为 5 且 `clinicAi.limit` SHALL 为 30

#### Scenario: 喂养额度用尽标记 degraded

- **WHEN** voice_ai used >= limit
- **THEN** 响应 `voiceAi.degraded` SHALL 为 true 且 `voiceAi.allowed` 语义为不可再计次消耗（App 以 degraded/used/limit 展示）

#### Scenario: wxId=0 拒绝

- **WHEN** 请求携带 wxId=0
- **THEN** 系统 SHALL 返回未授权/无效身份错误

### Requirement: voice internal quota check and consume APIs SHALL enforce wxId and feature semantics

`POST /voice/internal/api/ai-quota/check` MUST 要求有效 voice internal 密钥（与 voice-service 其它 internal API 一致）。body MUST 含 `wxId`（正整数）与 `feature`（`voice_ai` | `clinic_ai`）。响应 MUST 含 `allowed`（boolean）、`used`、`limit`，且 MUST 含 **`degraded`**（boolean）：当 `feature=clinic_ai` 或 `feature=voice_ai` 且 `used >= limit` 时 `degraded` MUST 为 true，此时 `allowed` MUST 为 false。`check` MUST NOT 修改用量。

`POST /voice/internal/api/ai-quota/consume` MUST 在 AI 成功返回后由 voice 本进程调用；成功时 MUST `INCR` 当月用量并返回 `{ used, limit }`。若扣减后 `used > limit`，系统 MUST 回滚该次 INCR 并返回超额错误。**degraded 对话路径 MUST NOT 调用 consume。**

#### Scenario: check 不扣减

- **WHEN** voice 在胖宝提问前调用 check feature=clinic_ai 且 used=29、limit=30
- **THEN** 响应 `allowed=true` 且 Redis 计数 SHALL 仍为 29

#### Scenario: voice_ai 用尽 check 返回 degraded

- **WHEN** check feature=voice_ai 且 used=limit
- **THEN** 响应 `allowed=false` 且 `degraded=true`
- **AND** Redis 计数 SHALL 不变

#### Scenario: consume 成功扣减

- **WHEN** voice 在 LLM 成功后调用 consume feature=voice_ai 且 used=4、limit=5
- **THEN** used SHALL 变为 5 且响应 used=5

#### Scenario: wxId 无效拒绝

- **WHEN** internal API 收到 wxId=0
- **THEN** 系统 SHALL 返回错误且 MUST NOT 读写的用量

## ADDED Requirements

### Requirement: voice feeding LLM SHALL 在 voice_ai 用尽时走 degraded 种子模型

voice-service 对 feature `voice_ai`：若 snapshot `allowed=true`，调用 Python 意图/LLM 前 MUST 使用 Admin `LoadProfile(LaneVoiceUnderstanding)` 的 provider/model；若 `allowed=false` 且 `degraded=true`，MUST 强制 `DegradedVoiceUnderstandingProfile`（智谱 **`glm-4.7-flash`**，与 `DefaultSeedProfile(LaneVoiceUnderstanding)` 一致）写入 Python `model` 配置，MUST NOT 写回 Admin DB lane，MUST NOT 返回 40302，成功 MUST NOT consume。本要求与 `ai-monthly-quota` 喂养条款一致，强调 Python 意图路径与闸门路径均覆盖。

#### Scenario: degraded 强制智谱 flash

- **WHEN** voice_ai used=limit 且用户发起冷启动意图分析
- **THEN** 发往 Python 的 model.provider/name MUST 为 zhipu / glm-4.7-flash（或等价种子）
- **AND** 成功后 used MUST 不变

#### Scenario: 额度内仍用 Admin profile

- **WHEN** voice_ai used < limit
- **THEN** 发往 Python 的 model MUST 来自当前 LaneVoiceUnderstanding profile（可为 Admin 配置的非种子模型）
