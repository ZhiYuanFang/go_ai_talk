## MODIFIED Requirements

### Requirement: ucg polish SHALL pre-check and consume quota locally

`POST /ucg/app/api/posts/polish` MUST 在调用上游 LLM（`LanePolish`）前于 **ucg-service 进程内**执行 polish check。若 `allowed=true`，MUST 经 `LoadProfile(LanePolish)` 调用上游；上游成功返回有效正文后 MUST 于本进程 consume；响应 MUST 含 `quotaDegraded: false`（或省略）。若 `allowed=false`（`used >= limit`），MUST **NOT** 返回 code **40302**；MUST 经 degraded 路径强制 `DefaultSeedProfile(LanePolish)`（智谱 **`glm-4.6v-flash`**）调用上游，MUST **NOT** consume；成功响应 MUST 含 **`quotaDegraded: true`**。check 通过后（含 degraded 路径）、调用上游前 MUST 经 polish lane 闸门；队列满 MUST 返回 **50301**「当前队列已满，请稍后重试」且 MUST NOT 调用上游、MUST NOT consume。参数错误、未配置 AI、上游失败、50301 MUST NOT 调用 consume。

#### Scenario: 额度内正常润笔

- **WHEN** 用户润笔 check 得到 used=2、limit=5
- **THEN** API SHALL 调用上游、consume 一次且 `quotaDegraded` 为 false 或省略

#### Scenario: 额度用尽 degraded 润笔

- **WHEN** 用户润笔 check 得到 used=5、limit=5
- **THEN** API SHALL **NOT** 返回 40302
- **AND** SHALL 返回润色正文且 `quotaDegraded=true`
- **AND** SHALL **NOT** consume 且 used SHALL 保持 5

#### Scenario: 上游失败不扣减

- **WHEN** allowed=true 但上游返回 5xx
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
