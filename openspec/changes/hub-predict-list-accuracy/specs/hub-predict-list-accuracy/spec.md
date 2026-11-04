## ADDED Requirements

### Requirement: Hub SHALL expose admin read-only 7-day predict accuracy by event root

The history-service MUST provide an Admin-authenticated read-only HTTP API that, given a `deviceNo`, returns near-7-calendar-day prediction accuracy aggregated per event root for that device. Scoring MUST match Flutter `trend_predict_accuracy` / `predict-list-accuracy-bars`: a row is an eligible sample only when stored `expectStartTime` (unix seconds) is &gt; 0 and actual start time is &gt; 0; a hit is when absolute difference between actual start and `expectStartTime` is ≤ 1800 seconds; day assignment MUST use the actual start instant in timezone `Asia/Shanghai`; day percent MUST be `round(100 * hits / samples)` when samples ≥ 1, otherwise that day MUST omit percent (null / no bar). The server MUST NOT fabricate `expect*` by replaying predictors. Cards SHOULD group by catalog root when device event catalog is available; otherwise group by history `eventId`. The API MUST NOT write Redis accuracy caches, MUST NOT send push, and MUST NOT alter App history write paths.

history-service MUST 提供 Admin 只读接口：按 `deviceNo` 返回近 7 个上海自然日、按事件根聚合的预测准确率。计分 MUST 对齐 Flutter：仅 `expectStartTime>0` 且实际开始&gt;0 为样本；命中为 `|实际−预计|≤1800` 秒；归日用 `Asia/Shanghai`；日准确率为 `round(100*hits/n)`，无样本日 MUST 无百分数。MUST NOT 重放伪造 expect*。有事件目录时 SHOULD 按根聚合，否则按 `eventId`。MUST NOT 写准确率 Redis、发推送或改 App 写路径。

#### Scenario: Day with samples returns percent

- **WHEN** Admin 查询某 `deviceNo` 且某根在某上海自然日有 2 条合格样本、1 条命中
- **THEN** 该日对该根的 `dailyPercents` MUST 为 `50`（或等价整数 50）

#### Scenario: Day without samples omits percent

- **WHEN** 某根在窗口某日零合格样本
- **THEN** 该日 percent MUST 为 null/省略
- **AND** MUST NOT 返回 `0` 冒充有样本

#### Scenario: Row without expectStartTime is not scored

- **WHEN** history 行缺少正 `expectStartTime`
- **THEN** 该行 MUST NOT 计入任何日的 samples/hits

### Requirement: Hub SHALL expose admin day summary of expect vs actual samples

The history-service MUST provide an Admin-authenticated read-only API to list eligible accuracy samples for one `deviceNo`, one event root (or fallback event id), and one calendar day in `Asia/Shanghai`. Each item MUST include `expectStartTime`, `actualStartTime` (unix seconds), and `hit` (boolean). The API MUST use the same eligibility and hit rules as the 7-day aggregate. Empty day MUST return an empty list (not an error solely for emptiness).

history-service MUST 提供 Admin 只读日明细：指定宝宝、事件根、上海自然日的合格样本列表；每条含预计开始、实际开始（unix 秒）与是否命中；规则与 7 日聚合一致；空日返回空列表。

#### Scenario: Day summary lists hit and miss

- **WHEN** 某日某根有一条命中与一条未命中合格样本
- **THEN** 响应列表 MUST 含两条，且 `hit` 分别为 true 与 false
- **AND** 各条 MUST 带对应的 expect/actual 开始时刻

### Requirement: Device history Hub page SHALL show a vertical accuracy card list like the App prediction list

The Hub static page `history.html` for a device MUST include a dedicated「预测准确率」section (separate from the Redis「预测上报」card) that presents a vertical list of event-root cards loaded from the Admin accuracy API. Each card MUST show near-7-day accuracy bars (height 0–100) with integer `%` labels on days that have percents, and MUST NOT draw a bar or `0%` on null days. On wide viewports the page MUST use a side-by-side layout (card list + day summary). Tapping a day bar with samples MUST show that day’s expect-vs-actual summary (预计开始 / 实际开始 / 命中) without navigating away; empty-day taps MUST be ignored; a control MUST return to the 7-day bars context. When no cards exist, the section MUST show an empty-state message.

`history.html` MUST 有独立「预测准确率」区（与 Redis 预测上报分开）：纵向事件根卡 + 近 7 日准确率柱（有 % 才画柱）；宽屏左右分栏；点有样本柱展示当日预计/实际/命中小结；空日不可进小结；须能返回近 7 日；无卡时展示空态。

#### Scenario: Bars mirror API null days

- **WHEN** API 某日 percent 为 null
- **THEN** Hub 该日槽位 MUST 不显示柱与 `0%`

#### Scenario: Tap bar shows summary columns

- **WHEN** 运维点击某卡有样本的日柱
- **THEN** 小结区 MUST 列出预计开始、实际开始、是否命中
- **AND** MUST NOT 仅依赖跳转其它路由作为主呈现
