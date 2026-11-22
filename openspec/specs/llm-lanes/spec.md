# llm-lanes

## Purpose

LLM 车道、配额、aimodel 退役、免费车道与选模闸门的行为规格。 Lane/quota admin and runtime gates shared across AI surfaces.

<!-- migrated-from: ai-model-admin-ui, ai-monthly-quota, aimodel-media-gen, aimodel-thinking-mode, app-ai-quota-degraded-ui, lane-free-model, llm-lane-admin, llm-lane-gate, retire-ai-quota-llm-aimodel, sim-llm-lane-admin, ucg-ai-quota, voice-ai-quota -->

## Requirements

### Requirement: gateway-app SHALL provide unified AI model and concurrency admin page

系统 MUST 提供独立 Admin 页面 **`/device/admin/ai-model-admin. 页面 MUST 在一屏内展示并编辑 **7 条 LLM lane** 的 `provider`、`model`、`maxInFlight`、`maxWaiters`：

| 分组 | lane 标识 | 后端 API |
|------|-----------|----------|
| Voice | `voiceUnderstanding`、`clinic` | `GET/PUT /voice/admin/api/llm-lanes` |
| UCG 润笔 | polish（`visionModel` 作 model） | `GET/PUT /ucg/admin/api/ai-config` |
| Sim | `simText`、`simVision`、`simImageGen`、`simVideoGen` | `GET/PUT /sim/admin/api/llm-lanes` |

页面 MUST 并行加载三域配置； 保存.

#### Scenario: 统一页加载七 lane

- **WHEN** 已鉴权管理员打开 ai-model-admin.html
- **THEN** 页面 MUST 展示 voice×2、ucg polish、sim×4 共七组模型与并发字段且值来自对应 GET API

#### Scenario: 统一页保存 voice lane

- **WHEN** 管理员修改 `clinic.maxInFlight` 并点击保存
- **THEN** 页面 MUST 调用 PUT `/voice/admin/api/llm-lanes` 且 MUST NOT 调用 sim scheduler reload

#### Scenario: 部分域保存失败

- **WHEN** voice PUT 成功而 sim PUT 返回 400
- **THEN** 页面 MUST 明确提示 sim 失败原因且 voice 成功状态 MUST 可见

### Requirement: Admin Hub SHALL link ai-model-admin

`resource/public/admin-modules.js` MUST 增加 `id: ai-model-admin` 模块入口，导航至 `/device/admin/ai-model-admin.html`，`showInNav: true`。Hub 登录后 MUST 可点击进入。

#### Scenario: Hub 导航可见 AI 模型与并发

- **WHEN** 管理员登录 Admin Hub
- **THEN** 模块列表 MUST 包含 ai-model-admin 入口

### Requirement: voice-admin ucg-admin sim-admin SHALL link to unified page instead of editing LLM

`voice-admin.html`、`ucg-admin.html`、`sim-admin.html` MUST **移除** LLM 模型/并发编辑 Tab 或表单控件。各页 MUST 保留指向 `/device/admin/ai-model-admin.html` 的可见链接（文案含「模型与并发」或等价）。各页原有非 LLM 职责（额度、Prompt、任务状态等）MUST 不变。

#### Scenario: voice-admin 无 LLM 编辑 Tab

- **WHEN** 运维打开 voice-admin.html
- **THEN** 页面 MUST NOT 含「LLM 车道」编辑 Tab 或 maxInFlight 输入框，且 MUST 含跳转统一页链接

#### Scenario: ucg-admin 无 polish 并发编辑

- **WHEN** 运维打开 ucg-admin「AI 配置」Tab
- **THEN** 页面 MUST NOT 含 provider/maxInFlight/maxWaiters/visionModel 编辑控件，且 MUST 含跳转统一页链接

#### Scenario: sim-admin 无 LLM lane 编辑

- **WHEN** 运维打开 sim-admin.html
- **THEN** 页面 MUST NOT 含 sim LLM lane 编辑表单，且 MUST 含跳转统一页链接

### Requirement: ai-model-admin static page MUST NOT count toward App usage stats

`/device/admin/ai-model-admin.html` 及本页触发的 voice/ucg/sim Admin PUT 为运维型接口，MUST NOT 计入 gateway-app App API 使用统计。

#### Scenario: 统一页保存不计入 usage

- **WHEN** 管理员从 ai-model-admin 保存 llm-lanes 成功
- **THEN** usage 统计 MUST NOT 递增 App API 计数

### Requirement: AI 模型与并发页展示 careAlert 与 free

`/device/admin/ai-model-admin.html` MUST 在 Voice 区域展示 **三条** lane：`voiceUnderstanding`、`clinic`、`careAlert`；对上述三条及 UCG `polish` MUST 提供 freeProvider/freeModel 编辑控件（允许清空）。Sim 四条 MUST NOT 展示 free 编辑。保存时 MUST 将 free 字段一并提交至对应域 Admin API。页面仍为运维型，MUST NOT 计入 App usage 统计。

#### Scenario: 页面加载三条 Voice lane

- **WHEN** 已鉴权管理员打开 ai-model-admin.html
- **THEN** 页面 MUST 展示 careAlert 块及各业务 lane 的 free 字段

#### Scenario: 清空 free 可保存

- **WHEN** 管理员清空某业务 lane 的 free 并保存成功
- **THEN** 后续非 premium 路径 MUST 按 omit/Python 自选语义工作

---

### Requirement: UCG polish SHALL pre-check quota and consume only on DashScope success

`POST /ucg/app/api/posts/polish` MUST 在调用上游 LLM（经 **ucg-service** `LanePolish` / aimodel）前于本进程执行 polish 额度 check；若 `allowed=false` MUST 返回 code **40302** 与 message **「本月额度已用完」** 且 MUST NOT 调用上游。额度 check 通过后、调用上游前 MUST 经 polish lane 闸门 `Acquire`；若队列满 MUST 返回 **50301**「当前队列已满，请稍后重试」且 MUST NOT 调用上游、MUST NOT consume。上游成功返回有效正文后 MUST 于本进程 consume。参数错误、未配置 AI、上游失败、50301 MUST NOT 调用 consume。

#### Scenario: 额度用尽

- **WHEN** 用户润笔 check 得到 used=5、limit=5
- **THEN** API SHALL 返回 40302 与「本月额度已用完」且 SHALL NOT 请求上游

#### Scenario: 上游失败不扣减

- **WHEN** check 通过但上游返回 5xx
- **THEN** 系统 SHALL NOT 调用 consume 且 used SHALL 不变

#### Scenario: 队列满不扣减

- **WHEN** check 通过但 polish lane 闸门返回队列满
- **THEN** API SHALL 返回 50301 且 SHALL NOT 调用 consume

### Requirement: Voice feeding AI SHALL require wxId and enforce quota around LLM calls

voice-service 在即将调用 LLM（**voiceUnderstanding** lane，含母婴理解、casual 流式、成长建议、历史问答等全部喂养 voice LLM 路径）前 MUST 解析 wxId>0（优先 `X-Internal-Wx-Id`，否则 device **user 域** internal API 由 deviceNo 反查，MUST NOT 经 ai-quota API）。 wxId≤0 MUST 返回 code **40301** 与登录引导文案，MUST NOT 调用 LLM。 LLM 调用前 MUST 于 **voice-service 进程内**对 feature **`voice_ai`** 执行 check； `allowed=false` MUST 返回 WS 错误帧 code **40302** message **「本月额度已用完」**。 额度 check 通过后、调用上游前 MUST 经 voiceUnderstanding lane 闸门 `Acquire`； 队列满 MUST.

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

### Requirement: Voice clinic AI SHALL enforce clinic_ai quota locally

voice-service 在处理 `/voice/clinic/ws` 的 `question` 并即将调用 LLM（**clinic** lane）前 MUST 解析 wxId>0。LLM 调用前 MUST 于 **voice-service 进程内**对 feature **`clinic_ai`** 执行 check；`allowed=false` MUST 返回 WS 40302。业务限流 42901 检查后、调用上游前 MUST 经 clinic lane 闸门 `Acquire`；队列满 MUST 返回 **50301** 且 MUST NOT 调用 LLM、MUST NOT consume。LLM 流式成功完成后 MUST consume。摘要刷新失败、rate limit、参数校验失败、50301 MUST NOT consume。

#### Scenario: 胖宝额度用尽

- **WHEN** check 得到 clinic_ai used=limit
- **THEN** WS SHALL 返回 40302 且 MUST NOT 调用 LLM

#### Scenario: 队列满不扣减胖宝额度

- **WHEN** clinic_ai check 与 42901 检查通过但 clinic 闸门队列满
- **THEN** WS SHALL 返回 50301 且 MUST NOT consume clinic_ai

---

### Requirement: aimodel SHALL expose sim lanes sharing Redis gates by model

`internal/services/aimodel` MUST 新增 Lane：`simText`、`simVision`、`simImageGen`、`simVideoGen`。`sim-user-service` 启动时 MUST 注册 `ProfileStore` 提供上述 lane 的 `provider`、`model`、`maxInFlight`、`maxWaiters`。

默认 model MUST 为：

| Lane | Model |
|------|-------|
| simText | glm-4.7-flash |
| simVision | glm-4.6v-flash |
| simImageGen | cogview-3-flash |
| simVideoGen | cogvideox-flash |

`Acquire` MUST 使用与现有服务相同的 Redis 键 `ai:llm:gate:{normalizedModel}:*`，使同 model 跨进程共用并发池。

#### Scenario: Shared gate with voice

- **WHEN** voice-service 与 sim-service 同时调用 `glm-4.7-flash`
- **THEN** 二者 MUST 竞争同一 inflight 上限

### Requirement: aimodel SHALL support image and video generation for sim

包 MUST 导出：

- `GenerateImage(ctx, lane, prompt) (result, err)` — CogView-3-Flash，POST 时 `Acquire`，释放后返回可下载 URL 或字节
- `SubmitVideoGeneration(ctx, lane, prompt) (taskID, err)` — CogVideoX-Flash 提交，POST 时 `Acquire`
- `PollVideoGeneration(ctx, taskID) (VideoPollResult, err)` — GET `/paas/v4/async-result/{taskID}` 轮询，MUST NOT 占用 inflight 槽

`PollVideoGeneration` 解析 MUST 支持智谱 **AsyncVideoGenerationResponse**：

- 状态：`task_status`（`PROCESSING`、`SUCCESS`、`FAIL`，大小写不敏感）； MAY fallback.

#### Scenario: Image generation acquires gate

- **WHEN** sim 调用 GenerateImage
- **THEN** 上游 HTTP 请求期间 MUST 持有 `cogview-3-flash` 槽位并在完成后释放

#### Scenario: Poll does not acquire gate

- **WHEN** sim 调用 PollVideoGeneration
- **THEN** MUST NOT 调用 `Acquire` inflight

#### Scenario: Official SUCCESS with video_result

- **WHEN** async-result 响应为 `{"task_status":"SUCCESS","video_result":[{"url":"https://example.com/a.mp4"}]}`
- **THEN** PollVideoGeneration MUST 返回 status=success 且 VideoURL 为上述 url

#### Scenario: PROCESSING without video_result

- **WHEN** 响应为 `{"task_status":"PROCESSING"}`
- **THEN** MUST 返回 status=processing 且 VideoURL 为空

#### Scenario: FAIL task

- **WHEN** 响应为 `{"task_status":"FAIL"}`
- **THEN** MUST 返回 status=failed

#### Scenario: Legacy video_url fallback

- **WHEN** 响应为 `{"task_status":"SUCCESS","video_url":"https://legacy.example/v.mp4"}` 且无 video_result
- **THEN** MUST 返回 status=success 且 VideoURL 为 legacy url

### Requirement: sim LLM text and vision SHALL use Invoke with sim lanes

昵称、文案、聊天、评论生成 MUST 经 `aimodel.Invoke`（或 `InvokeStream` 若需要）与对应 sim lane，MUST NOT 在 sim-service 内直连智谱 HTTP 绕过闸门。

#### Scenario: Comment uses simVision

- **WHEN** T2 生成评论
- **THEN** 调用 MUST 使用 `LaneSimVision` 且受 polish 同 model 池约束

### Requirement: aimodel ChatRequest SHALL support optional temperature for chat completions

`aimodel.ChatRequest` MUST 提供可选字段 `Temperature *float64`。当指针非 nil 时，`Invoke` / `InvokeStream`（及 `InvokeWithHeldProfile` 等等价路径）构建的上游 JSON MUST 包含 `temperature` 字段且值为所设浮点数。当指针为 nil 时，MUST NOT 向请求体写入 `temperature`（保持上游 provider 默认行为）。

#### Scenario: Temperature omitted by default

- **WHEN** 调用方构造 `ChatRequest` 且未设置 `Temperature`
- **THEN** 序列化后的请求体 MUST NOT 含 `temperature` 键

#### Scenario: Temperature explicitly set

- **WHEN** 调用方设置 `Temperature` 指向 `0.85`
- **THEN** 请求体 MUST 含 `"temperature": 0.85`

#### Scenario: Stream and non-stream parity

- **WHEN** 同一 `ChatRequest`（含 Temperature）分别用于流式与非流式
- **THEN** 二者 MUST 写入相同 `temperature` 值

---

### Requirement: aimodel SHALL 对智谱请求显式声明 thinking 模式且默认 disabled

`internal/services/aimodel` 在 `provider=zhipu` 的 chat/completions 请求体中 MUST **始终**包含 `thinking.type` 字段。当 `ChatRequest.ThinkingEnabled` 为 `false` 或未设置（零值）时 MUST 发送 `thinking.type=disabled`。当 `ThinkingEnabled` 为 `true` 时 MUST 发送 `thinking.type=enabled`。MUST NOT 因未启用 thinking 而省略 `thinking` 字段（避免 GLM-4.7 等模型上游默认开启 thinking）。

#### Scenario: sim 短文案默认关闭 thinking

- **WHEN** sim-user 经 `aimodel.Invoke(LaneSimText, ChatRequest{MaxTokens: 64})` 调用且未设置 `ThinkingEnabled`
- **THEN** 发往智谱的请求体 MUST 含 `"thinking": {"type": "disabled"}`

#### Scenario: clinic opt-in 开启 thinking

- **WHEN** voice clinic 经 `aimodel.InvokeStream` 调用且 `ThinkingEnabled=true`
- **THEN** 发往智谱的请求体 MUST 含 `"thinking": {"type": "enabled"}`

### Requirement: ChatRequest.MaxTokens SHALL 文档化与 thinking 的预算关系

`aimodel.ChatRequest` 的 `MaxTokens` 字段注释 MUST 说明：该值为上游 completion token 上限；当 thinking enabled 时 reasoning 与 content **共用**该预算，调用方不得将其仅理解为最终正文长度上限。

#### Scenario: 调用方阅读契约

- **WHEN** 开发者查阅 `internal/services/aimodel/request.go` 中 `ChatRequest`
- **THEN** MUST 可见 thinking 默认 disabled 语义及 MaxTokens 与 reasoning 共预算的说明

### Requirement: DeepSeek adapter SHALL 仅在 opt-in 时启用 thinking

`provider=deepseek` 时，仅当 `ChatRequest.ThinkingEnabled=true` MUST 写入 `extra_body.thinking`（或等价配置）与可选 `reasoning_effort`。`ThinkingEnabled=false` MUST NOT 写入 thinking 启用字段。

#### Scenario: 喂养 voice 未 opt-in

- **WHEN** voice 经 aimodel 调用 DeepSeek provider 且 `ThinkingEnabled=false`
- **THEN** 请求体 MUST NOT 含 thinking enabled 配置

---

### Requirement: Flutter SHALL 解析额度 API degraded 字段

`AiQuotaFeatureStatus`（`app/lib/data/ai_quota_models.dart`）MUST 扩展 **`degraded`** 布尔字段（JSON `degraded`，默认 false）。`VoiceAiQuotaStatus` 与 `PolishAiQuotaStatus` 的 fromJson MUST 解析各 feature 的 `degraded`。当 `degraded=true` 时，`remaining` 计算逻辑不变（仍为 `limit - used`）。

#### Scenario: 解析 polish degraded

- **WHEN** `/ucg/app/api/ai-quota` 返回 `polish: { used: 5, limit: 5, degraded: true }`
- **THEN** `PolishAiQuotaStatus.polish.degraded` SHALL 为 true

#### Scenario: 解析 clinic degraded

- **WHEN** `/voice/app/api/ai-quota` 返回 `clinicAi: { used: 10, limit: 10, degraded: true }`
- **THEN** `VoiceAiQuotaStatus.clinicAi.degraded` SHALL 为 true

### Requirement: AiQuotaRemainingHint SHALL 展示降速文案

`AiQuotaRemainingHint`（`app/lib/ui/widgets/ai_quota_remaining_hint.dart`）在 **`remaining=0` 且 `degraded=true`** 时 MUST 展示醒目降速文案，MUST NOT 仅展示「剩余 0 次」：

- **polish**：**「本月润笔额度已用完，已降速」**
- **clinicAi**：**「本月胖宝诊疗额度已用完，已降速」**
- **voiceAi**：保持现有 **「本月 AI 对话剩余 N 次」** 文案（`remaining=0` 时展示「剩余 0 次」）；degraded 字段不影响 voice hint 文案

降速文案 SHOULD 使用区别于普通 hint 的视觉权重（如略高对比度或 accent 色），具体样式由实现决定。

#### Scenario: 润笔降速 hint

- **WHEN** polish 额度 API 返回 used=5、limit=5、degraded=true
- **THEN** compose 页 hint SHALL 展示「本月润笔额度已用完，已降速」

#### Scenario: 胖宝降速 hint

- **WHEN** clinicAi 额度 API 返回 used=limit、degraded=true
- **THEN** 胖宝诊疗页 hint SHALL 展示「本月胖宝诊疗额度已用完，已降速」

#### Scenario: voice_ai 额度用尽 hint 不变

- **WHEN** voiceAi used=limit、degraded=true
- **THEN** hint SHALL 仍展示「本月 AI 对话剩余 0 次」（或等价 remaining 文案）

### Requirement: Flutter SHALL 收窄 40302 弹框至 voice_ai

`ai_quota_errors.dart` 及全局 40302 处理 MUST：**polish HTTP** 与 **clinic WS** 在仅额度用尽场景不再收到 40302，MUST NOT 弹「本月额度已用完」。**voice_ai WS** 收到 40302 MUST 仍弹框 **「本月额度已用完」**。40301 行为不变。

#### Scenario: voice 喂养 40302 仍弹框

- **WHEN** voice chat WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

#### Scenario: polish 成功 degraded 不弹框

- **WHEN** polish 返回 200 且 body 含 `quotaDegraded=true`
- **THEN** App SHALL NOT 弹 40302 额度用尽框

### Requirement: ucg_compose_screen SHALL 处理 quotaDegraded 可选提示

`ucg_compose_screen.dart` 在润笔成功响应 **`quotaDegraded=true`** 时 MAY 展示一次性 snackbar/toast（如「已切换至降速模式」）；`quotaDegraded=false` 或省略时 MUST NOT 展示该提示。该提示 MUST NOT 替代 `AiQuotaRemainingHint` 的持久降速文案。

#### Scenario: degraded 润笔可选 toast

- **WHEN** 用户触发润笔且响应 `quotaDegraded=true`
- **THEN** App MAY 展示一次性降速提示
- **AND** hint 区域 SHALL 同步反映 degraded 状态（刷新 quota provider 或本地标记）

### Requirement: pangbao_ai_screen SHALL 刷新 clinic 额度 degraded 状态

胖宝诊疗页（`pangbao_ai_screen.dart`）MUST 使用含 `degraded` 的额度 provider 驱动 `AiQuotaRemainingHintFeature.clinicAi`；额度用尽后用户仍可提问时，hint MUST 展示降速文案而非触发 40302 弹框。

#### Scenario: 额度用尽后仍可提问

- **WHEN** clinic_ai degraded=true 且用户发送 question 并成功收到 answer_done
- **THEN** UI SHALL NOT 展示 40302 弹框
- **AND** hint SHALL 展示「本月胖宝诊疗额度已用完，已降速」

### Requirement: App 额度状态与 VIP 有额度语义一致

App 所依赖的 voice/ucg `ai-quota` API 在账号为 VIP 时，MUST 对相关 feature 返回 `allowed=true`，MUST NOT 再以 40302「本月额度已用完」阻断 VIP 用户的 premium 路径。非 VIP 用户在额度用尽走 free/omit 时，API 可继续返回 `degraded=true`（或等价），但文案/弹框策略 SHOULD 与「免费模型降速」而非「功能不可用」对齐；Flutter 完整文案改版可作为跨仓跟随，本仓 MUST 保证字段语义可供客户端区分 VIP 与降速。

#### Scenario: VIP 拉取 voice ai-quota

- **WHEN** VIP 用户 GET `/voice/app/api/ai-quota`
- **THEN** `voiceAi`/`clinicAi`（及若暴露的 `careAlert`）MUST `allowed=true`

#### Scenario: 非 VIP 额尽仍可 degraded

- **WHEN** 非 VIP 用户某 feature used 已达 limit
- **THEN** 快照 MUST 标明不可再走 premium（如 `allowed=false` 且 `degraded=true`），客户端 MUST NOT 将其理解为 VIP

---

### Requirement: 业务 lane 可配置额度不足模型

对 lane `voiceUnderstanding`、`clinic`、`careAlert`、`polish`，Admin 配置 MUST 支持可选的免费/额度不足模型字段（`freeProvider`、`freeModel`，可为空）。非 premium 选模 MUST 读取上述字段：非空则作为下游模型；为空则 Go MUST omit 模型配置并由 Python（或 polish 侧约定的无模型覆盖语义）自行决定免费模型。Sim 四条 lane（`simText`/`simVision`/`simImageGen`/`simVideoGen`）MUST NOT 要求或持久化 free 字段。

#### Scenario: 保存 clinic free 模型

- **WHEN** 管理员 PUT llm-lanes 且 `clinic.freeProvider`/`freeModel` 为允许列表内有效值
- **THEN** 系统 MUST 持久化该 free 配置，供后续非 premium 诊所/tip 路径使用

#### Scenario: 清空 free 表示交 Python

- **WHEN** 管理员将某业务 lane 的 freeProvider 与 freeModel 存为空
- **THEN** 非 premium 且走 Python 的请求 MUST 省略 model 字段

#### Scenario: Sim 无 free

- **WHEN** 管理员 GET/PUT sim llm-lanes
- **THEN** API MUST NOT 要求客户端提交 free 字段；即便误传也 MUST NOT 改变 Sim 运行时选模契约（本变更不引入 Sim 额度策略）

### Requirement: 正式模型与 free 模型分离

lane 的正式 `provider`/`model` MUST 仅用于 premium 路径；MUST NOT 在非 premium 路径回退使用正式模型（除非 free 显式配置为相同值）。硬编码 `DegradedVoiceUnderstandingProfile` / `DegradedClinicProfile` / `DegradedPolishProfile` MUST NOT 再作为生产选模真相源。

#### Scenario: 额尽不再使用代码种子智谱

- **WHEN** 非 VIP 用户某 feature 额度用尽且该 lane free 配置指向运维设定的模型 A
- **THEN** 下游 MUST 使用模型 A，MUST NOT 忽略 free 而改用代码内 Degraded* 常量

---

### Requirement: voice-service SHALL 提供 LLM lane Admin API

voice-service MUST 提供 `GET /voice/admin/api/llm-lanes` 与 `PUT /voice/admin/api/llm-lanes`，认证 MUST 为 Header `X-Admin-Password` 等于 `voice.admin.password`（经 gateway-app 反代）。响应与请求 MUST 包含 `voiceUnderstanding` 与 `clinic` 两个子对象，各含 `provider`、`model`、`maxInFlight`、`maxWaiters`、`updatedAt`、`updatedBy`。GET MUST 返回 provider→model allowlist 供 Admin UI 下拉联动。PUT MUST 校验 allowlist 与正整数边界（`maxInFlight>=1`，`maxWaiters>=0`）。

#### Scenario: 管理员读取 lane 配置

- **WHEN** 运维携带正确口令 GET `/voice/admin/api/llm-lanes`
- **THEN** 响应 SHALL 含两 lane 当前 DB 配置与 allowlist

#### Scenario: 管理员更新 clinic 并发

- **WHEN** 运维 PUT `clinic.maxInFlight=5` 且 model 在 allowlist 内
- **THEN** voice-service MUST 持久化至 `ai_voice_voice` 且 MUST 失效 lane 缓存

#### Scenario: 口令错误拒绝

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 MUST 返回未授权且 MUST NOT 修改配置

### Requirement: ucg-service SHALL 扩展 AI 配置 Admin 以包含 polish lane 闸门

`GET/PUT /ucg/admin/api/ai-config` MUST 扩展支持 `provider`、`maxInFlight`、`maxWaiters`（与现有 `visionModel` / `maxImagesPerRequest` 并存或语义对齐为 polish lane 的 `model`）。PUT 成功后 MUST 失效 AI/lane 缓存（与现有 `InvalidateAIConfigCache` 一致）。

#### Scenario: 管理员更新润笔模型与缓冲池

- **WHEN** 运维 PUT `provider=zhipu`、`visionModel=glm-4.6v-flash`、`maxInFlight=1`、`maxWaiters=15`
- **THEN** ucg-service MUST 持久化且下一笔润笔 MUST 使用新 profile

### Requirement: DB 种子 MUST 默认为智谱三模型（方案 A）

首次 EnsureDefaultRows（或 migration 种子）时：若 lane 配置行不存在，系统 MUST 写入：`voiceUnderstanding` → `zhipu` / `glm-4.7-flash` / `maxInFlight=1` / `maxWaiters=20`；`clinic` → `zhipu` / `glm-4.1v-thinking-flash` / `maxInFlight=1` / `maxWaiters=10`；`polish` → `zhipu` / `glm-4.6v-flash` / `maxInFlight=1` / `maxWaiters=15`。YAML 中 DeepSeek/DashScope 默认值 MUST 保留作 DB 缺失时的冷启动兜底。

#### Scenario: 新环境首次启动

- **WHEN** voice 与 ucg 库尚无 lane 配置行
- **THEN** EnsureDefaultRows 后 Admin GET MUST 返回上述智谱默认

### Requirement: voice-admin 与 ucg-admin SHALL 提供 LLM lane 配置 UI

LLM lane 的模型与并发 Admin UI MUST 集中在 **`/device/admin/ai-model-admin.html`**（见 `ai-model-admin-ui`）。`voice-admin.html` 与 `ucg-admin.html` MUST NOT 再提供 LLM lane 编辑 Tab 或表单；MUST 仅链至统一页。后端 API（`GET/PUT /voice/admin/api/llm-lanes`、`GET/PUT /ucg/admin/api/ai-config` 之 polish 字段）MUST 不变，由统一页调用。

#### Scenario: voice LLM 经统一页保存

- **WHEN** 运维在 ai-model-admin 修改 voiceUnderstanding 并保存
- **THEN** 页面 MUST 调用 PUT `/voice/admin/api/llm-lanes` 且 voice-admin MUST NOT 含重复编辑 UI

#### Scenario: ucg polish 经统一页保存

- **WHEN** 运维在 ai-model-admin 修改 polish maxWaiters 并保存
- **THEN** 页面 MUST 调用 PUT `/ucg/admin/api/ai-config` 且 ucg-admin MUST NOT 含重复编辑 UI

### Requirement: LLM lane Admin API MUST NOT 计入 App usage 统计

`/voice/admin/api/llm-lanes` 与扩展后的 `/ucg/admin/api/ai-config` 为运维型 Admin API，MUST NOT 计入 gateway-app App API 使用统计。

#### Scenario: Admin 保存不计入 usage

- **WHEN** 管理员 PUT llm-lanes 成功
- **THEN** usage 统计 MUST NOT 递增该路径计数

### Requirement: Admin PUT 失效 lane 缓存 MUST NOT 导致进程崩溃

Admin 成功 PUT `/ucg/admin/api/ai-config` 或 `/voice/admin/api/llm-lanes` 后，系统 MUST 失效进程内 lane/profile 缓存以使下一笔 LLM 调用读取新配置；该失效路径 MUST NOT 因 `InvalidateLaneCache` 与 `ProfileStore.InvalidateCache` 互相调用而导致 stack overflow 或进程退出。

#### Scenario: ucg PUT ai-config 后进程保持运行

- **WHEN** 运维携带有效口令 PUT `/ucg/admin/api/ai-config` 且持久化成功
- **THEN** ucg-service 进程 MUST 保持运行且后续 GET 同一接口 MUST 返回更新后的配置

#### Scenario: voice PUT llm-lanes 后进程保持运行

- **WHEN** 运维携带有效口令 PUT `/voice/admin/api/llm-lanes` 且持久化成功
- **THEN** voice-service 进程 MUST 保持运行且后续 GET MUST 返回更新后的 lane 配置

### Requirement: voice llm-lanes 包含 careAlert 与 free 字段

`GET/PUT /voice/admin/api/llm-lanes` MUST 包含三个业务子对象：`voiceUnderstanding`、`clinic`、`careAlert`。每一子对象 MUST 含正式 `provider`/`model`/`maxInFlight`/`maxWaiters` 以及可选 `freeProvider`/`freeModel`（允许空字符串）。PUT MUST 校验正式模型 allowlist；当 free 非空时 MUST 同样校验 allowlist；free 全空 MUST 合法。

#### Scenario: PUT 写入 careAlert 与 free

- **WHEN** 管理员提交含 `careAlert` 正式模与空 free 的合法 PUT
- **THEN** 系统 MUST 持久化成功，后续 GET MUST 读出一致字段

#### Scenario: 非法 free 模型拒绝

- **WHEN** freeProvider/freeModel 非空但不在 allowlist
- **THEN** PUT MUST 失败且 MUST NOT 部分写入该 lane

### Requirement: ucg polish lane 支持 free 字段

ucg 侧 polish lane Admin 读写 MUST 支持与 voice 业务 lane 同语义的 `freeProvider`/`freeModel`（可空）。

#### Scenario: 保存 polish free

- **WHEN** 管理员保存 polish free 为允许列表内模型
- **THEN** 后续非 premium 润笔 MUST 能读到该配置

---

### Requirement: aimodel 统一入口 SHALL 默认关闭上游 thinking

凡经 `aimodel.Invoke` / `InvokeStream` / `InvokeWithHeldProfile` / `InvokeStreamWithHeldProfile` 发起的 chat/completions 请求，若调用方未将 `ChatRequest.ThinkingEnabled` 设为 `true`，aimodel MUST 按 provider 规则关闭 thinking（智谱 MUST 显式 `disabled`）。该默认语义 MUST 适用于所有 lane（含 `voiceUnderstanding`、`polish`、`simText`、`simVision` 等），MUST NOT 依赖上游模型默认 thinking 行为。

#### Scenario: 润笔未 opt-in

- **WHEN** `PolishPostText` 调用 `aimodel.Invoke(LanePolish, ...)` 且未设置 `ThinkingEnabled`
- **THEN** 智谱请求 MUST 显式 `thinking.type=disabled`

#### Scenario: voiceUnderstanding 闲聊

- **WHEN** 喂养语音 LLM 经 `LaneVoiceUnderstanding` 调用且未设置 `ThinkingEnabled`
- **THEN** 对智谱 provider 的请求 MUST 显式 `thinking.type=disabled`

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

### Requirement: careAlert 车道参与并发闸门

`careAlert` lane MUST 注册到与其它 lane 相同的 aimodel 并发闸门机制；`maxInFlight`/`maxWaiters` MUST 生效。同一 `provider+model` 跨 lane 共享池的既有语义 MUST 保持（含 free 模型若指向相同 provider+model）。

#### Scenario: careAlert 队列满

- **WHEN** careAlert 闸门等待队列已满且新的 premium 生成请求需要 Acquire
- **THEN** 服务 MUST 按既有 50301（或该路径等价错误）拒绝，MUST NOT 调用 Python 分析，MUST NOT consume `care_alert`

---

### Requirement: voice AI 额度与 lane_entitlement MUST 删除

voice-service MUST 删除 `ai_quota_*` 业务读写、App/Admin AI 额度 API、`lane_entitlement`（VIP∨额度∨硬件 → premium）及 care/growth 日限对 quota 的依赖。相关 Redis usage 键写入路径 MUST 停止。UCG 库中的 polish 额度删除见 `retire-ucg-polish`；本 requirement 覆盖 voice 库额度。

#### Scenario: voice App 额度 API 下线

- **WHEN** 客户端请求原 `GET /voice/app/api/ai-quota`
- **THEN** 服务端 MUST NOT 再返回 voiceAi/clinicAi/careAlert 额度业务响应（路由已删除或未注册）

### Requirement: llm_lane_config 与 Hub 选模 MUST 删除

系统 MUST 删除 voice `llm_lane_config` 的 Ensure/Admin GET-PUT、Hub「AI 模型与并发」中 Voice/UCG/Sim lane 编辑能力，以及业务请求向 Python/上游携带 provider/model 的选模逻辑。运维 MUST NOT 再通过该页配置理解/clinic/care/polish/sim 模型。

#### Scenario: llm-lanes Admin API 下线

- **WHEN** 运维请求原 `/voice/admin/api/llm-lanes` 或等价选模 Admin API
- **THEN** 服务端 MUST NOT 再提供可写的多 lane 选模配置接口

### Requirement: aimodel 多 lane 中台 MUST 退役

在 clinic/care/polish/sim/understanding 选模调用方删除后，系统 MUST 删除或使 `internal/services/aimodel` 不再被业务进程引用（含 `Acquire`/`Invoke*`/`ProfileStore` 注册）。`media_gen` 等仅 sim 使用的生成封装 MUST 随 sim 删除。

#### Scenario: 业务包不再 import aimodel

- **WHEN** 对 `internal/services/voice`、`internal/services/ucg`、`internal/services/simuser`（若仍存在）做 import 检查
- **THEN** MUST NOT 再存在对 `hello/internal/services/aimodel` 的业务引用（包本身可删除）

---

### Requirement: sim-user-service SHALL persist sim LLM lanes in database

`sim-user-service` MUST 在 `SIM_DB_LINK` 对应库维护表 **`sim_llm_lane_config`**（或语义等价名），主键 `lane`，取值 `simText`、`simVision`、`simImageGen`、`simVideoGen`。每行 MUST 含 `provider`、`model`、`max_in_flight`、`max_waiters`、`timeout_sec`（可选）、`updated_at`、`updated_by`。

`SimLLMLaneStore.Load(lane)` MUST 按优先级：**DB 行** > **环境变量**（`SIM_LLM_*`，仅 seed/迁移）> **代码默认种子**。EnsureSchema MUST 在无行时写入默认种子（与当前 env 默认语义一致）。

#### Scenario: 新环境 DB 种子

- **WHEN** sim 库尚无 `sim_llm_lane_config` 行且进程首次 EnsureSchema
- **THEN** 四 lane MUST 存在默认行且 Admin GET MUST 可读

#### Scenario: DB 覆盖 env

- **WHEN** DB 行与 `SIM_LLM_*` env 冲突
- **THEN** 运行时 MUST 使用 DB 值

### Requirement: sim-user-service SHALL provide sim LLM lane Admin API

sim-user-service MUST 提供 `GET /sim/admin/api/llm-lanes` 与 `PUT /sim/admin/api/llm-lanes`，鉴权 MUST 与现有 sim-admin 一致（Header `X-Admin-Password`）。响应与请求 MUST 含四 lane 子对象，各含 `provider`、`model`、`maxInFlight`、`maxWaiters`、`updatedAt`、`updatedBy`。GET MUST 返回 provider→model allowlist（含 zhipu 生图/生视频 model）。PUT MUST 校验 allowlist 与边界（`maxInFlight>=1`，`maxWaiters>=0`）。PUT 成功后 MUST 调用 `aimodel.InvalidateLaneCache()` 且 MUST NOT 触发 scheduler reload。

#### Scenario: 管理员读取 sim lanes

- **WHEN** 已鉴权 GET `/sim/admin/api/llm-lanes`
- **THEN** 响应 MUST 含四 lane 配置与 allowlist

#### Scenario: 管理员更新 simText 并发

- **WHEN** PUT `simText.maxInFlight=2` 且 model 在 allowlist
- **THEN** sim-service MUST 持久化且 MUST 失效 lane 缓存

#### Scenario: sim LLM PUT 不 reload scheduler

- **WHEN** 仅 PUT llm-lanes 成功
- **THEN** scheduler goroutine MUST NOT 因该 PUT 而 Stop/Start

### Requirement: aimodel allowlist SHALL include sim image and video generation models

`internal/services/aimodel` 的 `ProviderModels`（zhipu）MUST 包含 `cogview-3-flash`、`cogvideox-flash`（及 normalize 大小写变体），供 simImageGen/simVideoGen Admin PUT 校验通过。

#### Scenario: Admin 配置 sim 生图 model

- **WHEN** PUT simImageGen model=`cogview-3-flash`
- **THEN** 校验 MUST 通过且 MUST 持久化

### Requirement: sim LLM lane Admin API MUST NOT count toward App usage stats

`/sim/admin/api/llm-lanes` MUST NOT 计入 gateway-app App API 使用统计。

#### Scenario: sim llm-lanes PUT 不计 usage

- **WHEN** 管理员 PUT llm-lanes 成功
- **THEN** usage 统计 MUST NOT 递增

---

### Requirement: ucg-service SHALL be authoritative for polish quota configuration and usage

`ucg-service` MUST 在 **`ai_voice_ucg`** 库（GoFrame `database.default`，连接 `UCG_DB_LINK`）维护 **`polish`** feature 的 AI 月度额度全局默认与 per-wxId override。全局默认 MUST 包含 `polishMonthlyLimit`（初始 **5**）；Admin 可修改。per-wxId override MAY 覆盖 polish；未 override MUST 回退全局默认。月度用量 MUST 存 Redis，键格式 **`ai:usage:polish:{wxId}:{YYYYMM}`**，`YYYYMM` MUST 按 `Asia/Shanghai` 自然月生成。ucg-service MUST NOT 将 polish 配额配置或用量写入 device 或 voice 库表；MUST NOT 再转发 device internal ai-quota API。

#### Scenario: 全局润笔默认配置

- **WHEN** 管理员将全局 `polishMonthlyLimit` 设为 10
- **THEN** 无 override 的用户润笔上限 SHALL 为 10

#### Scenario: 单人 override 覆盖润笔

- **WHEN** wxId=1001 的 override 为 `polishMonthlyLimit=20`
- **THEN** 该用户润笔上限 SHALL 为 20

### Requirement: ucg polish SHALL pre-check and consume quota locally

`POST /ucg/app/api/posts/polish` MUST 在调用上游 LLM（`LanePolish`）前于 **ucg-service 进程内**执行 polish check；若 `allowed=false` MUST 返回 code **40302** 与 message **「本月额度已用完」** 且 MUST NOT 调用上游。check 通过后 MUST 经 polish lane 闸门；队列满 MUST 返回 **50301**。上游成功返回有效正文后 MUST 于本进程 consume。参数错误、未配置 AI、上游失败、50301 MUST NOT 调用 consume。

#### Scenario: 额度用尽

- **WHEN** 用户润笔 check 得到 used=5、limit=5
- **THEN** API SHALL 返回 40302 与「本月额度已用完」且 SHALL NOT 请求上游

#### Scenario: 上游失败不扣减

- **WHEN** check 通过但上游返回 5xx
- **THEN** 系统 SHALL NOT 调用 consume 且 used SHALL 不变

#### Scenario: 队列满

- **WHEN** check 通过但 polish 闸门队列满
- **THEN** API SHALL 返回 50301 且 MUST NOT consume

### Requirement: ucg App quota read API SHALL expose polish only

`GET /ucg/app/api/ai-quota` MUST 要求有效 Bearer 且 `X-Internal-Wx-Id > 0`（经 gateway-app 注入）。响应 MUST 为 `{ polish: { used, limit, degraded } }`，对应当月上海时区桶；**`degraded` MUST 为 `used >= limit`**。本接口 MUST NOT 返回 `voiceAi` 或 `clinicAi` 字段。

#### Scenario: 登录用户查询润笔额度

- **WHEN** wxId=1001 请求 `/ucg/app/api/ai-quota` 且当月润笔已用 2、上限 5
- **THEN** `polish.used` SHALL 为 2、`polish.limit` SHALL 为 5 且 `polish.degraded` SHALL 为 false

#### Scenario: 额度用尽 degraded 标记

- **WHEN** wxId=1001 当月润笔已用 5、上限 5
- **THEN** `polish.degraded` SHALL 为 true 且 `polish.used` SHALL 为 5

#### Scenario: wxId=0 拒绝

- **WHEN** 请求携带 wxId=0
- **THEN** 系统 SHALL 返回未授权/无效身份错误

### Requirement: ucg admin SHALL configure polish quota locally only

ucg-service MUST 提供 `GET/PUT /ucg/admin/api/ai-quota/default` 与 `GET/PUT /ucg/admin/api/ai-quota/user`（query/body 含 `wxId`），认证 MUST 为 Header `X-Admin-Password` 等于 `ucg.admin.password`。ucg-service MUST **本地**持久化 polish 配置至 `ai_voice_ucg`，MUST NOT 转发 device。PUT default MUST 仅接受 `polishMonthlyLimit`（正整数）。PUT user MUST 仅接受 optional `polishMonthlyLimit`；空值 SHALL 表示清除 override。Admin API MUST NOT 接受或返回 `voiceAiMonthlyLimit`、`clinicAiMonthlyLimit`。

#### Scenario: 管理员修改全局润笔默认

- **WHEN** 管理员 PUT default 为 polish=10
- **THEN** ucg 权威配置 SHALL 更新且新用户 check SHALL 使用 limit=10

#### Scenario: ucg admin 口令错误

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 修改配置

### Requirement: polish 额度与 VIP 共同策略

ucg 润笔（`polish` feature）MUST 使用与 voice 相同的 VIP∪额度共同策略：VIP 时 Allowed=true 且成功不计次；非 VIP 额度内使用 `polish` lane 正式模型并在成功后 consume；非 VIP 额尽时 MUST 使用该 lane 的 free 配置，free 为空时 MUST NOT 使用硬编码 `DegradedPolishProfile` 作为真相源（按 design 以无覆盖/默认上游语义处理，且 MUST 可观测）。

#### Scenario: VIP 润笔不计次

- **WHEN** VIP 用户润笔成功返回正文
- **THEN** `polish` 月用量 MUST NOT 增加，响应 MUST NOT 再以额度用尽为由失败

#### Scenario: 非 VIP 额尽走 free

- **WHEN** 非 VIP 且 polish 额度用尽，且 Admin 配置了 polish free 模型
- **THEN** 润笔 MUST 使用 free 模型完成（成功时可不计次），MUST NOT 返回 40302 阻断

---

### Requirement: voice-service SHALL be authoritative for voice_ai and clinic_ai quota configuration and usage

`voice-service` MUST 在 **`ai_voice_voice`** 库（GoFrame `database. default`，连接 `VOICE_DB_LINK`）维护 AI 月度额度全局默认与 per-wxId override，并 MUST 为 **`voice_ai`** 与 **`clinic_ai`** 两个 feature 独立计数。 全局默认 MUST 包含 `voiceAiMonthlyLimit`（初始 **5**）与 `clinicAiMonthlyLimit`（初始 **30**）； 未 override 的 feature MUST 回退全局默认。 月度用量 MUST 存 Redis，键格式 **`ai:usage:voice_ai:{wxId}:{YYYYMM}`** 与 **`ai:usage:clinic_ai:{wxId}:{YYYYMM}`**，其中 `YYYYMM` MUST 按 `Asia/Shanghai` 自然月生成。 voice-service MUST NOT 将.

#### Scenario: 全局默认独立配置

- **WHEN** 管理员将全局 `voiceAiMonthlyLimit` 设为 5 且 `clinicAiMonthlyLimit` 设为 30
- **THEN** 无 override 的用户喂养 AI 上限 SHALL 为 5、胖宝 AI 上限 SHALL 为 30

#### Scenario: 单人 override 覆盖单 feature

- **WHEN** wxId=1001 的 override 为 `clinicAiMonthlyLimit=50` 且未设 `voiceAiMonthlyLimit`
- **THEN** 该用户胖宝 AI 上限 SHALL 为 50 且喂养 AI SHALL 使用全局默认

### Requirement: voice internal quota check and consume APIs SHALL enforce wxId and feature semantics

`POST /voice/internal/api/ai-quota/check` MUST 要求有效 voice internal 密钥（与 voice-service 其它 internal API 一致）。body MUST 含 `wxId`（正整数）与 `feature`（`voice_ai` | `clinic_ai`）。响应 MUST 含 `allowed`（boolean）、`used`、`limit`。`check` MUST NOT 修改用量。

`POST /voice/internal/api/ai-quota/consume` MUST 在 AI 成功返回后由 voice 本进程调用；成功时 MUST `INCR` 当月用量并返回 `{ used, limit }`。若扣减后 `used > limit`，系统 MUST 回滚该次 INCR 并返回超额错误。

#### Scenario: check 不扣减

- **WHEN** voice 在胖宝提问前调用 check feature=clinic_ai 且 used=29、limit=30
- **THEN** 响应 `allowed=true` 且 Redis 计数 SHALL 仍为 29

#### Scenario: consume 成功扣减

- **WHEN** voice 在 LLM 成功后调用 consume feature=voice_ai 且 used=4、limit=5
- **THEN** used SHALL 变为 5 且响应 used=5

#### Scenario: wxId 无效拒绝

- **WHEN** internal API 收到 wxId=0
- **THEN** 系统 SHALL 返回错误且 MUST NOT 读写的用量

### Requirement: voice App quota read API SHALL expose voiceAi and clinicAi

`GET /voice/app/api/ai-quota` MUST 要求有效 Bearer 且 `X-Internal-Wx-Id > 0`（经 gateway-app 注入）。响应 MUST 为 `{ voiceAi: { used, limit }, clinicAi: { used, limit } }`，对应当月上海时区桶。本接口 MUST NOT 返回 `polish` 字段。

#### Scenario: 登录用户查询 voice 域额度

- **WHEN** wxId=1001 请求 `/voice/app/api/ai-quota` 且当月胖宝已用 5、上限 30
- **THEN** `clinicAi.used` SHALL 为 5 且 `clinicAi.limit` SHALL 为 30

#### Scenario: wxId=0 拒绝

- **WHEN** 请求携带 wxId=0
- **THEN** 系统 SHALL 返回未授权/无效身份错误

### Requirement: voice admin SHALL configure global default and per-user override locally

voice-service MUST 提供 `GET/PUT /voice/admin/api/ai-quota/default` 与 `GET/PUT /voice/admin/api/ai-quota/user`（query/body 含 `wxId`），以及 **`GET /voice/admin/api/ai-quota/users`**（用户额度分页列表，见 ADDED 要求）。 认证 MUST 为 Header `X-Admin-Password` 等于 `voice. voice-service MUST 本地持久化至 `ai_voice_voice`，MUST NOT 转发 device 或 ucg 写配额。 PUT default MUST 接受 `voiceAiMonthlyLimit` 与 `clinicAiMonthlyLimit`（正整数）。 PUT user MUST 接受 optional 两字段； 空值 SHALL 表示清除该 feature override。 运维浏览全量用户额度 MUST 使用 users.

#### Scenario: 管理员修改全局胖宝默认

- **WHEN** 管理员 PUT default 为 voiceAi=5、clinicAi=30
- **THEN** voice 权威配置 SHALL 更新且新用户 check clinic_ai SHALL 使用 limit=30

#### Scenario: voice admin 口令错误

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 修改配置

#### Scenario: 列表 API 与单人写并存

- **WHEN** 管理员先 GET users 再对其中一行 PUT user 修改 voiceAiMonthlyLimit
- **THEN** 后续 GET users 同 wx 行的 `voiceAi.limit` SHALL 反映新有效上限

### Requirement: voice feeding LLM SHALL 经 voiceUnderstanding lane 闸门

voice-service 对 feature `voice_ai` 的额度 check 通过后、调用任何喂养 voice LLM 前 MUST 经 `LaneVoiceUnderstanding` 闸门。队列满 MUST 返回 WS code **50301**。本要求与 `ai-monthly-quota` 中喂养 AI 条款一致，强调全部 LLM 路径（含 casual 流式与成长建议）无遗漏。

#### Scenario: 闲聊流式占用 voiceUnderstanding 闸门

- **WHEN** commit 后进入 `StreamCasualReplyWithBaiduTTS` 的 LLM 段
- **THEN** MUST 使用 voiceUnderstanding profile 的 model 闸门

### Requirement: voice admin SHALL list per-wx effective quota with identity fields

voice-service MUST 提供 `GET /voice/admin/api/ai-quota/users`，认证 MUST 为 Header `X-Admin-Password` 等于 `voice. Query MUST 支持：`page`（从 1）、`pageSize`（默认 20，最大 100）、**`deviceNo`**（可选，按设备号过滤）。 实现 MUST 经 **`DeviceAdmin(). ListWxPage`**（或等价 device 契约）取得真实 wx 分页全集（排除模拟号），MUST NOT 在 voice 进程直查 device/history 库表。 响应 MUST 为分页结构 `{ list, total, page, pageSize }`。 `list[]` 每项 MUST 含：

- `deviceNo`（string）
- `wxId`（int64，对应 wx. `limit` 为有效上限 = 正数 override ∨ 全局默认）
- `clinicAi`：`{ used, limit.

#### Scenario: 分页返回身份与双额度

- **WHEN** 管理员请求 `/voice/admin/api/ai-quota/users?page=1&pageSize=20` 且 device 域存在真实 wx
- **THEN** 响应 `list` 每项 SHALL 含 deviceNo、wxId、account、babyName 以及 voiceAi/clinicAi 的 used 与 limit

#### Scenario: 按 deviceNo 过滤

- **WHEN** 管理员请求 `deviceNo` 等于某已绑定设备号
- **THEN** 返回列表 SHALL 仅包含该设备号匹配的 wx 行（在契约过滤语义下）

#### Scenario: 无 override 时 limit 为全局默认

- **WHEN** 某 wx 无 clinic override 且全局 clinicAiMonthlyLimit=30
- **THEN** 该行 `clinicAi.limit` SHALL 为 30

#### Scenario: 有 override 时 limit 为覆盖值

- **WHEN** wxId=1001 的 voice override 为 10 且当月 voice 已用 3
- **THEN** 该行 `voiceAi.limit` SHALL 为 10 且 `voiceAi.used` SHALL 为 3

#### Scenario: 口令错误拒绝

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 返回用户额度列表

#### Scenario: 保存等于默认则清除 override

- **WHEN** 管理员 PUT user 将某 wx 的 clinicAiMonthlyLimit 设为与当前全局 clinic 默认相同的正整数
- **THEN** 该 wx 的 clinic override SHALL 被清除（后续 limit 跟随全局默认）

### Requirement: voice 域额度与 VIP 共同策略

`voice_ai` 与 `clinic_ai` 的 check/快照/App 状态 MUST 与 VIP 共同决策：VIP 时 MUST `Allowed=true` 且成功路径不计次。非 VIP 时保持月度 used/limit 语义；用尽后 MUST 允许继续走非 premium 选模（free/omit），MUST NOT 再依赖硬编码 Degraded 智谱种子作为唯一降级模型。

#### Scenario: 非 VIP 额度内喂养

- **WHEN** 非 VIP 且 `voice_ai` Allowed，完成喂养 LLM 成功
- **THEN** 系统 MUST 使用 `voiceUnderstanding` 正式模型，并 consume `voice_ai`

#### Scenario: 非 VIP 喂养额尽

- **WHEN** 非 VIP 且 `voice_ai` 用尽
- **THEN** 对话 MUST 仍可继续，选模 MUST 走 `voiceUnderstanding` free（或 omit），且该次成功 MUST NOT consume

### Requirement: tip 挂靠 clinic 额度与 lane

小贴士（TipStream）MUST 使用与 polish/诊疗相同的 VIP∪额度策略，feature 为 **`clinic_ai`**，lane 为 **`clinic`**（含 free）。MUST NOT 在无权益判定的情况下无脑传正式模型。

#### Scenario: tip 与诊疗共享 clinic_ai

- **WHEN** 非 VIP 用户 `clinic_ai` 仍有额度并成功生成 tip
- **THEN** 系统 MUST 按 clinic 正式模调用 Python，并 consume `clinic_ai`

#### Scenario: tip 在 clinic_ai 用尽时

- **WHEN** 非 VIP 且 `clinic_ai` 用尽后请求 tip
- **THEN** 系统 MUST 按 clinic free/omit 调用，MUST NOT 因 40302 单独拒绝 tip（与 clinic degraded 放行策略对齐）

### Requirement: 硬件语音通道特权

经 gateway 设备直连的 `/voice/chat/ws` 以及 MCP/internal text chat 设备通道，MUST 标记硬件特权并按 premium 选模（`voiceUnderstanding` 正式模），MUST NOT 计次。该特权 MUST NOT 自动授予 clinic WS、care-alert、tip、polish。

#### Scenario: MCP 文本对话走正式模不计次

- **WHEN** mcp-service 以配置 deviceNo 调用 internal text chat 成功
- **THEN** 意图分析 MUST 使用 `voiceUnderstanding` 正式模型，MUST NOT consume `voice_ai`

---
