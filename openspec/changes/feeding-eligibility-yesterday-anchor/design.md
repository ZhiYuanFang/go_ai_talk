## Context

现网：`history.GetFeedingDayStats` 扫描窗口内全部 `start_time` 后在 Go 按上海日聚合；`days[0]=今日`。`cash.CountConsecutiveEffectiveDays` 从今日起连续累计；资格结果按上海「请求日」Redis 缓存约 36h。含今日与日缓存冲突：上午未达标会钉住当日结果。

本变更在 **history-service** 改为 SQL 按日 COUNT，并将统计窗口平移为「昨天起 N 个已闭合上海日」；**cash-service** 契约同步为 `days[0]=昨天`。宿主边界不变（资格仍只在 cash）。Flutter 进度文案由客户端拼字段。

## Goals / Non-Goals

**Goals:**

- 锚点：上海昨日；今日不计入。
- history：DB 按日 COUNT；窗口 = `requiredDays` 个已闭合日。
- 跨上海日 0 点后，已闭合连续满额 → 立即 `qualified=true`。
- 客户端用 `effectiveDays` 等字段拼进度；不解释「今日不计入」。
- 规格增量改写 `feeding-effective-day-core` 与 `care-alert-feeding-eligibility`。

**Non-Goals:**

- 改场景配置表 / Admin API / 资格 HTTP 路径与字段名。
- 服务端强制改写进度 `message` 为权威进度文案。
- 在 device/voice 复制算法；新建测试文件；改 usage skip。
- 恢复或对齐旧「昨日有发生」闸门语义。

## Decisions

### D1 — 锚点与窗口（H+C 一致）

history `GetFeedingDayStats(deviceNo, days)`：

- `yesterdayStart` = 上海今日 00:00 的前一天 00:00。
- 时间半开区间：`[yesterdayStart-(days-1), todayStart)`（不含今日）。
- 返回 `Days`：`index0=昨天`，向过去排列，长度 = `days`；无记录日 count=0。

cash：`CountConsecutiveEffectiveDays` 仍从前向后扫，遇无效 break；注释与契约改为「从昨天起」。不在 cash 再丢弃「今日」行（history 已不含今日）。

**备选**：history 仍含今日、cash 跳过 `days[0]` → 否决（多拉无用日、契约双源）。

### D2 — SQL 按日 COUNT

使用 `device_no` + `start_time` 范围（命中 `idx_history_device_start`），`GROUP BY` 上海日历日，`COUNT(*)`。时区：**禁止依赖 `CONVERT_TZ` / mysql 时区表**（未装表时返回 NULL → 全 0 有效日）。改用固定东八算术：

`DATE_ADD('1970-01-01', INTERVAL ((start_time + 28800) DIV 86400) DAY)`

（上海无夏令时；与 Go `Asia/Shanghai` 切窗一致。）零条日由 Go 补全。

**备选**：`CONVERT_TZ(..., '+08:00')` → 否决（依赖时区表）。继续拉 `start_time` 列表 → 否决（本变更目标之一）。

### D3 — 跨日立即合格与缓存

缓存键仍含上海「请求日」`yyyyMMdd`（现有 `CashFeedingEligibilityKey`）。新上海日 → 新键 → miss → 重算；昨日已满门槛则当日首次请求即可合格。无需午夜主动失效 ticker。

### D4 — 客户端进度

- 权威数字：`effectiveDays` / `requiredDays` / `remainingDays`。
- 文案由 Flutter 拼接（如「已连续 X / Y 天」）；MUST NOT 要求声明「今日不计入」。
- 服务端 `message` 可保留激励文案，**非**进度权威。

### D5 — care_alert 产品语义

`care_alert_entry.requiredDays=2` 排除今日后几何上为「昨天+前天」连续有效，目的是促活解锁体验，与旧「昨日有发生」无关。规格中切断继承表述。

## Risks / Trade-offs

- [相对旧算法当日进度不可用] → 用户须等自然日闭合；与日缓存一致，属有意取舍。
- [MySQL 时区表缺失] → 已改 epoch+28800 算术，不依赖 CONVERT_TZ。
- [已缓存 effectiveDays=0] → 发版后依赖新上海日键或 Admin 改场景 bump `updated_at` 使资格键失效；否则当日可能仍命中旧 0。
- [BREAKING 资格时点] → App 仅展示字段，无协议字段变更；运营需知解锁可能推迟到「满额日的次日 0 点后」。

## Migration Plan

1. 先发 history（COUNT + 窗口平移），再发 cash（注释/契约；算法循环可不变）。
2. 同窗口发 Flutter 进度拼字段（若尚未用 `effectiveDays`）。
3. 回滚：回退 history 窗口与聚合方式 + cash 注释；客户端文案可保留字段展示。

## Open Questions

- （无阻塞项）进度文案具体句式由客户端产品定，规格只约束「展示已累计有效日、不解释今日排除」。
