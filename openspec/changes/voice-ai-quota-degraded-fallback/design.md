## Context

基线 v2.0.24 中：`clinic_ai` / `polish` 额度用尽走智谱降速（`Degraded*Profile`）、不计 consume、不返回 40302；`voice_ai` 仍硬阻断。喂养意图经 Python（`AnalyzeIntent` / Stream），Go 侧用 `aimodel.LoadProfile(LaneVoiceUnderstanding)` 填 `ModelCfg`。对话额度守卫集中在 `guardVoiceAIQuota` + `consumeVoiceAIQuotaOnSuccess`，入口为 `chatWithResult` 与 `HandleTranscriptForIntentStream`。澄清续聊（本地 pending cid）已免计，与本变更正交。

产品决策：**方案 A**——与 clinic 全量对齐，用尽后仍可完整喂养对话与 CRUD。

## Goals / Non-Goals

**Goals:**

- `voice_ai` 用尽时 `degraded=true`，对话继续，强制种子智谱 `glm-4.7-flash`，成功不计次。
- App hint 对 `voiceAi` 展示降速文案；不再因喂养额度用尽弹 40302。
- 澄清免计与 degraded 可叠加，行为可观测（日志标明 degraded）。

**Non-Goals:**

- 不改默认月度 limit、Admin API、Python intent 请求字段。
- 不为 degraded 新增独立 rate limit 或 Redis 键。
- 不限制 degraded 下的 CRUD（非方案 B）。
- 不改成长建议/历史问答若当前未走 `guardVoiceAIQuota` 的遗留路径（若存在，实现时对齐或在 tasks 中显式勾选确认；主路径以两入口为准）。

## Decisions

### 1. 镜像 clinic snapshot API，而非仅改 error 类型

- **选择**：`CheckVoiceAIQuota` 改为（或新增）返回 `AIQuotaSnapshot` 的守卫，用尽时 `Allowed=false` + `Degraded=true` 且 **不**返回 `VoiceAIQuotaError(40302)`；仅 `wxId≤0` 仍 40301。
- **理由**：与 `CheckClinicAIQuotaSnapshot` 一致，调用方可分支 `degraded` 选 profile 与是否 consume。
- **备选**：继续抛错再 catch——不利于流式入口与日志语义。

### 2. 降级通过覆盖传给 Python 的 ModelCfg，而非 Go 直连 LLM

- **选择**：新增 `aimodel.DegradedVoiceUnderstandingProfile()`（等同 `DefaultSeedProfile(LaneVoiceUnderstanding)`：`zhipu` / `glm-4.7-flash`）；degraded 时用该 profile 填 `AnalyzeIntentRequest.Model`，**不**改 Admin DB lane 行。
- **理由**：喂养 LLM 在 Python 侧；clinic 是 Go `aimodel.Invoke`，手段不同、语义相同（强制种子）。
- **备选**：临时改 Redis/DB profile——污染全局且多连接共享有竞态。

### 3. 额度内仍用 Admin `LoadProfile`；仅用尽强制种子

- **选择**：`allowed=true` 走现有 `LoadProfile`；仅 degraded 强制种子。
- **理由**：保留 Admin 在额度内切 DeepSeek 等的能力；用尽才降本。

### 4. consume 仅在 `allowed` 且意图成功时

- **选择**：`!clarifyResume && !degraded` 时才 `consumeVoiceAIQuotaOnSuccess`；意图失败不 consume（保持现状）。
- **理由**：对齐 clinic「成功且非 degraded 才扣」。

### 5. App：hint 驱动，去掉喂养用尽 40302

- **选择**：`voiceAi` 在 `remaining=0 && degraded=true` 时文案对齐 clinic（如「本月 AI 对话额度已用完，已降速」）；全局 40302 处理不再把喂养额度用尽当作主要场景（后端不再下发）。
- **理由**：与 `app-ai-quota-degraded-ui` 对 polish/clinic 的模式一致。
- **备选**：WS 成功帧带 `quotaDegraded`——喂养路径改动面大，hint + API `degraded` 足够。

### 6. internal check API

- **选择**：`POST /voice/internal/api/ai-quota/check` 对 `voice_ai` 用尽返回 `allowed=false` 且 **`degraded=true`**（扩展响应字段若尚无）；调用方若仍把 `allowed=false` 当硬失败，本变更以本进程对话路径为准，internal 契约同步文档化。
- **理由**：避免跨进程误解；当前喂养主路径已在 voice 进程内 check。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 降速模型误检导致错误喂养写库 | 接受方案 A 产品风险；可观测日志含 `quota_degraded=true`；后续可再开方案 B |
| 用尽后不计次导致智谱成本上升 | 与 clinic/polish 同类；依赖既有 lane 闸门；不新增无限重试 |
| 基线 `ai-monthly-quota` 与 `pangbao-ai-clinic` 对 clinic 描述不一致 | 本变更只改 voice_ai 相关条款；不借机重写 polish/clinic 历史矛盾 |
| Flutter 仓与 Go 仓分离 | tasks 明确 Flutter 勾选项；可先合后端再合 App hint |

## Migration Plan

1. 先部署 voice-service：用尽不再 40302，开始降速调用。
2. 同步发布 Flutter hint（否则用户可能只见「剩余 0 次」但仍可用——可接受短暂不一致）。
3. 回滚：恢复 `voice_ai` 硬阻断与 App 40302 文案即可；无 DB migration、无新 Redis 键。

## Open Questions

- （无阻塞项）方案 A 已确认。实现时核对成长建议/历史问答是否共用 `guardVoiceAIQuota`；若否，是否一并纳入本变更在 apply 时决定（默认：凡计 `voice_ai` 的 LLM 路径均应 degraded 对齐）。
