## Context

Flutter twin `live-activity-expect-lock` 约定方案 B：Go 推结构化字段，Swift 拼短文案。Imminent pending / fire payload 已含 `expectKind` 等；active-eta Redis 已含 `expectEndAt` + `historyId`。当前 `startImminentLiveActivityForWx` 只写 title/phase/nextAt；Timing start 只写 eventName/startTime/phase；`SyncPredictActiveEta` 注释写死 MUST NOT 推送。

跨服务：Timing start 在 **history-service**；ETA 在 **voice-service**。voice 已可经 `clients/push` 扇出 LA，不必回调 history。

## Goals / Non-Goals

**Goals:**

- Imminent start ContentState 透传可选 expect*。
- Timing ContentState 支持 `expectEndAt`；以 active-eta 同步后的 **update** 为灌入主路径。
- 有值才写键；0 / 空 omit；unix 秒；amount 浮点。
- 修正 active-eta「禁止一切推送」→ 允许 LA update，禁止 MQ/普通 alert。

**Non-Goals:**

- Go 拼「预计 HH:mm 结束」等中文（客户端职责）。
- 改 pending HTTP 契约字段名。
- Flutter Extension / Island。
- Timing start 时强依赖已有 ETA（竞态常见：先 start 后 ETA）。

## Decisions

1. **Imminent**  
   扩展 `startImminentLiveActivityForWx`（或等价）接收 expect 结构化参数；`expectKind` 为 `duration`/`amount` 时按维写入 ContentState。duration 仅当 `expectDurationSec>0`；amount 仅当 `expectAmount>0`；`expectUnit` 非空才写。不写 `expectLeafTitle` 进 LA。

2. **Timing + active-eta**  
   `SyncPredictActiveEta` 写 Redis 成功后：遍历 cleaned items，解析正 `historyId`，若 `expectEndAt>0`，对绑定用户列表 `FanoutLiveActivity` action=`update`，`attributesKind=timing`，`activityId=timing:{deviceNo}:{historyId}`，ContentState 至少含 `phase=timing`、`expectEndAt`；有 `title`/`startAt` 时可一并带上以免 update 覆盖丢字段（若 push 整包替换）。空列表仅清 Redis，不批量 end Timing。

3. **history Timing start/update**  
   不强制在 start 读 voice ETA；可选：若后续易取可带上。history `updateTimingLiveActivity` 可不写 `expectEndAt`；依赖随后 Flutter 再 sync active-eta 回灌。

4. **仍禁止** logo/color；时间字段秒级。

## Risks / Trade-offs

- [active-eta 改推送边界] → 规格明确仅 LA update，无 alert/MQ。  
- [update 整包替换丢字段] → ContentState 尽量带齐 eventName/startTime/expectEndAt。  
- [historyId 解析失败] → 跳过该条，打 debug/warn，不影响 Redis 成功。  
- [先 LA 后 ETA] → 副行延迟出现，可接受。

## Migration Plan

1. 与 Flutter twin 同发或 Go 稍后；缺字段客户端静默。  
2. 回滚：停写 ContentState 新键即可。

## Open Questions

- 无。
