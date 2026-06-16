## ADDED Requirements

### Requirement: ucg-service SHALL be authoritative for polish quota configuration and usage

`ucg-service` MUST 在 **`ai_voice_ucg`** 库（GoFrame `database.default`，连接 `UCG_DB_LINK`）维护 **`polish`** feature 的 AI 月度额度全局默认与 per-wxId override。全局默认 MUST 包含 `polishMonthlyLimit`（初始 **5**）；Admin 可修改。per-wxId override MAY 覆盖 polish；未 override MUST 回退全局默认。月度用量 MUST 存 Redis，键格式 **`ai:usage:polish:{wxId}:{YYYYMM}`**，`YYYYMM` MUST 按 `Asia/Shanghai` 自然月生成。ucg-service MUST NOT 将 polish 配额配置或用量写入 device 或 voice 库表；MUST NOT 再转发 device internal ai-quota API。

#### Scenario: 全局润笔默认配置

- **WHEN** 管理员将全局 `polishMonthlyLimit` 设为 10
- **THEN** 无 override 的用户润笔上限 SHALL 为 10

#### Scenario: 单人 override 覆盖润笔

- **WHEN** wxId=1001 的 override 为 `polishMonthlyLimit=20`
- **THEN** 该用户润笔上限 SHALL 为 20

### Requirement: ucg polish SHALL pre-check and consume quota locally

`POST /ucg/app/api/posts/polish` MUST 在调用 DashScope 前于 **ucg-service 进程内**执行 polish check；若 `allowed=false` MUST 返回 code **40302** 与 message **「本月额度已用完」** 且 MUST NOT 调用 DashScope。DashScope 成功返回有效正文后 MUST 于本进程 consume。参数错误、未配置 AI、DashScope 失败 MUST NOT 调用 consume。

#### Scenario: 额度用尽

- **WHEN** 用户润笔 check 得到 used=5、limit=5
- **THEN** API SHALL 返回 40302 与「本月额度已用完」且 SHALL NOT 请求 DashScope

#### Scenario: DashScope 失败不扣减

- **WHEN** check 通过但 DashScope 返回 5xx
- **THEN** 系统 SHALL NOT 调用 consume 且 used SHALL 不变

### Requirement: ucg App quota read API SHALL expose polish only

`GET /ucg/app/api/ai-quota` MUST 要求有效 Bearer 且 `X-Internal-Wx-Id > 0`（经 gateway-app 注入）。响应 MUST 为 `{ polish: { used, limit } }`，对应当月上海时区桶。本接口 MUST NOT 返回 `voiceAi` 或 `clinicAi` 字段。

#### Scenario: 登录用户查询润笔额度

- **WHEN** wxId=1001 请求 `/ucg/app/api/ai-quota` 且当月润笔已用 2、上限 5
- **THEN** `polish.used` SHALL 为 2 且 `polish.limit` SHALL 为 5

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
