## Why

喂养 AI（`voice_ai`）额度用尽后仍硬返回 **40302** 并阻断对话，而同仓的胖宝诊疗（`clinic_ai`）与润笔（`polish`）已在用尽后走智谱降速 fallback、不计次、不弹「本月额度已用完」。产品选择与 clinic **全量对齐（方案 A）**：额度用尽后仍可继续喂养对话与 CRUD 落地，强制 `voiceUnderstanding` 种子智谱模型，避免半途不可用。

## What Changes

- **`voice_ai` 用尽改为 degraded 路径**：不再因月度额度返回 WS/HTTP **40302**；继续调用 Python 意图分析，但强制使用 `DefaultSeedProfile(LaneVoiceUnderstanding)`（智谱 **`glm-4.7-flash`**），成功时 **MUST NOT** `consume`。
- **额度快照**：`CheckVoiceAIQuotaStore` 对 `voice_ai` 在 `used >= limit` 时 MUST 设 `degraded=true`（与 `clinic_ai` 一致）；契约注释同步更新。
- **对话入口**：`chatWithResult` 与 `HandleTranscriptForIntentStream`（及共用 `guard`/`consume`）按 snapshot 分支；澄清续聊免计语义保留且与 degraded 正交。
- **App UI（BREAKING 相对旧「用尽即弹框」预期）**：喂养路径不再因额度用尽下发 40302；`voiceAi` 在 `remaining=0` 且 `degraded=true` 时展示降速文案（对齐 clinic/polish）；全局 40302 弹框不再依赖喂养额度用尽场景。
- **App 读 API**：`GET /voice/app/api/ai-quota` 的 `voiceAi.degraded` MUST 在用尽时为 `true`（实现上已有字段则可直接赋值）。
- **非范围**：不改月度 limit 数值、Admin 配置面、澄清 cid 协议、Python intent 契约字段；不新增 Redis 键；不新增后台 ticker；不为 degraded 单独加 rate limit（沿用现有闸门/路径行为）。

## Capabilities

### New Capabilities

- （无）行为落在既有额度与 App 降速 UI 能力上扩展。

### Modified Capabilities

- `ai-monthly-quota`：喂养 AI 用尽从「40302 阻断」改为 degraded 继续调用；Flutter 40302 说明收窄，不再将喂养额度用尽列为必弹场景。
- `app-ai-quota-degraded-ui`：`voiceAi` 降速 hint 与「40302 仅 voice」条款改为与 clinic 一致的降速展示；喂养用尽不再弹 40302。
- `voice-ai-quota`：App/internal check 快照对 `voice_ai` 暴露 `degraded`；喂养 LLM 在用尽时强制种子 profile 且不计次。

## Impact

- **后端**：`internal/services/voice/ai_quota*.go`、`voice_chat_quota.go`、`voice_chat_understanding.go`、`voice_chat.go`；`internal/services/aimodel/degraded_profile.go`（新增 `DegradedVoiceUnderstandingProfile`）；`contracts/ai_quota_contracts.go` 注释。
- **Flutter（flutter_ai_talk）**：`AiQuotaRemainingHint` 的 voiceAi 文案；依赖 40302 的喂养弹框逻辑（用尽路径消失后以 hint 为准）。
- **基线对照**：复用/变更 v2.0.24 中 `ai-monthly-quota`、`app-ai-quota-degraded-ui`、`voice-ai-quota`，并与 `pangbao-ai-clinic` 中 clinic degraded 模式对称。
- **网关 / usage**：无新增 App HTTP 路由；不涉及 usage 统计 denylist 变更。
- **风险**：降速模型可能影响意图准确率与喂养写库质量；成本上用尽后不计次依赖智谱侧用量，与 clinic/polish 已接受的风险同类。
