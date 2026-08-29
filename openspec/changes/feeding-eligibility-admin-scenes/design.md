## Context

现网：`cash.GetUCGEligibility` 硬编码 `ucgRequiredDays=7`、`ucgMinRecordsPerDay=10`，并 `FetchFeedingDayStats(..., 14)`；资格按日 Redis 缓存。值得留意在 Flutter 用 `careAlertDailyFetchGate`（登录 + device + range 就绪 + **上海昨日有发生**），无服务端喂养连续日资格。

本变更在 **cash-service** 抽出共享有效日算法与按场景配置表；UCG 与值得留意资格判断 **均只写在 cash**；新增值得留意资格 API（cash）；Flutter 改闸门 UX；history 契约不变（仍按日条数，不做资格合成）。

## Goals / Non-Goals

**Goals:**

- 独立有效日 / 连续 streak 算法；场景阈值 Admin 可配且各场景自有 `minRecordsPerDay`。
- UCG / 值得留意均经该算法，且 **宿主均为 cash-service**；取数窗口 = 所需天数。
- 值得留意：cash 服务端资格 **替代**「昨日有发生」；客户端先资格后生成，未合格仍展示卡片+剩余天数。

**Non-Goals:**

- 在 device / voice / history 进程内实现或复制喂养资格天数判断。
- 改 history schema；VIP 短路；care-alert LLM/额度；新建测试文件；未经确认改 usage denylist。

## Decisions

### D0 — 资格判断宿主（强制）

| 能力 | 宿主 |
|------|------|
| 有效日 / 连续 streak 算法 | **cash-service** |
| UCG eligibility API | **cash-service** |
| 值得留意 feeding eligibility API | **cash-service** |
| 场景阈值表与 Admin | **cash-service** / `ai_voice_cash` |
| 按日 history 条数 | history-service（只计数，不算 qualified） |
| 值得留意生成/日列表 | device（不变）；仅资格通过后由客户端调用 |

### D1 — 配置表

`feeding_eligibility_scene`（`ai_voice_cash`）：

| 列 | 说明 |
|----|------|
| `scene_key` PK | `ucg_entry` / `care_alert_entry` |
| `required_days` | 连续有效日阈值 |
| `min_records_per_day` | 该场景日有效门槛 |
| `updated_at` | unix |

种子：`ucg_entry`(7,10)、`care_alert_entry`(2,10)。Admin 只更新已有行，禁止随意新建 scene（与功能编号类似，scene 与客户端约定）。

**备选**：yaml → 否决（要运维热改）。

### D2 — 核心算法

输入：按日 `[]{date,count}`（index0=今日，向过去）、`requiredDays`、`minRecordsPerDay`。

```
effective = 0
for d in days:
  if d.count >= minRecordsPerDay: effective++
  else: break
qualified = effective >= requiredDays
remaining = max(0, requiredDays - effective)
```

拉数：`FetchFeedingDayStats(deviceNo, requiredDays)`（不再固定 14）。算法包路径位于 `internal/services/cash/**`。

### D3 — API

- 保留 `GET /cash/app/api/ucg/eligibility` → scene `ucg_entry`（cash）。
- 新增 `GET /cash/app/api/care-alert/eligibility` → scene `care_alert_entry`（**cash**，字段同构）。
- Admin：`GET/POST /cash/admin/api/feeding-eligibility/scenes`。
- 网关：确认 `/cash/app/api/*` 反代覆盖；**不** Bearer 匿名豁免。
- usage：与 UCG 一样拟 skip；**实现前问负责人**。

### D4 — Redis

- 键按 device+上海日+scene（或配置版本）；Admin 写路径失效旧缓存。

### D5 — 客户端（Flutter）

- 停用「昨日有发生」作为值得留意拉取前提。
- 先调 **cash** care-alert eligibility；未合格展示卡片进度；合格再调 device care-alert。
- 客户端 MUST NOT 本地用 history range 自行判定权威 `qualified`。

### D6 — 服务边界

cash 经 `HISTORY_SERVICE_URL` 取数；不直查 history 库。device care-alert 生成不做资格天数判断。

## Risks / Trade-offs

- [改阈值后当日缓存仍旧] → 配置写路径 bump 版本进缓存键。
- [care_alert 默认 min=10 过严] → Admin 可调。
- [usage 未确认] → tasks「问负责人」门禁。

## Migration Plan

1. EnsureSchema + 部署 cash。
2. gateway-app 静态 Admin / skip（若已确认）。
3. 发 Flutter。
4. 回滚：回退代码；配置表可留。

## Open Questions

- usage skip 负责人确认。
- `care_alert_entry` 种子 min 是否默认 10（提案默认是）。
