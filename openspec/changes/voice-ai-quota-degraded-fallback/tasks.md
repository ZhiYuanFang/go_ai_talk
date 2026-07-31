## 1. 契约与降级 Profile

- [x] 1.1 更新 `AIQuotaSnapshot` / 相关注释：`voice_ai` 用尽时允许 `degraded=true`（不再写「恒为 false」）
- [x] 1.2 在 `aimodel` 新增 `DegradedVoiceUnderstandingProfile()`（zhipu / glm-4.7-flash，对齐 `DefaultSeedProfile(LaneVoiceUnderstanding)`），附中文注释

## 2. voice 额度 Store 与 Check

- [x] 2.1 `CheckVoiceAIQuotaStore`：`voice_ai` 在 `!allowed` 时设 `Degraded=true`（与 `clinic_ai` 对称）
- [x] 2.2 将喂养预检改为 snapshot 语义（仿 `CheckClinicAIQuotaSnapshot`）：用尽不返回 40302；仅 wxId≤0 返回 40301
- [x] 2.3 确认 `GET /voice/app/api/ai-quota` 与 internal check 响应带上 `voiceAi`/`degraded`（必要时补 DTO 字段）
- [x] 2.4 调整 `guardVoiceAIQuota`（或等价）：返回 degraded 标志；用尽时继续并 Mark 已 check

## 3. 对话路径接入 degraded

- [x] 3.1 `callDeepSeekUnifiedIntent` / Stream：支持传入强制 profile；degraded 时用 `DegradedVoiceUnderstandingProfile` 填 Python `ModelCfg`
- [x] 3.2 `chatWithResult`：按 snapshot 分支；degraded 不 consume；澄清免计保持；打 degraded 日志
- [x] 3.3 `HandleTranscriptForIntentStream`：与 3.2 同一额度/模型语义
- [x] 3.4 排查其它计 `voice_ai` 的 LLM 入口（成长建议/历史问答等）；凡走额度者 MUST 同样 degraded，否则在 PR 说明为何豁免

## 4. Flutter（flutter_ai_talk）

- [x] 4.1 `AiQuotaRemainingHint`：voiceAi 在 remaining=0 且 degraded=true 时展示降速文案
- [x] 4.2 确认喂养对话页绑定 voiceAi degraded hint；用尽后可继续对话且不弹 40302
- [x] 4.3 收窄/核对全局 40302 处理：喂养额度用尽路径不再依赖弹框

## 5. 验收与评审自检

- [x] 5.1 手工：额度内仍用 Admin profile 且成功 consume；用尽后继续对话、模型为种子智谱、used 不变（代码路径已对齐；运行时冒烟待部署后确认）
- [x] 5.2 手工：澄清续聊在用尽后仍可完成且不 consume（代码路径已对齐；运行时冒烟待部署后确认）
- [x] 5.3 对照 v2.0.24 变更能力：`ai-monthly-quota`、`app-ai-quota-degraded-ui`、`voice-ai-quota`
- [x] 5.4 确认无新增 App HTTP 路由 / 无 usage denylist 变更；无新 Redis 键；无新测试文件；关键路径含中文注释
