## 1. aimodel 核心包

- [x] 1.1 新增 `internal/services/aimodel`：`Lane` 枚举、`Profile`、`ErrQueueFull`（50301）、`CodeLLMQueueFull` 常量
- [x] 1.2 实现 Redis 按-model 闸门（`inflight`/`waiting`、Lua 或等价原子语义、持槽 TTL 兜底）
- [x] 1.3 实现 profile 加载：YAML 默认 + DB 覆盖 + ~60s 缓存与 `InvalidateLaneCache`
- [x] 1.4 实现 `provider_zhipu`、`provider_deepseek`、`provider_dashscope` 适配器（chat/completions + 流式 SSE）
- [x] 1.5 实现 `Invoke` / `InvokeStream` 统一入口（Acquire → HTTP → release）

## 2. voice-service 数据与 Admin API

- [x] 2.1 新增 `ai_voice_voice` 表 `llm_lane_config`（或等价）及 EnsureDefaultRows 种子 A（智谱三模型默认值仅 voice 两 lane）
- [x] 2.2 实现 `GET/PUT /voice/admin/api/llm-lanes`（口令认证、allowlist 校验、`updatedBy` 审计）
- [x] 2.3 注册 API 路由与 controller；确认 Admin API 不计入 usage 统计（proposal 已记录）

## 3. voice-service LLM 接入

- [x] 3.1 将 `callDeepSeekRaw` 及全部 `callDeepSeek*` 改为经 `LaneVoiceUnderstanding`（含 unifiedIntent、entity、action、directReply、growth、history）
- [x] 3.2 将 `streamCasualReplyWithBaiduTTS` 内 LLM HTTP 段改为 `InvokeStream(LaneVoiceUnderstanding)`
- [x] 3.3 更新 `event_child_pending.go`、`voice_chat.go` 直调成长建议路径
- [x] 3.4 将 `clinic_llm.go` 改为 `InvokeStream(LaneClinic)`，保留 thinking/answer delta 协议
- [x] 3.5 喂养/clinic 路径补齐检查顺序：额度/42901 → 闸门 → 上游；50301 映射 WS `error` 帧
- [x] 3.6 移除或废弃未使用的 `chatLimiter`/`sttLimiter` 中 LLM 相关死代码（STT limiter 保留）

## 4. ucg-service 数据与 Admin

- [x] 4.1 扩展 `ucg_ai_config`（或等价）存储 polish lane 的 `provider`、`maxInFlight`、`maxWaiters`；EnsureDefaultRows 种子 A（`glm-4.6v-flash`）
- [x] 4.2 扩展 `GET/PUT /ucg/admin/api/ai-config` 请求/响应字段；更新 `AllowedVisionModels` / allowlist 含智谱与原有 DashScope 模型
- [x] 4.3 将 `PolishPostText` 改为 `aimodel.Invoke(LanePolish)`；50301 映射 HTTP 业务码

## 5. Admin UI

- [x] 5.1 扩展 `resource/public/voice-admin.html`：Tab「LLM 车道」、provider→model 联动、保存调用 llm-lanes API
- [x] 5.2 扩展 `resource/public/ucg-admin.html` AI 配置 Tab：provider、maxInFlight、maxWaiters

## 6. 配置、环境与文档

- [x] 6.1 更新 `manifest/config/voice-chat.shared.yaml`（lane YAML 兜底，保留 deepseek 段）；`config.ucg-service.yaml` 注释对齐
- [x] 6.2 更新 `manifest/docker/.env.example` 增加 `GLM_API_KEY`；`docs/runbooks/release-deploy-and-run.md` 部署说明
- [x] 6.3 更新 `resource/public/privacy-policy.html`（可配置多供应商披露）
- [x] 6.4 核对 `manifest/docker/docker-compose.microservices.yml` 中 voice/ucg 环境变量注入位（如需 `GLM_API_KEY`）

## 7. 验收（手工）

- [ ] 7.1 test 环境：语音 commit 闲聊流式、意图解析、成长建议各一条；确认走 voiceUnderstanding 闸门
- [ ] 7.2 test 环境：胖宝 question 流式 thinking+answer；润笔 POST 多图
- [ ] 7.3 Admin 切回 deepseek-chat / qwen3-vl-plus preset 后下一笔请求生效（无需改代码）
- [ ] 7.4 压测或手工占满缓冲池，确认 50301 且额度不扣减
