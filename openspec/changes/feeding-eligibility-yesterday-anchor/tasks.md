## 1. history 按日 COUNT + 昨日锚点窗口

- [x] 1.1 改写 `GetFeedingDayStats`：时间窗为上海 `[yesterdayStart-(days-1), todayStart)`，返回 `days[0]=昨天` 向过去，零条日补 0
- [x] 1.2 查询改为 DB 按上海日 `GROUP BY` + `COUNT(*)`（时区固定东八），不再 Scan 全量 `start_time` 列表
- [x] 1.3 更新 `feeding_day_stats.go` / 内部 API 注释与 summary，标明「已闭合日、不含今日」

## 2. cash 资格契约对齐

- [x] 2.1 更新 `CountConsecutiveEffectiveDays` / `SynthesizeFeedingEligibility` / `GetFeedingEligibilityByScene` 中文注释与契约：`days[0]=昨天`
- [x] 2.2 确认仍 `FetchFeedingDayStats(..., requiredDays)`（窗口长度不变，语义已由 history 平移）；无需为「跳过今日」多拉一天
- [x] 2.3 确认按上海请求日的 Redis 资格键在跨日后 miss 重算，满足「0 点后立即合格」（无新增 ticker）

## 3. 客户端进度文案

- [ ] 3.1 Flutter UCG / 值得留意未合格态：用 `effectiveDays`、`requiredDays`、`remainingDays` 客户端拼接进度，展示已累计有效日
- [ ] 3.2 确认文案不强制包含「今日不计入」；不以服务端 `message` 为进度数字权威

## 4. 校验与文档

- [ ] 4.1 手工抽检：固定设备对比 SQL 按日 count 与资格 `effectiveDays`；跨日满额后首次请求 `qualified=true`
- [x] 4.2 如有 runbook / Hub 说明涉及「从今日起连续」，改为「从昨天起已闭合日」（若存在）
