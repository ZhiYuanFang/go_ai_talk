## Context

Flutter `live-activity-lock-ui`：Extension 按 Attributes.`eventId` 读 App Group `laEventVisuals` 取色/图，推送不带视觉字段；Timing 右侧大字依赖 ContentState.`startTime`（unix 秒）。

Go 现状：

- Imminent（`predict_live_activity.go`）：Attributes 已有 `eventId`、`nextAt`；ContentState 有 `nextAt`。
- Timing（`live_activity_timing.go`）：Attributes **无** `eventId`；`startTime` 直接传 `item.StartTime`（history 实体注释为时间戳，预期秒，须核对无毫秒混入）。

## Goals / Non-Goals

**Goals:**

- Timing start 固定 Attributes 含有效 `eventId`（history.EventId > 0 时写入；缺省不得伪造正 id）。
- 所有 LA 时间字段统一 unix **秒**。
- 载荷禁止 logo/color 类字段。

**Non-Goals:**

- 客户端 UI / App Group 索引。
- 改限流、清场、token、开关语义。
- 为旧客户端做字段双写兼容以外的特殊协议（旧 App 缺解码已兜底 0）。

## Decisions

1. **Timing Attributes 增加 `eventId`。**  
   在 `maybeStartTimingLiveActivity` 的 `attrs` 写入 `eventId`；函数增加 `eventID int64` 参数，由 `local.go` 落库成功路径传入 `item.EventId`。  
   *备选：仅放 ContentState — Flutter 固定 Attributes 已定读 attributes.eventId；否决。*

2. **时间单位：秒。**  
   与 `entity.History.StartTime` / pending `NextAt` 既有秒语义对齐；实现时若发现某调用点传毫秒，在该点 `/1000` 或改源头，禁止客户端猜单位。

3. **不推送视觉。**  
   构造 map 时不得加入 `logo`/`color`/`logoFile`；评审对照 Flutter `LIVE_ACTIVITY.md`。

4. **eventId=0。**  
   history 异常无事件 id 时仍可 start（计时文案可用），但 Attributes `eventId` 为 0；客户端走 BrandLogo/orb。可选：`eventId<=0` 打 Warning 日志，不阻断扇出。

## Risks / Trade-offs

- [历史毫秒脏数据] → 核对写入 StartTime 的所有路径；必要时对「过大」时间戳打点告警。  
- [仅 Go 先上 / 仅 App 先上] → App 缺 eventId 解码为 0 可兜底；Go 缺字段则无图。  
- [根/叶事件 id] → 与 history.EventId 一致即可（与 catalog 主键对齐）；不在本变更拆 root/leaf。

## Migration Plan

1. 合并本变更并与 Flutter `live-activity-lock-ui` 联调。  
2. 回滚：去掉 Timing `eventId` 键（App 仍兼容）。

## Open Questions

- 无。
