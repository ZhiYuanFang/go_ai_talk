## Context

预测临近离线推送由客户端同步 Redis pending，voice 延时 MQ 叫醒后在 `handlePredictImminentFire` 决策是否可见推送。现有闸：Redis 匹配 → 五分钟去重（SetNX 占位）→ 近 30 分钟 `start_time` 历史闸 → 推送。

history 域「进行中」权威为 `end_time=0`。现有 `ListHistoryFilter` 仅按 `start_time` 窗过滤，**无法**可靠表达「任意时长未闭合」；若用 `ignoreTimeRange` + limit 再客户端筛 `EndTime==0`，在「最新已闭合、更早仍开放」或大 limit 下既易漏判又增加 DB 扫描，不符合最少 DB 操作。

## Goals / Non-Goals

**Goals:**

- 叫醒时若同类事件（根含叶子）存在 `end_time=0`，跳过可见推送并 Ack，日志 `reason=in_progress`（或等价）。
- 进行中闸在去重占位之前；命中不写五分钟去重键。
- 保留 30 分钟 `start_time` 闸与既有去重语义。
- voice 仅经 history HTTP 契约查询；最少往返（存在性，limit 1）。

**Non-Goals:**

- 不改客户端 pending 同步协议（客户端仍可主动从 pending 移除进行中事件，作为额外优化）。
- 不改推送文案、bizType、角标策略。
- 不抑制**不同根**下的并行进行中（例如睡眠进行中仍可推换尿布）。
- 不新增 App 对外路由；不引入 Redis 新读缓存业务键（仅沿用现有 pushed/pending）。

## Decisions

### D1：进行中 = `end_time=0`，根展开叶子

- **选择**：与 history「已在进行中」一致；复用 `expandEventIDsForHistory`，任一展开 id 存在未闭合行即抑制整根推送。
- **备选**：仅精确匹配 pending 的 eventId — 否决（根 pending + 叶子计时会漏判）。

### D2：新增 history 存在性契约，而非滥用 ListHistoryFilter

- **选择**：在 `DeviceHistoryContract` 增加轻量方法（名例：`HasOpenHistory`），入参 `deviceNo` + `eventIds[]`，返回 bool；history-service 内部 `WHERE device_no=? AND event_id IN (...) AND end_time=0 LIMIT 1`（One/Exist，最少读）。HTTP 可为现有 internal/list 旁路新 GET（如 `/device/history/api/event/open-exists`）或等价；**additive**，不改已有 filter 响应结构。
- **备选 A**：`ListHistoryFilter(ignoreTimeRange)` 再滤 `EndTime==0` — 否决（漏判与多余扫描）。
- **备选 B**：voice 直查 history 库 — 禁止（服务边界）。

### D3：闸顺序 — 进行中 → 去重 → 30 分钟历史

```text
Redis match
  → HasOpenHistory(展开 ids)？ 是 → Ack skip（不占 dedup）
  → SetNX 五分钟去重
  → 近 30 分钟 start_time 历史？
  → push
```

- **选择**：进行中在 dedup 前，避免「睡着 skip」占满 5 分钟导致醒来后合法叫醒被挡。
- **备选**：与 history 闸同序（dedup 后）— 否决（产品已确认不想占键）。

### D4：查询失败语义

- **选择**：与现有 `history_check_err` 对齐 — **fail-closed**（不推、Ack）；因尚未占 dedup，短时间重投/下一叫醒仍可再试（若仍失败则继续 skip）。
- **备选**：fail-open 仍推 — 否决（误扰风险高于漏提醒）。

## Risks / Trade-offs

- **[history 新接口未部署而 voice 先上]** → 发版顺序：先 history-service 再 voice-service；或同发。
- **[事件目录拉取失败导致展开不全]** → 与现网 30 分钟闸相同：展开失败时至少含原始 eventId；记录日志。
- **[多叶子并行开放]** → `IN` + `LIMIT 1` 任一命中即可，行为正确。
- **[客户端未同步 pending]** → 服务端闸兜底，不依赖客户端删除。

## Migration Plan

1. 合并后先确保 history 契约/路由可用，再滚动 voice。
2. 观察日志：`in_progress` skip 次数；确认睡眠长会话不再误推。
3. 回滚：voice 去掉进行中闸即可；history 新接口可保留无害。

## Open Questions

- （无阻塞）HTTP 路径最终命名在实现时与现有 `/device/history/api/event/*` 风格对齐即可。
