## REMOVED Requirements

### Requirement: device-service SHALL be authoritative for AI monthly quota configuration and usage

**Reason**: AI 配额按域拆分至 voice-service（voice_ai、clinic_ai）与 ucg-service（polish）；device-service 完全退出配额业务，符合一服务一数据库原则。

**Migration**: 维护窗口内将 device `ai_quota_*` 表数据拆分迁移至 voice/ucg 库；删除 device DAO 与 API；App/Flutter 改调 `/voice/app/api/ai-quota` 与 `/ucg/app/api/ai-quota`。

### Requirement: Internal quota check and consume APIs SHALL enforce wxId and feature semantics

**Reason**: device internal ai-quota API 整体移除；各域在自有进程内 check/consume。

**Migration**: voice 使用 `/voice/internal/api/ai-quota/*`；ucg 使用进程内 polish check/consume；删除 `/device/internal/api/ai-quota/*`。

### Requirement: App quota read API SHALL expose used and limit for both features

**Reason**: 集中式 `GET /device/app/api/ai-quota` 废弃；改为分域 App API，gateway 不做 BFF 聚合。

**Migration**: Flutter 分别从 `/voice/app/api/ai-quota`（voiceAi、clinicAi）与 `/ucg/app/api/ai-quota`（polish）读取；移除对 `/device/app/api/ai-quota` 的调用。

### Requirement: UCG admin SHALL configure global default and per-user override via proxied APIs

**Reason**: ucg-admin 不再转发 device 管理三字段；ucg 仅本地管理 polish，voice 由 voice-admin 管理 voice/clinic。

**Migration**: ucg-admin 移除 voiceAi/clinicAi 字段；新增 voice-admin.html 管理 voice 域配额。

## MODIFIED Requirements

### Requirement: UCG polish SHALL pre-check quota and consume only on DashScope success

`POST /ucg/app/api/posts/polish` MUST 在调用 DashScope 前于 **ucg-service 进程内**执行 polish check（MUST NOT 经 device internal API）；若 `allowed=false` MUST 返回 code **40302** 与 message **「本月额度已用完」** 且 MUST NOT 调用 DashScope。DashScope 成功返回有效正文后 MUST 于 ucg-service 进程内 consume。参数错误、未配置 AI、DashScope 失败 MUST NOT 调用 consume。

#### Scenario: 额度用尽

- **WHEN** 用户润笔 check 得到 used=5、limit=5
- **THEN** API SHALL 返回 40302 与「本月额度已用完」且 SHALL NOT 请求 DashScope

#### Scenario: DashScope 失败不扣减

- **WHEN** check 通过但 DashScope 返回 5xx
- **THEN** 系统 SHALL NOT 调用 consume 且 used SHALL 不变

### Requirement: Voice feeding AI SHALL require wxId and enforce quota around LLM calls

voice-service 在即将调用 LLM（母婴 DeepSeek 或 casual 流式）前 MUST 解析 wxId>0（优先 `X-Internal-Wx-Id`，否则 device **user 域** internal API 由 deviceNo 反查，MUST NOT 经 ai-quota API）。wxId≤0 MUST 返回 code **40301** 与登录引导文案，MUST NOT 调用 LLM。LLM 调用前 MUST 于 **voice-service 进程内**对 feature **`voice_ai`** 执行 check；`allowed=false` MUST 返回 WS 错误帧 code **40302** message **「本月额度已用完」**。LLM 成功完成后 MUST consume。模式切换、规则回复、LLM 失败兜底、纯 ASR MUST NOT check 或 consume。

#### Scenario: 未登录不可用喂养 AI

- **WHEN** WS 会话 wxId 解析为 0 且用户 utterance 将触发 LLM
- **THEN** 系统 SHALL 返回 40301 且 SHALL NOT 调用 LLM

#### Scenario: 喂养 AI 额度用尽

- **WHEN** check 得到 voice_ai used=limit
- **THEN** WS SHALL 返回 40302「本月额度已用完」且 SHALL NOT 调用 LLM

#### Scenario: 模式切换不扣减

- **WHEN** 用户发送模式切换指令且不触发 LLM
- **THEN** 系统 SHALL NOT 调用 check 或 consume

### Requirement: Flutter client SHALL display quota exhaustion dialog

App 客户端（flutter_ai_talk）在收到 HTTP 40302 或 WS 错误帧 code=40302 MUST 弹框展示 **「本月额度已用完」**（含润笔、喂养 AI、胖宝 AI）。40301 MUST 引导用户登录，MUST NOT 使用额度用尽文案。额度展示 MUST 从分域 API 获取：voice 域（voiceAi、clinicAi）与 ucg 域（polish），MUST NOT 依赖 `/device/app/api/ai-quota`。

#### Scenario: 润笔超额弹框

- **WHEN** polish API 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

#### Scenario: 喂养 AI 超额弹框

- **WHEN** voice WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

#### Scenario: 胖宝 AI 超额弹框

- **WHEN** clinic WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

## ADDED Requirements

### Requirement: Voice clinic AI SHALL enforce clinic_ai quota locally

voice-service 在处理 `/voice/clinic/ws` 的 `question` 并即将调用 LLM 前 MUST 解析 wxId>0。LLM 调用前 MUST 于 **voice-service 进程内**对 feature **`clinic_ai`** 执行 check；`allowed=false` MUST 返回 WS 40302。LLM 流式成功完成后 MUST consume。摘要刷新失败、rate limit、参数校验失败 MUST NOT consume。

#### Scenario: 胖宝额度用尽

- **WHEN** check 得到 clinic_ai used=limit
- **THEN** WS SHALL 返回 40302 且 MUST NOT 调用 LLM
