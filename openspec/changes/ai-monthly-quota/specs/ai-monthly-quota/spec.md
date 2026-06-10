## ADDED Requirements

### Requirement: device-service SHALL be authoritative for AI monthly quota configuration and usage

`device-service` MUST 维护 AI 月度额度全局默认与 per-wxId override，并 MUST 为 `polish` 与 `voice_ai` 两个 feature 独立计数。全局默认 MUST 包含 `polishMonthlyLimit` 与 `voiceAiMonthlyLimit`（初始值均为 5，Admin 可独立修改，如润笔 10、喂养 AI 5）。per-wxId override MAY 单独覆盖任一 feature；未 override 的 feature MUST 回退全局默认。月度用量 MUST 存 Redis，键格式 `ai:usage:{feature}:{wxId}:{YYYYMM}`，其中 `YYYYMM` MUST 按 `Asia/Shanghai` 自然月生成。ucg-service 与 voice-service MUST 经 HTTP internal 契约访问额度，MUST NOT 直连 device 库表。

#### Scenario: 全局默认独立配置

- **WHEN** 管理员将全局 `polishMonthlyLimit` 设为 10 且 `voiceAiMonthlyLimit` 保持 5
- **THEN** 无 override 的用户润笔上限 SHALL 为 10、喂养 AI 上限 SHALL 为 5

#### Scenario: 单人 override 覆盖单 feature

- **WHEN** wxId=1001 的 override 为 `polishMonthlyLimit=20` 且未设 `voiceAiMonthlyLimit`
- **THEN** 该用户润笔上限 SHALL 为 20 且喂养 AI SHALL 使用全局默认

### Requirement: Internal quota check and consume APIs SHALL enforce wxId and feature semantics

`POST /device/internal/api/ai-quota/check` MUST 要求有效 `X-Device-Internal-Secret`。body MUST 含 `wxId`（正整数）与 `feature`（`polish` | `voice_ai`）。响应 MUST 含 `allowed`（boolean）、`used`、`limit`。`check` MUST NOT 修改用量。

`POST /device/internal/api/ai-quota/consume` MUST 在 AI 成功返回后由 ucg/voice 调用；成功时 MUST `INCR` 当月用量并返回 `{ used, limit }`。若扣减后 `used > limit`，系统 MUST 回滚该次 INCR 并返回超额错误。

#### Scenario: check 不扣减

- **WHEN** ucg 在润笔前调用 check 且 used=4、limit=5
- **THEN** 响应 `allowed=true` 且 Redis 计数 SHALL 仍为 4

#### Scenario: consume 成功扣减

- **WHEN** voice 在 LLM 成功后调用 consume 且 used=4、limit=5
- **THEN** used SHALL 变为 5 且响应 used=5

#### Scenario: wxId 无效拒绝

- **WHEN** internal API 收到 wxId=0
- **THEN** 系统 SHALL 返回错误且 MUST NOT 读写的用量

### Requirement: App quota read API SHALL expose used and limit for both features

`GET /device/app/api/ai-quota` MUST 要求有效 Bearer 且 `X-Internal-Wx-Id > 0`。响应 MUST 为 `{ polish: { used, limit }, voiceAi: { used, limit } }`，对应当月上海时区桶。

#### Scenario: 登录用户查询额度

- **WHEN** wxId=1001 请求 ai-quota 且当月润笔已用 2、上限 5
- **THEN** `polish.used` SHALL 为 2 且 `polish.limit` SHALL 为 5

#### Scenario: wxId=0 拒绝

- **WHEN** 请求携带 wxId=0
- **THEN** 系统 SHALL 返回未授权/无效身份错误

### Requirement: UCG polish SHALL pre-check quota and consume only on DashScope success

`POST /ucg/app/api/posts/polish` MUST 在调用 DashScope 前经 device internal 执行 check；若 `allowed=false` MUST 返回 code **40302** 与 message **「本月额度已用完」** 且 MUST NOT 调用 DashScope。DashScope 成功返回有效正文后 MUST 调用 consume。参数错误、未配置 AI、DashScope 失败 MUST NOT 调用 consume。

#### Scenario: 额度用尽

- **WHEN** 用户润笔 check 得到 used=5、limit=5
- **THEN** API SHALL 返回 40302 与「本月额度已用完」且 SHALL NOT 请求 DashScope

#### Scenario: DashScope 失败不扣减

- **WHEN** check 通过但 DashScope 返回 5xx
- **THEN** 系统 SHALL NOT 调用 consume 且 used SHALL 不变

### Requirement: Voice feeding AI SHALL require wxId and enforce quota around LLM calls

voice-service 在即将调用 LLM（母婴 DeepSeek 或 casual 流式）前 MUST 解析 wxId>0（优先 `X-Internal-Wx-Id`，否则 device internal 由 deviceNo 反查）。wxId≤0 MUST 返回 code **40301** 与登录引导文案，MUST NOT 调用 LLM。LLM 调用前 MUST check；`allowed=false` MUST 返回 WS 错误帧 code **40302** message **「本月额度已用完」**。LLM 成功完成后 MUST consume。模式切换、规则回复、LLM 失败兜底、纯 ASR MUST NOT check 或 consume。

#### Scenario: 未登录不可用喂养 AI

- **WHEN** WS 会话 wxId 解析为 0 且用户 utterance 将触发 LLM
- **THEN** 系统 SHALL 返回 40301 且 SHALL NOT 调用 LLM

#### Scenario: 喂养 AI 额度用尽

- **WHEN** check 得到 voice_ai used=limit
- **THEN** WS SHALL 返回 40302「本月额度已用完」且 SHALL NOT 调用 LLM

#### Scenario: 模式切换不扣减

- **WHEN** 用户发送模式切换指令且不触发 LLM
- **THEN** 系统 SHALL NOT 调用 check 或 consume

### Requirement: UCG admin SHALL configure global default and per-user override via proxied APIs

ucg-service MUST 提供 `GET/PUT /ucg/admin/api/ai-quota/default` 与 `GET/PUT /ucg/admin/api/ai-quota/user`（query/body 含 `wxId`），认证 MUST 为 Header `X-Admin-Password` 等于 `ucg.admin.password`。ucg-service MUST 转发至 device internal admin 接口，MUST NOT 本地持久化额度。PUT default MUST 接受 `polishMonthlyLimit` 与 `voiceAiMonthlyLimit`（正整数）。PUT user MUST 接受 optional 两字段；空值 SHALL 表示清除该 feature override。

`resource/public/ucg-admin.html`「AI 配置」Tab MUST 增加全局默认与 wxId override 表单，调用上述 ucg admin API。

#### Scenario: 管理员修改全局润笔默认

- **WHEN** 管理员 PUT default 为 polish=10、voiceAi=5
- **THEN** device 权威配置 SHALL 更新且新用户 check SHALL 使用新默认

#### Scenario: ucg admin 口令错误

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 修改配置

### Requirement: Flutter client SHALL display quota exhaustion dialog

App 客户端（flutter_ai_talk）在收到 HTTP 40302 或 WS 错误帧 code=40302 MUST 弹框展示 **「本月额度已用完」**。40301 MUST 引导用户登录，MUST NOT 使用额度用尽文案。

#### Scenario: 润笔超额弹框

- **WHEN** polish API 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

#### Scenario: 喂养 AI 超额弹框

- **WHEN** voice WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」
