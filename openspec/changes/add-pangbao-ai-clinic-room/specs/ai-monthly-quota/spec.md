## MODIFIED Requirements

### Requirement: device-service SHALL be authoritative for AI monthly quota configuration and usage

`device-service` MUST 维护 AI 月度额度全局默认与 per-wxId override，并 MUST 为 `polish`、`voice_ai` 与 **`clinic_ai`** 三个 feature 独立计数。全局默认 MUST 包含 `polishMonthlyLimit`、`voiceAiMonthlyLimit` 与 **`clinicAiMonthlyLimit`**（`clinic_ai` 初始值为 **30**，Admin 可独立修改；`polish`/`voice_ai` 初始值仍为 5）。per-wxId override MAY 单独覆盖任一 feature；未 override 的 feature MUST 回退全局默认。月度用量 MUST 存 Redis，键格式 `ai:usage:{feature}:{wxId}:{YYYYMM}`，其中 `feature` 含 `clinic_ai`，`YYYYMM` MUST 按 `Asia/Shanghai` 自然月生成。ucg-service 与 voice-service MUST 经 HTTP internal 契约访问额度，MUST NOT 直连 device 库表。

#### Scenario: 全局默认独立配置

- **WHEN** 管理员将全局 `polishMonthlyLimit` 设为 10、`voiceAiMonthlyLimit` 保持 5、`clinicAiMonthlyLimit` 保持 30
- **THEN** 无 override 的用户润笔上限 SHALL 为 10、喂养 AI 上限 SHALL 为 5、胖宝 AI 上限 SHALL 为 30

#### Scenario: 单人 override 覆盖单 feature

- **WHEN** wxId=1001 的 override 为 `clinicAiMonthlyLimit=50` 且未设其它 feature
- **THEN** 该用户胖宝 AI 上限 SHALL 为 50 且润笔/喂养 AI SHALL 使用全局默认

### Requirement: Internal quota check and consume APIs SHALL enforce wxId and feature semantics

`POST /device/internal/api/ai-quota/check` MUST 要求有效 `X-Device-Internal-Secret`。body MUST 含 `wxId`（正整数）与 `feature`（`polish` | `voice_ai` | **`clinic_ai`**）。响应 MUST 含 `allowed`（boolean）、`used`、`limit`。`check` MUST NOT 修改用量。

`POST /device/internal/api/ai-quota/consume` MUST 在 AI 成功返回后由 ucg/voice 调用；成功时 MUST `INCR` 当月用量并返回 `{ used, limit }`。若扣减后 `used > limit`，系统 MUST 回滚该次 INCR 并返回超额错误。

#### Scenario: check 不扣减

- **WHEN** voice 在胖宝提问前调用 check feature=clinic_ai 且 used=29、limit=30
- **THEN** 响应 `allowed=true` 且 Redis 计数 SHALL 仍为 29

#### Scenario: consume 成功扣减

- **WHEN** voice 在 Clinic LLM 成功后调用 consume feature=clinic_ai 且 used=29、limit=30
- **THEN** used SHALL 变为 30 且响应 used=30

#### Scenario: wxId 无效拒绝

- **WHEN** internal API 收到 wxId=0
- **THEN** 系统 SHALL 返回错误且 MUST NOT 读写的用量

### Requirement: App quota read API SHALL expose used and limit for both features

`GET /device/app/api/ai-quota` MUST 要求有效 Bearer 且 `X-Internal-Wx-Id > 0`。响应 MUST 为 `{ polish: { used, limit }, voiceAi: { used, limit }, clinicAi: { used, limit } }`，对应当月上海时区桶。

#### Scenario: 登录用户查询额度

- **WHEN** wxId=1001 请求 ai-quota 且当月胖宝已用 5、上限 30
- **THEN** `clinicAi.used` SHALL 为 5 且 `clinicAi.limit` SHALL 为 30

#### Scenario: wxId=0 拒绝

- **WHEN** 请求携带 wxId=0
- **THEN** 系统 SHALL 返回未授权/无效身份错误

### Requirement: UCG admin SHALL configure global default and per-user override via proxied APIs

ucg-service MUST 提供 `GET/PUT /ucg/admin/api/ai-quota/default` 与 `GET/PUT /ucg/admin/api/ai-quota/user`（query/body 含 `wxId`），认证 MUST 为 Header `X-Admin-Password` 等于 `ucg.admin.password`。ucg-service MUST 转发至 device internal admin 接口，MUST NOT 本地持久化额度。PUT default MUST 接受 `polishMonthlyLimit`、`voiceAiMonthlyLimit` 与 **`clinicAiMonthlyLimit`**（正整数）。PUT user MUST 接受 optional 三字段；空值 SHALL 表示清除该 feature override。

Device Admin UI（`resource/public` 下 ai-quota 配置页或 ucg-admin「AI 配置」Tab）MUST 增加 **胖宝 AI 月度次数** 第三字段，调用上述 admin API。

#### Scenario: 管理员修改全局胖宝默认

- **WHEN** 管理员 PUT default 为 polish=10、voiceAi=5、clinicAi=30
- **THEN** device 权威配置 SHALL 更新且新用户 check clinic_ai SHALL 使用 limit=30

#### Scenario: ucg admin 口令错误

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 修改配置

### Requirement: Flutter client SHALL display quota exhaustion dialog

App 客户端（flutter_ai_talk）在收到 HTTP 40302 或 WS 错误帧 code=40302 MUST 弹框展示 **「本月额度已用完」**（含润笔、喂养 AI、**胖宝 AI**）。40301 MUST 引导用户登录，MUST NOT 使用额度用尽文案。

#### Scenario: 胖宝 AI 超额弹框

- **WHEN** clinic WS 返回 code=40302
- **THEN** App SHALL 弹框「本月额度已用完」

## ADDED Requirements

### Requirement: Voice clinic AI SHALL enforce clinic_ai quota around LLM calls

voice-service 在处理 `/voice/clinic/ws` 的 `question` 并即将调用 LLM 前 MUST 解析 wxId>0。LLM 调用前 MUST 对 feature **`clinic_ai`** 执行 check；`allowed=false` MUST 返回 WS 40302。LLM 流式成功完成后 MUST consume。摘要刷新失败、rate limit、参数校验失败 MUST NOT consume。

#### Scenario: 胖宝额度用尽

- **WHEN** check 得到 clinic_ai used=limit
- **THEN** WS SHALL 返回 40302 且 MUST NOT 调用 LLM
