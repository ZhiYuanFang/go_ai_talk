## ADDED Requirements

### Requirement: 值得留意喂养资格 App API MUST 由 cash-service 提供且与 UCG 同构

**cash-service** MUST 提供登录+绑机可达的 `GET /cash/app/api/care-alert/eligibility`（或等价路径），按 `care_alert_entry` 配置在 cash 内合成并返回至少：`qualified`、`requiredDays`、`effectiveDays`、`remainingDays`，并 MAY 返回 `message`。`device_no` MUST 只信网关内部头。该路径 MUST NOT 加入 Bearer 匿名白名单。资格结果 MUST 可按日 Redis 缓存且 MUST NOT 落 MySQL 权威。history 失败 MUST fail-closed。device-service MUST NOT 作为该喂养天数资格的权威计算宿主。

#### Scenario: 连续满配置天数合格

- **WHEN** `care_alert_entry.requiredDays=2` 且连续有效日 ≥ 2
- **THEN** 响应 MUST `qualified=true`、`remainingDays=0`

#### Scenario: 不足返回剩余天数

- **WHEN** `requiredDays=2` 且 `effectiveDays=1`
- **THEN** 响应 MUST `qualified=false`、`remainingDays=1`

#### Scenario: 资格 API 归属 cash

- **WHEN** 客户端查询值得留意喂养是否达标
- **THEN** 请求 MUST 打到 cash App eligibility 路径，MUST NOT 依赖 device care-alert 接口返回喂养连续天数资格

### Requirement: 客户端 MUST 以 cash 服务端资格替代「昨日有发生」闸门

Flutter 值得留意展示流 MUST 先请求 **cash** care-alert eligibility。当 `qualified` 不为 true（含加载失败 fail-closed）时，客户端 MUST 仍展示「值得留意」卡片，并 MUST 展示基于 eligibility 的剩余/所需有效天数进度文案，且 MUST NOT 调用值得留意生成或日列表数据接口。当 `qualified=true` 时，客户端 MUST 再按原逻辑请求值得留意接口并展示内容。客户端 MUST NOT 再以「上海昨日有发生」作为是否可拉取值得留意数据的前提，MUST NOT 用本地 history 自行判定权威 `qualified`。

#### Scenario: 未合格展示进度不拉生成

- **WHEN** eligibility 返回 `qualified=false` 且 `remainingDays=1`
- **THEN** UI MUST 展示值得留意卡片与需再累计有效日的提示，且 MUST NOT 请求 care-alert daily/生成接口

#### Scenario: 合格后按原逻辑拉数

- **WHEN** eligibility 返回 `qualified=true`
- **THEN** 客户端 MUST 允许按既有逻辑请求值得留意数据接口并展示结果

#### Scenario: 资格失败 fail-closed

- **WHEN** eligibility HTTP 失败
- **THEN** 客户端 MUST NOT 当作已合格去请求生成接口
