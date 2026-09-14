## Context

gateway-app 已有 **App API 使用统计**（被动记 2xx HTTP → Redis day-hash，Hub「功能使用统计」）。产品需要 **页面/功能触达** 维度，API 路径无法等价映射。约束：仅 Redis、可丢；无 MQ/ticker；登录态；模拟用户跳过；report 不计入 API usage。探索结论已锁定 featureId 自由字符串、3s 限流、聚合+时间线 30 天、时间线每用户 ≤1 万条、条目仅事件名+服务端时间。

## Goals / Non-Goals

**Goals:**

- App 登录后可上报 `featureId`；gateway-app 写聚合 + 用户时间线。
- Admin/Hub「客户端使用统计」：按功能、按用户；用户下钻时间线。
- 限流、sim 跳过、TTL/条数硬顶可控。

**Non-Goals:**

- MySQL 落库、导出报表仓库、实时大屏。
- 服务端 feature 白名单 / 中文名 registry（一期直接展示 featureId）。
- 批量上报、客户端 `count`、`extra` JSON、客户端时间戳。
- 与 VIP/功能开通/计费联动；防作弊计费级强度。
- 强制本变更内完成 Flutter 全量埋点（可跨仓联调任务，非阻塞后端交付）。

## Decisions

### D1：宿主与路径

- **选择**：全部在 **gateway-app-server** 本机处理（与 `usagestats` 同进程同 Redis）。
  - App：`POST /device/app/api/client-usage/report`
  - Admin：`GET /device/admin/api/client-usage/*`（list/features、users、user 详情含 timeline、feature 下钻）
- **理由**：统计不落领域库；避免 device 反代吞路径（对齐现有 `/device/admin/api/usage/*` 本机绑定）。
- **备选**：device-service 落库 — 否决（用户明确仅 Redis）。

### D2：双写 — 聚合 + 时间线

- **选择**：一次成功 report 同时：
  1. **聚合**：日桶 Hash（全局 feature、per-wx feature、交叉 feature×wx），窗口 **≤30 天**；可选 lastAt Hash。
  2. **时间线**：per-wx 结构 append `{serverUnix, featureId}`；TTL/保留与聚合对齐 **30 天**；**LTRIM/裁剪至最多 10_000 条**。
- **理由**：Hub 列表要聚合；跟踪操作要时间线；条数硬顶防止满速刷爆。
- **备选**：仅时间线再聚合 — Admin 列表成本高，否决。

### D3：Redis 键（须 `cachekit` builder）

建议前缀与 API usage 隔离：

| 用途 | 键形态（示意） | TTL |
|------|----------------|-----|
| 日全局聚合 | `gw:featusage:d:{yyyyMMdd}:g` Hash field=`featureId` | 30d |
| 日用户聚合 | `gw:featusage:d:{yyyyMMdd}:w:{wxId}` | 30d |
| 日交叉 | `gw:featusage:d:{yyyyMMdd}:x` field=`featureId\x1f{wxId}` | 30d |
| lastAt | `gw:featusage:last:g` / `…:w:{wxId}` | 随写入刷新或 30d |
| 时间线 | `gw:featusage:tl:{wxId}` LIST 或 ZSET | key TTL 30d + 长度 ≤10000 |
| 限流 | `gw:featusage:rl:{wxId}` | 3s |

- **时间线结构倾向**：LIST（LPUSH + LTRIM 0 9999）成员 `"unix|featureId"`；或 ZSET score=unix、member 含 seq 防碰撞。实现选一种并在注释写明。
- **访问**：`cachekit.Default()` / WithObserver；禁止业务层 `g.Redis()`。
- **Redis 新空间**：产品已确认仅 Redis 可丢；属 usage 同族扩展，design 本决策即确认记录。

### D4：身份、跳过、限流

- 从网关登录态解析 **wxId**（与现有 usage 同源）；`wxId<=0` → **401/业务未登录错误**，不写。
- **模拟用户**：`IsSimulatedWx` 为真 → **不写**聚合/时间线（可 2xx 空成功以免客户端重试风暴，或明确 204；实现选「成功且不记」）。
- **限流**：同一 wxId **3 秒内仅 1 次成功写入**；超限 → **429**（或不计入的明确错误码）；不排队合并。
- featureId：trim 后非空；长度设合理上限（如 128）防巨型 field；无枚举校验。

### D5：usage 统计口径

- `POST /device/app/api/client-usage/report` **MUST** 加入 `maintenance_skip.go`（负责人确认：不计入 App API 使用统计）。
- Admin 读路径本就在 `/device/admin/api/` 结构性 skip。

### D6：Hub UI

- 新模块 id 如 `client-usage-stats`，标题 **「客户端使用统计」**，与「功能使用统计 / App API」并列，避免混名。
- 页能力：days=7|30（无 90）；按功能 | 按用户；用户详情展示时间线（服务端时间 + featureId）；hint 说明时间线可能因 1 万条裁剪短于次数合计、Redis 可丢。

### D7：异步写

- 对齐现有 usage：限流判定可同步；Redis 写入可 `RecordAsync` 风格 goroutine，**不得**拖慢业务响应语义（限流结果须在响应前确定）。

## Risks / Trade-offs

- [客户端可刷] → 3s 限流 + 非计费用途；接受。
- [featureId 拼写分裂] → 自由字符串约定；一期无白名单。
- [时间线裁剪 vs 聚合不一致] → UI hint；聚合仍反映窗口内次数。
- [Redis 丢失] → 产品接受；无补偿。
- [满速约 8h 顶满 1 万条] → 硬顶保护集群；运营看近况仍可用。

## Migration Plan

1. 发版 gateway-app（键 builder + report + Admin + Hub 静态页）。
2. Flutter 按需埋点调用 report（可后续发版）。
3. 回滚：下线路由/Hub 入口；Redis 键自然过期。

## Open Questions

无。探索已锁定：自由 featureId、必须登录、sim 跳过、report 不计入 API usage、3s 限流、聚合≤30d、时间线=事件名+服务端时间、每用户时间线≤1 万条。
