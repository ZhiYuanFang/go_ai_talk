## Why

现网喂养资格以「请求当日」为锚连续累计有效日，且 history 按日统计先拉窗口内全部 `start_time` 再在进程内聚合。当日未闭合日与按日 Redis 缓存冲突（上午未达标结果可能钉住整天），且拉全量时间戳成本偏高。需要改为按日 SQL COUNT，并以**上海昨日起**的已闭合自然日为锚；跨日 0 点后昨日凑满门槛须立即合格。进度文案由客户端用字段拼接，不向用户解释「今日不计入」。

## What Changes

- **锚点变更（BREAKING 相对现网资格语义）**：连续有效喂养日 MUST 以 **Asia/Shanghai 昨日** 为 `days[0]` 向前累计；**今日 MUST NOT** 计入有效日或 streak。
- **跨日立即合格**：上海日滚到新日后，若「昨天起」连续有效日已 ≥ `requiredDays`，当日首次资格请求 MUST 可返回 `qualified=true`（依赖按上海日缓存键自然失效，无额外 ticker）。
- **history 取数**：`feeding-day-stats` MUST 在 DB 侧按上海日 `COUNT` 聚合；窗口 = `requiredDays` 个**已闭合日**（昨天起往前），不再拉全量 `start_time` 行。
- **cash 算法契约**：`CountConsecutiveEffectiveDays` / 合成逻辑的输入约定改为 `days[0]=昨天`；宿主仍仅 cash-service。
- **客户端**：进度文案由客户端用 `effectiveDays` / `requiredDays` / `remainingDays` 拼接，告知已累计有效日；MUST NOT 强制声明「今日不计入」。服务端 `message` 非进度权威。
- **care_alert 动机澄清**：`care_alert_entry`（如 requiredDays=2）是促活/解锁更多体验的门槛，与历史「昨日有发生」数据质量闸门无继承关系；排除今日后实质为「昨天+前天」连续有效，属阈值几何结果而非旧闸门语义。
- **规格**：增量改写 `feeding-effective-day-core`（及客户端进度相关的 `care-alert-feeding-eligibility` 表述）；不改 Admin 场景表结构与 API 路径形状。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `feeding-effective-day-core`：锚点由「请求当日」改为「上海昨日」；history 按日 COUNT；窗口仍等于 `requiredDays` 但平移至已闭合日；跨日立即合格。
- `care-alert-feeding-eligibility`：进度展示改为客户端拼有效日字段；明确促活目的，切断与「昨日有发生」的语义关联；不要求向用户解释今日排除。

## Impact

- **进程**：`history-service`（`GetFeedingDayStats` / 内部 `feeding-day-stats`）；`cash-service`（资格合成注释与契约、缓存键仍按上海「请求日」）；Flutter（UCG / 值得留意进度文案拼字段）。
- **API**：路径与响应字段形状不变（仍 `qualified` / `requiredDays` / `effectiveDays` / `remainingDays`）；**BREAKING** 为同设备同日可能相对旧算法更晚或更早合格（不含今日进度、但已闭合满额则过 0 点即合格）。
- **非目标**：改 `feeding_eligibility_scene` 表；改 Admin；VIP 短路；新建测试文件；服务端强制改写进度 `message`；usage denylist 变更。
