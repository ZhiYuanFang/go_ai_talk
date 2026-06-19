## ADDED Requirements

### Requirement: 系统 SHALL 按 model 维度实施 Redis LLM 并发闸门

系统 MUST 为每条 LLM lane 的 **当前 `profile.model`** 维护独立 Redis 并发闸门。闸门 MUST 支持可配置 `maxInFlight`（同时占用上游 API 的槽位数，默认 1）与 `maxWaiters`（允许排队等待的请求数，缓冲池）。`maxInFlight` MUST 为不小于 1 的整数且 MUST NOT 在代码中写死为 1。闸门键 MUST 使用规范化 model 名（小写、trim），格式为 `ai:llm:gate:{model}:inflight` 与 `ai:llm:gate:{model}:waiting`（或 design 批准的等价原子语义）。不同 model MUST NOT 共用同一闸门池。

#### Scenario: 同 model 并发互斥

- **WHEN** `voiceUnderstanding` lane 配置 `model=glm-4.7-flash` 且 `maxInFlight=1`，且已有一条请求占用槽位
- **THEN** 第二条同 model 请求 MUST 进入等待队列（若 `waiting < maxWaiters`）且 MUST NOT 在上一条释放前发起上游 HTTP

#### Scenario: 不同 model 可并行

- **WHEN** `glm-4.7-flash` 槽位被 voiceUnderstanding 占用，且 clinic lane 配置 `model=glm-4.1v-thinking-flash`
- **THEN** clinic 请求 MUST 使用独立闸门池且 MAY 并行调用上游

#### Scenario: 换 model 使用新池

- **WHEN** Admin 将 `voiceUnderstanding.model` 从 `glm-4.7-flash` 改为 `glm-4.7-flashx`
- **THEN** 新请求 MUST 使用 `glm-4.7-flashx` 闸门键且 MUST NOT 与旧 model 池共享 inflight 计数

### Requirement: 等待队列满时 MUST 立即拒绝且不得调用上游

当 `waiting >= maxWaiters` 时，系统 MUST 在业务入口立即返回错误，message **「当前队列已满，请稍后重试」**，code **50301**。该路径 MUST NOT 调用上游 LLM API，且 MUST NOT consume 月度 AI 额度（`voice_ai` / `clinic_ai` / `polish`）。

#### Scenario: 润笔队列满

- **WHEN** `polish` lane 的 `glm-4.6v-flash` 缓冲池已满且用户请求 `POST /ucg/app/api/posts/polish`
- **THEN** API MUST 返回 50301 且 MUST NOT 请求上游

#### Scenario: 喂养语音队列满

- **WHEN** `voiceUnderstanding` lane 缓冲池已满且用户 commit 后将触发 LLM
- **THEN** voice WS MUST 返回 `error` 帧 code 50301 且 MUST NOT 调用 LLM

#### Scenario: 胖宝队列满

- **WHEN** `clinic` lane 缓冲池已满且客户端发送合法 `question`
- **THEN** clinic WS MUST 返回 `error` 帧 code 50301 且 MUST NOT 调用 LLM

### Requirement: aimodel 包 SHALL 提供 Lane 统一调用入口

`internal/services/aimodel` MUST 导出 lane 枚举 `voiceUnderstanding`、`clinic`、`polish` 及 `Invoke` / `InvokeStream`。业务代码 MUST 通过 lane 调用上游，MUST NOT 在业务层硬编码 provider endpoint 或 model 字符串。`Acquire` 成功至上游调用完全结束（含流式读毕或连接关闭）期间 MUST 持有闸门槽位，并在 `defer` 或等价路径释放。

#### Scenario: 闲聊流式持槽至流结束

- **WHEN** `streamCasualReplyWithBaiduTTS` 经 `InvokeStream(LaneVoiceUnderstanding)` 调用上游
- **THEN** 闸门槽位 MUST 从首个上游 HTTP 发起持有至 SSE 读取结束或 context 取消

### Requirement: 系统 SHALL 支持多 provider 适配器

aimodel MUST 支持至少三种 provider：`zhipu`、`deepseek`、`dashscope`。lane profile MUST 含 `provider` 字段；API Key MUST 仅从环境变量或进程配置读取，MUST NOT 存于 Admin DB。切换 provider MUST 仅需更新 lane profile（及对应 env 已配置），MUST NOT 修改业务 lane 调用点。

#### Scenario: Admin 切回 DeepSeek 喂养模型

- **WHEN** `voiceUnderstanding` profile 改为 `provider=deepseek`、`model=deepseek-chat` 且 `DEEPSEEK_API_KEY` 已配置
- **THEN** 下一笔喂养 LLM 请求 MUST 调用 DeepSeek endpoint 且 MUST 使用 `deepseek-chat` 闸门池

#### Scenario: provider 对应 key 缺失

- **WHEN** profile 为 `provider=zhipu` 但 `GLM_API_KEY` 未配置
- **THEN** 系统 MUST 返回明确配置错误且 MUST NOT 调用上游

### Requirement: voiceUnderstanding lane SHALL 覆盖喂养语音全部 LLM 路径

下列路径 MUST 经 `LaneVoiceUnderstanding` 调用上游：统一意图解析、实体/动作抽取、闲聊直答、闲聊流式 LLM 段、成长建议、历史问答，以及 `event_child_pending` 中的实体抽取。纯 ASR、纯 TTS、规则回复与模式切换（不触发 LLM）MUST NOT 经过该 lane。

#### Scenario: 成长建议走 voiceUnderstanding

- **WHEN** 用户触发成长建议且将调用 LLM
- **THEN** 系统 MUST 使用 `voiceUnderstanding` lane profile 的 model 与闸门

#### Scenario: 纯 ASR 不占用 LLM 闸门

- **WHEN** 用户仅进行语音转写且无 LLM 调用
- **THEN** 系统 MUST NOT 调用 `aimodel.Acquire`
