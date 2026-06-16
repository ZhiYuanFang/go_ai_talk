## ADDED Requirements

### Requirement: voice-service SHALL be authoritative for voice_ai and clinic_ai quota configuration and usage

`voice-service` MUST 在 **`ai_voice_voice`** 库（GoFrame `database.default`，连接 `VOICE_DB_LINK`）维护 AI 月度额度全局默认与 per-wxId override，并 MUST 为 **`voice_ai`** 与 **`clinic_ai`** 两个 feature 独立计数。全局默认 MUST 包含 `voiceAiMonthlyLimit`（初始 **5**）与 `clinicAiMonthlyLimit`（初始 **30**）；Admin 可独立修改。per-wxId override MAY 单独覆盖任一 feature；未 override 的 feature MUST 回退全局默认。月度用量 MUST 存 Redis，键格式 **`ai:usage:voice_ai:{wxId}:{YYYYMM}`** 与 **`ai:usage:clinic_ai:{wxId}:{YYYYMM}`**，其中 `YYYYMM` MUST 按 `Asia/Shanghai` 自然月生成。voice-service MUST NOT 将 voice/clinic 配额配置或用量写入 device 或 ucg 库表。

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

voice-service MUST 提供 `GET/PUT /voice/admin/api/ai-quota/default` 与 `GET/PUT /voice/admin/api/ai-quota/user`（query/body 含 `wxId`），认证 MUST 为 Header `X-Admin-Password` 等于 `voice.admin.password`（gateway 经 `VOICE_ADMIN_PASSWORD` 注入）。voice-service MUST 本地持久化至 `ai_voice_voice`，MUST NOT 转发 device 或 ucg。PUT default MUST 接受 `voiceAiMonthlyLimit` 与 `clinicAiMonthlyLimit`（正整数）。PUT user MUST 接受 optional 两字段；空值 SHALL 表示清除该 feature override。

#### Scenario: 管理员修改全局胖宝默认

- **WHEN** 管理员 PUT default 为 voiceAi=5、clinicAi=30
- **THEN** voice 权威配置 SHALL 更新且新用户 check clinic_ai SHALL 使用 limit=30

#### Scenario: voice admin 口令错误

- **WHEN** `X-Admin-Password` 无效
- **THEN** 系统 SHALL 返回未授权且 SHALL NOT 修改配置
