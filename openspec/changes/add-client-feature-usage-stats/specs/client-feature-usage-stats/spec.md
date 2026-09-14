## ADDED Requirements

### Requirement: App client-usage report MUST accept logged-in feature events

`gateway-app-server` MUST 提供 `POST /device/app/api/client-usage/report`（或 design 锁定的等价本机 App 路径）。请求体 MUST 包含非空字符串 `featureId`（客户端自由字符串；服务端 MUST NOT 强制枚举白名单）。系统 MUST 从登录态解析 `wxId`；当 `wxId <= 0` 时 MUST 拒绝写入且 MUST NOT 记入 Redis。当 `featureId` 为空或超过实现设定的最大长度时 MUST 拒绝。成功时 MUST 使用**服务端时间**记录事件，MUST NOT 信任客户端时间戳。

#### Scenario: 登录用户上报成功

- **WHEN** 已登录用户提交合法 `featureId` 且未触发限流且非模拟用户
- **THEN** 系统 SHALL 返回成功，并 SHALL 更新聚合与该用户时间线

#### Scenario: 未登录拒绝

- **WHEN** 请求无法解析出正整数 `wxId`
- **THEN** 系统 MUST 拒绝，且 MUST NOT 写入任何 client-usage Redis 键

### Requirement: Simulated users MUST NOT be recorded

当关联 `wxId` 被判定为模拟用户（与 App API usage 同源判定）时，client-usage report MUST NOT 写入聚合或时间线。

#### Scenario: 模拟用户上报不落库

- **WHEN** 模拟用户调用 report 且请求其余字段合法
- **THEN** 系统 SHALL NOT 增加聚合计数，SHALL NOT 追加时间线（响应可为成功空结果以免客户端重试）

### Requirement: Per-user rate limit of one successful report per 3 seconds

同一 `wxId` 在任意 3 秒窗口内 MUST 最多完成一次成功的 client-usage 写入。超出限制时 MUST 拒绝本次写入（例如 HTTP 429 或明确业务错误），MUST NOT 更新聚合或时间线。

#### Scenario: 3 秒内第二次上报被拒

- **WHEN** 同一用户在成功上报后 3 秒内再次调用 report
- **THEN** 系统 MUST 拒绝第二次写入，且 Redis 聚合/时间线 MUST NOT 因第二次请求增加

### Requirement: Redis aggregation window MUST be at most 30 days

系统 MUST 以日桶等形式维护按 `featureId`、按用户、以及功能×用户交叉的聚合计数，查询窗口 MUST 支持至多 **30 天**（例如 `days=7` 与 `days=30`）。聚合数据 MUST 仅存 Redis，MUST 经 `cachekit` 访问，键 MUST 由 platform builder 登记。产品接受 Redis 丢失导致统计丢失。

#### Scenario: 按功能查询近 7 天

- **WHEN** 管理员请求功能维度列表且 `days=7`
- **THEN** 系统 SHALL 返回窗口内各 `featureId` 的合计次数（及实现约定的最近时间字段，若有）

### Requirement: Per-user timeline MUST store event name and server time only

每个真实用户 MUST 可查询上报时间线。每条时间线条目 MUST 仅包含：`featureId`（事件名）与服务端 Unix 时间。MUST NOT 要求或持久化客户端自定义 `extra` 载荷作为一期范围。

#### Scenario: 用户时间线字段

- **WHEN** 管理员查询某 `wxId` 的时间线
- **THEN** 每条记录 SHALL 含 `featureId` 与服务端时间，且 SHALL NOT 依赖客户端提交的时间戳

### Requirement: Per-user timeline MUST cap at 10000 entries and retain about 30 days

单用户时间线 MUST 最多保留 **10000** 条；超出时 MUST 丢弃最旧条目。时间线保留策略 MUST 与约 **30 天** 有效期对齐（key TTL 与/或按日过期）。条数上限与时间窗口 MUST 同时生效。

#### Scenario: 超过一万条裁剪

- **WHEN** 某用户时间线已有 10000 条且再次成功上报
- **THEN** 系统 SHALL 追加新事件并 SHALL 移除最旧事件，使条数不超过 10000

### Requirement: Report API MUST NOT count toward App API usage stats

`POST /device/app/api/client-usage/report`（精确 METHOD+path）MUST 登记于 `maintenance_skip`（或等价 denylist），MUST NOT 计入 App API 使用统计。负责人已确认本接口不统计。

#### Scenario: report 不计入 API usage

- **WHEN** 用户成功调用 client-usage report
- **THEN** App API 使用统计对应计数 MUST NOT 因该请求递增

### Requirement: Admin client-usage APIs and Hub page

`gateway-app` MUST 提供需 Admin JWT 的读接口，至少支持：按功能聚合列表、按用户列表/下钻、单用户时间线。路径 MUST 由 gateway-app 本机处理，MUST NOT 直连他域数据库。Hub MUST 增加独立入口「客户端使用统计」，交互对齐现有 App API 使用统计页（功能/用户维度；时间范围不超过 30 天）。页面 MUST 提示数据仅 Redis、可丢失，且时间线可能因 10000 条上限短于次数合计。

#### Scenario: Hub 按用户查看时间线

- **WHEN** 管理员打开「客户端使用统计」并进入某用户详情
- **THEN** 系统 SHALL 展示该用户聚合摘要与时间线（事件名 + 服务端时间）
