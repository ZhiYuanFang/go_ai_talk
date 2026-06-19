## Why

当前 voice / clinic / 润笔三条 LLM 链路分别直连 DeepSeek 与 DashScope，进程内 `chatLimiter` 未接线，无按 model 的并发闸门；上游免费档模型并发通常为 1，高负载时易打穿合同或引发级联超时。需要统一引入 **按 model 分池** 的 Redis 等待队列，并在管理后台支持 **随时切换 provider/model 与并发/缓冲池**，同时保留 DeepSeek/DashScope 作为可回切预设。首次部署默认种子采用智谱三模型（方案 A）。

## What Changes

- 新增 `internal/services/aimodel`：`Lane` 枚举、`Profile`（provider/model/maxInFlight/maxWaiters）、Redis 按-model 闸门、`Invoke`/`InvokeStream` 统一入口及 `deepseek`/`zhipu`/`dashscope` 适配器。
- **voiceUnderstanding** lane 收敛喂养语音 **全部** LLM（意图/实体/动作、闲聊直答与流式、成长建议、历史问答等），替换现有 `callDeepSeek*` 与 `streamCasualReply` 中的上游 HTTP。
- **clinic** lane 收敛胖宝 `streamClinicLLM`；**polish** lane 收敛 `PolishPostText`。
- 每条 lane 独立闸门：默认 `maxInFlight=1`；`maxWaiters` 为缓冲池；等待队列满时 **立即** 返回 code **50301**「当前队列已满，请稍后重试」，且 **MUST NOT** 调用上游 LLM。
- 一期交付 Admin：`GET/PUT /voice/admin/api/llm-lanes`（voiceUnderstanding + clinic）；扩展 `GET/PUT /ucg/admin/api/ai-config`（polish lane 的 provider/model/并发/缓冲池）；扩展 `voice-admin.html` 与 `ucg-admin.html`。
- DB 种子（方案 A）：`voiceUnderstanding` → 智谱 `glm-4.7-flash`；`clinic` → `glm-4.1v-thinking-flash`；`polish` → `glm-4.6v-flash`；各 `maxInFlight=1`。YAML 保留 DeepSeek/DashScope 原文作冷启动兜底。
- 环境变量：`GLM_API_KEY`（新增）；保留 `DEEPSEEK_API_KEY`（或现有 yaml key）与 `UCG_DASHSCOPE_API_KEY` 以支持 Admin 回切。
- 更新隐私政策中 AI 供应商描述，与可配置 provider 一致。
- **BREAKING（运维）**：生产须配置 `GLM_API_KEY`；默认 LLM 供应商从 DeepSeek/DashScope 切至智谱（Admin 可回切）。

## Capabilities

### New Capabilities

- `llm-lane-gate`：按 model 的 Redis 并发闸门、50301 队列满语义、aimodel 统一调用与三 provider 适配。
- `llm-lane-admin`：voice/ucg Admin API 与页面动态配置 lane profile（provider、model、maxInFlight、maxWaiters）及 DB 种子。

### Modified Capabilities

- `ai-monthly-quota`：明确 50301 队列满时不 consume 额度、不调用 LLM；喂养/clinic/润笔检查顺序为额度/业务限流 → 闸门 → 上游。
- `pangbao-ai-clinic`：LLM 供应商可配置（非仅 DeepSeek）；队列满返回 50301；默认模型种子为智谱 thinking 模型。
- `voice-ai-quota`：喂养 voice LLM 经 aimodel lane；队列满 50301 语义与额度边界。
- `ucg-ai-quota`：润笔上游改为 aimodel lane（非硬编码 DashScope）；队列满 50301。
- `voice-admin-ui`：新增「LLM 车道」Tab；ucg-admin AI 配置 Tab 扩展 lane 闸门字段。
- `app-legal-docs`：修订 AI 供应商与模型可配置披露。
- `deepseek-history-redis-prefer`：历史读模型装配适用于「voiceUnderstanding lane 调用的 LLM」，不限定 DeepSeek 品牌。

## Impact

- **代码**：`internal/services/aimodel/**`；`internal/services/voice/voice_chat_deepseek.go`、`voice_chat_understanding.go`、`clinic_llm.go`、`internal/services/ucg/compose_ai.go`、`ai_config.go`；`internal/controller/voice_admin_*`、`ucg_admin_api.go`；`resource/public/voice-admin.html`、`ucg-admin.html`。
- **数据**：`ai_voice_voice` 新增 lane 配置表（或等价）；`ai_voice_ucg` 扩展 `ucg_ai_config` 或新表存 polish lane 闸门字段。
- **Redis**：新增 `ai:llm:gate:{model}:*` 键族（闸门计数，非读缓存；design 说明沿用 Redis 硬依赖）。
- **配置/env**：`manifest/config/voice-chat.shared.yaml` 增 lane 默认；`.env.example` 增 `GLM_API_KEY`；`docs/runbooks/release-deploy-and-run.md` 补充说明。
- **App API 使用统计**：新增 `/voice/admin/api/llm-lanes` 与扩展 `/ucg/admin/api/ai-config` 为运维型 Admin API，**不计入** usage 统计（与现有 voice/ucg admin 配额 API 同类）。
- **基线**：行为变更对照 `openspec/specs/v2.0.5/spec.md` 中 `ai-monthly-quota`、`pangbao-ai-clinic`、`voice-ai-quota`、`ucg-ai-quota`、`voice-admin-ui`、`app-legal-docs`。
