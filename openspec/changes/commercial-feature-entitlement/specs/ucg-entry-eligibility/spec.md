## ADDED Requirements

### Requirement: UCG 入场资格 API SHALL 按 device_no 返回连续有效喂养日结果

cash-service MUST 提供 `GET /cash/app/api/ucg/eligibility`，按网关注入的 `X-Internal-Device-No` 返回至少：`qualified`（bool）、`requiredDays`（固定 7）、`effectiveDays`（int）、`remainingDays`（int），并 MAY 返回可选文案字段。请求 MUST 经登录到达；`device_no` 为空时 MUST 返回错误，MUST NOT 伪造合格。该接口 MUST NOT 计入 App usage 统计（维护型 skip）。

#### Scenario: 连续满 7 有效日判定合格

- **WHEN** 某 `device_no` 按规则计算出的 `effectiveDays >= 7`
- **THEN** 响应 MUST 含 `qualified=true`、`requiredDays=7`、`remainingDays=0`

#### Scenario: 不足 7 日返回剩余天数

- **WHEN** `effectiveDays = 3`
- **THEN** 响应 MUST 含 `qualified=false`、`requiredDays=7`、`remainingDays=4`

#### Scenario: 缺少 device_no

- **WHEN** 请求未提供有效 `X-Internal-Device-No`
- **THEN** 系统 MUST 拒绝并返回错误，MUST NOT 返回 `qualified=true`

### Requirement: 有效喂养日 MUST 为上海日历日且连续统计且当日 history 不少于 10

有效喂养日判定 MUST 使用 `Asia/Shanghai`。某日有效当且仅当该 `device_no` 在该日内 history 记录数 ≥ 10。连续有效日 MUST 以请求当日为锚向前连续统计。资格计算 MUST 经 history-service HTTP 契约取数，MUST NOT 由 cash-service 直查 history 库。

#### Scenario: 单日满 10 条计为有效日

- **WHEN** 某上海日该设备 history 行数 = 10
- **THEN** 该日 MUST 计入有效喂养日

#### Scenario: 单日不足 10 条不计有效且中断连续

- **WHEN** 某上海日该设备 history 行数 = 9
- **THEN** 该日 MUST NOT 计入有效喂养日，且连续计数在该日中断

#### Scenario: 禁止跨库直查

- **WHEN** 实现资格计算
- **THEN** cash-service MUST 仅通过 history 内部 HTTP 获取按日计数，MUST NOT 连接 history 域数据库

### Requirement: UCG 入场资格 MUST NOT 受 VIP 或功能权益影响

资格计算与响应 MUST 仅基于喂养有效日规则。账号 VIP 状态、功能权益、邀请码/支付/广告开通 MUST NOT 将 `qualified` 置为 true 或缩短 `remainingDays`。客户端 MUST NOT 被本规格要求以 `isVip` 绕过入场门（服务端亦不提供 VIP 短路字段）。

#### Scenario: VIP 用户仍按喂养日计算

- **WHEN** 请求方账号为 VIP 但 `effectiveDays < 7`
- **THEN** 响应 MUST 为 `qualified=false`，且结果与非 VIP 在相同喂养数据下一致

### Requirement: 资格结果 MUST 按日缓存于 Redis 且不得落库

同一 `device_no` 在同一上海日内的资格结果 MUST 在缓存命中时直接返回，并 MUST 写入 Redis（经 `cachekit` + platform 键 builder）。MUST NOT 将资格结果作为权威数据写入 MySQL。MUST NOT 为此新增后台 ticker；miss 时在请求路径同步计算。history 调用失败时 MUST fail-closed（返回错误）。

#### Scenario: 同日第二次请求读缓存

- **WHEN** 某设备当日已成功计算并写入 Redis 后再次请求资格 API
- **THEN** 系统 MUST 返回与缓存一致的结果，且 MUST NOT 再次全量重算（除非缓存丢失）

#### Scenario: 资格不落 MySQL

- **WHEN** 资格计算完成
- **THEN** 系统 MUST NOT 将资格结果写入 MySQL 业务表

#### Scenario: Redis 经 platform

- **WHEN** 读写资格缓存
- **THEN** 代码 MUST 经 `cachekit`（含 Observer），键 MUST 来自 platform builder，MUST NOT 业务层直连 `g.Redis()`
