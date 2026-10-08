## Context

预测临近链路：客户端 `PUT /device/api/predict/imminent/pending` → voice 整表写 Redis → 对每条 `PublishDelayed`（`voice.delayed` + `x-delay` 毫秒）→ consumer 对账 Redis 后推送。

`rabbitmq_delayed_message_exchange` 对 delay 要求 `0 < Delay ≤ 4294967295`（约 49.7 天）；超出则 **nodelay 立刻路由**。疫苗等预约 `nextAt` 可达 2 年，当前 Go 侧用 `int64` 原样下发超大 `x-delay`，会触发误立刻推送。

预约表 `appointment_next.next_at`（BIGINT）与 pending JSON 已能存远点；本变更只修 **MQ 投递闸**，不改存储类型。

## Goals / Non-Goals

**Goals:**

- 防止 `delayMs` 超过插件可表达上限时立刻投递导致误推。
- Redis 仍完整保存远 `nextAt`；近端 delay 行为不变。
- 改动集中在 `SyncPredictImminentPending`，可观测、可回滚。

**Non-Goals:**

- 消费路径分段续约 / 二次 `PublishDelayed`（与现 Ack-only 契约冲突）。
- 修改 appointment API、消费闸（进行中/历史/去重）、推送文案。
- 强制本仓改 Flutter；仅约定「进窗后再 sync」。
- 新建后台 ticker 扫远预约补发 MQ。

## Decisions

### D1：上限常量对齐插件 `ERL_MAX_T`，可留安全余量

- **选择**：常量 `predictImminentMaxDelayMs = 4294967295`（或略小，如 `40 * 24 * 3600 * 1000` 若希望更保守）。默认取插件硬顶，避免「本可调度却被跳过」。
- **备选**：业务窗 30 天——更严，但会缩短可调度窗口，非必须。
- **理由**：根因是插件类型/定时器上限，常量应显式注释对齐 RabbitMQ 文档。

### D2：超上限只跳过 MQ，不失败整表同步

- **选择**：`delayMs > max` → 打 WARN 日志（含 eventId、delayMs 或「far」标记）、`continue`；其余条照常 Publish；函数仍返回已写入条数成功。
- **备选**：整表失败——客户端会重试，但仍无法用 MQ 表达，徒增噪音。
- **理由**：Redis 已是权威；远点本就不该依赖本次 MQ。

### D3：不在消费侧加「nextAt 必须临近」闸（本变更）

- **选择**：本变更只堵发布侧；消费侧既有 Redis 对账仍有效。
- **理由**：发布侧跳过后远点不会入队；若历史消息已误发，旧逻辑仍可能推一次——部署后新 sync 不再产生。可选后续加「`now < nextAt - lead - skew` 则 skip」作纵深，非本变更必须。

### D4：远预约补调度靠客户端再 sync

- **选择**：产品/客户端在进入可调度窗口（例如距 `fireAt` < ~45 天）时再次 PUT pending。
- **备选**：服务端分段续约——需改消费契约与 OpenSpec 背景任务语义，成本高。
- **理由**：与现「客户端驱动 pending」一致；本仓最小改动。

## Risks / Trade-offs

- **[长期无再 sync]** → 远预约永不推送。缓解：客户端进窗再 sync；Hub/文档备忘；Redis 仍有数据可供本地展示。
- **[上限取硬顶，Mnesia 长期堆积近 50 天消息]** → 既有插件风险，非本变更引入；远点不再入队反而减轻堆积。
- **[多实例并发 sync]** → 沿用现有 synclock；无关本闸。

## Migration Plan

1. 发布 `voice-service`（含新常量与跳过逻辑）。
2. 无需 DDL / Redis 迁移。
3. 回滚：去掉跳过逻辑即可；回滚后远点再次面临立刻投递风险。
4. 联调：构造 `nextAt = now + 60d` 与 `now + 2y`，确认 Redis 有条目、无立刻推送、日志含 skip；`nextAt = now + 10m` 仍正常延时推。

## Open Questions

- 客户端进窗再 sync 的具体阈值/触发点由 Flutter 仓确认（本变更不阻塞服务端落地）。
- 是否在消费侧加纵深「未到窗 skip」：默认不做；若线上仍见历史误投再开增量。
