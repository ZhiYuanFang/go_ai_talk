## Context

- **现状**：`history.event_unit` 列与 `enrichHistoryEventUnit` 已在 commit `827886d unit` 落地；写入前若请求体无 `EventUnit`，则调用 `lookupEventUnit(ctx, eventId)` 从 `dao.Event` 读取 `unit`。
- **架构约束**：微服务分库后 history-service 的 `default` 连接组仅指向 `HISTORY_DB_LINK`（`ai_voice_history_*`）；`event` 表权威在 device-service 的 `DEVICE_DB_LINK`（`ai_voice_device_*`）。`AGENTS.md` 与 `service-boundary-no-cross-db` 均禁止 history 直查他域表。
- **已有正确先例**：`ListEventOptions`、`GetBirthday` 等已通过 `internal/services/history/delegate_http.go` 调用 device internal HTTP。
- **调用方缺口**：voice 写入 history 未传 `EventUnit`；`history.html` 未展示/提交单位；`mergeHistoryUpdateFromReq` 漏合并 `EventUnit`；Redis 投影/outbox 载荷缺 `event_unit`。

## Goals / Non-Goals

**Goals:**

- 微服务部署下，新增/更新 history 时当 `event.unit` 非空，`history.event_unit` MUST 持久化为相同值（反规范化快照）。
- history-service 补全单位 MUST 经 device HTTP 契约，不得 `dao.Event` 直查。
- 列表、piece、WebSocket 通知、HTTP JSON 响应 MUST 暴露 `eventUnit` 供前端展示（如 `120 ml`）。
- voice 在已知 `entity.Event.Unit` 时 SHOULD 写入请求体，降低远程查主档频率。
- 投影/outbox 与 DB 字段对齐。

**Non-Goals:**

- 不修改 `event` 表结构或 admin 事件字典 UI（单位配置已存在）。
- 不对历史 NULL `event_unit` 做强制批量回填（可作为运维可选脚本，非本变更必须交付）。
- 不新增 `*_test.go`。
- 不新增 device internal「按 id 查单条 event」HTTP（除非 options 列表方案性能不可接受）。

## Decisions

### 1. 用 device event options 委托替代 `dao.Event` 直查

- **决定**：`lookupEventUnit` 改为调用 `delegateListEventOptions`（或封装 `resolveEventUnitFromDevice(ctx, eventID)`），在返回列表中按 `id` 匹配 `unit`；删除 history-service 内对 `dao.Event` 的 import/调用。
- **理由**：与 `canonicalEventNameForRow`、`ListEventOptions` 一致；无需新 HTTP 路由；符合服务边界。
- **备选**：新增 `GET /device/internal/api/event/get?eventId=` — 更精准但增加契约维护面；事件字典规模小（数十条）时全量 options 可接受。

### 2. 进程内短期缓存 event options（可选轻量优化）

- **决定**：在 `history_row.go` 或 `delegate_http.go` 增加带 TTL（如 60s）的 process-local cache，key 为 event options 版本或固定 key；写路径 miss 时委托 HTTP，避免每条 history 写入都打 device。
- **理由**：voice 连续记录、批量手动录入时会重复查同一 options 列表。
- **备选**：每次写入全量 HTTP — 实现最简单，先满足正确性；若 profiling 显示热点再加缓存（tasks 中列为可选子任务）。

### 3. voice 显式传 `EventUnit: event.Unit`

- **决定**：所有 `DeviceHistory().AddHistory` 构造 `entity.History` 处，当 `strings.TrimSpace(event.Unit) != ""` 时填入 `EventUnit`。
- **理由**：voice 已通过 device HTTP 持有完整 `entity.Event`；显式传参使 history 补全逻辑成为兜底而非唯一路径。
- **备选**：仅修 history 侧 — 仍依赖每次写入查 options，voice 侧零改动；双路径更稳健。

### 4. history.html 展示与提交

- **决定**：
  - 加载 event options 时在 `<option>` 上增加 `data-unit`；
  - 列表计数列展示为 `{eventNumber}{unit}`（unit 优先 `row.eventUnit`，回退 option 的 unit）；
  - 提交 payload 增加 `eventUnit`（从选中 option 的 `data-unit` 读取）。
- **理由**：前端不依赖二次猜单位；与 API 已有字段对齐。
- **备选**：纯后端补全 — 微服务修好后列表仍缺展示；用户体验不完整。

### 5. 投影与 outbox 补齐 `event_unit`

- **决定**：`historyProjectionEvent` 与 `EnqueueDomainOutbox` map 增加 `event_unit`/`eventUnit` 字段；`ApplyProjection` 写入 cache 时带上 `EventUnit`。
- **理由**：Redis 读模型与 DB 一致，避免 WS 推送后前端缓存缺单位。

### 6. 更新路径合并 `req.EventUnit`

- **决定**：`mergeHistoryUpdateFromReq` 设置 `item.EventUnit = strings.TrimSpace(req.EventUnit)`；仍走 `enrichHistoryEventUnit` 兜底。
- **理由**：修复已知遗漏，与 add 路径对称。

## Risks / Trade-offs

- **[Risk] event options 列表变大导致每次 lookup 拉全量** → 事件字典规模可控；后续可加按 id 内部 API 或 Redis 共享缓存。
- **[Risk] device-service 不可达时 unit 仍为空** → 与现有 `ListEventOptions` 失败语义一致；写入不应 silently 成功却假装有单位；日志 WARNING。
- **[Risk] 历史行 event_id 变更后 unit 与旧快照不一致** → 更新路径 re-enrich 以新 eventId 为准，符合反规范化「写入时刻快照」语义。
- **[Trade-off] 不回填旧数据** → 旧记录仍 NULL；用户可接受或另开运维任务。

## Migration Plan

1. 部署 history-service + voice-service + 静态资源（gateway-app 挂载 `history.html`）新版本。
2. 无需 DB migration（`event_unit` 列已存在）。
3. 验证：device 库 event 有 unit → 手动/语音新增 history → history 库 `event_unit` 非空 → 页面与 WS 可见。
4. 回滚：还原三处代码；已写入的非空 `event_unit` 可保留，无破坏性。

## Open Questions

- （无阻塞）是否在 admin 历史列表增加独立「单位」列 — 当前方案合并进数量列展示；若产品要独立列可在 apply 阶段调整 UI。
