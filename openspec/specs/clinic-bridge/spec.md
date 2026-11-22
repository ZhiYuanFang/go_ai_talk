# clinic-bridge

## Purpose

胖宝诊疗、tip 反馈退役、clinic 流式宿主与 tip 契约瘦身的行为规格。 Clinic bridge between voice host and Flutter without tip product flywheels.

<!-- migrated-from: clinic-tip-feedback-flutter, clinic-tip-feedback-voice-host, pangbao-ai-clinic, pangbao-ai-clinic-flutter, remove-clinic-tip-feedback, remove-device-tip, retire-clinic-growth-suggest, tip-contract-slim, tip-flutter-sse-go-dialect, tip-generate-voice-host, tip-python-derived-context, tip-streaming -->

## Requirements

### Requirement: tip 反馈字段与 URL 对齐 Go API

Flutter tip 客户端 SHALL 以 `POST /device/api/tip/feedback` 提交反馈，JSON Body 使用 `answerId`（string）与 `feedback`（int：1 或 -1），MUST NOT 继续使用 `tipId` / `feedbackResult`。

#### Scenario: tip submitFeedback 对齐
- **WHEN** 用户在首页小贴士面板点击 thumbs up/down 且本地已有非空 `answerId`
- **THEN** 客户端 SHALL 向 `/device/api/tip/feedback` 发送 `{ "answerId": "<id>", "feedback": 1|-1 }`
- **AND** 已反馈后 MUST NOT 允许再次更改

#### Scenario: 无 answerId 不提交
- **WHEN** 本地 `answerId` 为空或缺失
- **THEN** 客户端 MUST NOT 发起 tip feedback HTTP 请求

### Requirement: tip 流式 done 写入 answerId

Flutter tip 流式接收完成后 SHALL 将服务端返回的 `answerId` 写入 tip 状态，供反馈提交使用。该能力依赖包 B tip SSE；若 B 未完成，本变更 SHALL 补最小接线或在实现报告中标明 tip 反馈 blocker。

#### Scenario: done 携带 answerId
- **WHEN** tip SSE 流结束且 done 帧含 `answerId`
- **THEN** tip 状态 SHALL 保存该字符串，反馈按钮可用

#### Scenario: 包 B 阻塞时的降级
- **WHEN** 包 B 未完成导致 tip 无法获得 `answerId`
- **THEN** 实现报告 SHALL 明确 tip 反馈 blocker
- **AND** clinic 反馈与 Go Bind/反代/skip MUST 仍可完成

### Requirement: clinic 反馈 UI 与 HTTP

胖宝诊疗屏（`pangbao_ai_screen`）SHALL 在助手回答完成且存在 `answerId` 时展示 thumbs up/down，并经 HTTP `POST /device/api/clinic/feedback` 提交 `{answerId, feedback}`。

#### Scenario: clinic 回答完成后可反馈
- **WHEN** 某条助手消息已 `answer_done` 且 `answerId` 非空，且尚未反馈
- **THEN** UI SHALL 展示反馈按钮
- **AND** 用户点击后 SHALL POST `/device/api/clinic/feedback` 并标记该条已反馈

#### Scenario: 无 answerId 或已反馈
- **WHEN** `answerId` 为空，或该条 `feedbackGiven` 已为 true
- **THEN** UI MUST NOT 再次提交 clinic feedback

---

### Requirement: voice-service 注册 clinic/tip 反馈控制器

voice-service SHALL 在 HTTP 注册中 Bind `DeviceClinicFeedbackController`（或等价控制器），对外暴露已有 `POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback`。

#### Scenario: clinic feedback 可达
- **WHEN** 已鉴权客户端向 voice-service（经 gateway）发送 `POST /device/api/clinic/feedback`，Body 含合法 `answerId` 与 `feedback`（1 或 -1）
- **THEN** 控制器 SHALL 调用 `PythonAIClient.ClinicFeedback` 并将 Python 结果映射为 `{code, message, data}` 响应

#### Scenario: tip feedback 可达
- **WHEN** 已鉴权客户端向 voice-service（经 gateway）发送 `POST /device/api/tip/feedback`，Body 含合法 `answerId` 与 `feedback`（1 或 -1）
- **THEN** 控制器 SHALL 调用 `PythonAIClient.TipFeedback` 并将 Python 结果映射为 `{code, message, data}` 响应

#### Scenario: 非法 feedback 值
- **WHEN** 请求 Body 中 `feedback` 不是 1 或 -1
- **THEN** 控制器 SHALL 返回业务错误（如 code=400）且 MUST NOT 调用 Python

### Requirement: gateway 反代 clinic/tip API 到 VOICE

gateway-app SHALL 将 `/device/api/clinic/*` 与 `/device/api/tip/*` 反代至 `VOICE_API_PROXY`（voice 域路由模式），MUST NOT 反代至 device-service。

#### Scenario: clinic/tip API 走 voice 上游
- **WHEN** gateway-app 收到匹配 `/device/api/clinic/*` 或 `/device/api/tip/*` 的请求且 voice 代理已启用
- **THEN** 请求 SHALL 经 voice 反代中间件转发到 voice-service
- **AND** MUST NOT 进入 device 域 `/device/app/api/feedback/*` 反代路径

#### Scenario: 与 device 用户反馈路径隔离
- **WHEN** 客户端请求 `/device/app/api/feedback/*`
- **THEN** 行为 SHALL 保持既有 device 域反代，不受本变更影响

### Requirement: feedback 接口不计入 usage 统计

负责人已确认：`POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback` MUST NOT 计入 gateway-app App API 使用统计。实现 SHALL 在 `maintenance_skip.go` 以精确 `METHOD + path` 排除二者。

#### Scenario: clinic feedback 被 skip
- **WHEN** gateway-app 处理 `POST /device/api/clinic/feedback`
- **THEN** usagestats SHALL 将其视为维护型 API 并跳过计数

#### Scenario: tip feedback 被 skip
- **WHEN** gateway-app 处理 `POST /device/api/tip/feedback`
- **THEN** usagestats SHALL 将其视为维护型 API 并跳过计数

#### Scenario: tip generate 不在本包改统计策略
- **WHEN** 评估 `POST /device/tip/generate` 的 usage 策略
- **THEN** 本变更 MUST NOT 将其写入 `maintenance_skip.go`（统计属包 B）

### Requirement: 反馈路径与鉴权保持既有契约

反馈 HTTP 路径 MUST 保持为 `/device/api/clinic/feedback` 与 `/device/api/tip/feedback`；经 gateway-app 暴露时 MUST 要求登录（不加入 Bearer exempt）。

#### Scenario: 路径不变
- **WHEN** 查阅 `api/v1` 的 `g.Meta` 或对外文档
- **THEN** clinic/tip 反馈 path 仍为上述两路径，无破坏性改名

#### Scenario: 未登录被拒
- **WHEN** 无有效 Bearer 的客户端请求上述反馈 POST
- **THEN** gateway-app SHALL 按既有鉴权规则拒绝（401），不得因本变更加入白名单

---

### Requirement: voice-service SHALL 提供胖宝 AI 诊室 WebSocket handler

`voice-service` MUST 在路径 `/voice/clinic/ws` 注册 WebSocket handler（`BindHandler`），处理经 gateway-app 透传而来的客户端文本提问并流式返回 **clinic lane 配置的上游 LLM** 回答（默认种子为智谱 `glm-4.1v-thinking-flash`，Admin 可切回 DeepSeek 等）。用户可见功能名称为 **胖宝诊疗**；实现与配置仍使用 `clinic` / `clinic_ai` / `voice:clinic:*` 命名。该路径为**集群内业务端点**；App 对外入口 MUST 为 gateway-app-server 同路径透传（见 `gateway-ws-edge-proxy`）。实现 MUST NOT 将连接注册到 `VoiceWSManager`。实现 MUST NOT 提供 TTS 或音频上行能力（MVP 纯文本）。

#### Scenario: 经 gateway-app 透传后握手成功

- **WHEN** 客户端经 gateway-app 透传对 voice-service `/voice/clinic/ws` 完成 WebSocket Upgrade
- **THEN** voice-service SHALL 接受连接并等待首帧 `auth` JSON

### Requirement: Clinic WebSocket SHALL 以 wxId 为主键绑定身份

`/voice/clinic/ws` 的连接、限流与额度维度 MUST 以 **`wx.id`（正整数）** 为主键，与 `/voice/chat/ws` 以 `deviceNo` 注册 `VoiceWSManager` 的行为 MUST 显式不同。实现 MUST NOT 使用 `deviceNo` 反查 wxId 作为 clinic 鉴权路径（即 MUST NOT 调用 `ResolveVoiceWxID` 的 deviceNo fallback）。`deviceNo` MUST 与 JWT `device_no` claim 一致，并 MUST 透传给 Python Clinic 请求的 `device_no`；MUST NOT 再用于 Go 侧 history 摘要或宝宝画像拉取。

#### Scenario: 首帧 auth 绑定 wxId

- **WHEN** 客户端握手后发送 `type=auth` 且 JWT 含 `sub=1001` 与有效 `device_no`
- **THEN** 服务端 SHALL 解析 `wxId=1001` 并返回 `auth_ok`
- **AND** 后续限流/额度 MUST 使用 wxId=1001

#### Scenario: 未登录拒绝

- **WHEN** 客户端 `auth` 帧 JWT 无效或 `sub≤0`
- **THEN** 服务端 SHALL 返回 `error` code **40301** 且 MUST NOT 进入 `question` 处理

#### Scenario: deviceNo 与 JWT 不一致

- **WHEN** `auth` 帧 `deviceNo` 与 JWT `device_no` claim 不一致
- **THEN** 服务端 SHALL 返回参数错误 `error` 帧且 MUST NOT 返回 `auth_ok`

#### Scenario: 与 voice ball 连接互不踢线

- **WHEN** 同一 `deviceNo` 已建立 `/voice/chat/ws` 且同一 `wxId` 另建 `/voice/clinic/ws`
- **THEN** 两条连接 SHALL 同时保持，且 clinic 连接 MUST NOT 触发 `VoiceWSManager` 替换逻辑

### Requirement: Clinic WebSocket SHALL 使用规定的帧协议

握手成功后，客户端首帧 MUST 为 `type=auth`（含 `accessToken` 与 `deviceNo`）。 `auth_ok` 之后，客户端上行 MUST 支持 `type=question` 帧（含非空 `text` 与非空 **`turnId`** UUID）与 **`type=cancel`** 帧（含非空 **`turnId`**）。 服务端下行 MUST 支持 `auth_ok`、`thinking_delta`、`answer_delta`、`answer_done`、**`turn_cancelled`**、`error` 六种 `type`，MUST NOT 下发 `session_sync`。 流式下行帧（`thinking_delta`、`answer_delta`、`answer_done`）**MUST** 含与当前 turn 一致的 **`turnId`**。 `turn_cancelled` MUST 含 **`turnId`** 与 **`reason`**，取值为.

#### Scenario: 流式 thinking 与 answer 携带 turnId

- **WHEN** 客户端发送 `question` 且 `turnId=uuid-A` 且 LLM 流式返回 reasoning 与 content
- **THEN** 服务端 MUST 推送的 `thinking_delta`、`answer_delta` 与最终 `answer_done` 均 MUST 含 `turnId=uuid-A`

#### Scenario: auth 前拒绝 question

- **WHEN** 客户端未发送 `auth` 或 `auth` 未成功即发送 `question`
- **THEN** 服务端 SHALL 返回 `error` 帧且 MUST NOT 调用 LLM

#### Scenario: 空问题或缺少 turnId 拒绝

- **WHEN** 客户端发送 `question` 且 `text` 为空或仅空白，或缺少/空 `turnId`
- **THEN** 服务端 SHALL 返回 `error` 帧且 MUST NOT 调用 LLM

#### Scenario: cancel 中断当前 turn

- **WHEN** 客户端在 turn `uuid-A` 流式进行中发送 `cancel` 且 `turnId=uuid-A`
- **THEN** 服务端 MUST 取消该 turn 的 LLM 上下文
- **AND** MUST 下发 `turn_cancelled` 且 `turnId=uuid-A` 且 `reason=cancelled`
- **AND** MUST NOT 下发 `answer_done` 且 MUST NOT consume `clinic_ai`

#### Scenario: 新 question supersede 上一 turn

- **WHEN** turn `uuid-A` 仍在流式进行中且客户端发送新 `question` 且 `turnId=uuid-B`
- **THEN** 服务端 MUST 取消 turn A 的 LLM 上下文
- **AND** MUST 下发 `turn_cancelled` 且 `turnId=uuid-A` 且 `reason=superseded`
- **AND** MUST 开始处理 turn B

#### Scenario: WS 断开取消 active turn

- **WHEN** 连接存在 active LLM turn 且 WebSocket 读循环因关闭或错误退出
- **THEN** 服务端 MUST 取消该 turn 的 LLM 上下文
- **AND** MUST NOT consume `clinic_ai`

### Requirement: Clinic LLM SHALL 使用 deepseek-v4-pro 并启用 thinking

Clinic LLM 调用 MUST 使用模型 `deepseek-v4-pro`，MUST 启用 thinking（`extra_body.thinking` 或等价配置，`reasoning_effort: high`）。LLM 请求超时 MUST 为 **120 秒**，配置源 MUST 为 `config.voice-service.yaml` 的 `aiClinic` 块，MUST NOT 依赖 `voice-chat.shared.yaml` 的 voice ball 超时。

#### Scenario: thinking 流映射

- **WHEN** DeepSeek 返回 reasoning 流
- **THEN** 服务端 MUST 将其映射为 `thinking_delta` 帧

### Requirement: Clinic SHALL 强制执行 clinic_ai 月度额度（per wxId）

voice-service 在调用 Clinic LLM 前 MUST 使用 auth 已绑定的 `wxId>0`； `wxId≤0` MUST 返回 `error` code **40301**。 LLM 调用前 MUST 对 feature `clinic_ai` 以该 wxId 执行 check。 若 `allowed=true`，MUST 经 `LoadProfile(LaneClinic)` 调用 LLM； **仅** turn 以 **`answer_done` 成功结束** 时 MUST 以同一 wxId consume。 **若 `allowed=false`（`used >= limit`）**，MUST **NOT** 返回 code **40302**； MUST 经 **degraded** 路径调用 LLM，强制 `DefaultSeedProfile(LaneClinic)`（智谱 **`glm-4. 1v-thinking-flash`**），且 **`answer_done` 成功时 MUST NOT.

#### Scenario: 未登录

- **WHEN** wxId 解析为 0 且用户发送 `question`
- **THEN** WS SHALL 返回 40301 且 MUST NOT 调用 LLM

#### Scenario: clinic_ai 额度用尽 degraded 问答

- **WHEN** check 得到 clinic_ai used=limit
- **THEN** WS SHALL **NOT** 返回 40302
- **AND** SHALL 经 degraded 路径流式返回答案
- **AND** `answer_done` 后 MUST NOT consume clinic_ai

#### Scenario: 额度内成功完成扣减

- **WHEN** check 得到 used < limit 且 turn 完整流式结束并下发 `answer_done`
- **THEN** 系统 MUST consume `clinic_ai` 一次

#### Scenario: 用户 cancel 不扣额度

- **WHEN** turn 在流式过程中被 `cancel` 或 `superseded` 结束且未收到 `answer_done`
- **THEN** 系统 MUST NOT 对该 turn 调用 consume `clinic_ai`（含 degraded 路径）

### Requirement: Clinic SHALL 实施 Redis 限流（per wxId）

系统 MUST 对 clinic 提问路径实施 Redis 限流，键 **`voice:clinic:rate:{wxId}`**。限流计数 **MUST** 在 **`answer_done` 成功** 后递增；**cancelled**、**superseded**、**disconnected** 或失败 turn **MUST NOT** 递增限流计数。处理新 `question` 前 MUST 检查当前窗口计数；超限时 MUST 返回 WS `error` code **42901** 且 MUST NOT 调用 LLM。42901 与 **50301**（LLM lane 队列满）为不同语义：50301 MUST 在额度与 42901 检查通过后、调用上游前返回。

#### Scenario: 短时间频繁提问

- **WHEN** 同一 wxId 在限流窗口内已成功完成（`answer_done`）次数超过阈值
- **THEN** 下一次 `question` SHALL 返回 42901 且 MUST NOT 调用 LLM

#### Scenario: supersede 未完成 turn 不计入限流

- **WHEN** 用户在窗口内多次改问但均未产生 `answer_done`
- **THEN** 限流计数 MUST NOT 因 supersede 而额外递增

#### Scenario: 队列满返回 50301

- **WHEN** 42901 与 clinic_ai 检查通过但 clinic lane 闸门队列满
- **THEN** WS MUST 返回 50301 且 MUST NOT 调用 LLM

### Requirement: Clinic WS handler SHALL 非阻塞处理 question 并支持显式取消

`/voice/clinic/ws` 读循环 MUST NOT 同步阻塞在单条 `HandleQuestion` 直至 LLM 结束。每条合法 `question` MUST 在独立 goroutine 中处理，且连接 MUST 维护 at most one **active turn** 及其 `context.CancelFunc`。收到匹配 active turn 的 `cancel`、收到新 `question`（supersede）或连接关闭时 MUST 调用 cancel 中断 LLM/Python 流。用户显式 `cancel` 与 supersede MUST 分别对应下行 `turn_cancelled` 的 `reason=cancelled` 与 `reason=superseded`；连接关闭 MUST 仅取消上下文且 MUST NOT 依赖下行 `turn_cancelled`。

#### Scenario: 读循环可并发接收 cancel

- **WHEN** LLM 流式进行中且读循环收到 `cancel` 帧
- **THEN** handler MUST 在不等待 LLM 自然结束的情况下处理 cancel

#### Scenario: HandleQuestion 尊重 turn context

- **WHEN** turn context 已被 cancel
- **THEN** `HandleQuestion` MUST 停止 LLM 读流且 MUST NOT 写 `answer_done`

#### Scenario: cancel 在 thinking 阶段即时生效

- **WHEN** 服务端已开始下发 `thinking_delta` 且尚未进入 `answer_delta`，客户端发送匹配的 `cancel`
- **THEN** 服务端 MUST 取消 turn context 并下发 `turn_cancelled` 且 `reason=cancelled`

#### Scenario: cancel 在 answer 阶段即时生效

- **WHEN** 服务端已开始下发 `answer_delta` 且尚未 `answer_done`，客户端发送匹配的 `cancel`
- **THEN** 服务端 MUST 取消 turn context 并下发 `turn_cancelled` 且 `reason=cancelled`
- **AND** MUST NOT 随后下发该 turn 的 `answer_done`

### Requirement: Clinic LLM SHALL 经 ChatRequest 显式 opt-in thinking

胖宝诊疗 LLM 调用（`streamClinicLLM` / `streamClinicLLMHeld`）在构造 `aimodel.ChatRequest` 时 MUST 设置 `ThinkingEnabled=true`（或后续等价的显式 enabled 字段），以确保经 aimodel 层发送 `thinking: enabled` 而非继承默认 disabled。Clinic WS 下行 `thinking_delta` / `answer_delta` 协议 MUST 保持不变。

#### Scenario: clinic 流式仍返回 thinking

- **WHEN** 客户端发送合法 `question` 且 clinic lane 上游返回 reasoning 流
- **THEN** voice-service MUST 继续映射为 `thinking_delta` 帧且上游请求 MUST 为 thinking enabled

#### Scenario: clinic 不受 aimodel 默认 disabled 影响

- **WHEN** aimodel 全局默认 thinking 为 disabled
- **THEN** clinic 路径 MUST 仍为 enabled 且 MUST NOT 因零值误关 thinking

### Requirement: Clinic App quota read API SHALL expose clinic_ai degraded flag

`GET /voice/app/api/ai-quota` 响应中 `clinicAi` 对象 MUST 含 **`degraded`** 布尔字段；当 `clinic_ai` 的 `used >= limit` 时 MUST 为 `true`。`voiceAi` 对象 MAY 含同名字段（`used >= limit` 时为 true），但 voice_ai 用尽时 WS 仍 MUST 返回 40302，行为不变。

#### Scenario: clinic 额度用尽 API 标记

- **WHEN** wxId=1001 的 clinic_ai used=10、limit=10
- **THEN** `clinicAi.degraded` SHALL 为 true

#### Scenario: clinic 额度内

- **WHEN** wxId=1001 的 clinic_ai used=3、limit=10
- **THEN** `clinicAi.degraded` SHALL 为 false

### Requirement: Clinic LLM SHALL 经 clinic lane 与可配置 provider 调用

胖宝 `question` 处理中的 LLM MUST 经 `aimodel.InvokeStream(LaneClinic)` 调用；provider 与 model MUST 来自 Admin/DB profile，MUST NOT 硬编码 DeepSeek。thinking 流式下行语义（`thinking_delta` / `answer_delta`）MUST 保持与现有 WS 协议一致。

#### Scenario: 默认种子为智谱 thinking 模型

- **WHEN** 新部署且 DB 为种子 A 默认值
- **THEN** clinic lane MUST 使用 `provider=zhipu` 且 `model=glm-4.1v-thinking-flash`

#### Scenario: Admin 切回 DeepSeek

- **WHEN** Admin 将 clinic profile 改为 `provider=deepseek`、`model=deepseek-v4-pro`
- **THEN** 下一笔 `question` MUST 调用 DeepSeek 适配器且 MUST 使用 `deepseek-v4-pro` 闸门池

### Requirement: Clinic LLM stream SHALL emit answer_delta after thinking phase

当 Clinic 调用上游且 `ThinkingEnabled=true` 时，aimodel 流式层在已接收至少一个 reasoning/thinking 分片后，对上游单独到达的 `content` 分片 MUST 路由为 `answer`（触发 `OnAnswerDelta` 与 `answer_delta` WS 帧），MUST NOT 仅写入未订阅的 content 通道。

#### Scenario: reasoning 结束后 content 分片到达

- **WHEN** 上游先流式 `reasoning_content` 再流式纯 `content` 分片
- **THEN** voice-service MUST 向客户端发送 `answer_delta` 帧
- **AND** `answer_done` 的 `answer` 字段 MUST 非空（与上游正文一致）

#### Scenario: 非 thinking 闲聊路径不变

- **WHEN** `ThinkingEnabled=false` 且上游仅发送 `content`
- **THEN** 流式层 MUST 仍将 content 路由至 `OnContentDelta`（MUST NOT 误发 answer_delta）

### Requirement: Clinic Go path SHALL NOT cache or sync conversation history

voice-service Clinic WS MUST NOT 在 Redis 或其他 Go 进程内存储多轮对话 turns，MUST NOT 向客户端下发用于恢复 UI 历史的会话同步帧。对话 UI 历史由客户端本地负责；agent 多轮上下文由 Python 智能服务负责。Go Clinic 路径在成功 `auth_ok` 后 MUST 仅处理 `question`/`cancel` 与流式下行（`thinking_delta`/`answer_delta`/`answer_done`/`turn_cancelled`/`error`），并对 Python 调用仅透传提问所需字段（至少含 `question`、`device_no`、模型配置）。

#### Scenario: auth_ok 后无会话同步帧

- **WHEN** 客户端完成 `auth` 并收到 `auth_ok`
- **THEN** 服务端 MUST NOT 下发 `type=session_sync`
- **AND** 客户端 MAY 立即发送 `question`

#### Scenario: answer_done 不写 Go 会话缓存

- **WHEN** turn 以 `answer_done` 成功结束
- **THEN** voice-service MUST NOT 写入 `voice:clinic:session:{wxId}` 或等价 Go 侧对话缓存键

### Requirement: Clinic Go path SHALL NOT build feeding summary or baby profile for LLM

voice-service 在处理 Clinic `question` 时 MUST NOT 经 `DeviceHistory` 聚合近 7 天喂养摘要，MUST NOT 读写 `voice:clinic:summary:*`，MUST NOT 经 `DeviceProfile` 拉取宝宝画像并注入 Go 侧 LLM system context。喂养上下文与宝宝画像若需要，MUST 由 Python Clinic/companion 路径自行获取。

#### Scenario: question 不因摘要失败而阻断

- **WHEN** history 服务不可用
- **THEN** Clinic `question` 路径 MUST NOT 仅因「喂养摘要加载失败」返回 error（Go 不再依赖该摘要）

#### Scenario: question 不拉 DeviceProfile

- **WHEN** 客户端发送合法 `question`
- **THEN** voice-service Clinic 路径 MUST NOT 为拼装 LLM prompt 调用 DeviceProfile（本路径）

---

### Requirement: Flutter SHALL 提供胖宝 AI 诊室入口与页面

App（`flutter_ai_talk`）MUST 在首页 `home_immersive_header.dart` **新增** **胖宝** 入口（`Icons.pets`，tooltip「胖宝」），跳转至新页面 `pangbao_ai_screen.dart`。诊疗页 AppBar 标题 MUST 为 **「胖宝诊疗」**。首页品牌标题与 tooltip **MUST NOT** 改为「胖宝诊疗」（保持「胖宝」）。页面 MUST 支持文本输入提问并通过 WebSocket 展示流式 thinking 与 answer。首页 **MUST 保留** 原 **趋势** 入口（`Icons.insights`，tooltip「趋势」→ `/trends`），胖宝入口 **MUST NOT** 替换或隐藏趋势入口。

#### Scenario: 从首页进入胖宝诊疗

- **WHEN** 用户在首页点击胖宝入口（tooltip「胖宝」）
- **THEN** App SHALL 导航至胖宝诊疗页面且 AppBar 显示「胖宝诊疗」

#### Scenario: 从首页进入趋势

- **WHEN** 用户在首页点击趋势入口
- **THEN** App SHALL 导航至趋势图表页面（`/trends`）

### Requirement: Flutter SHALL 使用 ClinicWsClient 经 gateway-app 连接 WebSocket

App MUST 实现 `clinic_ws_client. 连接 URL MUST 使用 `wsClinicUrl` / `wsClinicUrlEffective`，默认由 `apiBaseUrl`（gateway-app-server 主机）推导为 `wss://{host}/voice/clinic/ws`（对齐 `wsVoiceAsrUrlEffective` 模式）。 客户端 **MUST NOT** 配置或连接 voice-service 内网地址。 连接成功后 MUST 先发送首帧 `type=auth`（`accessToken` + `deviceNo`，与 history WS / UCG chat 一致），收到 `auth_ok` 后方可发送 `question`。 每条 `question` MUST 含客户端生成的 UUID **`turnId`**。 客户端 MUST 支持发送 **`type=cancel`**（含 `turnId`）。 WS 生命周期：App 进入后台或离开诊疗页时 MUST **先** 对.

#### Scenario: 连接 gateway-app 而非 voice-service

- **WHEN** App 建立胖宝诊疗 WebSocket
- **THEN** 连接目标主机 MUST 与 `apiBaseUrl` 一致（gateway-app-server）
- **AND** 路径 MUST 为 `/voice/clinic/ws`（或含 `apiBaseUrl` path 前缀的等价路径）

#### Scenario: 首帧 auth

- **WHEN** clinic WS 握手成功
- **THEN** 客户端 MUST 发送 `auth` 帧且 MUST NOT 在收到 `auth_ok` 前发送 `question`

#### Scenario: question 携带 turnId

- **WHEN** 用户发送新问题
- **THEN** 客户端 MUST 生成新 UUID 作为 `turnId` 并随 `question` 上行

#### Scenario: 未登录不发 question

- **WHEN** 用户未登录（无有效 accessToken）
- **THEN** 客户端 MUST NOT 建立可提问的 clinic WS 连接；若服务端返回 40301 MUST 引导登录

#### Scenario: 离开页面显式 cancel

- **WHEN** 用户离开胖宝诊疗页或 App 进入后台且存在 active 流式 turn
- **THEN** 客户端 MUST 发送 `cancel`（含 active `turnId`）后再断开 WebSocket

#### Scenario: 流式展示回答

- **WHEN** 收到与 active `turnId` 一致的 `answer_delta` 序列后以 `answer_done` 结束
- **THEN** UI SHALL 逐字/逐段更新回答区域

#### Scenario: session_sync 恢复历史

- **WHEN** 收到 `session_sync` 且 `turns` 非空
- **THEN** App SHALL 将已完成轮次填充至聊天 `_items`（user 问 + assistant 答 + 免责声明）
- **AND** MUST NOT 为历史轮次渲染 thinking（服务端未提供）

#### Scenario: 丢弃 stale turnId 帧

- **WHEN** 收到 `thinking_delta` 或 `answer_delta` 且 `turnId` 不等于当前 active turn
- **THEN** 客户端 MUST 丢弃该帧且 MUST NOT 更新 UI

#### Scenario: turn_cancelled 清理进行中 UI

- **WHEN** 收到 `turn_cancelled` 且 `turnId` 为当前或刚结束的 active turn
- **THEN** UI MUST 清除该 turn 的进行中 thinking/answer 流式状态

### Requirement: Flutter SHALL 实现 thinking 展示交互规范

胖宝诊疗页 thinking 区域 MUST：默认最多可见 **5 行**，不足 5 行时弹性高度；流式过程中 MUST 自动滚动至**最新** thinking 行（折叠视口 **底对齐** 最新内容，对齐 `home_voice_message_strip` 内层 `jumpTo(maxScrollExtent)` 模式）；超过 5 行 MUST 折叠，用户点击可展开全部或在折叠区内局部滚动；折叠态 MUST 使用内层 `ScrollController` 且 **MUST NOT** 使用 `NeverScrollableScrollPhysics` 阻止底对齐滚动；用户手动上滑固定（pin scroll）后 MUST 停止自动滚动直至用户回到底部或点击「跟随最新」。

#### Scenario: 流式自动滚动至最新行

- **WHEN** thinking 流式追加且用户未 pin scroll
- **THEN** 折叠视口 SHALL 展示最新 thinking 行（非顶部旧行）
- **AND** 内层 scroll offset SHALL 跳转至 `maxScrollExtent`

#### Scenario: 用户 pin 后停止跟随

- **WHEN** 用户上滑 thinking 区域内层 scroll 以查看较早内容
- **THEN** 后续 `thinking_delta` MUST NOT 强制跳回底部，直至用户点击「跟随最新」或滚回底部

#### Scenario: 跟随最新恢复

- **WHEN** 用户已 pin 且点击「跟随最新」
- **THEN** thinking 区域 SHALL 跳至最新内容并恢复流式 auto-scroll

### Requirement: Flutter SHALL 展示免责声明

每条 AI 回答（`answer_done` 后）UI MUST 展示文案：**「本回答仅供参考，不能替代医生诊断」**。

#### Scenario: 回答完成后展示免责

- **WHEN** 一次问答流式完成
- **THEN** 该条回答下方 SHALL 显示上述免责声明

### Requirement: Flutter SHALL 使用独立胖宝 consent

胖宝功能 MUST 使用独立 consent 键 `pangbao_ai_consent_v1`，与首页喂养 AI consent 分离。首次进入胖宝页且未 consent 时 MUST 展示同意流程；未同意 MUST NOT 发送 `question`。

#### Scenario: 首次进入需 consent

- **WHEN** 用户首次打开胖宝页且本地无 `pangbao_ai_consent_v1`
- **THEN** App SHALL 展示 consent 对话框且 MUST NOT 发送 `auth`/`question` 直至用户同意

### Requirement: Flutter SHALL 展示 clinic_ai 额度

`ai_quota_models` MUST 扩展 `clinicAi` 字段。诊疗页额度 hint MUST 展示 **「本月胖宝诊疗剩余 N 次」**（或等价 `AiQuotaRemainingHintFeature.clinicAi` 文案）。收到 WS `error` code **40302** 或 HTTP 40302 MUST 弹框 **「本月额度已用完」**。code **40301** MUST 引导登录。

#### Scenario: 胖宝诊疗额度展示

- **WHEN** 用户进入胖宝诊疗页且额度 API 返回 clinicAi remaining=5
- **THEN** UI SHALL 展示「本月胖宝诊疗剩余 5 次」

#### Scenario: 胖宝额度用尽

- **WHEN** clinic WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

### Requirement: Flutter SHALL 允许流式过程中停止或改问

胖宝诊疗页在 LLM 流式（thinking 或 answer）进行中 **MUST NOT** 全局锁定文本输入。用户 **MUST** 能够：（a）发送新问题，以新 `turnId` supersede 当前 turn；和/或（b）通过停止控件发送 `cancel` 中断当前 turn。改问 **MUST** supersede 服务端上一 turn（由服务端下发 `turn_cancelled` reason=superseded）。

#### Scenario: 流式期间发送新问题

- **WHEN** thinking 或 answer 流式进行中且用户输入并发送新问题
- **THEN** 客户端 MUST 分配新 `turnId` 并发送 `question`
- **AND** UI MUST 展示新问题的 user 气泡并等待新 turn 的流式回复

#### Scenario: 流式期间点击停止

- **WHEN** 流式进行中且用户点击停止
- **THEN** 客户端 MUST 发送 `cancel` 且 `turnId` 为当前 active turn
- **AND** 收到 `turn_cancelled` 后 MUST 结束进行中 assistant 流式 UI

#### Scenario: 可选编辑用户气泡改问

- **WHEN** 实现 tap 用户气泡编辑（optional）
- **THEN** 预填问题文本后发送 MUST 使用新 `turnId` 并 supersede 当前 turn

---

### Requirement: 不得再暴露 clinic/tip HTTP 飞轮

voice-service MUST NOT 再 Bind 暴露 `POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback`。`DeviceClinicFeedbackController`（或仅承载上述两接口的等价控制器）MUST 删除。gateway-app MUST NOT 再为 `/device/api/clinic/*`、`/device/api/tip/*` 配置指向 voice 的反代（若该 pattern 仅服务上述反馈）。

#### Scenario: clinic feedback 不可达

- **WHEN** 客户端向 gateway 发起 `POST /device/api/clinic/feedback`
- **THEN** 请求 MUST NOT 由 clinic feedback 控制器处理

#### Scenario: tip feedback 不可达

- **WHEN** 客户端向 gateway 发起 `POST /device/api/tip/feedback`
- **THEN** 请求 MUST NOT 由 tip feedback 控制器处理

#### Scenario: gateway 反代清理

- **WHEN** 审查 `installVoiceProxyMiddleware`（或等价）绑定的 path 列表
- **THEN** MUST NOT 包含仅服务于 tip SSE、tip feedback、clinic feedback 的 `/device/tip/*`、`/device/api/tip/*`、`/device/api/clinic/*` pattern

### Requirement: care-alert 飞轮必须保留

本变更 MUST NOT 删除或下线 `POST /device/api/care-alert/feedback`。care-alert daily / delete item / feedback 仍由 **voice-service** 宿主处理，gateway MUST 继续将 `/device/api/care-alert/*` 反代至 voice。

#### Scenario: care-alert feedback 仍注册

- **WHEN** 审查 voice-service HTTP 注册与 care-alert 控制器
- **THEN** `POST /device/api/care-alert/feedback` MUST 仍由 care-alert 控制器处理，且服务层 `CareAlertFeedback`（及 Python CareAlertFeedback 客户端若原先存在）MUST 仍可用

#### Scenario: care-alert 反代仍在

- **WHEN** 审查 gateway voice 域反代 path 列表
- **THEN** MUST 仍包含 `/device/api/care-alert/*`

### Requirement: clinic WS 不受本变更影响

本变更 MUST NOT 移除或改坏 `/voice/clinic/ws`（及等价 clinic 连续对话 WebSocket）注册与反代。

#### Scenario: clinic WS 仍可达

- **WHEN** 审查 gateway WS 反代与 voice clinic WS 控制器
- **THEN** clinic WebSocket 路径 MUST 仍注册且可被反代

### Requirement: usage skip 与已删路径一致

`maintenance_skip.go`（或等价）MUST NOT 再保留已删除的 `POST /device/api/clinic/feedback` 与 `POST /device/api/tip/feedback` 排除项。MUST NOT 因本变更新增对 care-alert 路径的排除（除非负责人另行确认；本变更默认不动 care-alert usage 策略）。

#### Scenario: skip 列表无死路径

- **WHEN** 审查 usage 维护型排除列表
- **THEN** MUST NOT 出现已下线的 clinic/tip feedback 精确 path 条目

---

### Requirement: 服务端不得再暴露小贴士 SSE

voice-service 与 gateway-app MUST NOT 再对外提供 `POST /device/tip/generate`（及 `/device/tip/*` 反代）。`TipCtrl`（或等价）MUST NOT 再注册；`VoiceService.TipStream` 与仅服务于 tip 的 Python 客户端方法 MUST 删除或不可达。

#### Scenario: tip generate 不可达

- **WHEN** 客户端向 gateway 发起 `POST /device/tip/generate`
- **THEN** 请求 MUST NOT 由 tip SSE 控制器处理（典型为 404 或未绑定路由），且 voice-service HTTP Bind 列表 MUST NOT 包含 `NewTipCtrl()`（或等价）

#### Scenario: TipStream 从运行时移除

- **WHEN** 检索 voice 服务实现与 contracts 中的 tip 流式入口
- **THEN** MUST NOT 存在可被 HTTP 控制器调用的 `TipStream` 业务路径

### Requirement: 运维页不得再调用 tip generate

`resource/public/history.html`（及本仓内其它运维静态页若存在同等调用）MUST NOT 再请求 `/device/tip/generate`，相关调试 UI MUST 删除或改为不可用状态。

#### Scenario: history 运维无 tip 请求

- **WHEN** 审查 `resource/public/history.html` 源码
- **THEN** MUST NOT 出现对 `/device/tip/generate` 的 `fetch`/XHR 调用

---

### Requirement: 胖宝诊疗 clinic MUST 下线

系统 MUST 删除 `/voice/clinic/ws`（及等价 clinic HTTP）、`clinic_*` 服务实现、`clinic_ai` 额度字段与 Admin/App 展示、LaneClinic 配置，以及 Go 对 Python `/v1/clinic/stream` 的调用。gateway 对 clinic WS 的反代与鉴权豁免（若仅服务 clinic）MUST 同步清理。

#### Scenario: clinic WS 不可用

- **WHEN** 客户端尝试连接原 clinic WebSocket 路径
- **THEN** 服务端 MUST NOT 再注册 clinic 业务处理器

### Requirement: 聊天成长建议 MUST 删除

系统 MUST 删除喂养聊天中的「成长建议」生成路径（含 `callDeepSeekGrowthSuggestion` 及仅为其服务的 payload/意图分支），MUST NOT 再调用 Python `POST /v1/clinic` 同步陪伴接口。Hub/history 若仅用于运维读删历史 `suggest` 行，MUST 删除或停止暴露「成长建议」生成相关入口；历史行 MAY 留库。

#### Scenario: 不再触发成长建议 LLM

- **WHEN** 用户走喂养语音聊天主路径
- **THEN** 系统 MUST NOT 调用成长建议专用生成函数，MUST NOT 请求 `/v1/clinic`

---

### Requirement: Go App Tip API 请求体去掉月龄与当前时间

`POST /device/tip/generate` 的请求体（`DeviceTipGenerateReq` 或接线期等价结构）SHALL NOT 再包含 `babyAgeMonths`、`currentTime` 字段。相关 dc/注释 MUST NOT 再声称「0 则服务端按生日推算」或「0 则使用服务端当前时间」。

#### Scenario: 精简后的 App 请求体

- **WHEN** Flutter 或其它 App 客户端调用 `POST /device/tip/generate`
- **THEN** 请求 JSON SHALL 可仅含设备与事件相关字段（如 `deviceNo`、`eventId`、`eventName`）及既有其它非月龄/时间必填项
- **AND** 服务端 SHALL NOT 要求 `babyAgeMonths` 或 `currentTime`

#### Scenario: 假注释消除

- **WHEN** 评审 `api/v1/device_tip_http.go`（及控制器/服务注释）
- **THEN** SHALL NOT 存在「月龄为 0 由服务端推算」且实际未实现的误导描述

### Requirement: Go TipStream 与 PythonAIClient 去掉对应参数

`VoiceService.TipStream` 与 `PythonAIClient.TipStreamRequest` SHALL 删除 `babyAgeMonths`/`BabyAgeMonths`、`currentTime`/`CurrentTime`（及 JSON `baby_age_months`、`current_time`）。Go MUST NOT 再为 tip 补全 `currentTime` 后传给 Python。

#### Scenario: TipStream 签名瘦身

- **WHEN** 控制器调用 `TipStream` 发起小贴士流
- **THEN** 调用签名 SHALL 不再接收月龄与当前时间参数
- **AND** 发往 Python 的 JSON body SHALL 不含 `baby_age_months`、`current_time`

### Requirement: Python TipRequest 删除或废弃入站月龄与时间字段

`TipRequest` SHALL 删除 `baby_age_months`、`current_time` 字段，或将其标为废弃且非必填并忽略；权威月龄与时间上下文 MUST 来自服务端派生（见 `tip-python-derived-context`）。

#### Scenario: Tip 入站无月龄时间字段可成功

- **WHEN** Go 以 snake_case 发送 tip 请求且不含 `baby_age_months`、`current_time`
- **THEN** Python SHALL 接受该请求（其它必填合法时）
- **AND** SHALL 使用派生月龄与派生当前时间生成提示词

### Requirement: Flutter tip 调用去掉月龄与 currentTime

Flutter `tip_repository` / `tip_provider` / `home_screen` SHALL 不再计算或传递 `babyAgeMonths`，且请求体 SHALL NOT 包含 `currentTime`。

#### Scenario: home 触发 tip 不再传月龄

- **WHEN** `home_screen` 因历史 create 事件触发 tip 流式生成
- **THEN** 调用链 SHALL NOT 再传入 `babyAgeMonths`
- **AND** POST body SHALL NOT 包含 `babyAgeMonths` 或 `currentTime`

---

### Requirement: tip 请求路径为 /device/tip/generate

Flutter tip 客户端 SHALL 向网关基址下的 `/device/tip/generate` 发起 POST SSE 请求（不得使用缺少 `/device` 前缀的 `/tip/generate`）。

#### Scenario: URL 含 device tip 前缀

- **WHEN** `TipRepository.streamTip` 构建请求 URL
- **THEN** 路径 MUST 以 `/device/tip/generate` 结尾（相对 `apiBaseUrl`）

### Requirement: tip SSE 解析对齐 Go 方言

Flutter tip SSE 解析 SHALL 读取 `event:` 确定事件类型，`data:` 为纯文本内容；MUST 支持 thinking/answer/done/error，并在流结束识别 `data: [DONE]`。

#### Scenario: thinking/answer 增量

- **WHEN** 收到 `event: thinking`（或 `answer`）及随后 `data: <text>`
- **THEN** 客户端 MUST 将 data 纯文本作为该类型增量内容累积

#### Scenario: done 解析 answerId

- **WHEN** 收到 `event: done` 且 data 为含 `answerId` 的 JSON
- **THEN** 客户端 MUST 解析并保存 `answerId` 字符串，供后续反馈字段对齐使用

### Requirement: 状态模型使用 answerId

tip models / provider SHALL 使用 `String? answerId` 标识本次生成结果，MUST NOT 依赖从 SSE 流写入错误的 `tipId int`。

#### Scenario: 流结束后持有 answerId

- **WHEN** SSE done 事件成功解析 `answerId`
- **THEN** `TipContent`（或等价状态）MUST 更新为该 `answerId`，反馈按钮可用性可依据其非空判断

#### Scenario: 反馈 submit 字段对齐（可不完整飞轮）

- **WHEN** 调用方提交 tip 反馈
- **THEN** 请求体 SHOULD 使用 `answerId` 字符串字段对齐；完整 UI/后端反馈闭环属包 C，本包不要求反馈接口一定可用

---

### Requirement: TipCtrl 绑定到 voice-service

系统 SHALL 在 voice-service HTTP 注册路径中绑定 `NewTipCtrl()`（或等价），使 `POST /device/tip/generate` 由 voice 进程处理。

#### Scenario: voice 进程可路由 tip generate

- **WHEN** voice-service 启动并完成 `RegisterVoiceServiceHTTP`
- **THEN** `POST /device/tip/generate` MUST 由 `TipCtrl.Generate` 处理，且调用 `voice.TipStream` 写出 SSE

#### Scenario: 不绑定 feedback 控制器

- **WHEN** 本变更完成 voice 路由注册
- **THEN** 系统 MUST NOT Bind clinic tip feedback 相关控制器（留给包 C）

### Requirement: gateway 反代 /device/tip 到 voice

gateway-app SHALL 将 `/device/tip/*` 反向代理到 VOICE_API_PROXY（与现有 voice 域反代同一配置族）。

#### Scenario: App 请求经网关到达 voice

- **WHEN** 客户端向 gateway 发起 `POST /device/tip/generate`
- **THEN** 请求 MUST 被代理到 voice-service 对应路径，而非 history/device 本域处理

### Requirement: tip generate 需登录且计入 usage

`POST /device/tip/generate` SHALL 要求 Bearer 登录，且 MUST 计入 App API usage 统计。

#### Scenario: 不进入鉴权豁免

- **WHEN** 维护 `gateway_app_auth_exempt` 列表
- **THEN** `/device/tip/generate` MUST NOT 出现在 Bearer 豁免路径中

#### Scenario: 不进入 usage 维护排除

- **WHEN** 维护 `usagestats/maintenance_skip`
- **THEN** `/device/tip/generate` MUST NOT 被写入排除列表（即保持统计）

### Requirement: SSE 方言保持 Go 风格

`TipCtrl.Generate` SHALL 继续以 `event:` + data 纯文本写出 thinking/answer/done，done 的 data MUST 含 `answerId` 字段；最终 MUST 发送 `data: [DONE]`。

#### Scenario: 流式帧格式

- **WHEN** tip 流成功推送 thinking/answer/done
- **THEN** 每帧 MUST 含 `event: <name>` 与对应 `data:` 行，且 done 的 data 可解析出 `answerId`

---

### Requirement: Tip 月龄由 Python 根据生日自算

在 tip 生成图中，系统 SHALL 在 `fetch_baby_profile`（或等价画像拉取）完成之后，根据宝宝 `birthday` 与当前日期（时区 `Asia/Shanghai`）计算月龄，供提示词及依赖月龄的检索使用。调用方（Go/Flutter）MUST NOT 被要求提供 `baby_age_months` / `babyAgeMonths`。

#### Scenario: 有生日时写入计算出的月龄

- **WHEN** tip 图成功获取到可解析的 `birthday`
- **THEN** 系统 SHALL 按日历月差计算非负整数月龄（未来生日钳为 0）
- **AND** 提示词 SHALL 包含该月龄（例如「宝宝月龄：{n} 个月」）

#### Scenario: 无生日时提示「未知」而非 0 个月

- **WHEN** 画像缺失、无 `birthday` 或生日无法解析
- **THEN** 系统 SHALL NOT 将月龄表示为 `0` 个月以表示「未知」
- **AND** 提示词 SHALL 使用「宝宝月龄：未知」（或等价明确未知文案）
- **AND** 若知识检索依赖月龄，SHALL 跳过月龄过滤或使用不限月龄策略，MUST NOT 用 0 冒充新生儿月龄过滤

### Requirement: Tip 当前时间由 Python 在写提示词时生成

系统 SHALL 在构建 tip 提示词时生成当前时间上下文，时区 MUST 为 `Asia/Shanghai`。Tip 入站请求 MUST NOT 将 `current_time` / `currentTime` 作为必填字段。

#### Scenario: 提示词含上海时区当前时间

- **WHEN** tip 流式生成进入写提示词阶段
- **THEN** 提示词 SHALL 包含基于 `Asia/Shanghai` 的当前时间（可读本地时间；MAY 同时附带 Unix 秒）
- **AND** 该时间 SHALL 来自 Python 运行时时钟，而非请求体字段

#### Scenario: 请求体不再要求 current_time

- **WHEN** 调用方 POST tip stream 且 body **不含** `current_time` / `currentTime`
- **THEN** 系统 SHALL 仍能完成 tip 生成（在其它必填字段合法时）
- **AND** SHALL NOT 因缺少该字段返回 422

---

### Requirement: Tip streaming generate API

Go 服务端 SHALL 暴露 `POST /device/tip/generate` 接口，以 SSE 方式流式返回小贴士生成的思考过程与最终建议。

#### Scenario: Flutter client requests tip via SSE
- **WHEN** Flutter 客户端向 `/device/tip/generate` 发送 POST 请求（body 包含 deviceNo、eventId、eventName；月龄与当前时间由服务端派生，不再要求 babyAgeMonths / currentTime）
- **THEN** Go 服务端 SHALL 立即返回 SSE 响应头（Content-Type: text/event-stream、Cache-Control: no-cache）
- **AND** 按序推送 `thinking`、`answer`、`done` 三类 SSE 事件
- **AND** 最终以 `data: [DONE]` 结束流式响应

### Requirement: Tip thinking events streamed

Go 服务端 SHALL 将 Python AI 返回的思考过程增量实时推送给客户端，不得等全部完成后一次性返回。

#### Scenario: Thinking delta pushed to client
- **WHEN** Go 服务端从 Python `/v1/tip/stream` 收到 `thinking` 类型的 SSE 事件
- **THEN** Go 服务端 SHALL 立即将该 thinking delta 以 SSE `event: thinking` 帧转发给 Flutter 客户端
- **AND** 不得缓冲或延迟超过 100ms

### Requirement: Tip answer events streamed

Go 服务端 SHALL 将 Python AI 返回的建议内容增量实时推送给客户端。

#### Scenario: Answer delta pushed to client
- **WHEN** Go 服务端从 Python `/v1/tip/stream` 收到 `answer` 类型的 SSE 事件
- **THEN** Go 服务端 SHALL 立即将该 answer delta 以 SSE `event: answer` 帧转发给 Flutter 客户端

### Requirement: Tip done event with answer_id

Go 服务端 SHALL 在流式结束时返回 answer_id，供后续反馈使用。

#### Scenario: Done event includes answer_id
- **WHEN** Go 服务端从 Python `/v1/tip/stream` 收到 `done` 事件（含 answer_id）
- **THEN** Go 服务端 SHALL 向客户端推送 SSE `event: done` 帧，body 含 answer_id
- **AND** 紧接着推送 `data: [DONE]` 结束整个 SSE 连接

### Requirement: Tip API error handling

当 Python AI 服务不可用时，Go 服务端 SHALL 返回固定错误提示给客户端。

#### Scenario: Python AI unavailable during tip generate
- **WHEN** Go 服务端调用 Python `/v1/tip/stream` 失败（连接超时、5xx 错误等）
- **THEN** Go 服务端 SHALL 以 SSE `event: error` 帧向客户端推送错误消息："AI服务暂时不可用，请稍后再试"
- **AND** 紧接着推送 `data: [DONE]` 结束连接

---
