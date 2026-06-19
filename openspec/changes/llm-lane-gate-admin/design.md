## Context

- voice-service 内 LLM 分散在 `voice_chat_deepseek.go`、`voice_chat_understanding.go`、`clinic_llm.go`；`chatLimiter` 已初始化但未 `Acquire`；非流式 STT 另有 `sttLimiter`（与 LLM 无关，本变更不改动 ASR/TTS）。
- ucg-service 润笔经 `compose_ai.go` 直连 DashScope；`ucg_ai_config` 表已支持 Admin 改 `visionModel`（60s 内存缓存）。
- voice-admin 仅管额度；ucg-admin「AI 配置」管润笔模型与额度。
- 生产 Redis Cluster 已存在；闸门键为运行时协调，非业务读缓存（沿用 `cache-and-messaging-hard-dependencies` 硬依赖策略，无需新增「读缓存收益率」评估）。
- 探索阶段已确认：按 **model** 分池（非全局单槽）；队列满立即 50301；一期含 Admin；voiceUnderstanding 含全部喂养 voice LLM；DB 种子方案 A（智谱默认）。

## Goals / Non-Goals

**Goals:**

- 三条 lane（`voiceUnderstanding`、`clinic`、`polish`）统一经 `aimodel.Invoke` / `InvokeStream`。
- 按 **当前 profile.model** 的 Redis 闸门：`maxInFlight` 通用信号量（默认 1，可配 N）；`maxWaiters` 有界等待；满则 50301。
- provider 适配器：`deepseek`、`zhipu`、`dashscope`；业务代码只认 `Lane` 枚举。
- Admin 热更新 profile（DB + InvalidateCache）；YAML/env 冷启动兜底；种子 A 智谱三模型。
- 保留原供应商 preset（Admin 可切回 `deepseek-chat` / `deepseek-v4-pro` / `qwen3-vl-plus`）。

**Non-Goals:**

- 不改百度 ASR/TTS、不改 WS 帧协议、不改 Flutter。
- 不新增后台 ticker/reconciler。
- 不在 Admin 配置 API Key（仍 env）。
- 不实现跨 lane 全局单槽。
- 不新增 `*_test.go`。

## Decisions

### 1. 包边界：`internal/services/aimodel`

voice 与 ucg 共用包（同仓库、无跨进程 RPC）。导出：

- `LaneVoiceUnderstanding`、`LaneClinic`、`LanePolish`
- `LoadProfile(ctx, lane) (Profile, error)`
- `Acquire(ctx, profile) (release func(), err error)` — `err` 为 `ErrQueueFull` → 映射 50301
- `Invoke` / `InvokeStream` — 内部 Acquire → provider HTTP → release（覆盖整段流式生命周期）

**备选**：各服务内复制闸门 → 否决（无法统一语义）。

### 2. 闸门 Redis 键（按 model）

```
ai:llm:gate:{normalizedModel}:inflight   # 整数，上限 maxInFlight
ai:llm:gate:{normalizedModel}:waiting    # 整数，上限 maxWaiters
```

`normalizedModel` = 小写 trim 的 model 字符串。使用 Lua 或等价原子脚本：`waiting++` 前检查 `waiting < maxWaiters`；拿到 inflight 槽后 `waiting--`；释放时 `inflight--`。持槽期间对 inflight 侧设置 TTL 兜底（如 clinic 120s + buffer），防止进程崩溃泄漏。

换 model 时键空间自动切换，旧池自然排空。

### 3. Lane Profile 存储

| Lane | 宿主 | 存储 |
|------|------|------|
| voiceUnderstanding | voice-service | `ai_voice_voice.llm_lane_config`（lane 主键） |
| clinic | voice-service | 同上 |
| polish | ucg-service | 扩展 `ucg_ai_config` 单行或 `llm_lane_config` 等价字段 |

字段：`provider`、`model`、`maxInFlight`、`maxWaiters`、`updatedAt`、`updatedBy`。endpoint 首期按 provider 内置默认映射（zhipu/deepseek/dashscope 各一），**不**开放 Admin 改 endpoint。

加载顺序：**DB Admin 配置 > YAML 默认 > 代码兜底**。缓存 TTL ~60s，PUT 后 `InvalidateLaneCache()`。

**种子 A（EnsureDefaultRows）**：

| lane | provider | model | maxInFlight | maxWaiters |
|------|----------|-------|-------------|------------|
| voiceUnderstanding | zhipu | glm-4.7-flash | 1 | 20 |
| clinic | zhipu | glm-4.1v-thinking-flash | 1 | 10 |
| polish | zhipu | glm-4.6v-flash | 1 | 15 |

### 4. API Key 映射（env only）

| provider | env |
|----------|-----|
| zhipu | `GLM_API_KEY` |
| deepseek | 现有 `voiceChat.deepseek.apiKey` 或 `DEEPSEEK_API_KEY` env 覆盖 |
| dashscope | `UCG_DASHSCOPE_API_KEY` |

Admin 切 provider 时若 env 缺失，返回明确配置错误（运维可见），不静默 fallback 他域 key。

### 5. voice 接入点（voiceUnderstanding 全量）

统一替换以下路径的上游 HTTP（经 `aimodel`）：

- `callDeepSeekRaw`、`callDeepSeekUnifiedIntent`、`callDeepSeekEntityExtract`、`callDeepSeekActionExtract`
- `callDeepSeekDirectReply`、`callDeepSeekGrowthSuggestion`、`callDeepSeekHistoryReply`
- `streamCasualReplyWithBaiduTTS` 内 LLM `http.Do` 段（TTS 仍百度）
- `voice_chat.go` 直调 `callDeepSeekGrowthSuggestion`
- `event_child_pending.go` 的 `callDeepSeekEntityExtract`

`chatWithResult` / `HandleTranscriptForStreaming` 逻辑不变，仅底层 LLM 调用改道。

**检查顺序**（喂养）：`voice_ai` 额度 check → `Acquire(voiceUnderstanding)` → LLM；50301 **不** consume。clinic：`clinic` 业务限流 + `clinic_ai` 额度 → `Acquire(clinic)` → LLM。

### 6. Provider 适配差异

- **zhipu / deepseek**：OpenAI 兼容 `chat/completions`；clinic thinking：zhipu 用 `thinking: {type:enabled}`，deepseek 保留 `reasoning_effort` + `extra_body`；流式 delta 共用 `extractClinicStreamDeltas`（`reasoning_content` / `content`）。
- **dashscope polish**：保留现有多模态 `image_url` 消息结构，仅 endpoint/key/model 来自 profile。

### 7. Admin API

**voice-service**

- `GET /voice/admin/api/llm-lanes` — 返回两 lane + allowlist（provider→models）
- `PUT /voice/admin/api/llm-lanes` — body 含 `voiceUnderstanding` / `clinic` 子对象
- 认证：`X-Admin-Password` = `VOICE_ADMIN_PASSWORD`

**ucg-service**

- 扩展 `GET/PUT /ucg/admin/api/ai-config`：增加 `provider`、`maxInFlight`、`maxWaiters`（`visionModel` 语义对齐为 polish `model` 或并存字段，实现时保持 JSON 兼容）

**UI**

- `voice-admin.html`：Tab「LLM 车道」
- `ucg-admin.html`：AI 配置 Tab 增并发/缓冲池/provider 联动

### 8. 错误码

- 新增 `CodeLLMQueueFull = 50301`，message「当前队列已满，请稍后重试」
- WS：`{type:error, code:50301, message:...}`
- HTTP 润笔：`gcode` 50301 同等语义
- 与 40301/40302/42901 边界不变

### 9. 可演进性

- `maxInFlight` 为配置整数，非硬编码 1；后期 Admin 改 FlashX 模型 + 提高并发无需改代码。
- provider 字段支持回切 DeepSeek/DashScope。

## Risks / Trade-offs

- **[Risk] 智谱默认种子但 prod 未配 GLM_API_KEY** → `.env.example`、runbook、部署检查清单写明；启动时 lane 调用前校验 key 存在性。
- **[Risk] voiceUnderstanding 单 model 池：闲聊流式长时间占槽** → 符合按 model 限流语义；后期 Admin 抬 `maxInFlight` 或换 FlashX。
- **[Risk] 滚动发布时新旧 Pod profile 不一致** → 短窗口双池并存可接受；建议一次性保存 Admin 后滚动。
- **[Risk] Redis 闸门脚本 bug 导致泄漏或超发** → inflight TTL + 单测式手工验收清单；日志带 model/lane。
- **[Trade-off] 一期同时做 Admin + 三 provider + 全 voice LLM** → scope 大但边界清晰；tasks 分模块并行。

## Migration Plan

1. 合并代码前在 test 环境配置 `GLM_API_KEY`。
2. 执行 DB migration / EnsureDefaultRows（种子 A）。
3. 部署 voice-service、ucg-service；验证 Admin 读写字段。
4. 冒烟：语音 commit 闲聊流式、胖宝 question、润笔 POST。
5. 回滚：Admin 切回 deepseek/dashscope preset + 保留 env；或回滚镜像（DB 配置仍可读）。
6. 更新 `privacy-policy.html` 与 runbook。

## Open Questions

- （已关闭）默认种子：**A 智谱**。
- （已关闭）一期含 Admin；voice 全量 LLM。
- 生产 `maxWaiters` 初值是否需按 lane 再调优：design 采用种子表默认值，上线后靠 Admin 调整。
